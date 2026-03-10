/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package node

import (
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/winpath"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameInstallNodeByManual defines the action name.
	ActionNameInstallNodeByManual = "install_node_by_manual"
)

// NewActionInstallNodeByManual get a new action.
func NewActionInstallNodeByManual(capability *Capability) action.Definition {
	return &actionInstallNodeByManual{
		storageHostCredit:     capability.StorageHostCredit,
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		storageActionInstance: capability.StorageWorkflow,
		provider:              capability.DiscoverProvider,
		passwordVault:         capability.HostPasswordVault,
	}
}

// ActParamInstallNodeByManual ...
type ActParamInstallNodeByManual struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

// InstallParamsManual this struct defines the parameters for installing node.
type InstallParamsManual struct {
	InstallerPath   string
	Generation      types.Generation
	NodeRole        types.NodeRole
	CallbackSvrAddr string
	DownloadSvrAddr string
	NodeVersion     string
	DeployToken     string
	OperInstID      string
	BaseWorkDir     string
	BaseDeployDir   string
	AdditionArgs    []string
}

type actionInstallNodeByManual struct {
	storageHostCredit     credit.IStorageHostCredit
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	storageActionInstance workflow.IStorageActionInstance
	provider              discover.Provider
	passwordVault         creditvault.IHostPasswordVault
}

// Name returns the name of the action.
func (act *actionInstallNodeByManual) Name() string {
	return ActionNameInstallNodeByManual
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionInstallNodeByManual) DisplayNameZh() string {
	return "通过手动方式安装节点"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionInstallNodeByManual) DisplayNameEn() string {
	return "Install Node Manually"
}

// Version returns the version of the action.
func (act *actionInstallNodeByManual) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionInstallNodeByManual) Description() string {
	return "Generate manual installation commands for user to execute on the target machine"
}

// Timeout returns the timeout of the action.
func (act *actionInstallNodeByManual) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionInstallNodeByManual) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionInstallNodeByManual) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionInstallNodeByManual) DelayFn() func() {
	return func() {
		time.Sleep(5 * time.Second) // nolint: mnd
	}
}

