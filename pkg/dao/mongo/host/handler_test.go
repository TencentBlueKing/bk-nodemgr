/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package host ...
package host

import (
	"context"
	"os"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// testClient ...
func testClient(t *testing.T) Handler {
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

// Test_handler_ListAll ...
func Test_handler_ListAll(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "test")

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
			got, err := h.ListAll(ctx)
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

// UpsertMany upsert many hosts.
func Test_handler_UpsertMany(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "test")

	type args struct {
		ctx   context.Context
		hosts []*types.Host
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "nil ctx",
			args: args{
				ctx:   nil,
				hosts: nil,
			},
			wantErr: true,
		},
		{
			name: "nil hosts",
			args: args{
				ctx:   ctx,
				hosts: nil,
			},
			wantErr: true,
		},
		{
			name: "empty hosts",
			args: args{
				ctx:   ctx,
				hosts: []*types.Host{},
			},
			wantErr: true,
		},
		{
			name: "normal",
			args: args{
				ctx: ctx,
				hosts: []*types.Host{
					{
						TenantID: "single",
						CloudID:  1,
						BizID:    1,
						HostID:   1,
						InnerIP:  "127.0.0.1",
						Mac:      "123",
						OSType:   "centos",
					},
					{
						TenantID: "single",
						CloudID:  1,
						BizID:    2,
						HostID:   2,
						InnerIP:  "127.0.0.21",
						Mac:      "123",
						OSType:   "centos",
					},
					{
						TenantID: "single",
						CloudID:  1,
						BizID:    3,
						HostID:   3,
						InnerIP:  "127.0.0.3",
						Mac:      "123",
						OSType:   "centos",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "tenantID not match",
			args: args{
				ctx: ctx,
				hosts: []*types.Host{
					{
						TenantID: "test",
						CloudID:  1,
						BizID:    1,
						HostID:   1,
						InnerIP:  "127.0.0.1",
						Mac:      "123",
						OSType:   "centos",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			err := h.UpsertMany(tt.args.ctx, tt.args.hosts)
			if err != nil {
				t.Logf("UpsertMany() error = %v", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("UpsertMany() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
