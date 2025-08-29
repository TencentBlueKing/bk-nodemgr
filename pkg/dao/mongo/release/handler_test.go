/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package release

import (
	"context"
	"os"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
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

// Test_UpsertMany tests the upsert many.
func Test_UpsertMany(t *testing.T) {
	client := testClient(t)

	type args struct {
		releases []*types.Release
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				releases: []*types.Release{
					{
						Generation: types.Generation2,
						Type:       types.ReleaseTypeAgent,
						Version:    "v2.1.6-beta.1",
						Platform: platform.Platform{
							OS:   criteria.OSLinux,
							Arch: criteria.CPUArchAmd64,
						},
						Labels:      []string{"test"},
						ChangeLogEN: "test",
						ChangeLogZH: "test",
						FileName:    "v2.1.6-beta.1-linux-amd64.tgz",
						LocalDir:    "agent/",
						UpstreamDir: "agent/",
					},
					{
						Generation: types.Generation2,
						Type:       types.ReleaseTypeAgent,
						Version:    "v2.1.6-beta.1",
						Platform: platform.Platform{
							OS:   criteria.OSLinux,
							Arch: criteria.CPUArchArm64,
						},
						Labels:      []string{"test"},
						ChangeLogEN: "test",
						ChangeLogZH: "test",
						FileName:    "v2.1.6-beta.1-linux-arm64.tgz",
						LocalDir:    "agent/",
						UpstreamDir: "agent/",
					},
					{
						Generation: types.Generation2,
						Type:       types.ReleaseTypeAgent,
						Version:    "v2.1.6-beta.2",
						Platform: platform.Platform{
							OS:   criteria.OSWindows,
							Arch: criteria.CPUArchArm64,
						},
						Labels:      []string{"test"},
						ChangeLogEN: "test",
						ChangeLogZH: "test",
						FileName:    "v2.1.6-beta.2-windows-amd64.tgz",
						LocalDir:    "agent/",
						UpstreamDir: "agent/",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := client.UpsertMany(context.Background(), "", gen, tt.args.releases...); (err != nil) != tt.wantErr {
				t.Errorf("UpsertMany() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_Get tests the Get.
func Test_Get(t *testing.T) {
	client := testClient(t)

	type args struct {
		generation  types.Generation
		releaseType types.ReleaseType
		version     string
		platform    platform.Platform
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				generation:  types.Generation2,
				releaseType: types.ReleaseTypeAgent,
				version:     "v2.1.6-beta.1",
				platform: platform.Platform{
					OS:   criteria.OSLinux,
					Arch: criteria.CPUArchAmd64,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			release, err := client.Get(context.Background(), tt.args.releaseType, tt.args.generation, tt.args.platform, tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("got release: %v", release)
		})
	}
}

// Test_List tests the List.
func Test_List(t *testing.T) {
	client := testClient(t)

	type args struct {
		page types.Page
		opts []OptFn
	}
	tests := []struct {
		name      string
		args      args
		wantNum   int64
		wantTotal int64
		wantErr   bool
	}{
		{
			name: "filter by version",
			args: args{
				page: types.Page{Offset: 0, Limit: 10},
				opts: []OptFn{
					WithVersion("v2.1.6-beta.1"),
				},
			},
			wantNum:   2,
			wantTotal: 2,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			releases, total, err := client.List(context.Background(), "", gen, tt.args.page, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(releases) != int(tt.wantNum) {
				t.Errorf("List() gotNum = %v, want %v", len(releases), tt.wantNum)
				return
			}

			if total != tt.wantTotal {
				t.Errorf("List() gotTotal = %v, want %v", total, tt.wantTotal)
				return
			}

			for _, r := range releases {
				t.Logf("got release: %v", r)
			}
		})
	}
}

// Test_Delete tests the Delete.
func Test_Delete(t *testing.T) {
	client := testClient(t)

	type args struct {
		generation  types.Generation
		releaseType types.ReleaseType
		version     string
		platform    platform.Platform
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "delete by version",
			args: args{
				generation:  types.Generation2,
				releaseType: types.ReleaseTypeAgent,
				version:     "v2.1.6-beta.1",
				platform: platform.Platform{
					OS:   criteria.OSLinux,
					Arch: criteria.CPUArchAmd64,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := client.Delete(context.Background(), tt.args.releaseType, tt.args.generation, tt.args.platform, tt.args.version); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
