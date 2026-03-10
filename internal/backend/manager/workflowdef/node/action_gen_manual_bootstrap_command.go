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
	"time"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameGenManualBootstrapCommand defines the action name.
	ActionNameGenManualBootstrapCommand = "gen_manual_bootstrap_command"
)

// NewActionGenManualCommand get a new action.
func NewActionGenManualCommand(capability *Capability) action.Definition {
	return &actionGenManualBootstrapCommand{
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		storageActionInstance: capability.StorageWorkflow,
		provider:              capability.DiscoverProvider,
	}
}

// ActParamGenManualBootstrapCommand defines the action parameter.
type ActParamGenManualBootstrapCommand struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionGenManualBootstrapCommand struct {
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	storageActionInstance workflow.IStorageActionInstance
	provider              discover.Provider
}

// Name returns the name of the action.
func (act *actionGenManualBootstrapCommand) Name() string {
	return ActionNameGenManualBootstrapCommand
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionGenManualBootstrapCommand) DisplayNameZh() string {
	return "生成手动引导命令"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionGenManualBootstrapCommand) DisplayNameEn() string {
	return "Generate Manual Bootstrap Command"
}

// Version returns the version of the action.
func (act *actionGenManualBootstrapCommand) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionGenManualBootstrapCommand) Description() string {
	return "Generate manual commands for detecting system information (OS type, CPU architecture, etc.)"
}

// Timeout returns the timeout of the action.
func (act *actionGenManualBootstrapCommand) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionGenManualBootstrapCommand) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionGenManualBootstrapCommand) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionGenManualBootstrapCommand) DelayFn() func() {
	return func() {
		time.Sleep(5 * time.Second) // nolint: mnd
	}
}

// Do this func define what the action will do.
func (act *actionGenManualBootstrapCommand) Do(ctx *action.InstanceContext) error {
	param := new(ActParamGenManualBootstrapCommand)
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
	std.DeployInfo().BlockingActionName = ActionNameWaitDetectInfoByManual

	// generate install cmd.
	if err = act.generateInstallCMD(std); err != nil {
		return fmt.Errorf("failed to generate manual install bootstrap cmd: %w", err)
	}

	return nil
}

func (act *actionGenManualBootstrapCommand) generateInstallCMD(std *nodeUtils.NodeActionStandarder) error {
	callbackSvrEndpoints, downloadSvrEndpoints, err := act.selectServiceEndpoints(std)
	if err != nil {
		return fmt.Errorf("failed to select service urls: %w", err)
	}

	// generate server url.
	var callbackSvrAddress string
	var downloadSvrAddress string
	if len(std.DeployInfo().Host.Static.InnerIPList) > 0 {
		callbackSvrAddress = nodeUtils.SelectOneServerV4URL(callbackSvrEndpoints)
		downloadSvrAddress = nodeUtils.SelectOneServerV4URL(downloadSvrEndpoints)
	}
	if (callbackSvrAddress == "" || downloadSvrAddress == "") && len(std.DeployInfo().Host.Static.InnerIPV6List) > 0 {
		callbackSvrAddress = nodeUtils.SelectOneServerV6URL(callbackSvrEndpoints)
		downloadSvrAddress = nodeUtils.SelectOneServerV6URL(downloadSvrEndpoints)
	}
	if callbackSvrAddress == "" || downloadSvrAddress == "" {
		return fmt.Errorf("failed to select service urls")
	}

	osType, err := platform.NormalizeOS(std.DeployInfo().Host.Static.OSType)
	if err != nil {
		return fmt.Errorf("failed to normalize os type: %w", err)
	}

	var installCmd string
	switch osType {
	case criteria.OSLinux, criteria.OSDarwin:
		installCmd = "/bin/bash -c \"$(curl -fsSL " + callbackSvrAddress +
			"/api/v3/callback/workflow/node_install/get_manual_script/" +
			std.DeployInfo().Host.Static.OSType + "/" +
			std.InstanceData().OperationInstanceID + "?action=" + ActionNameGenManualBootstrapCommand + ")\""

	case criteria.OSWindows:
		installCmd = "powershell -ExecutionPolicy Bypass -Command \"& {Invoke-WebRequest -Uri '" + callbackSvrAddress +
			"/api/v3/callback/workflow/node_install/get_manual_script/" +
			std.DeployInfo().Host.Static.OSType + "/" +
			std.InstanceData().OperationInstanceID + "?action=" + ActionNameGenManualBootstrapCommand + "' -OutFile install.bat; .\\install.bat}\""

	default:
		return fmt.Errorf("unsupported os type: %s", std.DeployInfo().Host.Dynamic.NodeOsType)
	}

	// save commands to action instance private data
	if err := act.storageActionInstance.UpsertActionInstancePrivateData(
		std.Context(),
		std.InstanceData().OperationInstanceID,
		ActionNameGenManualBootstrapCommand,
		map[string]any{
			types.PDKeyManualInstallBootstrapCommandBash:       installCmd,
			types.PDKeyManualInstallCallbackAddress:            callbackSvrAddress,
			types.PDKeyManualInstallDownloadAddress:            downloadSvrAddress,
			types.PDKeyManualInstallActionNameReportDetectInfo: ActionNameWaitDetectInfoByManual,
			types.PDKeyManualInstallActionNameGetExecCommand:   ActionNameInstallNodeByManual,
		},
	); err != nil {
		return fmt.Errorf("failed to save manual install bootstrap command to private data: %w", err)
	}

	std.InstanceData().Log().
		Zh("已生成手动安装引导命令: %s", installCmd).
		En("generated manual install bootstrap command: %s", installCmd).
		Info()

	return nil
}

// selectServiceEndpoints selects service endpoints for download and callback servers.
// Returns: (callback services, download services, error).
func (act *actionGenManualBootstrapCommand) selectServiceEndpoints(std *nodeUtils.NodeActionStandarder) (
	[]discover.Endpoint, []discover.Endpoint, error) {

	if !std.DeployInfo().InstallOptions.DirectInstall {
		callbackSvrEndpoint, downloadSvrEndpoint, err := std.GetRelayEndpoints()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to build relay service urls: %w", err)
		}

		return callbackSvrEndpoint, downloadSvrEndpoint, nil
	}

	randSelector := discover.NewRandomSelector()
	callbackSvrEndpoint, err := act.provider.SelectEndpoints(
		discover.ServiceNameBackend,
		discover.EndpointNameBackendCallback,
		nodeUtils.DefaultEndpointSelectionCount,
		randSelector)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to select backend callback endpoint: %w", err)
	}

	downloadSvrEndpoint, err := act.provider.SelectEndpoints(
		discover.ServiceNameFile,
		discover.EndpointNameFileDownload,
		nodeUtils.DefaultEndpointSelectionCount,
		randSelector)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to select file download endpoint: %w", err)
	}

	return callbackSvrEndpoint, downloadSvrEndpoint, nil
}
