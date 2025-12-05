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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameResetNodeDynamic the action name.
	ActionNameResetNodeDynamic = "reset_node_dynamic"
)

// NewActionResetNodeDynamic create the action.
func NewActionResetNodeDynamic(capability *Capability) action.Definition {
	return &actionResetNodeDynamic{
		storageNodeDeployment: capability.StorageNode,
	}
}

// ActionParamResetNodeDynamic the action param.
type ActionParamResetNodeDynamic struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionResetNodeDynamic struct {
	storageNodeDeployment nodeStg.IDaoNodeDeployment
}

// Name returns the name of the action.
func (act *actionResetNodeDynamic) Name() string {
	return ActionNameResetNodeDynamic
}

// Version returns the version of the action.
func (act *actionResetNodeDynamic) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionResetNodeDynamic) Description() string {
	return "reset node dynamic"
}

// Timeout returns the timeout of the action.
func (act *actionResetNodeDynamic) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionResetNodeDynamic) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionResetNodeDynamic) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionResetNodeDynamic) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: funlen,nonamedreturns
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionResetNodeDynamic) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamResetNodeDynamic)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := nodeUtils.NewNodeActionStandarder(act.storageNodeDeployment)
	if err = std.Initialize(ctx, param.NodeActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	std.DeployInfo().Host.Dynamic.NodeRole = types.NodeRoleBlank
	std.DeployInfo().Host.Dynamic.NodeStatus = types.NodeStatusUnknown
	std.DeployInfo().Host.Dynamic.AgentID = ""

	// clear proxy info
	std.DeployInfo().Host.Dynamic.ProxyTags = nil
	std.DeployInfo().Host.Dynamic.ProxyClusterPort = 0
	std.DeployInfo().Host.Dynamic.ProxyDataPort = 0
	std.DeployInfo().Host.Dynamic.ProxyFilePort = 0
	std.DeployInfo().Host.Dynamic.ProxyInstallOriginUnitID = 0
	std.DeployInfo().Host.Dynamic.RelayDownloadPort = 0
	std.DeployInfo().Host.Dynamic.RelayCallbackPort = 0
	std.DeployInfo().Host.Dynamic.ProxyAccessDisabled = false

	if err := act.storageNodeDeployment.UpdateNodeDeploymentInfo(std.Context(), std.Token(), std.DeployInfo()); err != nil {
		return fmt.Errorf("update node deployment info failed: %w", err)
	}

	return nil
}
