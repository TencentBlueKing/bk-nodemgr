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
	callbackSvrAddress, err := act.selectCallbackSvr(std)
	if err != nil {
		return fmt.Errorf("failed to select callback server: %w", err)
	}
	callbackSvrAddress = "http://" + callbackSvrAddress

	downloadSvrAddress, err := act.selectDownloadSvr(std)
	if err != nil {
		return fmt.Errorf("failed to select download server: %w", err)
	}
	downloadSvrAddress = "http://" + downloadSvrAddress

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

	std.InstanceData().LogI(fmt.Sprintf("generated manual install bootstrap command: %s", installCmd))

	return nil
}

func (act *actionGenManualBootstrapCommand) selectCallbackSvr(std *nodeUtils.NodeActionStandarder) (string, error) {
	if !std.DeployInfo().InstallOptions.DirectInstall {
		return fmt.Sprintf("%s:%d", std.DeployInfo().RelayInfo.InnerIP, std.DeployInfo().RelayInfo.CallbackSvcPort), nil
	}

	randSelector := discover.NewRandomSelector()
	callbackSvrEndpoint, err := act.provider.GetEndpoint(
		discover.ServiceNameBackend,
		discover.EndpointNameBackendCallback,
		randSelector)
	if err != nil {
		return "", fmt.Errorf("failed to get backend callback endpoint: %w", err)
	}

	return callbackSvrEndpoint.GetIPV4Address(), nil
}

func (act *actionGenManualBootstrapCommand) selectDownloadSvr(std *nodeUtils.NodeActionStandarder) (string, error) {
	if !std.DeployInfo().InstallOptions.DirectInstall {
		return fmt.Sprintf("%s:%d", std.DeployInfo().RelayInfo.InnerIP, std.DeployInfo().RelayInfo.DownloadSvcPort), nil
	}

	randSelector := discover.NewRandomSelector()
	downloadSvrEndpoint, err := act.provider.GetEndpoint(
		discover.ServiceNameFile,
		discover.EndpointNameFileDownload,
		randSelector)
	if err != nil {
		return "", fmt.Errorf("failed to get backend file endpoint: %w", err)
	}

	return downloadSvrEndpoint.GetIPV4Address(), nil
}
