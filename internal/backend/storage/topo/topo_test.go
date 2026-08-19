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

// Package topo ...
package topo

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	mongohost "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/host"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// testClient ...
func testClient(t *testing.T) IStorage {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	nCtx := contextx.New(context.Background())
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

	s, err := NewStorage(mongoClient, "test")
	if err != nil {
		t.Fatal(err)
	}

	if err = s.Start(nCtx); err != nil {
		t.Fatal(err)
	}

	return s
}

// Test_storage_UpsertManyHost ...
func Test_storage_UpsertManyHost(t *testing.T) {
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("single"))

	type args struct {
		nCtx  contextx.IContext
		hosts []*types.Host
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "nil nCtx",
			args: args{
				nCtx:  nil,
				hosts: nil,
			},
			wantErr: true,
		},
		{
			name: "normal",
			args: args{
				nCtx: nCtx,
				hosts: []*types.Host{
					{
						TenantID: "single",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := testClient(t)
			if err := ds.UpsertManyHost(tt.args.nCtx, tt.args.hosts...); (err != nil) != tt.wantErr {
				t.Errorf("UpsertManyHost() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_convertHostConditionsToOptions_MatchesSetAndModuleInSameTopo(t *testing.T) {
	opts := convertHostConditionsToOptions(&types.HostCondition{
		StaticExactInclude: &types.HostStaticExactFields{
			SetID:    []int64{1, 2},
			ModuleID: []int64{10, 20},
		},
	})

	filter := bson.D{}
	for _, opt := range opts {
		filter = opt(filter)
	}

	if len(filter) != 1 {
		t.Fatalf("filter length = %d, want 1: %#v", len(filter), filter)
	}

	orElem := filter[0]
	if orElem.Key != "$or" {
		t.Fatalf("filter key = %s, want $or: %#v", orElem.Key, filter)
	}

	conditions, ok := orElem.Value.(bson.A)
	if !ok {
		t.Fatalf("$or value type = %T, want bson.A", orElem.Value)
	}
	if len(conditions) != 4 {
		t.Fatalf("$or condition length = %d, want 4: %#v", len(conditions), conditions)
	}

	for _, condition := range conditions {
		conditionDoc, ok := condition.(bson.D)
		if !ok {
			t.Fatalf("$or condition type = %T, want bson.D", condition)
		}
		if len(conditionDoc) != 1 || conditionDoc[0].Key != mongohost.FieldKeyStaticTopo {
			t.Fatalf("condition = %#v, want single topo elemMatch", conditionDoc)
		}
		elemMatch, ok := conditionDoc[0].Value.(bson.M)["$elemMatch"].(bson.D)
		if !ok {
			t.Fatalf("topo condition value = %#v, want $elemMatch bson.D", conditionDoc[0].Value)
		}
		if len(elemMatch) != 2 {
			t.Fatalf("elemMatch length = %d, want set_id and module_id: %#v", len(elemMatch), elemMatch)
		}
	}
}

func Test_convertHostConditionsToOptions_DynamicAgentIDNotEmpty(t *testing.T) {
	opts := convertHostConditionsToOptions(&types.HostCondition{
		DynamicAgentIDNotEmpty: true,
	})

	filter := bson.D{}
	for _, opt := range opts {
		filter = opt(filter)
	}

	expected := bson.D{{Key: mongohost.FieldKeyDynamicAgentID, Value: bson.M{"$gt": ""}}}
	if !reflect.DeepEqual(filter, expected) {
		t.Fatalf("filter = %#v, want %#v", filter, expected)
	}
}
