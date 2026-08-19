/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package node

import (
	"errors"
	"fmt"
	"path"
	"time"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
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
		storageNetworkUnit:    capability.StorageTopo,
		provider:              capability.DiscoverProvider,
		passwordVault:         capability.HostPasswordVault,
	}
}

// ActParamInstallNodeByManual ...
type ActParamInstallNodeByManual struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionInstallNodeByManual struct {
	storageHostCredit     credit.IStorageHostCredit
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	storageActionInstance workflow.IStorageActionInstance
	storageNetworkUnit    topoStg.IStorageNetworkUnit
	provider              discover.IProvider
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
func (act *actionInstallNodeByManual) DelayFn(_ int) func() {
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
	if err := std.SaveBlockingActionName(ActionNameWaitInstallerComplete); err != nil {
		return fmt.Errorf("failed to save blocking action name: %w", err)
	}

	// generate install cmd.
	if err = act.generateInstallCMD(std); err != nil {
		return fmt.Errorf("failed to generate manual install exec cmd: %w", err)
	}

	err = saveWaitInstallerPrivateData(
		std.Context(), act.storageActionInstance, std.InstanceData().OperationInstanceID,
		false, true)
	if err != nil {
		return err
	}

	return nil
}

func (act *actionInstallNodeByManual) generateInstallCMD(std *nodeUtils.NodeActionStandarder) error {
	// format installer tool name
	toolName, err := tool.FormatInstallerName(std.DeployInfo().Host.Dynamic.NodeOsType, std.DeployInfo().Host.Dynamic.NodeCPUArch)
	if err != nil {
		return fmt.Errorf("failed to format installer tool name: %w", err)
	}

	endpointSource := nodeUtils.SelectInstallEndpointSource(std)
	callbackEndpoints, downloadEndpoints, err := nodeUtils.GenerateNodeInstallerServerEndpoints(std, act.provider, endpointSource)
	if err != nil {
		return fmt.Errorf("failed to generate node installer server endpoints: %w", err)
	}

	// generate download URL and commands
	var installerPath string
	if std.DeployInfo().Host.Dynamic.NodeOsType == criteria.OSWindows {
		installerPath = winpath.Clean(winpath.Join(std.DeployInfo().InstallerRuntime.WorkDir, toolName))
	} else {
		installerPath = path.Clean(path.Join(std.DeployInfo().InstallerRuntime.WorkDir, toolName))
	}

	// build install params
	installParams := &installer.NodeInstallParams{
		NodeCommonParams: installer.NodeCommonParams{
			DeployEnv:     system.GetEnv(),
			Generation:    int(std.DeployInfo().Host.Dynamic.NodeGeneration),
			NodeRole:      string(std.DeployInfo().Host.Dynamic.NodeRole),
			BaseWorkDir:   std.DeployInfo().InstallerRuntime.BaseWorkDir,
			BaseDeployDir: std.DeployInfo().BaseRuntime.BaseDeployDir,
		},
		InstallerPath:   installerPath,
		DownloadSvrAddr: nodeUtils.BuildServerURLs(downloadEndpoints...),
		CallbackSvrAddr: nodeUtils.BuildServerURLs(callbackEndpoints...),
		DeployToken:     std.Token(),
		NodeVersion:     std.DeployInfo().Host.Dynamic.NodeVersion,
		OperInstID:      std.InstanceData().OperationInstanceID,
		LogToStd:        true,
		RenewGSEProc:    std.DeployInfo().InstallOptions.RenewGSEProc,
		RenewGSETask:    std.DeployInfo().InstallOptions.RenewGSETask,
	}

	if !std.DeployInfo().InstallOptions.ReRegister && std.DeployInfo().Host.Dynamic.AgentID != "" {
		installParams.AdditionArgs = append(installParams.AdditionArgs,
			fmt.Sprintf("--agent_id %s", std.DeployInfo().Host.Dynamic.AgentID))
	}

	// generate manual installation exec commands
	installCmd, err := act.buildCMD(installParams, std.DeployInfo().Host.Dynamic.NodeOsType)
	if err != nil {
		return fmt.Errorf("failed to build manual install cmd: %w", err)
	}
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

func (act *actionInstallNodeByManual) buildCMD(param *installer.NodeInstallParams, osType criteria.OSType) (string, error) {
	var installCmd string
	var err error
	if osType == criteria.OSWindows {
		_, installCmd, err = param.ToWindowsScriptManual()
	} else {
		_, installCmd, err = param.ToUnixScriptManual()
	}
	if err != nil {
		return "", fmt.Errorf("failed to render node manual install script: %w", err)
	}

	return installCmd, nil
}
