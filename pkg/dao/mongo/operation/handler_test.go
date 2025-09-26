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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// testClient ...
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

// Test_handler_List tests the List method of the handler
func Test_handler_List(t *testing.T) {
	// prepare test data
	h := testClient(t)
	baseOp1 := &operation.Operation{
		TriggerID:   "trigger_base",
		OperationID: "operation_base1",
		InstanceIDs: []string{},
		Definition: &operation.DefinitionSnapshot{
			SnapshotName:               "order_processing1",
			SnapshotActionDefNames:     []string{"validate1", "charge1"},
			SnapshotDefaultParameters:  operation.Param{Timeout: 10 * time.Second},
			SnapshotExtraExecutionName: "extra_execution1",
		},
		Param: operation.Param{
			ParentOperationID: "parent_operation1",
			Timeout:           1 * time.Second,
			InitContent:       map[string]any{"key": "value1"},
		},
	}

	baseOp2 := &operation.Operation{
		TriggerID:   "trigger_base",
		OperationID: "operation_base2",
		InstanceIDs: []string{"instance_3", "instance_4"},
		Definition: &operation.DefinitionSnapshot{
			SnapshotName:               "order_processing2",
			SnapshotActionDefNames:     []string{"validate2", "charge2"},
			SnapshotDefaultParameters:  operation.Param{Timeout: 20 * time.Second},
			SnapshotExtraExecutionName: "extra_execution2",
		},
		Param: operation.Param{
			ParentOperationID: "parent_operation2",
			Timeout:           2 * time.Second,
			InitContent:       map[string]any{"key": "value2"},
		},
	}

	if err := h.Upsert(context.Background(), baseOp1); err != nil {
		t.Fatalf("prepare data failed: %v", err)
	}

	if err := h.Upsert(context.Background(), baseOp2); err != nil {
		t.Fatalf("prepare data failed: %v", err)
	}
	tests := []struct {
		nCtx      context.Context
		name      string
		page      types.Page
		opts      []OptFn
		wantCount int
		wantErr   bool
	}{
		{
			name:      "normal list",
			page:      types.Page{Offset: 0, Limit: 10},
			opts:      nil,
			wantCount: 2,
			wantErr:   false,
			nCtx:      context.Background(),
		},
		{
			name:      "filter by non-existent trigger",
			page:      types.Page{Offset: 0, Limit: 10},
			opts:      []OptFn{WithTriggerID("invalid_trigger")},
			wantCount: 0,
			wantErr:   false,
			nCtx:      context.Background(),
		},
		{
			name:      "invalid page params",
			page:      types.Page{Offset: -1, Limit: 0},
			wantCount: 0,
			wantErr:   true,
			nCtx:      context.Background(),
		},
		{
			name:      "nil nCtx",
			page:      types.Page{Offset: -1, Limit: 0},
			wantCount: 0,
			wantErr:   true,
			nCtx:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops, total, err := h.List(tt.nCtx, tt.page, tt.opts...)

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if int64(tt.wantCount) != total {
					t.Errorf("List() total = %v, want %v", total, tt.wantCount)
				}

				if len(ops) != tt.wantCount {
					t.Errorf("List() result count = %v, want %v", len(ops), tt.wantCount)
				}
			}
		})
	}
}

// InvalidFilterFn provides an invalid filter function for testing purposes.
func InvalidFilterFn() OptFn {
	return func(filter bson.D) bson.D {
		return append(filter, bson.E{
			Key: "$and",
			Value: []bson.D{
				{{Key: FieldKeyTriggerID, Value: "non_existent_value_1"}},
				{{Key: FieldKeyTriggerID, Value: "non_existent_value_2"}},
			},
		})
	}
}
