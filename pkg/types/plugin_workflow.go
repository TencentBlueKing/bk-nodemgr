/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package types

import (
	"fmt"
	"time"
)

// PluginWorkflow represents a plugin workflow.
type PluginWorkflow struct {
	WorkflowID string
	TriggerID  string
	Type       PluginWorkflowType
	HostIDs    []int64
	Operator   string

	OperateTime time.Time
	FinishTime  time.Time
	Status      PluginWorkflowStatus
}

// PluginWorkflowType represents the operation type of a plugin workflow.
type PluginWorkflowType string

const (
	// PluginWorkflowTypeInstallOfficial is the operation type for install official.
	PluginWorkflowTypeInstallOfficial PluginWorkflowType = "install_official"

	// PluginWorkflowTypeUpgradeOfficial is the operation type for upgrade official.
	PluginWorkflowTypeUpgradeOfficial PluginWorkflowType = "upgrade_official"

	// PluginWorkflowTypeReconfigOfficial is the operation type for reconfig official.
	PluginWorkflowTypeReconfigOfficial PluginWorkflowType = "reconfig_official"

	// PluginWorkflowTypeRestartOfficial is the operation type for restart official.
	PluginWorkflowTypeRestartOfficial PluginWorkflowType = "restart_official"
)

// Validate validates the plugin workflow type.
func (pluginWorkflowType PluginWorkflowType) Validate() error {
	switch pluginWorkflowType {
	case PluginWorkflowTypeInstallOfficial,
		PluginWorkflowTypeUpgradeOfficial,
		PluginWorkflowTypeReconfigOfficial,
		PluginWorkflowTypeRestartOfficial:
		return nil
	default:
		return fmt.Errorf("invalid plugin workflow type: %s", pluginWorkflowType)
	}
}

// PluginWorkflowStatus represents the status of a plugin workflow.
type PluginWorkflowStatus string

const (
	// PluginWorkflowStatusRunning is the status when the workflow is running.
	PluginWorkflowStatusRunning PluginWorkflowStatus = "running"

	// PluginWorkflowStatusSuccess is the status when the workflow is successful.
	PluginWorkflowStatusSuccess PluginWorkflowStatus = "success"

	// PluginWorkflowStatusFailed is the status when the workflow has failed.
	PluginWorkflowStatusFailed PluginWorkflowStatus = "failed"

	// PluginWorkflowStatusPartialFailed is the status when the workflow has partially failed.
	PluginWorkflowStatusPartialFailed PluginWorkflowStatus = "partial_failed"
)

// Validate validates the plugin workflow status.
func (pluginWorkflowStatus PluginWorkflowStatus) Validate() error {
	switch pluginWorkflowStatus {
	case PluginWorkflowStatusRunning,
		PluginWorkflowStatusSuccess,
		PluginWorkflowStatusFailed,
		PluginWorkflowStatusPartialFailed:
		return nil
	default:
		return fmt.Errorf("invalid plugin workflow status: %s", pluginWorkflowStatus)
	}
}
