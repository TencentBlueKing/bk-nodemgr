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

	"github.com/google/uuid"
)

// OperInstTimeoutDefault default OperInst timeout.
const OperInstTimeoutDefault = 10 * time.Minute

// newOperationDef creates a new operationDef.
func newOperationDef(name string) *operationDef {
	return &operationDef{
		name:       name,
		actionDefs: make([]ActionDef, 0),
	}
}

// operationDef represents a operationDef definition.
type operationDef struct {
	name       string
	actionDefs []ActionDef
}

// Name returns the name of the operationDef.
func (def *operationDef) Name() string {
	return def.name
}

// Next appends a new action to the operationDef.
func (def *operationDef) Next(actionDef ActionDef) *operationDef {
	def.actionDefs = append(def.actionDefs, actionDef)

	return def
}

// OperInstIDPrefix ...
const OperInstIDPrefix = "oper-inst"

// NewInstance creates a new OperInst.
func (def *operationDef) NewInstance(triggerID string, timeout time.Duration) (*OperInst, error) {
	if err := def.Validate(); err != nil {
		return nil, err
	}

	if timeout == 0 {
		timeout = OperInstTimeoutDefault
	}

	inst := &OperInst{
		data: &OperInstData{
			TriggerID:         triggerID,
			OperInstID:        fmt.Sprintf("%s-%s", OperInstIDPrefix, uuid.NewString()),
			OperDefName:       def.name,
			ActionNames:       make([]string, len(def.actionDefs)),
			ActionInstDataMap: make(map[string]*ActionInstData),
			Timeout:           timeout,
			InitContent:       make(map[string]any),
			Lifecycle: &Lifecycle{
				CreatedAt: time.Now().Local(),
				State:     OperInstStateInit,
			},
		},
		operationDef: def,
	}

	for index, action := range def.actionDefs {
		inst.data.ActionNames[index] = action.Name()
		inst.data.ActionInstDataMap[action.Name()] = &ActionInstData{
			OperInstID: inst.data.OperInstID,
			Name:       action.Name(),
			Index:      index,
			State:      ActionInstStatePending,
			Messages:   make([]Message, 0),
			Content:    make(map[string]any),
		}
	}

	return inst, nil
}

// Validate validates the operationDef.
func (def *operationDef) Validate() error {
	if len(def.actionDefs) == 0 {
		return errors.New("empty operationDef definition")
	}

	for i, action := range def.actionDefs {
		if action == nil {
			return fmt.Errorf("action is nil, index(%d)", i)
		}
	}

	return nil
}
