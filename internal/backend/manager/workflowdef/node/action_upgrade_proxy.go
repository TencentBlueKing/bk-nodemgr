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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameUpgradeProxy defines the action name.
	ActionNameUpgradeProxy = "upgrade_proxy"

	upgradeProxyScriptTimeout = 10 * time.Minute
)

// NewActionUpgradeProxy get a new action.
func NewActionUpgradeProxy(capability *Capability) action.Definition {
	return &actionUpgradeProxy{
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		gseHandler:            capability.GSEHandler,
		storageActionInstance: capability.StorageWorkflow,
	}
}

// ActionParamUpgradeProxy defines the action param.
type ActionParamUpgradeProxy struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionUpgradeProxy struct {
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	gseHandler            gse.IHandler
	storageActionInstance workflow.IStorageActionInstance
}

// Name returns the name of the action.
func (act *actionUpgradeProxy) Name() string {
	return ActionNameUpgradeProxy
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionUpgradeProxy) DisplayNameZh() string {
	return "升级 Proxy"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionUpgradeProxy) DisplayNameEn() string {
	return "Upgrade Proxy"
}

// Version returns the version of the action.
func (act *actionUpgradeProxy) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionUpgradeProxy) Description() string {
	return "upgrade proxy"
}

// Timeout returns the timeout of the action.
func (act *actionUpgradeProxy) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionUpgradeProxy) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionUpgradeProxy) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionUpgradeProxy) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: funlen,nonamedreturns
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionUpgradeProxy) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamUpgradeProxy)
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

	// select matching tools.
	toolName, err := tool.FormatInstallerName(std.DeployInfo().Host.Dynamic.NodeOsType, std.DeployInfo().Host.Dynamic.NodeCPUArch)
	if err != nil {
		return err
	}

	proxyEndpoint := discover.Endpoint{
		IPV4: std.DeployInfo().Host.Static.InnerIPList[0],
		IPV6: std.DeployInfo().Host.Static.InnerIPV6List[0],
		Port: int(std.DeployInfo().Host.Dynamic.RelayCallbackPort),
	}

	upgradeParams := &installer.NodeUpgradeParams{
		NodeCommonParams: installer.NodeCommonParams{
			DeployEnv:     system.GetEnv(),
			Generation:    int(std.DeployInfo().Host.Dynamic.NodeGeneration),
			NodeRole:      string(std.DeployInfo().Host.Dynamic.NodeRole),
			BaseWorkDir:   std.DeployInfo().InstallerRuntime.BaseWorkDir,
			BaseDeployDir: std.DeployInfo().BaseRuntime.BaseDeployDir,
		},
		InstallWorkDir:    std.DeployInfo().InstallerRuntime.WorkDir,
		InstallerFileName: toolName,
		CallbackSvrAddr:   nodeUtils.BuildServerURLs(proxyEndpoint),
		DeployToken:       std.Token(),
		NodeVersion:       std.DeployInfo().Host.Dynamic.NodeVersion,
		OperInstID:        std.InstanceData().OperationInstanceID,
		SkipDownload:      true,
	}

	err = saveWaitInstallerPrivateData(
		std.Context(), act.storageActionInstance, std.InstanceData().OperationInstanceID,
		false, true)
	if err != nil {
		return err
	}

	// exec upgrade command
	if std.DeployInfo().Host.Dynamic.NodeOsType == criteria.OSWindows {
		return act.doUpgradeWindows(std, upgradeParams)
	}

	return act.doUpgradeUnix(std, upgradeParams)
}

func (act *actionUpgradeProxy) doUpgradeUnix(std *nodeUtils.NodeActionStandarder, param *installer.NodeUpgradeParams) error {
	_, upgradeCmd, err := param.ToUnixScript()
	if err != nil {
		return fmt.Errorf("failed to render proxy upgrade script: %w", err)
	}

	std.InstanceData().Log().
		Zh("升级 Proxy 命令: %s", upgradeCmd).
		En("upgrade proxy cmd: %s", upgradeCmd).
		Info()

	taskID, err := act.gseHandler.ExecuteScript(std.Context(),
		types.ScriptTypeBash,
		fmt.Sprintf(
			`mkdir -p %s && cd %s && echo "%s" > upgrade.sh && sh upgrade.sh`,
			param.InstallWorkDir,
			param.InstallWorkDir,
			upgradeCmd),
		upgradeProxyScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: std.DeployInfo().Host.Dynamic.AgentID,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to execute proxy upgrade script: %w", err)
	}

	std.InstanceData().Log().
		Zh("升级 Proxy task-id: %s", taskID).
		En("upgrade proxy task-id: %s", taskID).
		Info()

	return nil
}

func (act *actionUpgradeProxy) doUpgradeWindows(std *nodeUtils.NodeActionStandarder, param *installer.NodeUpgradeParams) error {
	_, upgradeCmd, err := param.ToWindowsScript()
	if err != nil {
		return fmt.Errorf("failed to render proxy upgrade script: %w", err)
	}

	std.InstanceData().Log().
		Zh("升级 Proxy 命令: %s", upgradeCmd).
		En("upgrade proxy cmd: %s", upgradeCmd).
		Info()

	taskID, err := act.gseHandler.ExecuteScript(std.Context(),
		types.ScriptTypeBat,
		fmt.Sprintf(
			`cd %s && %s`,
			param.InstallWorkDir,
			upgradeCmd),
		upgradeProxyScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: std.DeployInfo().Host.Dynamic.AgentID,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to execute proxy upgrade script: %w", err)
	}

	std.InstanceData().Log().
		Zh("升级 Proxy task-id: %s", taskID).
		En("upgrade proxy task-id: %s", taskID).
		Info()

	return nil
}
