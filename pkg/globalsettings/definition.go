/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
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
	// OperSyncHostMaxConcurrencyNum defines the max concurrency setting for sync host operations.
	OperSyncHostMaxConcurrencyNum = "oper_sync_host_max_concurrency_num"
	// DefaultOperSyncHostMaxConcurrencyNum defines the default max concurrency for sync host operations.
	DefaultOperSyncHostMaxConcurrencyNum = "100"
)

// PreDefinition returns the definition of global settings.
func PreDefinition() []*types.GlobalSettings {
	return []*types.GlobalSettings{
		{
			SettingName: CleanTriggerIntervalSecond,
			Value:       scheduler.Every1m,
		},
		{
			SettingName: OperSyncHostMaxConcurrencyNum,
			Value:       DefaultOperSyncHostMaxConcurrencyNum,
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
