/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cptemplate

import (
	"context"
	"os"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

// Test_UpsertMany tests UpsertMany.
func Test_UpsertMany(t *testing.T) {
	tenant.SetMode(tenant.ModeMultiple)
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("test"))

	type args struct {
		nCtx                 context.Context
		configPolicyTemplate []*types.ConfigPolicyTemplate
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "invalid nCtx",
			args: args{
				nCtx: nil,
			},
			wantErr: true,
		},
		{
			name: "invalid template",
			args: args{
				nCtx:                 nCtx,
				configPolicyTemplate: nil,
			},
			wantErr: true,
		},
		{
			name: "normal",
			args: args{
				nCtx: nCtx,
				configPolicyTemplate: []*types.ConfigPolicyTemplate{
					&types.ConfigPolicyTemplate{
						TenantID: "test",
						ID:       1,
						Template: "{}",
					},
					&types.ConfigPolicyTemplate{
						TenantID: "test",
						ID:       2,
						Template: "{\"a\":\"b\"}",
					},
					&types.ConfigPolicyTemplate{
						TenantID: "test",
						ID:       3,
						Template: "{\"c\":\"d\"}",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			err := h.UpsertMany(tt.args.nCtx, tt.args.configPolicyTemplate...)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpsertMany() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

// Test_Get tests Get.
func Test_Get(t *testing.T) {
	tenant.SetMode(tenant.ModeMultiple)
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("test"))

	type args struct {
		nCtx           context.Context
		configPolicyID int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "invalid nCtx",
			args: args{
				nCtx: nil,
			},
			wantErr: true,
		},
		{
			name: "invalid id",
			args: args{
				nCtx:           nCtx,
				configPolicyID: -1,
			},
			wantErr: true,
		},
		{
			name: "normal-1",
			args: args{
				nCtx:           nCtx,
				configPolicyID: 1,
			},
			wantErr: false,
		},

		{
			name: "normal-2",
			args: args{
				nCtx:           nCtx,
				configPolicyID: 2,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			_, err := h.Get(tt.args.nCtx, tt.args.configPolicyID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

// Test_DeleteMany tests DeleteMany.
func Test_DeleteMany(t *testing.T) {
	tenant.SetMode(tenant.ModeMultiple)
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("test"))

	type args struct {
		nCtx           context.Context
		configPolicyID []int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "invalid nCtx",
			args: args{
				nCtx: nil,
			},
			wantErr: true,
		},
		{
			name: "invalid ids",
			args: args{
				nCtx:           nCtx,
				configPolicyID: nil,
			},
			wantErr: true,
		},
		{
			name: "normal-1",
			args: args{
				nCtx:           nCtx,
				configPolicyID: []int64{1, 2, 3},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			if err := h.DeleteMany(tt.args.nCtx, tt.args.configPolicyID...); (err != nil) != tt.wantErr {
				t.Errorf("DeleteMany() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
