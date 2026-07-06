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
	// ActionNameUninstallNode defines the action name.
	ActionNameUninstallNode = "uninstall_node"
	// ActionNameUninstallNodeSkipReport defines the action name for uninstall without installer report.
	ActionNameUninstallNodeSkipReport = "uninstall_node_skip_report"

	uninstallScriptTimeout = 10 * time.Minute
)

// NewActionUninstallNode get a new action.
func NewActionUninstallNode(capability *Capability) action.Definition {
	return newActionUninstallNode(capability, ActionNameUninstallNode, false)
}

// NewActionUninstallNodeSkipReport get a new action that skips installer report.
func NewActionUninstallNodeSkipReport(capability *Capability) action.Definition {
	return newActionUninstallNode(capability, ActionNameUninstallNodeSkipReport, true)
}

func newActionUninstallNode(capability *Capability, name string, skipCallback bool) action.Definition {
	return &actionUninstallNode{
		name:                  name,
		skipCallback:          skipCallback,
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		storageNetworkUnit:    capability.StorageTopo,
		gseHandler:            capability.GSEHandler,
		provider:              capability.DiscoverProvider,
		storageActionInstance: capability.StorageWorkflow,
	}
}

// ActionParamUninstallNode defines the action param.
type ActionParamUninstallNode struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionUninstallNode struct {
	name                  string
	skipCallback          bool
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	storageNetworkUnit    topoStg.IStorageNetworkUnit
	gseHandler            gse.IHandler
	provider              discover.Provider
	storageActionInstance workflow.IStorageActionInstance
}

// Name returns the name of the action.
func (act *actionUninstallNode) Name() string {
	return act.name
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionUninstallNode) DisplayNameZh() string {
	return "卸载节点"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionUninstallNode) DisplayNameEn() string {
	return "Uninstall Node"
}

// Version returns the version of the action.
func (act *actionUninstallNode) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionUninstallNode) Description() string {
	return "uninstall node"
}

// Timeout returns the timeout of the action.
func (act *actionUninstallNode) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionUninstallNode) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionUninstallNode) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionUninstallNode) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: funlen,nonamedreturns
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionUninstallNode) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamUninstallNode)
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

	// select matching tools.
	toolName, err := tool.FormatInstallerName(std.DeployInfo().Host.Dynamic.NodeOsType, std.DeployInfo().Host.Dynamic.NodeCPUArch)
	if err != nil {
		return err
	}

	uninstallParams := &installer.NodeUninstallParams{
		NodeCommonParams: installer.NodeCommonParams{
			DeployEnv:     system.GetEnv(),
			Generation:    int(std.DeployInfo().Host.Dynamic.NodeGeneration),
			NodeRole:      string(std.DeployInfo().Host.Dynamic.NodeRole),
			BaseWorkDir:   std.DeployInfo().InstallerRuntime.BaseWorkDir,
			BaseDeployDir: std.DeployInfo().BaseRuntime.BaseDeployDir,
		},
		InstallWorkDir:    std.DeployInfo().InstallerRuntime.WorkDir,
		InstallerFileName: toolName,
		DeployToken:       std.Token(),
		OperInstID:        std.InstanceData().OperationInstanceID,
		SkipCallback:      act.skipCallback,
	}

	if !act.skipCallback {
		// let the callback server known which action to mark and log.
		if err := std.SaveBlockingActionName(ActionNameWaitInstallerComplete); err != nil {
			return fmt.Errorf("failed to save blocking action name: %w", err)
		}

		callbackEndpoints, _, err := nodeUtils.GenerateNodeInstallerServerEndpoints(
			std, act.provider, nodeUtils.NodeInstallerEndpointSourceServer)
		if err != nil {
			return fmt.Errorf("failed to generate node installer server endpoints: %w", err)
		}

		uninstallParams.CallbackSvrAddr = nodeUtils.BuildServerURLs(callbackEndpoints...)
		err = saveWaitInstallerPrivateData(
			std.Context(), act.storageActionInstance, std.InstanceData().OperationInstanceID,
			false, false)
		if err != nil {
			return err
		}
	}

	// exec uninstall command
	if std.DeployInfo().Host.Dynamic.NodeOsType == criteria.OSWindows {
		return act.doUninstallWindows(std, uninstallParams)
	}

	return act.doUninstallUnix(std, uninstallParams)
}

// nolint: perfsprint
func (act *actionUninstallNode) doUninstallUnix(std *nodeUtils.NodeActionStandarder, param *installer.NodeUninstallParams) error {
	_, uninstallCmd, err := param.ToUnixScript()
	if err != nil {
		return fmt.Errorf("failed to render node uninstall script: %w", err)
	}
	std.InstanceData().Log().
		Zh("卸载节点命令: %s", uninstallCmd).
		En("uninstall node cmd: %s", uninstallCmd).
		Info()

	taskID, err := act.gseHandler.ExecuteScript(std.Context(),
		types.ScriptTypeBash,
		fmt.Sprintf(
			`mkdir -p %s && cd %s && echo "%s" > uninstall.sh && sh uninstall.sh`,
			param.InstallWorkDir,
			param.InstallWorkDir,
			uninstallCmd),
		uninstallScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: std.DeployInfo().Host.Dynamic.AgentID,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to execute uninstall script: %w", err)
	}
	std.InstanceData().Log().
		Zh("卸载节点 task-id: %s", taskID).
		En("uninstall node task-id: %s", taskID).
		Info()

	return nil
}

// nolint: perfsprint
func (act *actionUninstallNode) doUninstallWindows(std *nodeUtils.NodeActionStandarder, param *installer.NodeUninstallParams) error {
	_, uninstallCmd, err := param.ToWindowsScript()
	if err != nil {
		return fmt.Errorf("failed to render node uninstall script: %w", err)
	}
	std.InstanceData().Log().
		Zh("卸载节点命令: %s", uninstallCmd).
		En("uninstall node cmd: %s", uninstallCmd).
		Info()

	taskID, err := act.gseHandler.ExecuteScript(std.Context(),
		types.ScriptTypeBat,
		fmt.Sprintf(
			`cd %s && %s`,
			param.InstallWorkDir,
			uninstallCmd),
		uninstallScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: std.DeployInfo().Host.Dynamic.AgentID,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to execute uninstall script: %w", err)
	}
	std.InstanceData().Log().
		Zh("卸载节点 task-id: %s", taskID).
		En("uninstall node task-id: %s", taskID).
		Info()

	return nil
}
