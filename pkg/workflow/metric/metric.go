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

// Package metric provides metric for workflow.
package metric

import (
	"errors"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

var (
	// nolint: gochecknoglobals
	monitor *Monitor
)

// OperInstMetricData defines the metric data of operation instance.
type OperInstMetricData struct {
	param Param

	startTime time.Time
}

// Start start the timer.
func (md *OperInstMetricData) Start() *OperInstMetricData {
	md.startTime = time.Now()

	return md
}

// End end the timer.
func (md *OperInstMetricData) End(err error) {
	md.param.ProcessDuration = time.Since(md.startTime)
	md.param.Err = err

	monitor.RecordOperationInstanceLaunched(md.param)
}

// NewOperationInstanceLaunch create a new OperInstMetricData.
func NewOperationInstanceLaunch(data *operation.InstanceBriefData) *OperInstMetricData {
	return &OperInstMetricData{
		param: Param{
			OperationName: data.Metadata.OperationDefName,
		},
	}
}

// OperationInstanceProcessed record the operation instance processed.
func OperationInstanceProcessed(data *operation.InstanceBriefData) {
	monitor.RecordOperationInstanceProcessed(Param{
		OperationName:   data.Metadata.OperationDefName,
		OperationState:  data.Lifecycle.State,
		ProcessDuration: time.Since(data.Lifecycle.StartedAt),
	})
}

// ActionMetricData defines the metric data of action.
type ActionMetricData struct {
	param Param

	startTime time.Time
}

// Start start the timer.
func (md *ActionMetricData) Start() *ActionMetricData {
	md.startTime = time.Now()

	monitor.RecordActionReceived(md.param)

	return md
}

// End end the timer.
func (md *ActionMetricData) End(lifecycle *action.Lifecycle) {
	md.param.ActionState = lifecycle.State
	md.param.ProcessDuration = time.Since(md.startTime)

	monitor.RecordActionProcessed(md.param)
}

// NewActionProcess create a new ActionMetricData.
func NewActionProcess(data *action.InstanceData) *ActionMetricData {
	return &ActionMetricData{
		param: Param{
			OperationName: data.OperationDefName,
			ActionName:    data.Name,
		},
	}
}

// ActionNotRegistered record the action not registered.
func ActionNotRegistered(actionName string) {
	monitor.RecordActionNotResigter(Param{
		ActionName: actionName,
		Err:        errors.New("action not registered"),
	})
}

// ActionDataNotFound record the action data not found.
func ActionDataNotFound(actionName string) {
	monitor.RecordActionDataNotFound(Param{
		ActionName: actionName,
		Err:        errors.New("action data not found"),
	})
}

// nolint: gochecknoinits
func init() {
	monitor = NewMonitor("workflow").Enable()
}
