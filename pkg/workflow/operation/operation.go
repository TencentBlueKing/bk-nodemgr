/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operation provide operation related structs.
package operation

import (
	"fmt"
	"time"
)

const (
	// MaxInstanceNum is the max instance num.
	MaxInstanceNum = 100
)

// RetryMode defines the retry mode.
type RetryMode string

const (
	// RetryModeAll means retry all failed actions.
	RetryModeAll RetryMode = "ALL"
	// RetryModePartial means retry partial actions.
	RetryModePartial RetryMode = "PARTIAL"
)

// RetryFlag defines the retry flag.
type RetryFlag struct {
	Mode             RetryMode
	SourceInstanceID string
	RetryInstanceID  string
}

// Operation defines the operation in workflow.
type Operation struct {
	TriggerID   string
	OperationID string
	Definition  Definition
	InstanceIDs []string
	Param       Param
	RetryFlags  []RetryFlag
	CreateTime  time.Time
	// the latest instance brief data.
	// currently support lifecycle, last inst action and instance-id.
	LatestInstBriefData *InstanceBriefData
}

// CheckEnforceability checks the enforceability of operation.
func (o *Operation) CheckEnforceability() error {
	if len(o.InstanceIDs) >= MaxInstanceNum {
		return fmt.Errorf("operation can not be executed, instances num %d", len(o.InstanceIDs))
	}

	return nil
}

// GetLastInstanceID gets the last instance id.
func (o *Operation) GetLastInstanceID() string {
	if len(o.InstanceIDs) == 0 {
		return ""
	}

	return o.InstanceIDs[len(o.InstanceIDs)-1]
}

// GetNonLastInstanceIDs gets the non-last instance ids.
func (o *Operation) GetNonLastInstanceIDs() []string {
	if len(o.InstanceIDs) <= 1 {
		return []string{}
	}

	return o.InstanceIDs[:len(o.InstanceIDs)-1]
}

// GetLastRetryFlag gets the last retry flag.
func (o *Operation) GetLastRetryFlag() *RetryFlag {
	if len(o.RetryFlags) == 0 {
		return nil
	}

	return &o.RetryFlags[len(o.RetryFlags)-1]
}

// Param defines the operation param.
type Param struct {
	ParentOperationID string
	ParentOperInstID  string
	Timeout           time.Duration
	InitContent       map[string]any
	RetryStartPoint   map[string]bool
}

// InstanceStatusDistribution represents the operation instance status distribution for a trigger.
type InstanceStatusDistribution struct {
	NotInitedCount int64            // not inited count
	StatusMap      map[string]int64 // status -> count
}
