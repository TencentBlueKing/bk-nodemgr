/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operengine ...
package operengine

import (
	"errors"
	"fmt"
	"time"
)

// ActionInstData action instance.
type ActionInstData struct {
	TriggerID   string
	OperInstID  string
	OperationID string
	OperDefName string
	Name        string
	Index       int
	Messages    []Message
	Content     map[string]any
	PrivateData map[string]any
	Lifecycle   *ActInstLifeCycle
}

// Message ...
type Message struct {
	Time time.Time
	Text string
}

// ActInstLifeCycle define the lifecycle of the action instance.
type ActInstLifeCycle struct {
	State     ActionInstState
	StartedAt time.Time
	EndedAt   time.Time
	StoppedAt time.Time
}

// Info ...
func (data *ActionInstData) Info() string {
	return fmt.Sprintf("operation-inst-id(%s), index(%d), action-name(%s)", data.OperInstID, data.Index, data.Name)
}

// Log log messages.
func (data *ActionInstData) Log(messages ...string) {
	for _, message := range messages {
		data.Messages = append(data.Messages, Message{
			Time: time.Now(),
			Text: message,
		})
	}
}

// OperInstData OperInst data.
type OperInstData struct {
	// the below fields should be written only once
	TriggerID         string
	OperInstID        string
	OperDefName       string
	OperationID       string
	ActionNames       []string
	ParentOperationID string
	Timeout           time.Duration
	InitContent       map[string]any

	// the below fields can be changed
	ActionInstDataMap map[string]*ActionInstData
	Lifecycle         *Lifecycle
}

// Lifecycle describes the lifecycle of the operation instance.
type Lifecycle struct {
	State     OperInstState
	CreatedAt time.Time
	StartedAt time.Time
	EndedAt   time.Time
	StoppedAt time.Time
}

// Validate Lifecycle.
func (l *Lifecycle) Validate() error {
	if err := l.State.Validate(); err != nil {
		return fmt.Errorf("lifecycle state validate failed, err(%w)", err)
	}

	return nil
}

// Validate the operation.
func (data *OperInstData) Validate() error {
	if data.OperInstID == "" {
		return errors.New("operation instance id is empty")
	}

	if data.OperDefName == "" {
		return errors.New("operationDef name is empty")
	}

	if len(data.ActionNames) == 0 {
		return errors.New("action names is empty")
	}

	if err := data.Lifecycle.Validate(); err != nil {
		return fmt.Errorf("operation instance lifecycle validate failed, err(%w)", err)
	}

	return nil
}

// OperInst is a operationDef instance.
type OperInst struct {
	operationDef *operationDef
	data         *OperInstData
}

// Validate  the operation.
func (o *OperInst) Validate() error {
	if err := o.data.Validate(); err != nil {
		return fmt.Errorf("operation instance data validate failed, err(%w)", err)
	}

	if o.operationDef.name != o.data.OperDefName {
		return errors.New("operationDef name not match")
	}

	if len(o.operationDef.actionDefs) != len(o.data.ActionNames) {
		return errors.New("action count not match")
	}

	return nil
}

// GetActionInstData get the action instance.
func (o *OperInst) GetActionInstData(actionName string) (*ActionInstData, error) {
	actionData, ok := o.data.ActionInstDataMap[actionName]
	if !ok {
		return nil, fmt.Errorf("action not found, name(%s)", actionName)
	}

	return actionData, nil
}
