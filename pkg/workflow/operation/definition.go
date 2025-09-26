/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operation describes the operation of workflow.
package operation

import "github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"

// Definition represents an operation.
type Definition interface {
	// Name returns the name of the operation definition.
	Name() string

	// ActionDefNames returns the names of the action definitions, in order.
	ActionDefNames() []string

	// DefaultParameters returns the default parameters of the operation.
	DefaultParameters() Param

	// ExtraExecutionName returns the extra execution name for the operation.
	ExtraExecutionName() string
}

// DefinitionSnapshot represents a snapshot of an operation definition.
type DefinitionSnapshot struct {
	SnapshotName               string
	SnapshotActionDefNames     []string
	SnapshotDefaultParameters  Param
	RetryStartPoint            map[string]bool
	SnapshotExtraExecutionName string
}

// Name returns the name of the operation definition.
func (ds *DefinitionSnapshot) Name() string {
	return ds.SnapshotName
}

// ActionDefNames returns the names of the action definitions, in order.
func (ds *DefinitionSnapshot) ActionDefNames() []string {
	return ds.SnapshotActionDefNames
}

// DefaultParameters returns the default parameters of the operation.
func (ds *DefinitionSnapshot) DefaultParameters() Param {
	return ds.SnapshotDefaultParameters
}

// ExtraExecutionName returns the extra execution name for the operation.
func (ds *DefinitionSnapshot) ExtraExecutionName() string {
	return ds.SnapshotExtraExecutionName
}

// ExtraExecution represents an extra execution for an operation.
type ExtraExecution interface {
	// Name returns the name of the extra execution.
	Name() string

	Do(nCtx contextx.IContext, instance *InstanceBriefData) error
}
