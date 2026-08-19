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

package workflow

import (
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

func TestTriggerHandler_InstantiateOperation_PartialCreateFailure(t *testing.T) {
	createErr := errors.New("create operation instance failed")
	successOper := &triggerHandlerTestOperationCtl{
		operationID: "oper-success",
		instance:    &triggerHandlerTestOperationInstanceCtl{id: "oper-inst-success"},
	}
	failedOper := &triggerHandlerTestOperationCtl{
		operationID: "oper-failed",
		createErr:   createErr,
	}
	trigCtl := &triggerHandlerTestTriggerCtl{
		triggerID:  "trigger-partial-create-failure",
		operations: []IOperationCtl{failedOper, successOper},
	}
	handler := &triggerHandler{}

	err := handler.instantiateOperation(contextx.Background(), trigCtl, types.Page{Limit: 10})
	if err != nil {
		t.Fatalf("instantiateOperation() error = %v, want nil", err)
	}

	if got := failedOper.createCalls.Load(); got != 1 {
		t.Fatalf("failed operation create calls = %d, want 1", got)
	}
	if got := successOper.createCalls.Load(); got != 1 {
		t.Fatalf("success operation create calls = %d, want 1", got)
	}
}

func TestTriggerHandler_InstantiateOperation_ListFailure(t *testing.T) {
	listErr := errors.New("list need instantiate operation failed")
	trigCtl := &triggerHandlerTestTriggerCtl{
		triggerID: "trigger-list-failure",
		listErr:   listErr,
	}
	handler := &triggerHandler{}

	err := handler.instantiateOperation(contextx.Background(), trigCtl, types.Page{Limit: 10})
	if !errors.Is(err, listErr) {
		t.Fatalf("instantiateOperation() error = %v, want wrapping %v", err, listErr)
	}
}

func TestTriggerHandler_InstantiateOperation_Panic(t *testing.T) {
	panicOper := &triggerHandlerTestOperationCtl{
		operationID: "oper-panic",
		panicValue:  "panic while creating operation instance",
	}
	trigCtl := &triggerHandlerTestTriggerCtl{
		triggerID:  "trigger-panic",
		operations: []IOperationCtl{panicOper},
	}
	handler := &triggerHandler{}

	err := handler.instantiateOperation(contextx.Background(), trigCtl, types.Page{Limit: 10})
	if err == nil {
		t.Fatalf("instantiateOperation() error = nil, want panic error")
	}
	if !strings.Contains(err.Error(), "go pool panic:") {
		t.Fatalf("instantiateOperation() error = %v, want go pool panic marker", err)
	}
}

func TestTriggerHandler_DoOnceTrigger_PartialCreateFailureContinues(t *testing.T) {
	createErr := errors.New("create operation instance failed")
	initInstance := &triggerHandlerTestOperationInstanceCtl{id: "oper-inst-success"}
	successOper := &triggerHandlerTestOperationCtl{
		operationID: "oper-success",
		instance:    initInstance,
	}
	failedOper := &triggerHandlerTestOperationCtl{
		operationID: "oper-failed",
		createErr:   createErr,
	}
	trigCtl := &triggerHandlerTestTriggerCtl{
		triggerID: "trigger-once-partial-create-failure",
		operations: []IOperationCtl{
			failedOper,
			successOper,
		},
		instances: []IOperationInstanceCtl{initInstance},
	}
	handler := &triggerHandler{}

	instances, err := handler.doOnceTrigger(contextx.Background(), trigCtl)
	if err != nil {
		t.Fatalf("doOnceTrigger() error = %v, want nil", err)
	}
	if got := trigCtl.listInstancesCalls.Load(); got != 1 {
		t.Fatalf("ListOperationInstances calls = %d, want 1", got)
	}
	if len(instances) != 1 || instances[0].GetOperationInstanceID() != "oper-inst-success" {
		t.Fatalf("doOnceTrigger() instances = %v, want successful init instance", instances)
	}
}

type triggerHandlerTestTriggerCtl struct {
	triggerID string
	metadata  trigger.Metadata

	operations []IOperationCtl
	instances  []IOperationInstanceCtl
	listErr    error

	listNeedCalls      atomic.Int32
	listInstancesCalls atomic.Int32
}

func (ctl *triggerHandlerTestTriggerCtl) GetTriggerID() string {
	return ctl.triggerID
}

func (ctl *triggerHandlerTestTriggerCtl) GetTriggerCategory() trigger.Category {
	return trigger.CategoryOnce
}

func (ctl *triggerHandlerTestTriggerCtl) GetTriggerMetadata() trigger.Metadata {
	if ctl.metadata != nil {
		return ctl.metadata
	}

	return &trigger.MetadataOnce{}
}

func (ctl *triggerHandlerTestTriggerCtl) GetLastTriggeredAt() time.Time {
	return time.Time{}
}

func (ctl *triggerHandlerTestTriggerCtl) IsActive() bool {
	return true
}

func (ctl *triggerHandlerTestTriggerCtl) ActivateTrigger(_ contextx.IContext) error {
	return nil
}

func (ctl *triggerHandlerTestTriggerCtl) InactivateTrigger(_ contextx.IContext) error {
	return nil
}

func (ctl *triggerHandlerTestTriggerCtl) TryInactivateTrigger(_ contextx.IContext) error {
	return nil
}

func (ctl *triggerHandlerTestTriggerCtl) CreateOperation(
	_ contextx.IContext, _ operation.Definition, _ operation.Param,
) (IOperationCtl, error) {
	return nil, nil
}

func (ctl *triggerHandlerTestTriggerCtl) GetOperation(_ contextx.IContext, _ string) (IOperationCtl, error) {
	return nil, nil
}

func (ctl *triggerHandlerTestTriggerCtl) ListOperation(_ contextx.IContext, _ ...string) ([]IOperationCtl, error) {
	return nil, nil
}

func (ctl *triggerHandlerTestTriggerCtl) UpdateLastTriggeredTime(_ contextx.IContext) error {
	return nil
}

func (ctl *triggerHandlerTestTriggerCtl) UpdateMetadata(_ contextx.IContext, _ trigger.Metadata) error {
	return nil
}

func (ctl *triggerHandlerTestTriggerCtl) UpdateOperationRetryFlag(
	_ contextx.IContext, _ operation.RetryMode, _ ...string,
) error {
	return nil
}

func (ctl *triggerHandlerTestTriggerCtl) ListNeedInstantiateOperation(
	_ contextx.IContext, page types.Page,
) ([]IOperationCtl, error) {
	ctl.listNeedCalls.Add(1)
	if ctl.listErr != nil {
		return nil, ctl.listErr
	}

	if page.Offset >= len(ctl.operations) {
		return make([]IOperationCtl, 0), nil
	}

	end := min(page.Offset+page.Limit, len(ctl.operations))
	operations := make([]IOperationCtl, end-page.Offset)
	copy(operations, ctl.operations[page.Offset:end])

	return operations, nil
}

func (ctl *triggerHandlerTestTriggerCtl) ListOperationInstances(
	_ contextx.IContext, _ types.Page, _ ...operation.State,
) ([]IOperationInstanceCtl, error) {
	ctl.listInstancesCalls.Add(1)
	instances := make([]IOperationInstanceCtl, len(ctl.instances))
	copy(instances, ctl.instances)

	return instances, nil
}

type triggerHandlerTestOperationCtl struct {
	operationID string
	instance    IOperationInstanceCtl
	createErr   error
	panicValue  any

	createCalls atomic.Int32
}

func (ctl *triggerHandlerTestOperationCtl) GetOperationID() string {
	return ctl.operationID
}

func (ctl *triggerHandlerTestOperationCtl) GetLastOperationInstance(
	_ contextx.IContext,
) (IOperationInstanceCtl, error) {
	return nil, nil
}

func (ctl *triggerHandlerTestOperationCtl) CreateOperationInstance(
	_ contextx.IContext,
) (IOperationInstanceCtl, error) {
	ctl.createCalls.Add(1)
	if ctl.panicValue != nil {
		panic(ctl.panicValue)
	}
	if ctl.createErr != nil {
		return nil, ctl.createErr
	}
	if ctl.instance != nil {
		return ctl.instance, nil
	}

	return &triggerHandlerTestOperationInstanceCtl{id: fmt.Sprintf("%s-inst", ctl.operationID)}, nil
}

func (ctl *triggerHandlerTestOperationCtl) GetOperationInstance(
	_ contextx.IContext, _ string,
) (IOperationInstanceCtl, error) {
	return nil, nil
}

type triggerHandlerTestOperationInstanceCtl struct {
	id string
}

func (ctl *triggerHandlerTestOperationInstanceCtl) GetOperationInstanceID() string {
	return ctl.id
}

func (ctl *triggerHandlerTestOperationInstanceCtl) LaunchOperationInstance(_ contextx.IContext) error {
	return nil
}

func (ctl *triggerHandlerTestOperationInstanceCtl) TerminateOperationInstance(_ contextx.IContext) error {
	return nil
}
