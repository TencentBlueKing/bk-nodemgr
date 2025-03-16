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
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	// OperInstMaxNum define a max instance number of an operation.
	OperInstMaxNum = 100

	// OperationPrefix define the prefix of operation id.
	OperationPrefix = "O"
)

// NewOperation creates a new operation.
func NewOperation(triggerID string, defSnapshot OperDefSnapshot) *Operation {
	return &Operation{
		TriggerID:   triggerID,
		OperationID: fmt.Sprintf("%s-%s", OperationPrefix, uuid.New().String()),
		DefSnapshot: defSnapshot,
		OperInstIDs: []string{},
	}
}

// Operation ...
type Operation struct {
	TriggerID   string
	OperationID string
	DefSnapshot OperDefSnapshot
	OperInstIDs []string
}

// CheckEnforceability check if an operation can be executed.
func (operation *Operation) CheckEnforceability() error {
	// check the operation instance length.
	if len(operation.OperInstIDs) > OperInstMaxNum {
		return fmt.Errorf("operation can not be executed, operation-inst-length(%d)", len(operation.OperInstIDs))
	}

	return nil
}

// getLatestOperInstID ...
func (operation *Operation) getLatestOperInstID() string {
	if len(operation.OperInstIDs) == 0 {
		return ""
	}

	return operation.OperInstIDs[len(operation.OperInstIDs)-1]
}

// OperDefSnapshot defines the snapshot of operationDef.
type OperDefSnapshot struct {
	OperDefName string
	ActionNames []string
}

// OperInstParam ...
type OperInstParam struct {
	// Timeout define the timeout of operation instance.
	Timeout time.Duration

	// InitContent define the init content of operation instance.
	InitContent map[string]any

	// ParentOperationID define the parent operation id.
	ParentOperationID string
}
