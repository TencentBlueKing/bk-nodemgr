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
	"time"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameUnbindAgentHostRel defines the action name.
	ActionNameUnbindAgentHostRel = "unbind_agent_host_rel"
)

// NewActionUnbindAgentHostRel get a new action.
func NewActionUnbindAgentHostRel(capability *Capability) action.Definition {
	return &actionUnbindAgentHostRel{
		IUnbindHostAgent:      capability.CMDBHandler,
		storageHost:           capability.StorageTopo,
		storageNodeDeployment: capability.StorageNode,
	}
}

// ActParamUnbindAgentHostRel defines the action param.
type ActParamUnbindAgentHostRel struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionUnbindAgentHostRel struct {
	cmdb.IUnbindHostAgent
	storageHost           topoStg.IStorageHost
	storageNodeDeployment nodeStg.IDaoNodeDeployment
}

// Name returns the name of the action.
func (act *actionUnbindAgentHostRel) Name() string {
	return ActionNameUnbindAgentHostRel
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionUnbindAgentHostRel) DisplayNameZh() string {
	return "解除 Agent 与主机关联"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionUnbindAgentHostRel) DisplayNameEn() string {
	return "Unbind Agent-Host Relation"
}

// Version returns the version of the action.
func (act *actionUnbindAgentHostRel) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionUnbindAgentHostRel) Description() string {
	return "unbind agent host relation"
}

// Timeout returns the timeout of the action.
func (act *actionUnbindAgentHostRel) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionUnbindAgentHostRel) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionUnbindAgentHostRel) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionUnbindAgentHostRel) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionUnbindAgentHostRel) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamUnbindAgentHostRel)
	err = conv.MapToStruct(ctx.Data.Content, param)
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

	// this is a special case, when the deployment is reverted, the host id is not in the host table.
	if err := act.checkHostExist(std.Context(), std.DeployInfo()); err != nil {
		return err
	}

	if std.DeployInfo().Host.Dynamic == nil || std.DeployInfo().Host.Dynamic.AgentID == "" {
		std.InstanceData().Log().
			Zh("跳过解除主机与Agent关联, 主机ID(%d), agent-id为空", std.DeployInfo().Host.HostID).
			En("skip unbinding host agent relation, host-id(%d), agent-id is empty", std.DeployInfo().Host.HostID).
			Info()
		logger.G.Sys().Ctx(std.Context()).
			With("host-id", std.DeployInfo().Host.HostID, "agent-id", "").
			Info("skip unbinding host agent relation because agent id is empty")

		return nil
	}

	agentID := std.DeployInfo().Host.Dynamic.AgentID
	if err := act.UnbindHostAgent(std.Context(), &std.DeployInfo().Host); err != nil {
		return fmt.Errorf("unbind host agent relation failed: %w", err)
	}
	logger.G.Sys().Ctx(std.Context()).
		With("host-id", std.DeployInfo().Host.HostID, "agent-id", agentID).
		Info("successfully unbind host agent relation from cmdb")

	std.DeployInfo().Host.Dynamic.AgentID = ""
	if err := act.storageHost.UpdateHostDynamicFields(
		std.Context(), types.HostDynamicFields{AgentID: true}, &std.DeployInfo().Host,
	); err != nil {
		std.DeployInfo().Host.Dynamic.AgentID = agentID
		return fmt.Errorf("unbind host agent relation failed: %w", err)
	}
	logger.G.Sys().Ctx(std.Context()).
		With("host-id", std.DeployInfo().Host.HostID, "agent-id", agentID).
		Info("successfully unbind host agent relation from db")

	std.InstanceData().Log().
		Zh("解除主机与Agent关联成功, 主机ID(%d), agent-id(%s)", std.DeployInfo().Host.HostID, agentID).
		En("successfully unbind host agent relation, host-id(%d), agent-id(%s)", std.DeployInfo().Host.HostID, agentID).
		Info()

	logger.G.Sys().Ctx(std.Context()).
		With("host-id", std.DeployInfo().Host.HostID, "agent-id", agentID).
		Info("successfully unbind host agent relation")

	return nil
}

func (act *actionUnbindAgentHostRel) checkHostExist(nCtx contextx.IContext, info *types.DeploymentInfo) error {
	daoHost, err := act.storageHost.GetHostByID(nCtx, info.Host.HostID)
	if err != nil {
		return fmt.Errorf("get host info failed: %w", err)
	}

	if daoHost == nil {
		return fmt.Errorf("host not found, host-id(%d)", info.Host.HostID)
	}

	return nil
}
