/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operation ...
package operation

import (
	"context"
	"os"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// testClient ...
func testClient(t *testing.T) Storage {
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

// Test_storage_UpsertOperation tests the UpsertOperation method of the storage
func Test_storage_UpsertOperation(t *testing.T) {
	type args struct {
		ctx       context.Context
		operation *operation.Operation
	}

	baseOperation := &operation.Operation{
		OperationID: "test-op-id",
		TriggerID:   "test-trigger-id",
		Definition:  &operation.DefinitionSnapshot{SnapshotName: "test-def"},
	}

	tests := []struct {
		name        string
		args        args
		wantErr     bool
		preInsert   bool
		validateKey string
	}{
		{
			name: "normal",
			args: args{
				ctx:       context.Background(),
				operation: baseOperation,
			},
			wantErr:     false,
			validateKey: "OperationID",
		},
		{
			name: "update operation",
			args: args{
				ctx: context.Background(),
				operation: &operation.Operation{
					OperationID: "test-op-id",
					TriggerID:   "updated-trigger-id",
				},
			},
			preInsert:   true,
			wantErr:     false,
			validateKey: "TriggerID",
		},
		{
			name: "nil operation",
			args: args{
				ctx:       context.Background(),
				operation: nil,
			},
			wantErr: true,
		},
		{
			name: "nil ctx",
			args: args{
				ctx:       nil,
				operation: baseOperation,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testClient(t)
			ctx := tt.args.ctx

			if tt.preInsert {
				if err := s.UpsertOperation(context.Background(), baseOperation); err != nil {
					t.Fatalf("err: %v", err)
				}
			}

			err := s.UpsertOperation(ctx, tt.args.operation)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpsertOperation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.args.operation != nil {
				got, err := s.GetOperation(context.Background(), tt.args.operation.OperationID)
				if err != nil {
					t.Errorf("err: %v", err)
					return
				}

				switch tt.validateKey {
				case "OperationID":
					if got.OperationID != tt.args.operation.OperationID {
						t.Errorf("OperationID no match, got = %v, want %v", got.OperationID, tt.args.operation.OperationID)
					}
				case "TriggerID":
					if got.TriggerID != tt.args.operation.TriggerID {
						t.Errorf("TriggerID no match, got = %v, want %v", got.TriggerID, tt.args.operation.TriggerID)
					}
				}
			}

		})
	}
}

// Test_storage_GetOperation ...
func Test_storage_GetOperation(t *testing.T) {
	type args struct {
		ctx         context.Context
		operationID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:         context.Background(),
				operationID: "35fa1c8a-3089-4a89-9b45-100398698dd2",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testClient(t)
			got, err := s.GetOperation(tt.args.ctx, tt.args.operationID)
			if err != nil {
				t.Logf("GetOperation() error = %v", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOperation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("got: %+v", got)
		})
	}
}

// Test_storage_GetOperation ...
func Test_storage_ListOperation(t *testing.T) {
	type args struct {
		ctx         context.Context
		operationID string
	}
	tests := []struct {
		name       string
		args       args
		wantCount  int
		wantErr    bool
		trigger_id string
	}{
		{
			name: "normal",
			args: args{
				ctx:         context.Background(),
				operationID: "35fa1c8a-3089-4a89-9b45-100398698dd2",
			},
			trigger_id: "trigger_base",
			wantCount:  2,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testClient(t)
			got, _, err := s.ListOperation(tt.args.ctx, types.Page{Offset: 0, Limit: 10}, tt.trigger_id)
			if err != nil {
				t.Logf("GetOperation() error = %v", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOperation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("got: %+v", got)
		})
	}
}

func Test_storage_ListEmptyOperation(t *testing.T) {
	type args struct {
		ctx         context.Context
		operationID string
	}
	tests := []struct {
		name       string
		args       args
		wantCount  int
		wantErr    bool
		trigger_id string
	}{
		{
			name: "normal",
			args: args{
				ctx:         context.Background(),
				operationID: "35fa1c8a-3089-4a89-9b45-100398698dd2",
			},
			trigger_id: "trigger_base",
			wantCount:  2,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testClient(t)
			got, _, err := s.ListOperation(tt.args.ctx, types.Page{Offset: 0, Limit: 10}, tt.trigger_id)
			if err != nil {
				t.Logf("GetOperation() error = %v", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOperation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("got: %+v", got)
		})
	}
}

// Test_storage_ListEmptyOpera ...
func Test_storage_ListEmptyOpera(t *testing.T) {
	type args struct {
		ctx         context.Context
		operationID string
	}
	tests := []struct {
		name       string
		args       args
		wantCount  int
		wantErr    bool
		trigger_id string
	}{
		{
			name: "normal",
			args: args{
				ctx:         context.Background(),
				operationID: "35fa1c8a-3089-4a89-9b45-100398698dd2",
			},
			trigger_id: "trigger_base",
			wantCount:  1,
			wantErr:    false,
		},
		{
			name: "normal",
			args: args{
				ctx:         context.Background(),
				operationID: "35fa1c8a-3089-4a89-9b45-100398698dd2",
			},
			trigger_id: "special_trigger",
			wantCount:  1,
			wantErr:    false,
		},
		{
			name: "error",
			args: args{
				ctx:         context.Background(),
				operationID: "35fa1c8a-3089-4a89-9b45-100398698dd2",
			},
			trigger_id: "no_exists_trigger",
			wantCount:  0,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testClient(t)
			got, _, err := s.ListEmptyOperation(tt.args.ctx, types.Page{Offset: 0, Limit: 10}, tt.trigger_id)
			if err != nil {
				t.Logf("GetOperation() error = %v", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOperation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("got: %+v", got)
		})
	}
}
