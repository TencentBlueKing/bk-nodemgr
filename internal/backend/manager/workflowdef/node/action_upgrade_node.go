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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameUpgradeNode defines the action name.
	ActionNameUpgradeNode = "upgrade_node"

	upgradeScriptTimeout = 10 * time.Minute
)

// NewActionUpgradeNode get a new action.
func NewActionUpgradeNode(capability *Capability) action.Definition {
	return &actionUpgradeNode{
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		gseHandler:            capability.GSEHandler,
		provider:              capability.DiscoverProvider,
	}
}

// ActionParamUpgradeNode defines the action param.
type ActionParamUpgradeNode struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

// UpgradeParams this struct defines the parameters for upgrading agent.
type UpgradeParams struct {
	AgentID          string
	InstallerName    string
	InstallerWorkDir string
	Generation       types.Generation
	NodeRole         types.NodeRole
	DownloadSvrAddr  string
	CallbackSvrAddr  string
	NodeVersion      string
	DeployToken      string
	OperInstID       string
	BaseWorkDir      string
	BaseDeployDir    string
	AdditionArgs     []string
}

type actionUpgradeNode struct {
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	gseHandler            gse.IHandler
	provider              discover.Provider
}

// Name returns the name of the action.
func (act *actionUpgradeNode) Name() string {
	return ActionNameUpgradeNode
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionUpgradeNode) DisplayNameZh() string {
	return "升级节点"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionUpgradeNode) DisplayNameEn() string {
	return "Upgrade Node"
}

// Version returns the version of the action.
func (act *actionUpgradeNode) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionUpgradeNode) Description() string {
	return "upgrade node"
}

// Timeout returns the timeout of the action.
func (act *actionUpgradeNode) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionUpgradeNode) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionUpgradeNode) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionUpgradeNode) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: funlen,nonamedreturns
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionUpgradeNode) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamUpgradeNode)
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

	var callbackSvrAddr string

	if std.DeployInfo().Host.Dynamic.NodeRole == types.NodeRoleProxy {
		// proxy upgrade: use proxy's own relay callback address,
		// since proxy is the relay in its own network unit.
		proxyEndpoint := discover.Endpoint{
			IPV4: std.DeployInfo().Host.Dynamic.AdvertiseIP,
			IPV6: std.DeployInfo().Host.Dynamic.AdvertiseIPV6,
			Port: int(std.DeployInfo().Host.Dynamic.RelayCallbackPort),
		}
		callbackSvrAddr = nodeUtils.BuildServerURLs(proxyEndpoint)
	} else {
		// direct-link agent upgrade: use discover callback endpoints.
		callbackEndpoints, err := act.provider.SelectEndpoints(
			discover.ServiceNameBackend,
			discover.EndpointNameBackendCallback,
			nodeUtils.DefaultEndpointSelectionCount,
			discover.NewRoundRobinSelector())
		if err != nil {
			return fmt.Errorf("failed to select backend callback endpoints: %w", err)
		}

		callbackSvrAddr = nodeUtils.BuildServerURLs(callbackEndpoints...)
	}

	upgradeParams := &UpgradeParams{
		AgentID:          std.DeployInfo().Host.Dynamic.AgentID,
		InstallerName:    toolName,
		InstallerWorkDir: std.DeployInfo().InstallerRuntime.WorkDir,
		NodeVersion:      std.DeployInfo().Host.Dynamic.NodeVersion,
		Generation:       std.DeployInfo().Host.Dynamic.NodeGeneration,
		NodeRole:         std.DeployInfo().Host.Dynamic.NodeRole,
		CallbackSvrAddr:  callbackSvrAddr,
		DeployToken:      std.Token(),
		OperInstID:       std.InstanceData().OperationInstanceID,
		BaseWorkDir:      std.DeployInfo().InstallerRuntime.BaseWorkDir,
		BaseDeployDir:    std.DeployInfo().BaseRuntime.BaseDeployDir,
	}

	if err := std.UpdateInstanceDataContent(ActionWaitInstallerComplete{
		NodeActionStandardParam: param.NodeActionStandardParam,
		EnsureAgentID:           true,
	}); err != nil {
		return fmt.Errorf("failed to update instance data content: %w", err)
	}

	// exec upgrade command
	if std.DeployInfo().Host.Dynamic.NodeOsType == criteria.OSWindows {
		return act.doUpgradeWindows(std, upgradeParams)
	}

	return act.doUpgradeUnix(std, upgradeParams)
}

