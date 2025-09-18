/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package globalsettings provides storage for global settings.
package globalsettings

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
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

var once = sync.Once{}

// prepareData for all tests.
func prepareData(t *testing.T, ctx context.Context) {
	testDatas := []*types.GlobalSettings{
		{
			SettingName: "test1",
			Value:       "value1",
		},
		{
			SettingName: "test2",
			Value:       "value2",
		},
	}

	once.Do(func() {
		h := testClient(t)
		err := h.Upsert(ctx, testDatas...)
		if err != nil {
			t.Errorf("prepareData() error = %v", err)
		}
	})
}

// Test_handler_UpsertMany tests the UpsertMany method of the handler.
func Test_handler_UpsertMany(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "test")
	prepareData(t, ctx)

	type args struct {
		ctx context.Context
		gs  []*types.GlobalSettings
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx: ctx,
				gs: []*types.GlobalSettings{
					{
						SettingName: "test3",
						Value:       "value3",
					},
					{
						SettingName: "test2",
						Value:       "value4",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)

			err := h.Upsert(tt.args.ctx, tt.args.gs...)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpsertMany() error = %v, wantErr: %v", err, tt.wantErr)
				return
			}
		})
	}
}

// Test_handler_Get tests the Create method of the handler.
func Test_handler_Get(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "test")
	prepareData(t, ctx)

	type args struct {
		ctx  context.Context
		name string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    *types.GlobalSettings
	}{
		{
			name: "normal",
			args: args{
				ctx:  ctx,
				name: "test1",
			},
			wantErr: false,
			want: &types.GlobalSettings{
				SettingName: "test1",
				Value:       "value1",
			},
		},
		{
			name: "not found",
			args: args{
				ctx:  ctx,
				name: "test11",
			},
			wantErr: false,
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)

			got, err := h.Get(tt.args.ctx, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr: %v", err, tt.wantErr)
				return
			}

			if got != nil && got.Value != tt.want.Value {
				t.Errorf("Get() error, got %v, want: %v", got, tt.want)
			}

			t.Logf("Get() got = %v", got)
		})
	}
}

// Test_handler_List tests the List method of the handler.
func Test_handler_List(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "test")
	prepareData(t, ctx)

	type args struct {
		ctx  context.Context
		page types.Page
		opts []OptFn
	}
	tests := []struct {
		name    string
		args    args
		want    []*types.GlobalSettings
		wantNum int64
		wantErr bool
	}{
		{
			name: "filter by setting name",
			args: args{
				ctx: ctx,
				page: types.Page{
					Offset: 0,
					Limit:  2,
					Sort:   "",
				},
				opts: []OptFn{
					WithSettingName("test1"),
				},
			},
			want: []*types.GlobalSettings{
				{
					SettingName: "test1",
					Value:       "value1",
				},
			},
			wantNum: 1,
			wantErr: false,
		},
		{
			name: "all settings",
			args: args{
				ctx: ctx,
				page: types.Page{
					Offset: 0,
					Limit:  10,
				},
				opts: nil,
			},
			wantNum: 3,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)

			got, gotNum, err := h.List(tt.args.ctx, tt.args.page, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantNum > 0 && tt.wantNum != gotNum {
				t.Errorf("List() num = %d, wantNum %d", gotNum, tt.wantNum)
				return
			}
			t.Logf("List() num = %d", gotNum)
			for _, v := range got {
				t.Logf("List() got = %v", v)
			}
		})
	}
}

// Test_handler_Count tests the Count method of the handler.
func Test_handler_Count(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "test")
	prepareData(t, ctx)

	type args struct {
		ctx  context.Context
		opts []OptFn
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "filter by setting name",
			args: args{
				ctx: ctx,
				opts: []OptFn{
					WithSettingName("test1"),
				},
			},
			want:    1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)

			got, err := h.Count(tt.args.ctx, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Count() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want > 0 && got != tt.want {
				t.Errorf("Count() got = %d, want %d", got, tt.want)
				return
			}
			t.Logf("Count() got = %d", got)
		})
	}
}

// Test_handler_Exist tests the Exist method of the handler.
func Test_handler_Exist(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "test")
	prepareData(t, ctx)

	type args struct {
		ctx  context.Context
		name string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "exist",
			args: args{
				ctx:  ctx,
				name: "test1",
			},
			want: true,
		},
		{
			name: "nonexist",
			args: args{
				ctx:  ctx,
				name: "test9",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)

			got, err := h.Exist(tt.args.ctx, tt.args.name)
			if err != nil {
				t.Errorf("Exist() error = %v", err)
				return
			}

			if tt.want != got {
				t.Errorf("Exist() got = %v, want %v", got, tt.want)
				return
			}
			t.Logf("Exist() got = %v", got)
		})
	}
}

func Test_handler_DeleteMany(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "test")
	prepareData(t, ctx)

	type args struct {
		ctx   context.Context
		names []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx: ctx,
				names: []string{
					"test1",
					"test2",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)

			err := h.Delete(tt.args.ctx, tt.args.names...)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteMany() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
