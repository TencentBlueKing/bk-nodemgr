/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package plugin

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IStorage defines the storage interface.
type IStorage interface {
	basestorage.Interface

	IDaoPluginDeployment
	IDaoPluginWorkflow
}

// IDaoPluginDeployment defines the plugin deployment dao interface.
type IDaoPluginDeployment interface {
	// Create create plugin deployment.
	Create(ctx contextx.IContext, pluginDeployment *types.PluginDeployment) error

	// UpdatePluginDeploymentInfo update plugin deployment info.
	UpdatePluginDeploymentInfo(ctx contextx.IContext, token string, pluginDeploymentInfo *types.PluginDeploymentInfo) error

	// GetInfo get plugin deployment info.
	GetInfo(ctx contextx.IContext, token string) (*types.PluginDeploymentInfo, error)
}

// IDaoPluginWorkflow defines the dao interface.
type IDaoPluginWorkflow interface {
	// GetPluginWorkflow gets a plugin workflow by workflow-id.
	GetPluginWorkflow(ctx contextx.ITenantContext, workflowID string) (*types.PluginWorkflow, error)

	// CreatePluginWorkflow creates a new plugin workflow.
	CreatePluginWorkflow(ctx contextx.ITenantContext, workflow *types.PluginWorkflow) error

	// UpdatePluginWorkflowStatus updates the status of a plugin workflow.
	UpdatePluginWorkflowStatus(ctx contextx.ITenantContext, workflowID string, status types.PluginWorkflowStatus) error
}
