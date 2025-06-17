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

// testHandler ...
func testHandler(t *testing.T) IHandler {
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

	return New(mongoClient.Database(os.Getenv("MONGO_DATABASE")), logger.LoggerDefault{})
}

// Test_handler_Upsert ...
func Test_handler_Upsert(t *testing.T) {
	now := time.Now()
	type args struct {
		ctx  context.Context
		data *operation.InstanceData
	}

	mockData := func(id, trigger, def string, actions []string, state string, content map[string]any) *operation.InstanceData {
		return &operation.InstanceData{
			InstanceBriefData: operation.InstanceBriefData{
				Metadata: &operation.InstanceMetadata{
					TriggerID:           trigger,
					OperationInstanceID: id,
					OperationDefName:    def,
					ActionNames:         actions,
				},
				Lifecycle: &operation.Lifecycle{State: operation.State(state)},
			},
			ActionInstanceDataMap: map[string]*action.InstanceData{
				actions[0]: {
					TriggerID:           trigger,
					OperationInstanceID: id,
					Name:                actions[0],
					Content:             content,
					Lifecycle:           &action.Lifecycle{State: action.State(state)},
				},
			},
		}
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "single_action",
			args: args{
				ctx: context.Background(),
				data: mockData(
					"operation-inst-7bd49883-bcc9-4776-80ff-d3d37ca4143f",
					"trigger-1",
					"operation-def-1",
					[]string{"action-1"},
					"success",
					map[string]any{"biz": "1"},
				),
			},
		},
		{
			name: "complex_workflow",
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
							InitContent:         map[string]any{"order_id": "ORD-1001"},
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
							Content:             map[string]any{"items": []string{"item-1", "item-2"}},
							Lifecycle: &action.Lifecycle{
								State:     "success",
								StartedAt: now.Add(-44 * time.Minute),
								EndedAt:   now.Add(-43 * time.Minute),
							},
						},
					},
				},
			},
		},
		{
			name: "multi_action",
			args: args{
				ctx: context.Background(),
				data: &operation.InstanceData{
					InstanceBriefData: operation.InstanceBriefData{
						Metadata: &operation.InstanceMetadata{
							TriggerID:           "trigger-002",
							OperationInstanceID: "op-instance-002",
							OperationDefName:    "def-inventory-check",
							ActionNames:         []string{"check-stock", "update-inventory"},
						},
						Lifecycle: &operation.Lifecycle{State: "running"},
					},
					ActionInstanceDataMap: map[string]*action.InstanceData{
						"check-stock": {
							Name:                "check-stock",
							TriggerID:           "trigger-002",
							OperationInstanceID: "op-instance-002",
							Lifecycle:           &action.Lifecycle{State: "success"},
						},
						"update-inventory": {
							Name:                "update-inventory",
							TriggerID:           "trigger-002",
							OperationInstanceID: "op-instance-002",
							Content:             map[string]any{"sku": "SKU-1234", "quantity": 40},
							Lifecycle:           &action.Lifecycle{State: "running"},
						},
					},
				},
			},
		},
		{
			name: "no_actions",
			args: args{
				ctx: context.Background(),
				data: &operation.InstanceData{
					InstanceBriefData: operation.InstanceBriefData{
						Metadata:  &operation.InstanceMetadata{TriggerID: "trigger-003"},
						Lifecycle: &operation.Lifecycle{State: "success"},
					},
					ActionInstanceDataMap: map[string]*action.InstanceData{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testHandler(t)
			if err := h.Upsert(tt.args.ctx, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Upsert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_handler_FindOne ...
func Test_handler_FindOne(t *testing.T) {
	type args struct {
		ctx  context.Context
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
				ctx: context.Background(),
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
			got, err := h.FindOne(tt.args.ctx, tt.args.opts...)
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
		ctx  context.Context
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
				ctx:  context.Background(),
				opts: []OptFn{WithOperInstID("operation-inst-7bd49883-bcc9-4776-80ff-d3d37ca4143f")},
			},
			wantErr: false,
		},
		{
			name: "normal_2",
			args: args{
				ctx:  context.Background(),
				opts: []OptFn{WithOperInstID("op-instance-001")},
			},
			wantErr: false,
		},
		{
			name: "normal_3",
			args: args{
				ctx:  context.Background(),
				opts: []OptFn{WithOperInstID("op-instance-002")},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testHandler(t)
			got, err := h.FindOneWithoutActionData(tt.args.ctx, tt.args.opts...)
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
		ctx  context.Context
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
				ctx:  context.Background(),
				opts: []OptFn{WithOperInstID("operation-inst-7bd49883-bcc9-4776-80ff-d3d37ca4143f")},
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name: "normal_2",
			args: args{
				ctx:  context.Background(),
				opts: []OptFn{},
			},
			wantErr:   false,
			wantCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testHandler(t)
			num, err := h.Count(tt.args.ctx, tt.args.opts...)
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

// Test_ListFullData ...
func Test_ListFullData(t *testing.T) {
	type args struct {
		ctx  context.Context
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
				ctx:  context.Background(),
				opts: []OptFn{WithOperInstID("operation-inst-7bd49883-bcc9-4776-80ff-d3d37ca4143f")},
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name: "normal_2",
			args: args{
				ctx:  context.Background(),
				opts: []OptFn{},
			},
			wantErr:   false,
			wantCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testHandler(t)
			got, num, err := h.ListFullData(tt.args.ctx, types.Page{}, tt.args.opts...)
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
			for _, data := range got {
				t.Logf("Test_ListFullData() got =%+v \n", data)
			}

		})
	}
}

// Test_ListWithoutActInst ...
func Test_ListWithoutActInst(t *testing.T) {
	type args struct {
		ctx  context.Context
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
				ctx:  context.Background(),
				opts: []OptFn{WithOperInstID("operation-inst-7bd49883-bcc9-4776-80ff-d3d37ca4143f")},
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name: "normal_2",
			args: args{
				ctx:  context.Background(),
				opts: []OptFn{},
			},
			wantErr:   false,
			wantCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testHandler(t)
			got, num, err := h.ListWithoutActInst(tt.args.ctx, types.Page{}, tt.args.opts...)
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
			for _, data := range got {
				t.Logf("Test_ListFullData() got =%+v,actions:%+v \n", data, data.Metadata.ActionNames)
			}

		})
	}
}

// FindOneWithoutActionData ...
func Test_handler_FindOneWithoutActionData(t *testing.T) {
	type args struct {
		ctx  context.Context
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
				ctx: context.Background(),
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
			got, err := h.FindOneWithoutActionData(tt.args.ctx, tt.args.opts...)
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
		ctx        context.Context
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
				ctx:        context.Background(),
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
				ctx:        context.Background(),
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
			err := h.UpdateLifeCycle(tt.args.ctx, tt.args.OperInstID, tt.args.LifeCycle)
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