// Do this func define what the action will do.
// To ensure readability, this action uses fmt.Sprintf to concatenate characters.
// nolint: perfsprint,funlen
func (act *actionInstallNodeByManual) Do(ctx *action.InstanceContext) error {
	param := new(ActParamInstallNodeByManual)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := nodeUtils.NewNodeActionStandarder(act.storageNodeDeployment, act.storageHost)
	if err = std.Initialize(ctx, param.NodeActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	// let the callback server known which action to mark and log.
	std.DeployInfo().BlockingActionName = ActionNameWaitInstallerComplete

	// generate install cmd.
	if err = act.generateInstallCMD(std); err != nil {
		return fmt.Errorf("failed to generate manual install exec cmd: %w", err)
	}

	// update instance data content for next action
	if err := std.UpdateInstanceDataContent(ActionWaitInstallerComplete{
		NodeActionStandardParam: param.NodeActionStandardParam,
		EnsureAgentID:           true,
	}); err != nil {
		return fmt.Errorf("failed to update instance data content: %w", err)
	}

	return nil
}

func (act *actionInstallNodeByManual) generateInstallCMD(std *nodeUtils.NodeActionStandarder) error {
	callbackSvrAddress, downloadSvrAddress, err := act.selectServiceURLs(std)
	if err != nil {
		return fmt.Errorf("failed to select service urls: %w", err)
	}

	// get deploy constant
	deployConstant, err := deployconstant.GetNodeDeployConf(std.DeployInfo().Host.Dynamic.NodeGeneration, std.DeployInfo().Host.Dynamic.NodeOsType)
	if err != nil {
		return fmt.Errorf("failed to get deploy constant: %w", err)
	}

	// format installer tool name
	toolName, err := tool.FormatInstallerName(std.DeployInfo().Host.Dynamic.NodeOsType, std.DeployInfo().Host.Dynamic.NodeCPUArch)
	if err != nil {
		return fmt.Errorf("failed to format installer tool name: %w", err)
	}

	// generate download URL and commands
	var installerPath string
	if std.DeployInfo().Host.Dynamic.NodeOsType == criteria.OSWindows {
		installerPath = winpath.Clean(winpath.Join(std.DeployInfo().InstallerWorkDir, toolName))
	} else {
		installerPath = path.Clean(path.Join(std.DeployInfo().InstallerWorkDir, toolName))
	}

	// build install params
	installParams := &InstallParamsManual{
		NodeVersion:     std.DeployInfo().Host.Dynamic.NodeVersion,
		Generation:      std.DeployInfo().Host.Dynamic.NodeGeneration,
		InstallerPath:   installerPath,
		NodeRole:        std.DeployInfo().Host.Dynamic.NodeRole,
		CallbackSvrAddr: callbackSvrAddress,
		DownloadSvrAddr: downloadSvrAddress,
		DeployToken:     std.Token(),
		OperInstID:      std.InstanceData().OperationInstanceID,
		BaseWorkDir:     deployConstant.BaseWorkDir,
		BaseDeployDir:   deployConstant.BaseDeployDir,
	}

	if !std.DeployInfo().InstallOptions.ReRegister && std.DeployInfo().Host.Dynamic.AgentID != "" {
		installParams.AdditionArgs = append(installParams.AdditionArgs,
			fmt.Sprintf("--agent_id %s", std.DeployInfo().Host.Dynamic.AgentID))
	}

	// generate manual installation exec commands
	installCmd := act.buildCMD(installParams, std.DeployInfo().Host.Dynamic.NodeOsType)

	// save commands to action instance private data
	if err = act.storageActionInstance.UpsertActionInstancePrivateData(
		std.Context(),
		std.InstanceData().OperationInstanceID,
		ActionNameInstallNodeByManual,
		map[string]any{
			types.PDKeyManualInstallInstallerPath: installerPath,
			types.PDKeyManualInstallExecCommand:   installCmd,
		},
	); err != nil {
		return fmt.Errorf("failed to save manual install exec command to private data: %w", err)
	}

	std.InstanceData().Log().
		Zh("已生成手动安装执行命令: %s", installCmd).
		En("generated manual install exec command: %s", installCmd).
		Info()

	return nil
}

// To ensure readability, this action uses fmt.Sprintf to concatenate characters.
// nolint: perfsprint
func (act *actionInstallNodeByManual) buildCMD(param *InstallParamsManual, osType criteria.OSType) string {
	args := []string{
		fmt.Sprintf("--deploy_env %s", system.GetEnv()),
		fmt.Sprintf("--generation %d", param.Generation),
		fmt.Sprintf("--node_role %s", param.NodeRole),
		fmt.Sprintf("--base_work_dir %s", param.BaseWorkDir),
		fmt.Sprintf("--base_deploy_dir %s", param.BaseDeployDir),
		fmt.Sprintf("--dlsvr_addr %s", param.DownloadSvrAddr),
		fmt.Sprintf("--cbsvr_addr %s", param.CallbackSvrAddr),
		fmt.Sprintf("--deploy_token %s", param.DeployToken),
		fmt.Sprintf("--node_version %s", param.NodeVersion),
		fmt.Sprintf("--oper_inst_id %s", param.OperInstID),
		"--log_to_std",
	}
	if len(param.AdditionArgs) > 0 {
		args = append(args, param.AdditionArgs...)
	}

	// windows bat command.
	if osType == criteria.OSWindows {
		installCmd := fmt.Sprintf("cd %s && %s %s %s",
			winpath.Join(param.BaseWorkDir, system.GetEnv()),
			param.InstallerPath,
			installer.NodeCmdFullInstall,
			strings.Join(args, " "))

		return installCmd
	}

	// unix shell command.
	installCmd := fmt.Sprintf("%s %s %s", param.InstallerPath, installer.NodeCmdFullInstall, strings.Join(args, " "))

	return installCmd
}

// selectServiceURLs selects service URLs for download and callback servers.
// Returns: (callbackURLs, downloadURLs, error).
func (act *actionInstallNodeByManual) selectServiceURLs(std *nodeUtils.NodeActionStandarder) (string, string, error) {
	if !std.DeployInfo().InstallOptions.DirectInstall {
		callbackEndpoints, downloadEndpoints, err := std.GetRelayEndpoints()
		if err != nil {
			return "", "", fmt.Errorf("failed to get relay endpoints: %w", err)
		}

		return nodeUtils.BuildServerURLs(callbackEndpoints...), nodeUtils.BuildServerURLs(downloadEndpoints...), nil
	}

	randSelector := discover.NewRandomSelector()
	callbackSvrEndpoint, err := act.provider.SelectEndpoints(
		discover.ServiceNameBackend,
		discover.EndpointNameBackendCallback,
		nodeUtils.DefaultEndpointSelectionCount,
		randSelector)
	if err != nil {
		return "", "", fmt.Errorf("failed to select backend callback endpoint: %w", err)
	}

	downloadSvrEndpoint, err := act.provider.SelectEndpoints(
		discover.ServiceNameFile,
		discover.EndpointNameFileDownload,
		nodeUtils.DefaultEndpointSelectionCount,
		randSelector)
	if err != nil {
		return "", "", fmt.Errorf("failed to select file download endpoint: %w", err)
	}

	return nodeUtils.BuildServerURLs(callbackSvrEndpoint...), nodeUtils.BuildServerURLs(downloadSvrEndpoint...), nil
}
