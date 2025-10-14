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
	"fmt"
	"time"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameTryReuseAgentID defines the action name.
	ActionNameTryReuseAgentID = "try_reuse_agent_id"
)

// NewActionTryReuseAgentID ...
func NewActionTryReuseAgentID(capability *Capability) action.Definition {
	return &TryReuseAgentID{
		storageHost:           capability.StorageTopo,
		storageNodeDeployment: capability.StorageNode,
	}
}

// ActParamTryReuseAgentID ...
type ActParamTryReuseAgentID struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

// TryReuseAgentID ...
type TryReuseAgentID struct {
	storageHost           topoStg.IStorageHost
	storageNodeDeployment nodeStg.IDaoNodeDeployment
}

// Name returns the name of the action.
func (act *TryReuseAgentID) Name() string {
	return ActionNameTryReuseAgentID
}

// Version returns the version of the action.
func (act *TryReuseAgentID) Version() string {
	return "v1.0.0"
}

// Description returns the description of the action.
func (act *TryReuseAgentID) Description() string {
	return "This will determine if the AgentID needs to be reused"
}

// Timeout returns the timeout of the action.
func (act *TryReuseAgentID) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *TryReuseAgentID) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *TryReuseAgentID) MaxRetryCount() uint {
	return 3 //nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *TryReuseAgentID) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *TryReuseAgentID) Do(ctx *action.InstanceContext) error {
	param := new(ActParamTryReuseAgentID)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := nodeUtils.NewNodeActionStandarder(act.storageNodeDeployment)
	if err = std.Initialize(ctx, param.NodeActionStandardParam); err != nil {
		return err
	}

	// To reduce the frequency of cache invalidation in downstream systems, reuse the AgentID as much as possible.
	// Notice: Since there will be a large number of if judgments here,
	// it is recommended that when adding a new judgment, it is recommended to use the principle of fast ending, i.e.,
	// if the current judgment is not satisfied, just go back.

	// force re-register the agentID.
	if std.DeployInfo().InstallOptions.ReRegister {
		ctx.Data.LogI("force re-register, will not reuse agent id")
		logger.G.Sys().Info("force re-register, will not reuse agent id")

		return nil
	}

	// try to reuse the agentID.
	hosts, count, err := act.storageHost.ListHost(std.Context(), types.Page{
		Offset: 0,
		Limit:  1,
	}, &types.HostCondition{
		ExactInclude: &types.HostExactFields{
			NetworkAreaID: []int64{std.DeployInfo().Host.Static.NetworkAreaID},
			Addressing:    []types.Addressing{std.DeployInfo().Host.Static.Addressing},
			InnerIP:       []string{std.DeployInfo().Host.Static.InnerIP},
		},
	})
	if err != nil {
		return fmt.Errorf("reuse agent id failed: %w", err)
	}

	// not match host, can't reuse.
	// maybe: host don't exist, or host 's network area changed.
	if count == 0 {
		ctx.Data.LogE("not match host, can't reuse agent id")
		logger.G.Sys().Info("not match host, can't reuse agent id")

		return nil
	}

	std.DeployInfo().Host.Dynamic.AgentID = hosts[0].Dynamic.AgentID

	if err := act.storageNodeDeployment.UpdateNodeDeploymentInfo(ctx.Ctx, param.Token, std.DeployInfo()); err != nil {
		return fmt.Errorf("update node deployment info failed: %w", err)
	}

	ctx.Data.LogI(fmt.Sprintf("find agent id, try reuse it, agent-id(%s)", std.DeployInfo().Host.Dynamic.AgentID))
	logger.G.Sys().With("agent-id", std.DeployInfo().Host.Dynamic.AgentID).Info("find agent id, try reuse it")

	return nil
}
