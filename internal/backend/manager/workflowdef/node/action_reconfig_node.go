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
	// ActionNameReconfigNode defines the action name.
	ActionNameReconfigNode = "reconfig_node"

	reconfigScriptTimeout = 10 * time.Minute
)

// NewActionReconfigNode get a new action.
func NewActionReconfigNode(capability *Capability) action.Definition {
	return &actionReconfigNode{
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		gseHandler:            capability.GSEHandler,
		provider:              capability.DiscoverProvider,
		storageActionInstance: capability.StorageWorkflow,
	}
}

// ActionParamReconfigNode defines the action param.
type ActionParamReconfigNode struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionReconfigNode struct {
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	gseHandler            gse.IHandler
	provider              discover.Provider
	storageActionInstance workflow.IStorageActionInstance
}

// Name returns the name of the action.
func (act *actionReconfigNode) Name() string {
	return ActionNameReconfigNode
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionReconfigNode) DisplayNameZh() string {
	return "重新配置节点"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionReconfigNode) DisplayNameEn() string {
	return "Reconfigure Node"
}

// Version returns the version of the action.
func (act *actionReconfigNode) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionReconfigNode) Description() string {
	return "reconfig node"
}

// Timeout returns the timeout of the action.
func (act *actionReconfigNode) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionReconfigNode) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionReconfigNode) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionReconfigNode) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: funlen,nonamedreturns
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionReconfigNode) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamReconfigNode)
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
		// notice: in the normal case, this branch will not be entered, because the reconfig of Proxy will use ActionNameReconfigProxy.
		// proxy reconfig: use proxy's own relay callback address,
		// since proxy is the relay in its own network unit.
		proxyEndpoint := discover.Endpoint{
			Port: int(std.DeployInfo().Host.Dynamic.RelayCallbackPort),
		}

		if len(std.DeployInfo().Host.Static.InnerIPList) > 0 {
			proxyEndpoint.IPV4 = std.DeployInfo().Host.Static.InnerIPList[0]
		}

		if len(std.DeployInfo().Host.Static.InnerIPV6List) > 0 {
			proxyEndpoint.IPV6 = std.DeployInfo().Host.Static.InnerIPV6List[0]
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

	reconfigParams := &installer.NodeReconfigParams{
		NodeCommonParams: installer.NodeCommonParams{
			DeployEnv:     system.GetEnv(),
			Generation:    int(std.DeployInfo().Host.Dynamic.NodeGeneration),
			NodeRole:      string(std.DeployInfo().Host.Dynamic.NodeRole),
			BaseWorkDir:   std.DeployInfo().InstallerRuntime.BaseWorkDir,
			BaseDeployDir: std.DeployInfo().BaseRuntime.BaseDeployDir,
		},
		InstallWorkDir:    std.DeployInfo().InstallerRuntime.WorkDir,
		InstallerFileName: toolName,
		CallbackSvrAddr:   callbackSvrAddr,
		DeployToken:       std.Token(),
		OperInstID:        std.InstanceData().OperationInstanceID,
	}

	err = saveWaitInstallerPrivateData(
		std.Context(), act.storageActionInstance, std.InstanceData().OperationInstanceID,
		false, true)
	if err != nil {
		return err
	}

	// exec reconfig command
	if std.DeployInfo().Host.Dynamic.NodeOsType == criteria.OSWindows {
		return act.doReconfigWindows(std, reconfigParams)
	}

	return act.doReconfigUnix(std, reconfigParams)
}

// nolint: perfsprint
func (act *actionReconfigNode) doReconfigUnix(std *nodeUtils.NodeActionStandarder, param *installer.NodeReconfigParams) error {
	_, reconfigCmd, err := param.ToUnixScript()
	if err != nil {
		return fmt.Errorf("failed to render node reconfig script: %w", err)
	}
	std.InstanceData().Log().
		Zh("重新配置节点命令: %s", reconfigCmd).
		En("reconfig node cmd: %s", reconfigCmd).
		Info()

	taskID, err := act.gseHandler.ExecuteScript(std.Context(),
		types.ScriptTypeBash,
		fmt.Sprintf(
			`mkdir -p %s && cd %s && echo "%s" > reconfig.sh && sh reconfig.sh`,
			param.InstallWorkDir,
			param.InstallWorkDir,
			reconfigCmd),
		reconfigScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: std.DeployInfo().Host.Dynamic.AgentID,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to execute reconfig script: %w", err)
	}
	std.InstanceData().Log().
		Zh("重新配置节点 task-id: %s", taskID).
		En("reconfig node task-id: %s", taskID).
		Info()

	return nil
}

// nolint: perfsprint
func (act *actionReconfigNode) doReconfigWindows(std *nodeUtils.NodeActionStandarder, param *installer.NodeReconfigParams) error {
	_, reconfigCmd, err := param.ToWindowsScript()
	if err != nil {
		return fmt.Errorf("failed to render node reconfig script: %w", err)
	}
	std.InstanceData().Log().
		Zh("重新配置节点命令: %s", reconfigCmd).
		En("reconfig node cmd: %s", reconfigCmd).
		Info()

	taskID, err := act.gseHandler.ExecuteScript(std.Context(),
		types.ScriptTypeBat,
		fmt.Sprintf(
			`cd %s && %s`,
			param.InstallWorkDir,
			reconfigCmd),
		reconfigScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: std.DeployInfo().Host.Dynamic.AgentID,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to execute reconfig script: %w", err)
	}
	std.InstanceData().Log().
		Zh("重新配置节点 task-id: %s", taskID).
		En("reconfig node task-id: %s", taskID).
		Info()

	return nil
}