func (act *actionUpgradeNode) doUpgradeUnix(std *nodeUtils.NodeActionStandarder, param *UpgradeParams) error {
	upgradeParams := &installer.NodeUpgradeParams{
		NodeCommonParams: installer.NodeCommonParams{
			DeployEnv:     system.GetEnv(),
			Generation:    int(param.Generation),
			NodeRole:      string(param.NodeRole),
			BaseWorkDir:   param.BaseWorkDir,
			BaseDeployDir: param.BaseDeployDir,
			AdditionArgs:  param.AdditionArgs,
		},
		InstallWorkDir:    param.InstallerWorkDir,
		InstallerFileName: param.InstallerName,
		CallbackSvrAddr:   param.CallbackSvrAddr,
		DeployToken:       param.DeployToken,
		NodeVersion:       param.NodeVersion,
		OperInstID:        param.OperInstID,
		SkipDownload:      true,
	}
	_, upgradeCmd, err := upgradeParams.ToUnixScript()
	if err != nil {
		return fmt.Errorf("failed to render node upgrade script: %w", err)
	}

	std.InstanceData().Log().
		Zh("升级节点命令: %s", upgradeCmd).
		En("upgrade node cmd: %s", upgradeCmd).
		Info()

	taskID, err := act.gseHandler.ExecuteScript(std.Context(),
		types.ScriptTypeBash,
		fmt.Sprintf(
			`mkdir -p %s && cd %s && echo "%s" > upgrade.sh && sh upgrade.sh`,
			param.InstallerWorkDir,
			param.InstallerWorkDir,
			upgradeCmd),
		upgradeScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: param.AgentID,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to execute upgrade script: %w", err)
	}
	std.InstanceData().Log().
		Zh("升级节点 task-id: %s", taskID).
		En("upgrade node task-id: %s", taskID).
		Info()

	return nil
}

func (act *actionUpgradeNode) doUpgradeWindows(std *nodeUtils.NodeActionStandarder, param *UpgradeParams) error {
	upgradeParams := &installer.NodeUpgradeParams{
		NodeCommonParams: installer.NodeCommonParams{
			DeployEnv:     system.GetEnv(),
			Generation:    int(param.Generation),
			NodeRole:      string(param.NodeRole),
			BaseWorkDir:   param.BaseWorkDir,
			BaseDeployDir: param.BaseDeployDir,
			AdditionArgs:  param.AdditionArgs,
		},
		InstallWorkDir:    param.InstallerWorkDir,
		InstallerFileName: param.InstallerName,
		CallbackSvrAddr:   param.CallbackSvrAddr,
		DeployToken:       param.DeployToken,
		NodeVersion:       param.NodeVersion,
		OperInstID:        param.OperInstID,
		SkipDownload:      true,
	}
	_, upgradeCmd, err := upgradeParams.ToWindowsScript()
	if err != nil {
		return fmt.Errorf("failed to render node upgrade script: %w", err)
	}

	std.InstanceData().Log().
		Zh("升级节点命令: %s", upgradeCmd).
		En("upgrade node cmd: %s", upgradeCmd).
		Info()

	taskID, err := act.gseHandler.ExecuteScript(std.Context(),
		types.ScriptTypeBat,
		fmt.Sprintf(
			`cd %s && %s`,
			param.InstallerWorkDir,
			upgradeCmd),
		upgradeScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: param.AgentID,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to execute upgrade script: %w", err)
	}
	std.InstanceData().Log().
		Zh("升级节点 task-id: %s", taskID).
		En("upgrade node task-id: %s", taskID).
		Info()

	return nil
}
