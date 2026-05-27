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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// PluginWorkflow represents a plugin workflow.
type PluginWorkflow struct {
	TenantID   string
	WorkflowID string
	TriggerID  string
	Type       PluginWorkflowType
	HostIDs    []int64
	BizIDs     []int64
	Operator   string

	OperateTime time.Time
	FinishTime  time.Time
	Status      PluginWorkflowStatus
}

// PluginWorkflowType represents the operation type of a plugin workflow.
type PluginWorkflowType string

const (
	// PluginWorkflowTypeInstall is the operation type for install plugin.
	PluginWorkflowTypeInstall PluginWorkflowType = "install_plugin"

	// PluginWorkflowTypeUpgrade is the operation type for upgrade plugin.
	PluginWorkflowTypeUpgrade PluginWorkflowType = "upgrade_plugin"

	// PluginWorkflowTypeUninstall is the operation type for uninstall plugin.
	PluginWorkflowTypeUninstall PluginWorkflowType = "uninstall_plugin"

	// PluginWorkflowTypeReconfig is the operation type for reconfig plugin.
	PluginWorkflowTypeReconfig PluginWorkflowType = "reconfig_plugin"

	// PluginWorkflowTypeApplyPluginSubConfig is the operation type for apply plugin sub config.
	PluginWorkflowTypeApplyPluginSubConfig = "apply_plugin_subconfig"

	// PluginWorkflowTypeRestart is the operation type for restart plugin.
	PluginWorkflowTypeRestart PluginWorkflowType = "restart_plugin"

	// PluginWorkflowTypeStop is the operation type for stop plugin.
	PluginWorkflowTypeStop PluginWorkflowType = "stop_plugin"

	// PluginWorkflowTypePluginEnsureV2 is the operation type for ensure plugin v2.
	PluginWorkflowTypePluginEnsureV2 PluginWorkflowType = "ensure_plugin_v2"

	// PluginWorkflowTypeUninstallV2 is the operation type for uninstall plugin v2.
	PluginWorkflowTypeUninstallV2 PluginWorkflowType = "uninstall_plugin_v2"

	// PluginWorkflowTypeMigrateV2 is the operation type for migrate plugin v2.
	PluginWorkflowTypeMigrateV2 PluginWorkflowType = "migrate_plugin_v2"
)

// Validate validates the plugin workflow type.
func (pluginWorkflowType PluginWorkflowType) Validate() error {
	switch pluginWorkflowType {
	case PluginWorkflowTypeInstall,
		PluginWorkflowTypeUpgrade,
		PluginWorkflowTypeUninstall,
		PluginWorkflowTypeReconfig,
		PluginWorkflowTypeApplyPluginSubConfig,
		PluginWorkflowTypeRestart,
		PluginWorkflowTypeStop:
		return nil
	default:
		return fmt.Errorf("invalid plugin workflow type: %s", pluginWorkflowType)
	}
}

// PluginWorkflowTypeListToStringList convert plugin workflow type list to string list.
func PluginWorkflowTypeListToStringList(types []PluginWorkflowType) []string {
	strList := make([]string, 0, len(types))
	for _, t := range types {
		strList = append(strList, string(t))
	}

	return strList
}

// StringListToPluginWorkflowTypeList convert string list to plugin workflow type list.
func StringListToPluginWorkflowTypeList(strList []string) []PluginWorkflowType {
	types := make([]PluginWorkflowType, 0, len(strList))
	for _, str := range strList {
		types = append(types, PluginWorkflowType(str))
	}

	return types
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

// PluginWorkflowStatusListToStringList convert plugin workflow status list to string list.
func PluginWorkflowStatusListToStringList(statuses []PluginWorkflowStatus) []string {
	strList := make([]string, 0, len(statuses))
	for _, s := range statuses {
		strList = append(strList, string(s))
	}

	return strList
}

// StringListToPluginWorkflowStatusList convert string list to plugin workflow status list.
func StringListToPluginWorkflowStatusList(strList []string) []PluginWorkflowStatus {
	statuses := make([]PluginWorkflowStatus, 0, len(strList))
	for _, str := range strList {
		statuses = append(statuses, PluginWorkflowStatus(str))
	}

	return statuses
}

// GetFinishedPluginWorkflowStatus returns the finished plugin workflow status.
func GetFinishedPluginWorkflowStatus() []PluginWorkflowStatus {
	return []PluginWorkflowStatus{
		PluginWorkflowStatusSuccess,
		PluginWorkflowStatusFailed,
		PluginWorkflowStatusPartialFailed,
	}
}

// PluginWorkflowListOperationResult operation list result.
type PluginWorkflowListOperationResult struct {
	HostID        int64
	BizID         int64
	NetworkAreaID int64
	NetworkUnitID int64
	InnerIPList   []string
	InnerIPV6List []string
	PluginName    string
	PluginVersion string

	OperationID     string
	OperInstanceIDs []string

	Operator   string
	CreateTime time.Time

	// LastInstanceBriefData is the last instance brief data.
	// currently support life-cycle and latest action.
	LastInstanceBriefData *operation.InstanceBriefData
}

// PluginWorkflowOperationRetryParam validates the retry param.
type PluginWorkflowOperationRetryParam struct {
	WorkflowID   string
	OperationIDs []string
	RetryMode    operation.RetryMode
}

// PluginWorkflowOperationTerminateParam validates the terminate param.
type PluginWorkflowOperationTerminateParam struct {
	WorkflowID   string
	OperationIDs []string
}
