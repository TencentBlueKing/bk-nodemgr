//go:build integration

/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package packageworkflow

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/testsuite/support"
)

func testClient(t *testing.T) IHandler {
	t.Helper()

	_, db := support.RequireMongoDatabase(t)
	return New(db)
}

func testContext() contextx.IContext {
	return contextx.New(context.Background(), contextx.WithTenantID(uuid.NewString()))
}

func testPackageWorkflow(nCtx contextx.IContext) *types.PackageWorkflow {
	return &types.PackageWorkflow{
		TenantID:    nCtx.TenantID(),
		WorkflowID:  uuid.NewString(),
		TriggerID:   uuid.NewString(),
		Type:        types.PackageWorkflowTypeImport,
		FileName:    "gse_plugin-1.0.0.tgz",
		Operator:    "admin",
		OperateTime: time.Date(2026, 8, 5, 10, 0, 0, 0, time.UTC),
		Status:      types.PackageWorkflowStatusRunning,
	}
}

func TestHandler_CreateAndGet(t *testing.T) {
	h := testClient(t)
	nCtx := testContext()
	want := testPackageWorkflow(nCtx)

	if err := h.Create(nCtx, want); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := h.Get(nCtx, want.WorkflowID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	assertPackageWorkflowEqual(t, got, want)

	status, err := h.GetStatus(nCtx, want.WorkflowID)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}
	if status != want.Status {
		t.Fatalf("GetStatus() = %s, want %s", status, want.Status)
	}
}

func TestHandler_CountAndList(t *testing.T) {
	h := testClient(t)
	nCtx := testContext()
	runningWorkflow := testPackageWorkflow(nCtx)
	successWorkflow := testPackageWorkflow(nCtx)
	successWorkflow.FileName = "gse_agent-2.0.0.tgz"
	successWorkflow.Operator = "system"
	successWorkflow.Status = types.PackageWorkflowStatusSuccess

	for _, workflow := range []*types.PackageWorkflow{runningWorkflow, successWorkflow} {
		if err := h.Create(nCtx, workflow); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	tests := []struct {
		name      string
		opts      []OptFn
		wantTotal int64
	}{
		{
			name:      "all workflows",
			wantTotal: 2,
		},
		{
			name:      "by workflow ID",
			opts:      []OptFn{WithWorkflowID(runningWorkflow.WorkflowID)},
			wantTotal: 1,
		},
		{
			name:      "by trigger ID",
			opts:      []OptFn{WithTriggerID(successWorkflow.TriggerID)},
			wantTotal: 1,
		},
		{
			name:      "by status",
			opts:      []OptFn{WithStatus(types.PackageWorkflowStatusRunning)},
			wantTotal: 1,
		},
		{
			name:      "by operator",
			opts:      []OptFn{WithOperator("system")},
			wantTotal: 1,
		},
		{
			name: "by operate time range",
			opts: []OptFn{WithOperateTimeRange(types.TimeRange{
				StartTime: runningWorkflow.OperateTime.Add(-time.Minute),
				EndTime:   runningWorkflow.OperateTime.Add(time.Minute),
			})},
			wantTotal: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := h.Count(nCtx, tt.opts...)
			if err != nil {
				t.Fatalf("Count() error = %v", err)
			}
			if count != tt.wantTotal {
				t.Fatalf("Count() = %d, want %d", count, tt.wantTotal)
			}

			got, total, err := h.List(nCtx, types.UnlimitedPage(), tt.opts...)
			if err != nil {
				t.Fatalf("List() error = %v", err)
			}
			if total != tt.wantTotal {
				t.Fatalf("List() total = %d, want %d", total, tt.wantTotal)
			}
			if int64(len(got)) != tt.wantTotal {
				t.Fatalf("List() len = %d, want %d", len(got), tt.wantTotal)
			}
		})
	}
}

func TestHandler_UpdateStatusAndFinishTime(t *testing.T) {
	h := testClient(t)
	nCtx := testContext()
	workflow := testPackageWorkflow(nCtx)

	if err := h.Create(nCtx, workflow); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := h.UpdateStatus(nCtx, workflow.WorkflowID, types.PackageWorkflowStatusSuccess); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	finishTime := time.Date(2026, 8, 5, 10, 10, 0, 0, time.UTC)
	if err := h.UpdateFinishTime(nCtx, workflow.WorkflowID, finishTime); err != nil {
		t.Fatalf("UpdateFinishTime() error = %v", err)
	}

	got, err := h.Get(nCtx, workflow.WorkflowID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Status != types.PackageWorkflowStatusSuccess {
		t.Fatalf("Status = %s, want %s", got.Status, types.PackageWorkflowStatusSuccess)
	}
	if !got.FinishTime.Equal(finishTime) {
		t.Fatalf("FinishTime = %s, want %s", got.FinishTime, finishTime)
	}
}

func assertPackageWorkflowEqual(t *testing.T, got, want *types.PackageWorkflow) {
	t.Helper()

	if got == nil {
		t.Fatal("got nil package workflow")
	}
	if got.TenantID != want.TenantID {
		t.Fatalf("TenantID = %s, want %s", got.TenantID, want.TenantID)
	}
	if got.WorkflowID != want.WorkflowID {
		t.Fatalf("WorkflowID = %s, want %s", got.WorkflowID, want.WorkflowID)
	}
	if got.TriggerID != want.TriggerID {
		t.Fatalf("TriggerID = %s, want %s", got.TriggerID, want.TriggerID)
	}
	if got.Type != want.Type {
		t.Fatalf("Type = %s, want %s", got.Type, want.Type)
	}
	if got.FileName != want.FileName {
		t.Fatalf("FileName = %s, want %s", got.FileName, want.FileName)
	}
	if got.Operator != want.Operator {
		t.Fatalf("Operator = %s, want %s", got.Operator, want.Operator)
	}
	if !got.OperateTime.Equal(want.OperateTime) {
		t.Fatalf("OperateTime = %s, want %s", got.OperateTime, want.OperateTime)
	}
	if got.Status != want.Status {
		t.Fatalf("Status = %s, want %s", got.Status, want.Status)
	}
}
