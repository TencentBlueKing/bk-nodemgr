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
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameSyncNodeInfo defines the action name.
	ActionNameSyncNodeInfo = "sync_node_info"
)

// NewActionSyncNodeInfo get a new action.
func NewActionSyncNodeInfo(capability *Capability) action.Definition {
	return &actionSyncNodeInfo{
		gseClient:             capability.GSEHandler,
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
	}
}

// ActParamSyncNodeInfo ...
type ActParamSyncNodeInfo struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionSyncNodeInfo struct {
	gseClient             gse.IHandler
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
}

// Name returns the name of the action.
func (act *actionSyncNodeInfo) Name() string {
	return ActionNameSyncNodeInfo
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionSyncNodeInfo) DisplayNameZh() string {
	return "同步节点信息"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionSyncNodeInfo) DisplayNameEn() string {
	return "Sync Node Info"
}

// Version returns the version of the action.
func (act *actionSyncNodeInfo) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionSyncNodeInfo) Description() string {
	return "sync node info to db"
}

// Timeout returns the timeout of the action.
func (act *actionSyncNodeInfo) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionSyncNodeInfo) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionSyncNodeInfo) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionSyncNodeInfo) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionSyncNodeInfo) Do(ctx *action.InstanceContext) error {
	param := new(ActParamSyncNodeInfo)
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

	agentInfos, err := act.gseClient.ListAgentInfo(std.Context(), std.DeployInfo().Host.Dynamic.AgentID)
	if err != nil {
		return err
	}

	if len(agentInfos) != 1 {
		return fmt.Errorf("get agent info error, agent-id(%s), aget-infos(%v)", std.DeployInfo().Host.Dynamic.AgentID, agentInfos)
	}

	agentInfo := agentInfos[0]
	std.DeployInfo().Host.Dynamic.NodeCPUArch, err = platfmt.NormalizeArch(string(agentInfo.Arch))
	if err != nil {
		return fmt.Errorf("normalize arch error, agent-id(%s), arch(%s), err(%v)",
			std.DeployInfo().Host.Dynamic.AgentID, agentInfo.Arch, err)
	}

	std.DeployInfo().Host.Dynamic.NodeOsType, err = platfmt.NormalizeOS(string(agentInfo.OSType))
	if err != nil {
		return fmt.Errorf("normalize os error, agent-id(%s), os-type(%s), err(%v)",
			std.DeployInfo().Host.Dynamic.AgentID, agentInfo.OSType, err)
	}

	if err := act.storageNodeDeployment.UpdateNodeDeploymentInfo(std.Context(), param.Token, std.DeployInfo()); err != nil {
		return fmt.Errorf("update info error, agent-id(%s), info(%v)", std.DeployInfo().Host.Dynamic.AgentID, err)
	}

	std.InstanceData().Log().
		Zh("同步节点信息成功, agent-id(%s), os-type(%s), cpu-arch(%s)",
			std.DeployInfo().Host.Dynamic.AgentID, std.DeployInfo().Host.Dynamic.NodeOsType, std.DeployInfo().Host.Dynamic.NodeCPUArch).
		En("sync node info successfully, agent-id(%s), os-type(%s), cpu-arch(%s)",
			std.DeployInfo().Host.Dynamic.AgentID, std.DeployInfo().Host.Dynamic.NodeOsType, std.DeployInfo().Host.Dynamic.NodeCPUArch).
		Info()

	return nil
}
