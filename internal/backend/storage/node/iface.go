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
	"context"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IStorage defines the storage interface.
type IStorage interface {
	basestorage.Interface

	IDaoNodeDeployment
	IDaoNodeWorkflow
}

// IDaoNodeDeployment defines the node deployment dao interface.
type IDaoNodeDeployment interface {
	// CreateNodeDeployment createNodeDeployment a node deployment.
	CreateNodeDeployment(ctx context.Context, nodeDeployment *types.NodeDeployment) error

	// GetNodeDeploymentNodeConf get gse agent setting.
	GetNodeDeploymentNodeConf(ctx context.Context, token string) (*types.NodeConf, error)

	// SetNodeDeploymentNodeConf set gse agent setting.
	SetNodeDeploymentNodeConf(ctx context.Context, token string, conf *types.NodeConf) error

	// GetNodeDeploymentInfo get node deployment info.
	GetNodeDeploymentInfo(ctx context.Context, token string) (*types.DeploymentInfo, error)

	// UpdateNodeDeploymentInfo update node deployment info.
	UpdateNodeDeploymentInfo(ctx context.Context, token string, info *types.DeploymentInfo) error
}

// IDaoNodeWorkflow define the node workflow dao interface.
type IDaoNodeWorkflow interface {
	// ListNodeWorkflow lists node workflow by page and conditions.
	ListNodeWorkflow(ctx context.Context, page types.Page, conditions ...*types.NodeWorkflowCondition) (
		[]*types.NodeWorkflow, int64, error)

	// CountNodeWorkflow counts node workflow by conditions.
	CountNodeWorkflow(ctx context.Context, conditions ...*types.NodeWorkflowCondition) (int64, error)

	// DistinctNodeWorkflow distincts node workflow fields.
	DistinctNodeWorkflow(
		ctx context.Context, request types.NodeWorkflowDistinctRequest, conditions ...*types.NodeWorkflowCondition) (
		*types.NodeWorkflowDistinctResult, error)

	// GetNodeWorkflow gets a node workflow by workflow-id.
	GetNodeWorkflow(ctx context.Context, workflowID string) (*types.NodeWorkflow, error)

	// CreateNodeWorkflow creates a new node workflow.
	CreateNodeWorkflow(ctx context.Context, workflow *types.NodeWorkflow) error

	// UpdateNodeWorkflowStatus updates the status of a node workflow.
	UpdateNodeWorkflowStatus(ctx context.Context, workflowID string, status types.NodeWorkflowStatus) error
}
