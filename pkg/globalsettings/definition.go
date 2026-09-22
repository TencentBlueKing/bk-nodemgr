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

// Package globalsettings provides a singleton for global settings.
package globalsettings

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/scheduler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	// CleanTriggerIntervalSecond defines the interval of cleaning trigger records.
	CleanTriggerIntervalSecond = "clean_trigger_interval_second"
	// NetworkUnitSegmentRules defines the setting name for network unit segment rules.
	NetworkUnitSegmentRules = "networkunit_segment_rules"
	// PluginCompatibilityModePolicy defines the setting name for plugin compatibility mode policy.
	PluginCompatibilityModePolicy = "plugin_compatibility_mode_policy"

	// OperSyncAgentInfoMaxConcurrencyNum defines the max concurrency setting for sync agent info operations.
	OperSyncAgentInfoMaxConcurrencyNum = "oper_sync_agent_info_max_concurrency_num"
	// OperSyncAgentInfoMaxConcurrencyNumDefault defines the default max concurrency for sync agent info operations.
	OperSyncAgentInfoMaxConcurrencyNumDefault = "100"

	// OperSyncAgentStateMaxConcurrencyNum defines the max concurrency setting for sync agent state operations.
	OperSyncAgentStateMaxConcurrencyNum = "oper_sync_agent_state_max_concurrency_num"
	// OperSyncAgentStateMaxConcurrencyNumDefault defines the default max concurrency for sync agent state operations.
	OperSyncAgentStateMaxConcurrencyNumDefault = "100"

	// OperSyncAlivePluginProcessInfoMaxConcurrencyNum defines the max concurrency setting for sync alive plugin process info operations.
	OperSyncAlivePluginProcessInfoMaxConcurrencyNum = "oper_sync_alive_plugin_process_info_max_concurrency_num"
	// OperSyncAlivePluginProcessInfoMaxConcurrencyNumDefault defines the default max concurrency for sync alive plugin process info operations.
	OperSyncAlivePluginProcessInfoMaxConcurrencyNumDefault = "100"

	// OperCorrectUnknownProcessStatusMaxConcurrencyNum defines the max concurrency setting for process status correction operations.
	OperCorrectUnknownProcessStatusMaxConcurrencyNum = "oper_correct_unknown_process_status_max_concurrency_num"
	// OperCorrectUnknownProcessStatusMaxConcurrencyNumDefault defines the default max concurrency for process status correction operations.
	OperCorrectUnknownProcessStatusMaxConcurrencyNumDefault = "100"

	// OperCleanOrphanProcessPageSize defines the maximum number of processes scanned per cleanup run.
	OperCleanOrphanProcessPageSize = "oper_clean_orphan_process_page_size"
	// OperCleanOrphanProcessPageSizeDefault defines the default cleanup scan limit.
	OperCleanOrphanProcessPageSizeDefault = "1000"

	// OperSyncHostMaxConcurrencyNum defines the max concurrency setting for sync host operations.
	OperSyncHostMaxConcurrencyNum = "oper_sync_host_max_concurrency_num"
	// OperSyncHostMaxConcurrencyNumDefault defines the default max concurrency for sync host operations.
	OperSyncHostMaxConcurrencyNumDefault = "100"
)

// PreDefinition returns the definition of global settings.
func PreDefinition() []*types.GlobalSettings {
	return []*types.GlobalSettings{
		{
			SettingName: CleanTriggerIntervalSecond,
			Value:       scheduler.Every1m,
		},
		{
			SettingName: PluginCompatibilityModePolicy,
			Value:       `{"enabled_plugins":["bkmonitorbeat"],"disabled_biz":[]}`,
		},
		{
			SettingName: OperSyncAgentInfoMaxConcurrencyNum,
			Value:       OperSyncAgentInfoMaxConcurrencyNumDefault,
		},
		{
			SettingName: OperSyncAgentStateMaxConcurrencyNum,
			Value:       OperSyncAgentStateMaxConcurrencyNumDefault,
		},
		{
			SettingName: OperSyncAlivePluginProcessInfoMaxConcurrencyNum,
			Value:       OperSyncAlivePluginProcessInfoMaxConcurrencyNumDefault,
		},
		{
			SettingName: OperCorrectUnknownProcessStatusMaxConcurrencyNum,
			Value:       OperCorrectUnknownProcessStatusMaxConcurrencyNumDefault,
		},
		{
			SettingName: OperCleanOrphanProcessPageSize,
			Value:       OperCleanOrphanProcessPageSizeDefault,
		},
		{
			SettingName: OperSyncHostMaxConcurrencyNum,
			Value:       OperSyncHostMaxConcurrencyNumDefault,
		},
	}
}

var (
	errUninitialized = errors.New("global settings storage not initialized")
	errNonexist      = errors.New("global settings not exist")
)

// ErrUninitialized returns an error indicating that the global settings storage is not initialized.
func ErrUninitialized() error {
	return errUninitialized
}

// ErrNonexist returns an error indicating that the global settings do not exist.
func ErrNonexist() error {
	return errNonexist
}
