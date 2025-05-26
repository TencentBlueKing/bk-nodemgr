/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operinstdata ...
package operinstdata

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// testClient ...
func testClient(t *testing.T) IStorage {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	mongoClient, err := mongo.Connect(
		ctx,
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

	s, err := NewStorage(mongoClient, "test", logger.LoggerDefault{})
	if err != nil {
		t.Fatal(err)
	}

	if err = s.Start(ctx); err != nil {
		t.Fatal(err)
	}

	return s
}

// Test_storage_UpsertOperInstData ...
// NOCC: golint/fnsize.
func Test_storage_UpsertOperInstData(t *testing.T) {
	now := time.Now()
	type args struct {
		ctx  context.Context
		data *operation.InstanceData
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx: context.Background(),
				data: &operation.InstanceData{
					InstanceBriefData: operation.InstanceBriefData{
						Metadata: &operation.InstanceMetadata{
							TriggerID:           "trigger-1",
							OperationInstanceID: "operation-inst-7bd49883-bcc9-4776-80ff-d3d37ca4143f",
							OperationDefName:    "operation-def-1",
							OperationID:         "",
							ActionNames:         []string{"action-1"},
							ParentOperationID:   "",
							Timeout:             0,
							InitContent:         map[string]any{},
						},
						Lifecycle: &operation.Lifecycle{
							State:     "success",
							StartedAt: time.Time{},
							CreatedAt: time.Time{},
							EndedAt:   time.Time{},
							StoppedAt: time.Time{},
						},
					},
					ActionInstanceDataMap: map[string]*action.InstanceData{
						"action-1": {
							TriggerID:           "trigger-1",
							OperationInstanceID: "operation-inst-7bd49883-bcc9-4776-80ff-d3d37ca4143f",
							Name:                "action-1",
							Index:               0,
							Messages:            nil,
							Content: map[string]any{
								"biz": "1",
							},
							Lifecycle: &action.Lifecycle{
								State:     "success",
								CreatedAt: time.Time{},
								StartedAt: time.Time{},
								EndedAt:   time.Time{},
								StoppedAt: time.Time{},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "normal",
			args: args{
				ctx: context.Background(),
				data: &operation.InstanceData{
					InstanceBriefData: operation.InstanceBriefData{
						Metadata: &operation.InstanceMetadata{
							TriggerID:           "trigger-001",
							OperationInstanceID: "op-instance-001",
							OperationDefName:    "def-order-process",
							ActionNames:         []string{"validate-order"},
							Timeout:             30 * time.Minute,
							InitContent: map[string]any{
								"order_id": "ORD-1001",
							},
						},
						Lifecycle: &operation.Lifecycle{
							State:     "success",
							CreatedAt: now.Add(-1 * time.Hour),
							StartedAt: now.Add(-44 * time.Minute),
							EndedAt:   now.Add(-40 * time.Minute),
						},
					},
					ActionInstanceDataMap: map[string]*action.InstanceData{
						"validate-order": {
							TriggerID:           "trigger-001",
							OperationInstanceID: "op-instance-001",
							Name:                "validate-order",
							Index:               0,
							Content: map[string]any{
								"items": []string{"item-1", "item-2"},
							},
							Lifecycle: &action.Lifecycle{
								State:     "success",
								StartedAt: now.Add(-44 * time.Minute),
								EndedAt:   now.Add(-43 * time.Minute),
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple_actions_mixed_status",
			args: args{
				ctx: context.Background(),
				data: &operation.InstanceData{
					InstanceBriefData: operation.InstanceBriefData{
						Metadata: &operation.InstanceMetadata{
							TriggerID:           "trigger-001",
							OperationInstanceID: "op-instance-002",
							OperationDefName:    "def-inventory-check",
							ActionNames:         []string{"check-stock", "update-inventory"},
							ParentOperationID:   "parent-op-001",
							Timeout:             1 * time.Hour,
							InitContent:         map[string]any{},
						},
						Lifecycle: &operation.Lifecycle{
							State:     "running",
							CreatedAt: now.Add(-30 * time.Minute),
							StartedAt: now.Add(-24 * time.Minute),
						},
					},
					ActionInstanceDataMap: map[string]*action.InstanceData{
						"check-stock": {
							Name:                "check-stock",
							TriggerID:           "trigger-002",
							OperationInstanceID: "op-instance-002",
							Index:               0,
							Lifecycle: &action.Lifecycle{
								State:     "success",
								StartedAt: now.Add(-24 * time.Minute),
								EndedAt:   now.Add(-20 * time.Minute),
							},
						},
						"update-inventory": {
							Name:                "update-inventory",
							TriggerID:           "trigger-002",
							OperationInstanceID: "op-instance-002",
							Index:               1,
							Content: map[string]any{
								"sku":      "SKU-1234",
								"quantity": 40,
							},
							Lifecycle: &action.Lifecycle{
								State:     "running",
								StartedAt: now.Add(-18 * time.Minute),
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty_action_data",
			args: args{
				ctx: context.Background(),
				data: &operation.InstanceData{
					InstanceBriefData: operation.InstanceBriefData{
						Metadata: &operation.InstanceMetadata{
							TriggerID:           "trigger-001",
							OperationInstanceID: "op-instance-003",
							OperationDefName:    "operation-def-1",
							OperationID:         "",
							ActionNames:         []string{},
							ParentOperationID:   "",
							Timeout:             0,
							InitContent:         map[string]any{},
						},
						Lifecycle: &operation.Lifecycle{
							State:     "success",
							StartedAt: time.Time{},
							CreatedAt: time.Time{},
							EndedAt:   time.Time{},
							StoppedAt: time.Time{},
						},
					},
					ActionInstanceDataMap: map[string]*action.InstanceData{},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testClient(t)
			if err := s.UpsertOperationInstanceData(tt.args.ctx, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("UpsertOperInstData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_storage_GetActionInstanceData ...
func Test_storage_GetActionInstanceData(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		wantErr         bool
		operaInstanceId string
		actionName      string
	}{
		{
			name:            "normal",
			operaInstanceId: "op-instance-002",
			actionName:      "check-stock",
			wantErr:         false,
		},
		{
			name:            "normal",
			operaInstanceId: "op-instance-003",
			actionName:      "check-stock",
			wantErr:         true,
		},
		{
			name:            "error",
			operaInstanceId: "op-instance-005",
			actionName:      "check-stock",
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testClient(t)
			got, err := s.GetActionInstanceData(ctx, tt.operaInstanceId, tt.actionName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetActionInstanceData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("got = %#v\n", got)
		})
	}
}

// GetActionInstanceLifecycle ...
func Test_storage_GetActionInstanceLifecycle(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		wantErr         bool
		operaInstanceId string
		actionName      string
	}{
		{
			name:            "normal",
			operaInstanceId: "op-instance-002",
			actionName:      "check-stock",
			wantErr:         false,
		},
		{
			name:            "normal",
			operaInstanceId: "op-instance-003",
			actionName:      "check-stock",
			wantErr:         true,
		},
		{
			name:            "error",
			operaInstanceId: "op-instance-005",
			actionName:      "check-stock",
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testClient(t)
			got, err := s.GetActionInstanceLifecycle(ctx, tt.operaInstanceId, tt.actionName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetActionInstanceLifecycle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("got = %#v\n", got)

		})
	}
}

// UpdateActionInstanceLifecycle ...
func Test_storage_UpdateActionInstanceLifecycle(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		wantErr         bool
		operaInstanceId string
		actionName      string
		lifecycle       *action.Lifecycle
	}{
		{
			name:            "normal",
			operaInstanceId: "op-instance-002",
			actionName:      "check-stock-test",
			lifecycle: &action.Lifecycle{
				State:   action.StatePending,
				EndedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name:            "error-1",
			operaInstanceId: "op-instance-004",
			actionName:      "check-stock",
			lifecycle: &action.Lifecycle{
				State:   action.StatePending,
				EndedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name:            "error-2",
			operaInstanceId: "op-instance-006",
			actionName:      "check-stock",
			lifecycle: &action.Lifecycle{
				State:   action.StatePending,
				EndedAt: time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testClient(t)
			err := s.UpdateActionInstanceLifecycle(ctx, tt.operaInstanceId, tt.actionName, tt.lifecycle)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateActionInstanceLifecycle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

// PushActionInstanceMessage ...
func Test_storage_PushActionInstanceMessage(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		wantErr         bool
		operaInstanceId string
		actionName      string
		messgae         *action.Message
	}{
		{
			name:            "normal",
			operaInstanceId: "op-instance-002",
			actionName:      "check-stock",
			messgae: &action.Message{
				Time: time.Now(),
				Text: "test1",
			},
			wantErr: false,
		},
		{
			name:            "error-1",
			operaInstanceId: "op-instance-003",
			actionName:      "check-stock-003",
			messgae: &action.Message{
				Time: time.Now(),
				Text: "test2",
			},
			wantErr: true,
		},
		{
			name:            "error-2",
			operaInstanceId: "op-instance-005",
			actionName:      "check-stock-004",
			messgae: &action.Message{
				Time: time.Now(),
				Text: "test3",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testClient(t)
			err := s.PushActionInstanceMessage(ctx, tt.operaInstanceId, tt.actionName, *tt.messgae)
			if (err != nil) != tt.wantErr {
				t.Errorf("PushActionInstanceMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

// GetOperationInstanceFullData ...
func Test_storage_GetOperationInstanceFullData(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		wantErr         bool
		operaInstanceId string
	}{
		{
			name:            "normal",
			operaInstanceId: "op-instance-002",
			wantErr:         false,
		},
		{
			name:            "normal-1",
			operaInstanceId: "op-instance-003",
			wantErr:         false,
		},
		{
			name:            "error-2",
			operaInstanceId: "op-instance-005",
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testClient(t)
			got, err := s.GetOperationInstanceFullData(ctx, tt.operaInstanceId)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOperationInstanceFullData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil {
				t.Logf("got:%+v,actions:%+v,actionMap:%+v", got, got.Metadata.ActionNames, got.ActionInstanceDataMap)
			}
		})
	}
}

// GetOperationInstanceBriefData ...
func Test_storage_GetOperationInstanceBriefData(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		wantErr         bool
		operaInstanceId string
	}{
		{
			name:            "normal",
			operaInstanceId: "op-instance-002",
			wantErr:         false,
		},
		{
			name:            "normal-1",
			operaInstanceId: "op-instance-003",
			wantErr:         false,
		},
		{
			name:            "error-2",
			operaInstanceId: "op-instance-005",
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testClient(t)
			got, err := s.GetOperationInstanceBriefData(ctx, tt.operaInstanceId)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOperationInstanceBriefData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil {
				t.Logf("got:%+v,actions:%+v", got, got.Metadata.ActionNames)
			}
		})
	}
}

// ListOperationInstanceBriefData ...
func Test_storage_ListOperationInstanceBriefData(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		wantErr   bool
		triggerID string
		wantCount int64
	}{
		{
			name:      "normal",
			triggerID: "trigger-001",
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "normal_2",
			triggerID: "trigger-1",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "error_2",
			triggerID: "trigger-10",
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testClient(t)
			got, err := s.ListOperationInstanceBriefData(ctx, types.Page{}, tt.triggerID, operation.StateRunning, operation.StateSuccess)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListOperationInstanceBriefData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != int(tt.wantCount) {
				t.Errorf("ListOperationInstanceBriefData() count = %v, wantCount %v", len(got), int(tt.wantCount))
				return
			}

			for _, v := range got {
				t.Logf("got:%+v,actions:%+v,instanceId:%+v", v, v.Metadata.ActionNames, v.Metadata.OperationInstanceID)
			}

		})
	}
}

// CountOperationInstance ...
func Test_storage_CountOperationInstance(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		wantErr   bool
		triggerID string
		wantCount int64
	}{
		{
			name:      "normal",
			triggerID: "trigger-001",
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "normal_2",
			triggerID: "trigger-1",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "zero",
			triggerID: "trigger-10",
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testClient(t)
			got, err := s.CountOperationInstance(ctx, tt.triggerID, operation.StateRunning, operation.StateSuccess)
			if (err != nil) != tt.wantErr {
				t.Errorf("CountOperationInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantCount {
				t.Errorf("CountOperationInstance() count = %v, wantCount %v", got, int(tt.wantCount))
				return
			}

		})
	}
}

// UpdateOperationInstanceLifecycle ...
func Test_storage_UpdateOperationInstanceLifecycle(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		wantErr     bool
		operationID string
		lifecycle   *operation.Lifecycle
	}{
		{
			name:        "normal",
			operationID: "op-instance-001",
			lifecycle: &operation.Lifecycle{
				State:     operation.StateFailed,
				CreatedAt: time.Time{},
				EndedAt:   time.Time{},
				StoppedAt: time.Time{},
			},
			wantErr: false,
		},
		{
			name:        "normal_2",
			operationID: "op-instance-002",
			lifecycle: &operation.Lifecycle{
				State:     operation.StateFailed,
				CreatedAt: time.Time{},
				EndedAt:   time.Time{},
				StoppedAt: time.Time{},
			},
			wantErr: false,
		},
		{
			name:        "error_2",
			operationID: "op-instance-004",
			lifecycle: &operation.Lifecycle{
				State:     operation.StateInit,
				CreatedAt: time.Time{},
				EndedAt:   time.Time{},
				StoppedAt: time.Time{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testClient(t)
			err := s.UpdateOperationInstanceLifecycle(ctx, tt.operationID, tt.lifecycle)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateOperationInstanceLifecycle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
