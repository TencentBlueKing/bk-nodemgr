/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package utils

import (
	"fmt"

	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

// NewNodeActionStandarder creates a new NodeActionStandarder.
func NewNodeActionStandarder(
	storageNodeDeployment nodeStg.IDaoNodeDeployment,
	storageHost topoStg.IStorageHost,
) *NodeActionStandarder {

	return &NodeActionStandarder{
		storageNodeDeployment: storageNodeDeployment,
		storageHost:           storageHost,
	}
}

// NodeActionStandarder defines the standard parameters of node action.
type NodeActionStandarder struct {
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost

	instanceContext *action.InstanceContext
	param           NodeActionStandardParam

	ctx  contextx.IContext
	info *types.DeploymentInfo
}

// Initialize initializes the NodeActionStandarder.
func (std *NodeActionStandarder) Initialize(instanceContext *action.InstanceContext, param NodeActionStandardParam) error {
	std.instanceContext = instanceContext
	std.param = param

	var err error
	std.info, err = std.storageNodeDeployment.GetNodeDeploymentInfo(std.instanceContext.Ctx, std.param.Token)
	if err != nil {
		return fmt.Errorf("failed to get node deployment info: %w", err)
	}

	std.ctx = contextx.From(std.instanceContext.Ctx, contextx.WithTenantID(std.info.Host.TenantID), contextx.WithBKUsername(std.param.Operator))

	return nil
}

// Save saves the NodeActionStandarder data.
func (std *NodeActionStandarder) Save() error {
	if err := std.storageNodeDeployment.UpdateNodeDeploymentInfo(std.instanceContext.Ctx, std.param.Token, std.info); err != nil {
		return fmt.Errorf("failed to update node deployment info: %w", err)
	}

	return nil
}

// DeployInfo returns the node deployment info.
func (std *NodeActionStandarder) DeployInfo() *types.DeploymentInfo {
	return std.info
}

// Context returns the tenant user context.
func (std *NodeActionStandarder) Context() contextx.IContext {
	return std.ctx
}

// Token returns the token.
func (std *NodeActionStandarder) Token() string {
	return std.param.Token
}

// Operator returns the operator.
func (std *NodeActionStandarder) Operator() string {
	return std.param.Operator
}

// InstanceData returns the instance data.
func (std *NodeActionStandarder) InstanceData() *action.InstanceData {
	return std.instanceContext.Data
}

// ResetInstanceDataContext resets the instance data content.
func (std *NodeActionStandarder) ResetInstanceDataContext() {
	std.instanceContext.Data.Content = conv.StructToMapIgnoreError(std.param)
}

// UpdateInstanceDataContent updates the instance data content.
func (std *NodeActionStandarder) UpdateInstanceDataContent(obj any) error {
	return std.instanceContext.Data.UpdateContent(obj)
}

// NodeActionStandardParam defines the standard parameters of node action.
type NodeActionStandardParam struct {
	Token    string `json:"token"`
	Operator string `json:"operator"`
}
