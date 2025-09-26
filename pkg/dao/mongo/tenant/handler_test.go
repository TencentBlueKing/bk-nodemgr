/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package tenant ...
package tenant

import (
	"context"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// testClient ...
func testClient(t *testing.T) Handler {
	nCtx := context.Background()
	mongoClient, err := mongo.Connect(
		nCtx,
		&options.ClientOptions{
			Hosts: []string{
				"mongo.dev.com:27017",
			},
			Auth: &options.Credential{
				Username:      "root",
				Password:      "mongo_root1",
				AuthSource:    "admin",
				AuthMechanism: "SCRAM-SHA-256",
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	return New(mongoClient.Database("test"))
}

// Test_handler_ListAll ...
func Test_handler_ListAll(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "test",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.ListAll(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("ListAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, v := range got {
				t.Logf("ListAll() got = %v", v)
			}
		})
	}
}

// Test_handler_Upsert ...
func Test_handler_Upsert(t *testing.T) {
	type args struct {
		tenant *types.Tenant
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				tenant: &types.Tenant{
					ID:     "0",
					Name:   "test",
					Status: false,
				},
			},
			wantErr: false,
		},
		{
			name: "test2",
			args: args{
				tenant: &types.Tenant{
					ID:     "1",
					Name:   "test",
					Status: false,
				},
			},
			wantErr: false,
		},
		{
			name: "test3",
			args: args{
				tenant: &types.Tenant{
					ID:     "2",
					Name:   "test",
					Status: false,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			if err := h.Upsert(context.Background(), tt.args.tenant); (err != nil) != tt.wantErr {
				t.Errorf("Upsert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
