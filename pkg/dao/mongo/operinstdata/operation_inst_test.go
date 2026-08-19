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

// Package operinstdata ...
package operinstdata

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/common"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// testHandler ...
func testHandler(t *testing.T) IHandler {
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

// Test_handler_Upsert ...
func Test_handler_Upsert(t *testing.T) {
	now := time.Now()
	type args struct {
		nCtx context.Context
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
				nCtx: context.Background(),
				data: &operation.InstanceData{
					InstanceBriefData: operation.InstanceBriefData{
						Metadata: &operation.InstanceMetadata{
							TriggerID:              "trigger-1",
							OperationInstanceID:    "operation-inst-7bd49883-bcc9-4776-80ff-d3d37ca4143f",
							OperationDefName:       "operation-def-1",
							OperationID:            "",
							ActionNames:            []string{"action-1"},
							ParentOperationID:      "",
							Timeout:                0,
							InitContent:            map[string]any{},
							ExtraExecutionName:     "extra-execution",
							ExtraExecutionMessages: make([]common.Message, 0),
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
				nCtx: context.Background(),
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
							ExtraExecutionName:     "extra-execution",
							ExtraExecutionMessages: make([]common.Message, 0),
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
				nCtx: context.Background(),
				data: &operation.InstanceData{
					InstanceBriefData: operation.InstanceBriefData{
						Metadata: &operation.InstanceMetadata{
							TriggerID:              "trigger-002",
							OperationInstanceID:    "op-instance-002",
							OperationDefName:       "def-inventory-check",
							ActionNames:            []string{"check-stock", "update-inventory"},
							ParentOperationID:      "parent-op-001",
							Timeout:                1 * time.Hour,
							InitContent:            map[string]any{},
							ExtraExecutionName:     "extra-execution",
							ExtraExecutionMessages: make([]common.Message, 0),
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
				nCtx: context.Background(),
				data: &operation.InstanceData{
					InstanceBriefData: operation.InstanceBriefData{
						Metadata: &operation.InstanceMetadata{
							TriggerID:              "trigger-003",
							OperationInstanceID:    "op-instance-003",
							OperationDefName:       "operation-def-1",
							OperationID:            "",
							ActionNames:            []string{},
							ParentOperationID:      "",
							Timeout:                0,
							InitContent:            map[string]any{},
							ExtraExecutionName:     "extra-execution",
							ExtraExecutionMessages: make([]common.Message, 0),
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
			h := testHandler(t)
			if err := h.Upsert(tt.args.nCtx, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Upsert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_handler_FindOne ...
func Test_handler_FindOne(t *testing.T) {
	type args struct {
		nCtx context.Context
		opts []OptFn
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				nCtx: context.Background(),
				opts: []OptFn{
					WithOperInstID("op-instance-002"),
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testHandler(t)
			got, err := h.FindOne(tt.args.nCtx, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindOne() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("got = %#v\n", got)
			for name, data := range got.ActionInstanceDataMap {
				t.Logf("action = %s, data = %#v\n", name, data)
			}
		})
	}
}

// Test_handler_FindOneOperInstDataWithoutActionData ...
func Test_handler_FindOneOperInstDataWithoutActionData(t *testing.T) {
	type args struct {
		nCtx context.Context
		opts []OptFn
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal_1",
			args: args{
				nCtx: context.Background(),
				opts: []OptFn{WithOperInstID("operation-inst-7bd49883-bcc9-4776-80ff-d3d37ca4143f")},
			},
			wantErr: false,
		},
		{
			name: "normal_2",
			args: args{
				nCtx: context.Background(),
				opts: []OptFn{WithOperInstID("op-instance-001")},
			},
			wantErr: false,
		},
		{
			name: "normal_3",
			args: args{
				nCtx: context.Background(),
				opts: []OptFn{WithOperInstID("op-instance-002")},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testHandler(t)
			got, err := h.FindOneWithoutActionData(tt.args.nCtx, tt.args.opts...)
			if err != nil {
				t.Logf("FindOneWithoutActionData() error = %v", err)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("FindOneWithoutActionData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("FindOneWithoutActionData() got = %+v\n", got)

		})
	}
}

// Test_handler_Count ...
func Test_handler_Count(t *testing.T) {
	type args struct {
		nCtx context.Context
		opts []OptFn
	}

	tests := []struct {
		name      string
		args      args
		wantErr   bool
		wantCount int64
	}{
		{
			name: "normal_1",
			args: args{
				nCtx: context.Background(),
				opts: []OptFn{WithOperInstID("operation-inst-7bd49883-bcc9-4776-80ff-d3d37ca4143f")},
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name: "normal_2",
			args: args{
				nCtx: context.Background(),
				opts: []OptFn{},
			},
			wantErr:   false,
			wantCount: 218,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testHandler(t)
			num, err := h.Count(tt.args.nCtx, tt.args.opts...)
			if err != nil {
				t.Logf("Test_handler_Count() error = %v", err)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("Test_handler_Count() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantCount != num {
				t.Errorf("Test_handler_Count() got = %v, wantCount %v", num, tt.wantCount)
			}
			t.Logf("Test_handler_Count() count = %+v\n", num)

		})
	}
}

// Test_ListWithoutActInst ...
func Test_ListWithoutActInst(t *testing.T) {
	type args struct {
		nCtx context.Context
		opts []OptFn
	}

	tests := []struct {
		name      string
		args      args
		wantErr   bool
		wantCount int64
	}{
		{
			name: "normal_1",
			args: args{
				nCtx: context.Background(),
				opts: []OptFn{WithOperInstID("operation-inst-7bd49883-bcc9-4776-80ff-d3d37ca4143f")},
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name: "normal_2",
			args: args{
				nCtx: context.Background(),
				opts: []OptFn{},
			},
			wantErr:   false,
			wantCount: 218,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testHandler(t)
			_, num, err := h.ListWithoutActInst(tt.args.nCtx, types.Page{}, tt.args.opts...)
			if err != nil {
				t.Logf("Test_ListFullData() error = %v", err)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("Test_ListFullData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantCount != num {
				t.Errorf("Test_ListFullData() got = %v, wantCount %v", num, tt.wantCount)
			}

		})
	}
}

// FindOneWithoutActionData ...
func Test_handler_FindOneWithoutActionData(t *testing.T) {
	type args struct {
		nCtx context.Context
		opts []OptFn
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				nCtx: context.Background(),
				opts: []OptFn{
					WithOperInstID("op-instance-002"),
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testHandler(t)
			got, err := h.FindOneWithoutActionData(tt.args.nCtx, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindOne() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("got =%+v\n,actions:%+v \n", got, got.Metadata.ActionNames)
			for name, data := range got.ActionInstanceDataMap {
				t.Logf("action = %s, data = %+v\n", name, data)
			}
		})
	}
}

func Test_UpdateLifeCycle(t *testing.T) {
	type args struct {
		nCtx       context.Context
		OperInstID string
		LifeCycle  *operation.Lifecycle
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal_1",
			args: args{
				nCtx:       context.Background(),
				OperInstID: "operation-inst-7bd49883-bcc9-4776-80ff-d3d37ca4143f",
				LifeCycle: &operation.Lifecycle{
					State:     operation.StateTerminated,
					StartedAt: time.Now(),
					EndedAt:   time.Now(),
				},
			},
			wantErr: false,
		},
		{
			name: "normal_2",
			args: args{
				nCtx:       context.Background(),
				OperInstID: "op-instance-002",
				LifeCycle: &operation.Lifecycle{
					State:     operation.StateFailed,
					StartedAt: time.Now(),
					EndedAt:   time.Now(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testHandler(t)
			err := h.UpdateLifeCycle(tt.args.nCtx, tt.args.OperInstID, tt.args.LifeCycle)
			if err != nil {
				t.Logf("Test_ListFullData() error = %v", err)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("Test_ListFullData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}

// Test_UpdateExecMessages tests the UpdateExecMessages method.
func Test_UpdateExecMessages(t *testing.T) {
	type args struct {
		nCtx       context.Context
		OperInstID string
		ExecName   string
		Messages   []common.Message
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal_1",
			args: args{
				nCtx:       context.Background(),
				OperInstID: "operation-inst-7bd49883-bcc9-4776-80ff-d3d37ca4143f",
				Messages: []common.Message{
					{
						Time:  time.Now(),
						Text:  "This is a test message",
						Level: "INFO",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "normal_2",
			args: args{
				nCtx:       context.Background(),
				OperInstID: "op-instance-002",
				Messages: []common.Message{
					{
						Time:  time.Now(),
						Text:  "This is a test message",
						Level: "INFO",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testHandler(t)
			err := h.UpdateExtraExecutionMessages(tt.args.nCtx, tt.args.OperInstID, tt.args.Messages...)
			if err != nil {
				t.Logf("Test_UpdateExecMessages() error = %v", err)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("Test_UpdateExecMessages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}
