/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package processconfig

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"sync"
	"testing"
)

// testClient ...
func testClient(t *testing.T) IHandler {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	nCtx := contextx.Background()
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

var once = sync.Once{}

// prepareData for all tests.
func prepareData(t *testing.T, nCtx contextx.IContext) {
	once.Do(func() {
		// pre insert.
		h := testClient(t)
		err := h.UpsertMany(nCtx,
			&types.ProcessConfig{
				ProcessID: "test-process-id-1",
				Name:      "test-process-name-1",
				Content:   "test-content-1",
				MD5:       "test-md5-1",
			},
			&types.ProcessConfig{
				ProcessID: "test-process-id-2",
				Name:      "test-process-name-2",
				Content:   "test-content-2",
				MD5:       "test-md5-2",
			},
			&types.ProcessConfig{
				ProcessID: "test-process-id-3",
				Name:      "test-process-name-3",
				Content:   "test-content-3",
				MD5:       "test-md5-3",
			},
		)
		if err != nil {
			t.Errorf("prepareData() error = %v", err)
		}
	})
}

// Test_handler_Create ...
func Test_handler_Create(t *testing.T) {
	nCtx := contextx.From(contextx.Background(), contextx.WithTenantID("test-tenant-id"))
	prepareData(t, nCtx)

	type args struct {
		nCtx   contextx.IContext
		config *types.ProcessConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "create process config",
			args: args{
				nCtx: nCtx,
				config: &types.ProcessConfig{
					ProcessID: "test-process-id",
					Name:      "test-process-name",
					MD5:       "test-md5",
					Content:   "test-content",
				},
			},
			wantErr: false,
		},
	}
	h := testClient(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := h.Create(tt.args.nCtx, tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}

			t.Logf("Create() success for name(%s), processID(%s)", tt.args.config.Name, tt.args.config.ProcessID)
		})
	}
}

// Test_handler_UpsertMany ...
func Test_handler_UpsertMany(t *testing.T) {
	nCtx := contextx.From(contextx.Background(), contextx.WithTenantID("test-tenant-id"))
	prepareData(t, nCtx)

	type args struct {
		nCtx   contextx.IContext
		config []*types.ProcessConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "upsert many process config",
			args: args{
				nCtx: nCtx,
				config: []*types.ProcessConfig{
					{
						ProcessID: "test-process-id-1",
						Name:      "test-process-name-1",
						Content:   "updated-test-content-1",
						MD5:       "updated-test-md5-1",
					},
					{
						ProcessID: "test-process-id-4",
						Name:      "test-process-name-4",
						Content:   "test-content-4",
						MD5:       "test-md5-4",
					},
				},
			},
			wantErr: false,
		},
	}
	h := testClient(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := h.UpsertMany(tt.args.nCtx, tt.args.config...); (err != nil) != tt.wantErr {
				t.Errorf("UpsertMany() error = %v, wantErr %v", err, tt.wantErr)
			}

			t.Logf("UpsertMany() success")
		})
	}
}

// Test_handler_DeleteMany ...
func Test_handler_DeleteMany(t *testing.T) {
	nCtx := contextx.From(contextx.Background(), contextx.WithTenantID("test-tenant-id"))
	prepareData(t, nCtx)

	type args struct {
		nCtx contextx.IContext
		opts []base.OptFn
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "delete many process config",
			args: args{
				nCtx: nCtx,
				opts: []base.OptFn{
					WithName("test-process-name-2", "test-process-name-3"),
					WithProcessID("test-process-id-2", "test-process-id-3"),
				},
			},
			wantErr: false,
		},
	}
	h := testClient(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := h.DeleteMany(tt.args.nCtx, tt.args.opts...); (err != nil) != tt.wantErr {
				t.Errorf("DeleteMany() error = %v, wantErr %v", err, tt.wantErr)
			}

			t.Logf("DeleteMany() success")
		})
	}
}

// Test_handler_Get ...
func Test_handler_Get(t *testing.T) {
	nCtx := contextx.From(contextx.Background(), contextx.WithTenantID("test-tenant-id"))
	prepareData(t, nCtx)

	type args struct {
		nCtx      contextx.IContext
		name      string
		processID string
	}
	tests := []struct {
		name    string
		args    args
		want    *types.ProcessConfig
		wantErr bool
	}{
		{
			name: "get process config",
			args: args{
				nCtx:      nCtx,
				name:      "test-process-name-1",
				processID: "test-process-id-1",
			},
			want: &types.ProcessConfig{
				ProcessID: "test-process-id-1",
				Name:      "test-process-name-1",
			},
			wantErr: false,
		},
	}
	h := testClient(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.Get(tt.args.nCtx, tt.args.name, tt.args.processID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !compareProcessConfig(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}

			t.Logf("Get() success")
		})
	}
}

// compareProcessConfig compare two ProcessConfig.
func compareProcessConfig(a, b *types.ProcessConfig) bool {
	if a == nil || b == nil {
		return a == b
	}

	return a.Name == b.Name &&
		a.ProcessID == b.ProcessID
}

// Test_handler_List ...
func Test_handler_List(t *testing.T) {
	nCtx := contextx.From(contextx.Background(), contextx.WithTenantID("test-tenant-id"))
	prepareData(t, nCtx)

	type args struct {
		nCtx contextx.IContext
		opts []base.OptFn
	}
	tests := []struct {
		name    string
		args    args
		wantLen int64
		wantErr bool
	}{
		{
			name: "list process config",
			args: args{
				nCtx: nCtx,
				opts: []base.OptFn{},
			},
			wantLen: 3,
			wantErr: false,
		},
	}
	h := testClient(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, cnt, err := h.List(tt.args.nCtx, types.UnlimitedPage(), tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if cnt != tt.wantLen {
				t.Errorf("List() got len = %v, want %v", len(got), tt.wantLen)
			}

			t.Logf("List() success")
		})
	}
}
