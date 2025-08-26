/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package action

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/common"
)

// InstanceContext represents the context of an action instance.
type InstanceContext struct {
	// Context is the context of the action
	Ctx contextx.IContext

	// Data is the action being executed
	Data *InstanceData
}

// InstanceData represents the data of an action instance.
type InstanceData struct {
	TriggerID           string
	OperationID         string
	OperationDefName    string
	OperationInstanceID string

	Name        string
	Index       int
	TotalIndex  int
	Messages    []common.Message
	Content     map[string]any
	PrivateData map[string]any
	Lifecycle   *Lifecycle
}

// Info gets info string.
func (data *InstanceData) Info() string {
	return fmt.Sprintf("operation-inst-id(%s), index(%d), action-name(%s)",
		data.OperationID, data.Index, data.Name)
}

// LogI logs messages.
func (data *InstanceData) LogI(messages ...string) {
	for _, message := range messages {
		data.Messages = append(data.Messages, common.Message{
			Time:  time.Now(),
			Text:  message,
			Level: "INFO",
		})
	}
}

// LogW logs error messages.
func (data *InstanceData) LogW(messages ...string) {
	for _, message := range messages {
		data.Messages = append(data.Messages, common.Message{
			Time:  time.Now(),
			Text:  message,
			Level: "WARN",
		})
	}
}

// LogE logs error messages.
func (data *InstanceData) LogE(messages ...string) {
	for _, message := range messages {
		data.Messages = append(data.Messages, common.Message{
			Time:  time.Now(),
			Text:  message,
			Level: "ERROR",
		})
	}
}

// IsFirst checks if the action instance is the first one.
func (data *InstanceData) IsFirst() bool {
	return data.Index == 0
}

// IsLast checks if the action instance is the last one.
func (data *InstanceData) IsLast() bool {
	return data.Index == data.TotalIndex-1
}

// NeedExecuted checks if the action instance should be executed.
func (data *InstanceData) NeedExecuted() error {
	switch data.Lifecycle.State {
	case StateSuccess:
		return common.ErrActionAlreadySucceeded()

	case StateSkipped:
		return common.ErrActionSkipped()

	case StatePending:
		return nil

	case StateFailed, StateTimeout, StateTerminated:
		return fmt.Errorf("operation instance has completed. oper-inst-id(%s), action-name(%s), state(%s)",
			data.OperationID, data.Name, data.Lifecycle.State)

	case StateRunning:
		return fmt.Errorf("action is running. oper-inst-id(%s), action-name(%s), state(%s)",
			data.OperationID, data.Name, data.Lifecycle.State)

	default:
		return fmt.Errorf("unexpected action state. oper-inst-id(%s), action-name(%s), state(%s)",
			data.OperationID, data.Name, data.Lifecycle.State)
	}
}

// Lifecycle describes the lifecycle of an action instance.
type Lifecycle struct {
	State     State
	CreatedAt time.Time
	StartedAt time.Time
	EndedAt   time.Time
	StoppedAt time.Time
}

// Start starts the action instance lifecycle.
func (life *Lifecycle) Start() {
	life.State = StateRunning
	life.StartedAt = time.Now()
}

// EndWithErr ends the action instance with error.
func (life *Lifecycle) EndWithErr(err error) {
	life.EndedAt = time.Now()

	if err != nil {
		life.State = StateFailed

		return
	}

	life.State = StateSuccess
}

// EndWithTimeout ends the action instance with timeout.
func (life *Lifecycle) EndWithTimeout() {
	life.EndedAt = time.Now()
	life.State = StateTimeout
}

// EndWithTerminated ends the action instance with terminated.
func (life *Lifecycle) EndWithTerminated() {
	life.EndedAt = time.Now()
	life.State = StateTerminated
}

// State defines action instance state.
type State string

const (
	// StatePending action instance state pending.
	StatePending State = "pending"

	// StateRunning action instance state running.
	StateRunning State = "running"

	// StateSuccess action instance state success.
	StateSuccess State = "success"

	// StateFailed action instance state failed.
	StateFailed State = "failed"

	// StateTimeout action instance state timeout.
	StateTimeout State = "timeout"

	// StateSkipped action instance state skipped.
	StateSkipped State = "skipped"

	// StateTerminated action instance state terminated.
	StateTerminated State = "terminated"

	// StateUnknown action instance state unknown.
	StateUnknown State = "unknown"
)

// Validate ActionInstState.
func (state State) Validate() error {
	switch state {
	case StatePending, StateRunning, StateSuccess, StateFailed,
		StateTimeout, StateSkipped, StateTerminated, StateUnknown:
		return nil

	default:
		return fmt.Errorf("invalid action state, state(%s)", state)
	}
}

// StateListToStringList converts a state list to a string list.
func StateListToStringList(states []State) []string {
	data := make([]string, len(states))
	for idx, state := range states {
		data[idx] = string(state)
	}

	return data
}

// StringListToStateList converts a string list to a state list.
func StringListToStateList(states []string) []State {
	data := make([]State, len(states))
	for idx, state := range states {
		data[idx] = State(state)
	}

	return data
}
