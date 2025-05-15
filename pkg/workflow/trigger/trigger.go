/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package trigger

import (
	"time"
)

// Trigger defines the trigger, which will handle the workflow.
type Trigger struct {
	TriggerID       string
	Category        Category
	Metadata        Metadata
	State           State
	CreatedAt       time.Time
	UpdatedAt       time.Time
	LastTriggeredAt time.Time
}

// Category represents the category of a trigger.
type Category string

const (
	// CategoryOnce represents a trigger that is executed only once.
	CategoryOnce Category = "once"

	// CategoryPeriodic represents a trigger that is executed periodically.
	CategoryPeriodic Category = "periodic"

	// CategoryOrdered represents a trigger that is executed in order.
	CategoryOrdered Category = "ordered"
)

// State represents the state of a trigger.
type State string

const (
	// StateInit represents the initial state of a trigger.
	StateInit State = "init"

	// StateRunning represents a trigger that is currently running.
	StateRunning State = "running"

	// StateTerminated represents a trigger that has been terminated.
	StateTerminated State = "terminated"
)
