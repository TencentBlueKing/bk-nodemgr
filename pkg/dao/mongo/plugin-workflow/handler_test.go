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

package pluginworkflow

import (
	"context"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var instance struct {
	TenantID string
	sync.Once
}

func tenantID() string {
	instance.Once.Do(func() {
		instance.TenantID = uuid.NewString()
	})

	return instance.TenantID
}

// testClient creates a test handler instance
func testClient(t *testing.T) IHandler {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	nCtx := context.Background()
	mongoClient, err := mongo.Connect(
		nCtx,
		&options.ClientOptions{
			Hosts: []string{
				os.Getenv("MONGO_ADDRESS"),
			},
			Auth: &options.Credential{
				Username:      os.Getenv("MONGO_USER"),
				Password:      os.Getenv("MONGO_PASSWORD"),
				AuthSource:    os.Getenv("MONGO_AUTH_SOURCE"),
				AuthMechanism: os.Getenv("MONGO_AUTH_MECHANISM"),
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	return New(mongoClient.Database(os.Getenv("MONGO_DATABASE")))
}

// TestHandler_Create test handler Create
func TestHandler_Create(t *testing.T) {
	type args struct {
		nCtx     contextx.IContext
		workflow *types.PluginWorkflow
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal test",
			args: args{
				nCtx: contextx.New(context.Background(), contextx.WithTenantID(tenantID())),
				workflow: &types.PluginWorkflow{
					WorkflowID:  uuid.NewString(),
					TriggerID:   uuid.NewString(),
					Type:        types.PluginWorkflowTypeInstall,
					HostIDs:     []int64{1001, 1002},
					Operator:    "test-operator",
					OperateTime: time.Now(),
					Status:      types.PluginWorkflowStatusRunning,
				},
			},
			wantErr: false,
		},
		{
			name: "missing tenant ID",
			args: args{
				nCtx: contextx.New(context.Background()),
				workflow: &types.PluginWorkflow{
					WorkflowID:  uuid.NewString(),
					TriggerID:   uuid.NewString(),
					Type:        types.PluginWorkflowTypeInstall,
					HostIDs:     []int64{1001, 1002},
					Operator:    "test-operator",
					OperateTime: time.Now(),
					Status:      types.PluginWorkflowStatusRunning,
				},
			},
			wantErr: true,
		},
		{
			name: "empty trigger ID",
			args: args{
				nCtx: contextx.New(context.Background(), contextx.WithTenantID(tenantID())),
				workflow: &types.PluginWorkflow{
					WorkflowID:  uuid.NewString(),
					TriggerID:   "",
					Type:        types.PluginWorkflowTypeInstall,
					HostIDs:     []int64{1001, 1002},
					Operator:    "test-operator",
					OperateTime: time.Now(),
					Status:      types.PluginWorkflowStatusRunning,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			if err := h.Create(tt.args.nCtx, tt.args.workflow); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestHandler_Count test handler Count
func TestHandler_Count(t *testing.T) {
	h := testClient(t)
	testTenantID := uuid.NewString() // 使用唯一的租户ID

	// Create a test workflow first
	nCtx := contextx.New(context.Background(), contextx.WithTenantID(testTenantID))
	testWorkflow := &types.PluginWorkflow{
		WorkflowID:  uuid.NewString(),
		TriggerID:   uuid.NewString(),
		Type:        types.PluginWorkflowTypeInstall,
		HostIDs:     []int64{1001, 1002},
		Operator:    "test-operator-" + uuid.NewString()[:8], // 使用唯一的操作者
		OperateTime: time.Now(),
		Status:      types.PluginWorkflowStatusRunning,
	}

	err := h.Create(nCtx, testWorkflow)
	if err != nil {
		t.Fatalf("Failed to create test workflow: %v", err)
	}

	type args struct {
		nCtx contextx.IContext
		opts []OptFn
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "count all workflows",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{},
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "count by workflow ID",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{WithWorkflowID(testWorkflow.WorkflowID)},
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "count by status",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{WithStatus(types.PluginWorkflowStatusRunning)},
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "count by type",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{WithType(types.PluginWorkflowTypeInstall)},
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "count by host IDs",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{WithHostIDs(1001)},
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "count by operator",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{WithOperator(testWorkflow.Operator)},
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "count by non-existent workflow",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{WithWorkflowID("non-existent-workflow-id")},
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "missing tenant ID",
			args: args{
				nCtx: contextx.New(context.Background()),
				opts: []OptFn{},
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.Count(tt.args.nCtx, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Count() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Count() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestHandler_List test handler List
func TestHandler_List(t *testing.T) {
	h := testClient(t)
	testTenantID := uuid.NewString() // 使用唯一的租户ID

	// Create test workflows first
	nCtx := contextx.New(context.Background(), contextx.WithTenantID(testTenantID))
	testWorkflow1 := &types.PluginWorkflow{
		WorkflowID:  uuid.NewString(),
		TriggerID:   uuid.NewString(),
		Type:        types.PluginWorkflowTypeInstall,
		HostIDs:     []int64{1001, 1002},
		Operator:    "test-operator-1",
		OperateTime: time.Now(),
		Status:      types.PluginWorkflowStatusRunning,
	}
	testWorkflow2 := &types.PluginWorkflow{
		WorkflowID:  uuid.NewString(),
		TriggerID:   uuid.NewString(),
		Type:        types.PluginWorkflowTypeUpgrade,
		HostIDs:     []int64{1003, 1004},
		Operator:    "test-operator-2",
		OperateTime: time.Now(),
		Status:      types.PluginWorkflowStatusRunning,
	}

	err := h.Create(nCtx, testWorkflow1)
	if err != nil {
		t.Fatalf("Failed to create test workflow 1: %v", err)
	}
	err = h.Create(nCtx, testWorkflow2)
	if err != nil {
		t.Fatalf("Failed to create test workflow 2: %v", err)
	}

	// Update testWorkflow2 status to Success for testing status filtering
	err = h.UpdateStatus(nCtx, testWorkflow2.WorkflowID, types.PluginWorkflowStatusSuccess)
	if err != nil {
		t.Fatalf("Failed to update test workflow 2 status: %v", err)
	}

	type args struct {
		nCtx contextx.IContext
		page types.Page
		opts []OptFn
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "list all workflows",
			args: args{
				nCtx: nCtx,
				page: types.Page{
					Offset: 0,
					Limit:  10,
				},
				opts: []OptFn{},
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "list with pagination",
			args: args{
				nCtx: nCtx,
				page: types.Page{
					Offset: 0,
					Limit:  1,
				},
				opts: []OptFn{},
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "list by type",
			args: args{
				nCtx: nCtx,
				page: types.Page{
					Offset: 0,
					Limit:  10,
				},
				opts: []OptFn{WithType(types.PluginWorkflowTypeInstall)},
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "list by status",
			args: args{
				nCtx: nCtx,
				page: types.Page{
					Offset: 0,
					Limit:  10,
				},
				opts: []OptFn{WithStatus(types.PluginWorkflowStatusRunning)},
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "missing tenant ID",
			args: args{
				nCtx: contextx.New(context.Background()),
				page: types.Page{
					Offset: 0,
					Limit:  10,
				},
				opts: []OptFn{},
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := h.List(tt.args.nCtx, tt.args.page, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				expectedCount := int(tt.want)
				if len(got) != expectedCount {
					t.Errorf("List() got count = %v, want %v", len(got), expectedCount)
				}
			}
		})
	}
}

// TestHandler_Get test handler Get
func TestHandler_Get(t *testing.T) {
	h := testClient(t)
	testTenantID := tenantID()

	// Create a test workflow first
	nCtx := contextx.New(context.Background(), contextx.WithTenantID(testTenantID))
	testWorkflow := &types.PluginWorkflow{
		WorkflowID:  uuid.NewString(),
		TriggerID:   uuid.NewString(),
		Type:        types.PluginWorkflowTypeInstall,
		HostIDs:     []int64{1001, 1002},
		Operator:    "test-operator",
		OperateTime: time.Now(),
		Status:      types.PluginWorkflowStatusRunning,
	}

	err := h.Create(nCtx, testWorkflow)
	if err != nil {
		t.Fatalf("Failed to create test workflow: %v", err)
	}

	type args struct {
		nCtx       contextx.IContext
		workflowID string
	}
	tests := []struct {
		name    string
		args    args
		want    *types.PluginWorkflow
		wantErr bool
	}{
		{
			name: "get by workflow ID",
			args: args{
				nCtx:       nCtx,
				workflowID: testWorkflow.WorkflowID,
			},
			want: &types.PluginWorkflow{
				WorkflowID: testWorkflow.WorkflowID,
				TriggerID:  testWorkflow.TriggerID,
				Type:       types.PluginWorkflowTypeInstall,
				HostIDs:    testWorkflow.HostIDs,
				Operator:   testWorkflow.Operator,
				Status:     types.PluginWorkflowStatusRunning,
			},
			wantErr: false,
		},
		{
			name: "get non-existent workflow",
			args: args{
				nCtx:       nCtx,
				workflowID: "non-existent-workflow-id",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "empty workflow ID",
			args: args{
				nCtx:       nCtx,
				workflowID: "",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "missing tenant ID",
			args: args{
				nCtx:       contextx.New(context.Background()),
				workflowID: testWorkflow.WorkflowID,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.Get(tt.args.nCtx, tt.args.workflowID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != nil {
				if got.WorkflowID != tt.want.WorkflowID ||
					got.TriggerID != tt.want.TriggerID ||
					got.Type != tt.want.Type ||
					!reflect.DeepEqual(got.HostIDs, tt.want.HostIDs) ||
					got.Operator != tt.want.Operator ||
					got.Status != tt.want.Status {
					t.Errorf("Get() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// TestHandler_UpdateStatus test handler UpdateStatus
func TestHandler_UpdateStatus(t *testing.T) {
	h := testClient(t)
	testTenantID := tenantID()

	// Create a test workflow first
	nCtx := contextx.New(context.Background(), contextx.WithTenantID(testTenantID))
	testWorkflow := &types.PluginWorkflow{
		WorkflowID:  uuid.NewString(),
		TriggerID:   uuid.NewString(),
		Type:        types.PluginWorkflowTypeInstall,
		HostIDs:     []int64{1001, 1002},
		Operator:    "test-operator",
		OperateTime: time.Now(),
		Status:      types.PluginWorkflowStatusRunning,
	}

	err := h.Create(nCtx, testWorkflow)
	if err != nil {
		t.Fatalf("Failed to create test workflow: %v", err)
	}

	type args struct {
		nCtx       contextx.IContext
		workflowID string
		status     types.PluginWorkflowStatus
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "update status to success",
			args: args{
				nCtx:       nCtx,
				workflowID: testWorkflow.WorkflowID,
				status:     types.PluginWorkflowStatusSuccess,
			},
			wantErr: false,
		},
		{
			name: "update status to failed",
			args: args{
				nCtx:       nCtx,
				workflowID: testWorkflow.WorkflowID,
				status:     types.PluginWorkflowStatusFailed,
			},
			wantErr: false,
		},
		{
			name: "empty workflow ID",
			args: args{
				nCtx:       nCtx,
				workflowID: "",
				status:     types.PluginWorkflowStatusSuccess,
			},
			wantErr: true,
		},
		{
			name: "missing tenant ID",
			args: args{
				nCtx:       contextx.New(context.Background()),
				workflowID: testWorkflow.WorkflowID,
				status:     types.PluginWorkflowStatusSuccess,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := h.UpdateStatus(tt.args.nCtx, tt.args.workflowID, tt.args.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				// Verify the update
				updated, err := h.Get(tt.args.nCtx, tt.args.workflowID)
				if err != nil {
					t.Errorf("Failed to get updated workflow: %v", err)
				} else if updated.Status != tt.args.status {
					t.Errorf("UpdateStatus() status = %v, want %v", updated.Status, tt.args.status)
				}
			}
		})
	}
}

// TestHandler_UpdateFinishTime test handler UpdateFinishTime
func TestHandler_UpdateFinishTime(t *testing.T) {
	h := testClient(t)
	testTenantID := tenantID()

	// Create a test workflow first
	nCtx := contextx.New(context.Background(), contextx.WithTenantID(testTenantID))
	testWorkflow := &types.PluginWorkflow{
		WorkflowID:  uuid.NewString(),
		TriggerID:   uuid.NewString(),
		Type:        types.PluginWorkflowTypeInstall,
		HostIDs:     []int64{1001, 1002},
		Operator:    "test-operator",
		OperateTime: time.Now(),
		Status:      types.PluginWorkflowStatusRunning,
	}

	err := h.Create(nCtx, testWorkflow)
	if err != nil {
		t.Fatalf("Failed to create test workflow: %v", err)
	}

	type args struct {
		nCtx       contextx.IContext
		workflowID string
		finishTime time.Time
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "update finish time",
			args: args{
				nCtx:       nCtx,
				workflowID: testWorkflow.WorkflowID,
				finishTime: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "empty workflow ID",
			args: args{
				nCtx:       nCtx,
				workflowID: "",
				finishTime: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "zero finish time",
			args: args{
				nCtx:       nCtx,
				workflowID: testWorkflow.WorkflowID,
				finishTime: time.Time{},
			},
			wantErr: true,
		},
		{
			name: "missing tenant ID",
			args: args{
				nCtx:       contextx.New(context.Background()),
				workflowID: testWorkflow.WorkflowID,
				finishTime: time.Now(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := h.UpdateFinishTime(tt.args.nCtx, tt.args.workflowID, tt.args.finishTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateFinishTime() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				// Verify the update
				updated, err := h.Get(tt.args.nCtx, tt.args.workflowID)
				if err != nil {
					t.Errorf("Failed to get updated workflow: %v", err)
				} else if updated.FinishTime.IsZero() && !tt.args.finishTime.IsZero() {
					t.Errorf("UpdateFinishTime() finish time not updated")
				}
			}
		})
	}
}

// TestHandler_DistinctPluginWorkflowType test handler DistinctPluginWorkflowType
func TestHandler_DistinctPluginWorkflowType(t *testing.T) {
	h := testClient(t)
	testTenantID := tenantID()

	// Create test workflows first
	nCtx := contextx.New(context.Background(), contextx.WithTenantID(testTenantID))
	testWorkflow1 := &types.PluginWorkflow{
		WorkflowID:  uuid.NewString(),
		TriggerID:   uuid.NewString(),
		Type:        types.PluginWorkflowTypeInstall,
		HostIDs:     []int64{1001},
		Operator:    "test-operator",
		OperateTime: time.Now(),
		Status:      types.PluginWorkflowStatusRunning,
	}
	testWorkflow2 := &types.PluginWorkflow{
		WorkflowID:  uuid.NewString(),
		TriggerID:   uuid.NewString(),
		Type:        types.PluginWorkflowTypeUpgrade,
		HostIDs:     []int64{1002},
		Operator:    "test-operator",
		OperateTime: time.Now(),
		Status:      types.PluginWorkflowStatusRunning,
	}

	err := h.Create(nCtx, testWorkflow1)
	if err != nil {
		t.Fatalf("Failed to create test workflow 1: %v", err)
	}
	err = h.Create(nCtx, testWorkflow2)
	if err != nil {
		t.Fatalf("Failed to create test workflow 2: %v", err)
	}

	type args struct {
		nCtx contextx.IContext
		opts []OptFn
	}
	tests := []struct {
		name    string
		args    args
		want    []types.PluginWorkflowType
		wantErr bool
	}{
		{
			name: "distinct all types",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{},
			},
			want:    []types.PluginWorkflowType{types.PluginWorkflowTypeInstall, types.PluginWorkflowTypeUpgrade},
			wantErr: false,
		},
		{
			name: "distinct by status",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{WithStatus(types.PluginWorkflowStatusRunning)},
			},
			want:    []types.PluginWorkflowType{types.PluginWorkflowTypeInstall, types.PluginWorkflowTypeUpgrade},
			wantErr: false,
		},
		{
			name: "missing tenant ID",
			args: args{
				nCtx: contextx.New(context.Background()),
				opts: []OptFn{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.DistinctPluginWorkflowType(tt.args.nCtx, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("DistinctPluginWorkflowType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("DistinctPluginWorkflowType() got count = %v, want %v", len(got), len(tt.want))
				}
				// Check if all expected types are present
				for _, expectedType := range tt.want {
					found := false
					for _, gotType := range got {
						if gotType == expectedType {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("DistinctPluginWorkflowType() missing type %v", expectedType)
					}
				}
			}
		})
	}
}

// TestHandler_DistinctPluginWorkflowBkHostID test handler DistinctPluginWorkflowBkHostID
func TestHandler_DistinctPluginWorkflowBkHostID(t *testing.T) {
	h := testClient(t)
	testTenantID := tenantID()

	// Create test workflows first
	nCtx := contextx.New(context.Background(), contextx.WithTenantID(testTenantID))
	testWorkflow1 := &types.PluginWorkflow{
		WorkflowID:  uuid.NewString(),
		TriggerID:   uuid.NewString(),
		Type:        types.PluginWorkflowTypeInstall,
		HostIDs:     []int64{1001, 1002},
		Operator:    "test-operator",
		OperateTime: time.Now(),
		Status:      types.PluginWorkflowStatusRunning,
	}
	testWorkflow2 := &types.PluginWorkflow{
		WorkflowID:  uuid.NewString(),
		TriggerID:   uuid.NewString(),
		Type:        types.PluginWorkflowTypeUpgrade,
		HostIDs:     []int64{1002, 1003},
		Operator:    "test-operator",
		OperateTime: time.Now(),
		Status:      types.PluginWorkflowStatusRunning,
	}

	err := h.Create(nCtx, testWorkflow1)
	if err != nil {
		t.Fatalf("Failed to create test workflow 1: %v", err)
	}
	err = h.Create(nCtx, testWorkflow2)
	if err != nil {
		t.Fatalf("Failed to create test workflow 2: %v", err)
	}

	type args struct {
		nCtx contextx.IContext
		opts []OptFn
	}
	tests := []struct {
		name    string
		args    args
		want    []int64
		wantErr bool
	}{
		{
			name: "distinct all host IDs",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{},
			},
			want:    []int64{1001, 1002, 1003},
			wantErr: false,
		},
		{
			name: "missing tenant ID",
			args: args{
				nCtx: contextx.New(context.Background()),
				opts: []OptFn{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.DistinctPluginWorkflowBkHostID(tt.args.nCtx, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("DistinctPluginWorkflowBkHostID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("DistinctPluginWorkflowBkHostID() got count = %v, want %v", len(got), len(tt.want))
				}
				// Check if all expected host IDs are present
				for _, expectedHostID := range tt.want {
					found := false
					for _, gotHostID := range got {
						if gotHostID == expectedHostID {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("DistinctPluginWorkflowBkHostID() missing host ID %v", expectedHostID)
					}
				}
			}
		})
	}
}

// TestHandler_DistinctPluginWorkflowOperator test handler DistinctPluginWorkflowOperator
func TestHandler_DistinctPluginWorkflowOperator(t *testing.T) {
	h := testClient(t)
	testTenantID := uuid.NewString() // 使用唯一的租户ID

	// Create test workflows first
	nCtx := contextx.New(context.Background(), contextx.WithTenantID(testTenantID))
	testWorkflow1 := &types.PluginWorkflow{
		WorkflowID:  uuid.NewString(),
		TriggerID:   uuid.NewString(),
		Type:        types.PluginWorkflowTypeInstall,
		HostIDs:     []int64{1001},
		Operator:    "operator-1",
		OperateTime: time.Now(),
		Status:      types.PluginWorkflowStatusRunning,
	}
	testWorkflow2 := &types.PluginWorkflow{
		WorkflowID:  uuid.NewString(),
		TriggerID:   uuid.NewString(),
		Type:        types.PluginWorkflowTypeUpgrade,
		HostIDs:     []int64{1002},
		Operator:    "operator-2",
		OperateTime: time.Now(),
		Status:      types.PluginWorkflowStatusRunning,
	}

	err := h.Create(nCtx, testWorkflow1)
	if err != nil {
		t.Fatalf("Failed to create test workflow 1: %v", err)
	}
	err = h.Create(nCtx, testWorkflow2)
	if err != nil {
		t.Fatalf("Failed to create test workflow 2: %v", err)
	}

	type args struct {
		nCtx contextx.IContext
		opts []OptFn
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "distinct all operators",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{},
			},
			want:    []string{"operator-1", "operator-2"},
			wantErr: false,
		},
		{
			name: "missing tenant ID",
			args: args{
				nCtx: contextx.New(context.Background()),
				opts: []OptFn{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.DistinctPluginWorkflowOperator(tt.args.nCtx, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("DistinctPluginWorkflowOperator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("DistinctPluginWorkflowOperator() got count = %v, want %v", len(got), len(tt.want))
				}
				// Check if all expected operators are present
				for _, expectedOperator := range tt.want {
					found := false
					for _, gotOperator := range got {
						if gotOperator == expectedOperator {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("DistinctPluginWorkflowOperator() missing operator %v", expectedOperator)
					}
				}
			}
		})
	}
}

// TestHandler_DistinctPluginWorkflowStatus test handler DistinctPluginWorkflowStatus
func TestHandler_DistinctPluginWorkflowStatus(t *testing.T) {
	h := testClient(t)
	testTenantID := uuid.NewString() // 使用唯一的租户ID

	// Create test workflows first
	nCtx := contextx.New(context.Background(), contextx.WithTenantID(testTenantID))
	testWorkflow1 := &types.PluginWorkflow{
		WorkflowID:  uuid.NewString(),
		TriggerID:   uuid.NewString(),
		Type:        types.PluginWorkflowTypeInstall,
		HostIDs:     []int64{1001},
		Operator:    "test-operator",
		OperateTime: time.Now(),
		Status:      types.PluginWorkflowStatusRunning,
	}
	testWorkflow2 := &types.PluginWorkflow{
		WorkflowID:  uuid.NewString(),
		TriggerID:   uuid.NewString(),
		Type:        types.PluginWorkflowTypeUpgrade,
		HostIDs:     []int64{1002},
		Operator:    "test-operator",
		OperateTime: time.Now(),
		Status:      types.PluginWorkflowStatusRunning,
	}

	err := h.Create(nCtx, testWorkflow1)
	if err != nil {
		t.Fatalf("Failed to create test workflow 1: %v", err)
	}
	err = h.Create(nCtx, testWorkflow2)
	if err != nil {
		t.Fatalf("Failed to create test workflow 2: %v", err)
	}

	// Update testWorkflow2 status to Success for testing distinct statuses
	err = h.UpdateStatus(nCtx, testWorkflow2.WorkflowID, types.PluginWorkflowStatusSuccess)
	if err != nil {
		t.Fatalf("Failed to update test workflow 2 status: %v", err)
	}

	type args struct {
		nCtx contextx.IContext
		opts []OptFn
	}
	tests := []struct {
		name    string
		args    args
		want    []types.PluginWorkflowStatus
		wantErr bool
	}{
		{
			name: "distinct all statuses",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{},
			},
			want:    []types.PluginWorkflowStatus{types.PluginWorkflowStatusRunning, types.PluginWorkflowStatusSuccess},
			wantErr: false,
		},
		{
			name: "distinct by type",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{WithType(types.PluginWorkflowTypeInstall)},
			},
			want:    []types.PluginWorkflowStatus{types.PluginWorkflowStatusRunning},
			wantErr: false,
		},
		{
			name: "missing tenant ID",
			args: args{
				nCtx: contextx.New(context.Background()),
				opts: []OptFn{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.DistinctPluginWorkflowStatus(tt.args.nCtx, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("DistinctPluginWorkflowStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("DistinctPluginWorkflowStatus() got count = %v, want %v", len(got), len(tt.want))
				}
				// Check if all expected statuses are present
				for _, expectedStatus := range tt.want {
					found := false
					for _, gotStatus := range got {
						if gotStatus == expectedStatus {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("DistinctPluginWorkflowStatus() missing status %v", expectedStatus)
					}
				}
			}
		})
	}
}
