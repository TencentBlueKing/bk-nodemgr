/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operation defines the operation instance and its lifecycle.
package operation

import (
	"fmt"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/common"
)

// InstanceMetadata defines the metadata of operation instance.
type InstanceMetadata struct {
	TriggerID              string
	OperationInstanceID    string
	OperationDefName       string
	OperationID            string
	ActionNames            []string
	ParentOperationID      string
	Index                  int
	Timeout                time.Duration
	InitContent            map[string]any
	ExtraExecutionName     string
	ExtraExecutionMutex    sync.Mutex
	ExtraExecutionMessages []common.Message
}

// SetMessage sets the extra execution message for the operation instance.
func (data *InstanceMetadata) SetMessage(zhMsg, enMsg, level string) {
	data.ExtraExecutionMutex.Lock()
	defer data.ExtraExecutionMutex.Unlock()

	data.ExtraExecutionMessages = append(data.ExtraExecutionMessages, common.Message{
		Time:   time.Now(),
		TextZh: zhMsg,
		TextEn: enMsg,
		Level:  level,
	})
}

// Log returns a new LogBuilder for chainable bilingual logging.
func (data *InstanceMetadata) Log() *common.LogBuilder {
	return common.NewLogBuilder(data)
}

// InstanceBriefData defines the brief data of operation instance.
type InstanceBriefData struct {
	// the below fields should be written only once.
	Metadata *InstanceMetadata

	// the latest action brief data.
	LatestActionInstBriefData *action.InstanceBriefData

	// the below fields can be changed.
	Lifecycle *Lifecycle
}

// InstanceData defines the data of operation instance.
type InstanceData struct {
	InstanceBriefData

	// the below fields can be changed.
	ActionInstanceDataMap map[string]*action.InstanceData
}

// Lifecycle describes the lifecycle of the operation instance.
type Lifecycle struct {
	State     State
	CreatedAt time.Time
	StartedAt time.Time
	EndedAt   time.Time
	StoppedAt time.Time
}

// InstanceStatus defines the status of operation instance.
type InstanceStatus struct {
	TriggerID           string
	OperationInstanceID string
	State               State
	Index               int
	OperationID         string
}

// IsRunning checks if the lifecycle is running.
func (life *Lifecycle) IsRunning() bool {
	return life.State == StateRunning
}

// IsTerminated checks if the lifecycle is terminated.
func (life *Lifecycle) IsTerminated() bool {
	return life.State == StateTerminated
}

// Launch lauch the action instance lifecycle.
func (life *Lifecycle) Launch() {
	life.State = StateLaunched
	life.StartedAt = time.Now()
}

// Start starts the action instance lifecycle.
func (life *Lifecycle) Start() {
	life.State = StateRunning
}

// EndByActionState ends the action instance lifecycle.
func (life *Lifecycle) EndByActionState(lastActionInstState action.State) {
	var state State

	switch lastActionInstState {
	case action.StateSuccess, action.StateSkipped:
		state = StateSuccess
	case action.StateFailed:
		state = StateFailed
	case action.StateTimeout:
		state = StateTimeout
	case action.StateTerminated:
		state = StateTerminated
	case action.StateRunning:
		state = StateRunning
	case action.StatePending:
		// TODO: implement me.
		state = StateFailed
	default:
		state = StateFailed
	}

	life.End(state)
}

// End ends the action instance lifecycle.
func (life *Lifecycle) End(state State) {
	life.EndedAt = time.Now()
	life.State = state
}

// NeedExecutedExtraExecution checks if the lifecycle needs to execute extra execution.
func (life *Lifecycle) NeedExecutedExtraExecution() bool {
	switch life.State {
	case StateLaunched, StateSuccess, StateFailed, StateTimeout, StateTerminated:
		return true
	default:
		return false
	}
}

// State defines operation instance state.
type State string

const (
	// StateInit operation instance state init.
	StateInit State = "init"

	// StateLaunched operation instance state launched.
	StateLaunched State = "launched"

	// StateRunning operation instance state running.
	StateRunning State = "running"

	// StateSuccess operation instance state success.
	StateSuccess State = "success"

	// StateFailed operation instance state failed.
	StateFailed State = "failed"

	// StateTimeout operation instance state timeout.
	StateTimeout State = "timeout"

	// StateTerminated operation instance state terminated.
	StateTerminated State = "terminated"
)

// Validate checks if the state is a valid operation instance state.
func (state State) Validate() error {
	switch state {
	case StateInit, StateLaunched, StateRunning,
		StateSuccess, StateFailed, StateTimeout, StateTerminated:
		return nil
	default:
		return fmt.Errorf("invalid operation instance state. state(%s)", state)
	}
}

// CheckStateFinished checks if the state is finished.
func CheckStateFinished(state State) bool {
	switch state {
	case StateSuccess, StateFailed, StateTimeout, StateTerminated:
		return true
	default:
		return false
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

// GetAllStates returns all states.
func GetAllStates() []State {
	return []State{
		StateInit,
		StateLaunched,
		StateRunning,
		StateSuccess,
		StateFailed,
		StateTimeout,
		StateTerminated,
	}
}
