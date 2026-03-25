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
		storageNetworkArea:    capability.StorageTopo,
		storageNetworkUnit:    capability.StorageTopo,
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
	storageNetworkArea    topoStg.IStorageNetworkArea
	storageNetworkUnit    topoStg.IStorageNetworkUnit
	storageNodeDeployment nodeStg.IDaoNodeDeployment
}

// Name returns the name of the action.
func (act *TryReuseAgentID) Name() string {
	return ActionNameTryReuseAgentID
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *TryReuseAgentID) DisplayNameZh() string {
	return "尝试复用 Agent ID"
}

// DisplayNameEn returns the English display name of the action.
func (act *TryReuseAgentID) DisplayNameEn() string {
	return "Try Reuse Agent ID"
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
	std := nodeUtils.NewNodeActionStandarder(act.storageNodeDeployment, act.storageHost)
	if err = std.Initialize(ctx, param.NodeActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	networkArea, err := act.storageNetworkArea.GetNetworkArea(std.Context(), std.DeployInfo().Host.Static.NetworkAreaID)
	if err != nil {
		return fmt.Errorf("failed to get network area: %w", err)
	}

	networkUnit, err := act.storageNetworkUnit.GetNetworkUnit(std.Context(), std.DeployInfo().Host.Dynamic.NetworkUnitID)
	if err != nil {
		return fmt.Errorf("failed to get network unit: %w", err)
	}

	// log some basic information.
	std.InstanceData().Log().
		Zh("所属管控区域: (%d)%s", networkArea.ID, networkArea.Name).
		En("Network Area: (%d)%s", networkArea.ID, networkArea.Name).
		Info()
	std.InstanceData().Log().
		Zh("所属管控单元: (%d)%s", networkUnit.ID, networkUnit.Name).
		En("Network Unit: (%d)%s", networkUnit.ID, networkUnit.Name).
		Info()

	if std.DeployInfo().Host.Static.Addressing == types.AddressingStatic {
		std.InstanceData().Log().Zh("IP寻址: 静态").En("Addressing: Static").Info()
	} else {
		std.InstanceData().Log().Zh("IP寻址: 动态").En("Addressing: Dynamic").Info()
	}

	if len(std.DeployInfo().Host.Static.InnerIPList) > 0 {
		std.InstanceData().Log().
			Zh("内网IPV4: %v", std.DeployInfo().Host.Static.InnerIPList).
			En("Inner IPV4: %v", std.DeployInfo().Host.Static.InnerIPList).
			Info()
	}
	if len(std.DeployInfo().Host.Static.InnerIPV6List) > 0 {
		std.InstanceData().Log().
			Zh("内网IPV6: %v", std.DeployInfo().Host.Static.InnerIPV6List).
			En("Inner IPV6: %v", std.DeployInfo().Host.Static.InnerIPV6List).
			Info()
	}

	// To reduce the frequency of cache invalidation in downstream systems, reuse the AgentID as much as possible.
	// Notice: Since there will be a large number of if judgments here,
	// it is recommended that when adding a new judgment, it is recommended to use the principle of fast ending, i.e.,
	// if the current judgment is not satisfied, just go back.

	// force re-register the agentID.
	if std.DeployInfo().InstallOptions.ReRegister {
		std.InstanceData().Log().
			Zh("强制重新注册agent-id").
			En("force re-register agent-id").
			Info()
		logger.G.Sys().Info("force re-register agent id, will not reuse agent id")

		return nil
	}

	// try to reuse the agentID.
	hosts, count, err := act.storageHost.ListHost(std.Context(), types.Page{
		Offset: 0,
		Limit:  1,
	}, &types.HostCondition{
		StaticExactInclude: &types.HostStaticExactFields{
			NetworkAreaID: []int64{std.DeployInfo().Host.Static.NetworkAreaID},
			Addressing:    []types.Addressing{std.DeployInfo().Host.Static.Addressing},
			InnerIP:       std.DeployInfo().Host.Static.InnerIPList,
		},
	})
	if err != nil {
		return fmt.Errorf("reuse agent id failed: %w", err)
	}

	// not match host, can't reuse.
	// maybe: host don't exist, or host 's network area changed.
	if count == 0 {
		std.InstanceData().Log().
			Zh("未匹配到主机, 无法复用 agent-id").
			En("not match host, can't reuse agent-id").
			Warn()
		logger.G.Sys().Info("not match host, can't reuse agent id")

		return nil
	}

	std.DeployInfo().Host.Dynamic.AgentID = hosts[0].Dynamic.AgentID

	std.InstanceData().Log().
		Zh("复用已存在的agent-id: %s", std.DeployInfo().Host.Dynamic.AgentID).
		En("reuse existing agent-id: %s", std.DeployInfo().Host.Dynamic.AgentID).
		Info()
	logger.G.Sys().With("agent-id", std.DeployInfo().Host.Dynamic.AgentID).Info("find agent id, try reuse it")

	return nil
}
