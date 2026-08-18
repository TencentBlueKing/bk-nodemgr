//go:build integration

/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package packagedeployment

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/testsuite/support"
)

func testClient(t *testing.T) IHandler {
	t.Helper()

	_, db := support.RequireMongoDatabase(t)
	return New(db)
}

func testContext() contextx.IContext {
	return contextx.New(context.Background())
}

func testPackageDeployment() *types.PackageDeployment {
	token := uuid.NewString()
	return &types.PackageDeployment{
		Token: token,
		Info: &types.PackageDeploymentInfo{
			UploadID: uuid.NewString(),
			ImportPluginPkgOptions: types.PackageImportPluginPkgOptions{
				FileSourceType: types.FileSourceTypeDownload,
				FileSource:     "https://example.com/packages/gse_plugin-1.0.0.tgz",
				MD5:            "4d96767dd3f2c09e19101d0da9ce251f",
				PluginName:     "gse_plugin",
				PluginPkgName:  "gse_plugin.tgz",
				Version:        "1.0.0",
				Platforms: []platfmt.Platform{
					platfmt.NewPlatform(criteria.OSLinux, criteria.CPUArchAmd64),
				},
			},
		},
	}
}

func TestHandler_CreateAndGetPackageDeploymentInfo(t *testing.T) {
	h := testClient(t)
	nCtx := testContext()
	want := testPackageDeployment()

	if err := h.CreatePackageDeployment(nCtx, want); err != nil {
		t.Fatalf("CreatePackageDeployment() error = %v", err)
	}

	got, err := h.GetPackageDeploymentInfo(nCtx, want.Token)
	if err != nil {
		t.Fatalf("GetPackageDeploymentInfo() error = %v", err)
	}

	assertPackageDeploymentInfoEqual(t, got, want.Info)
}

func TestHandler_ListPackageDeployment(t *testing.T) {
	h := testClient(t)
	nCtx := testContext()
	pluginDeployment := testPackageDeployment()
	agentDeployment := testPackageDeployment()
	agentDeployment.Info.UploadID = uuid.NewString()
	agentDeployment.Info.ImportPluginPkgOptions.FileSource = "https://example.com/packages/gse_agent-2.0.0.tgz"
	agentDeployment.Info.ImportPluginPkgOptions.PluginName = "gse_agent"
	agentDeployment.Info.ImportPluginPkgOptions.PluginPkgName = "gse_agent.tgz"
	agentDeployment.Info.ImportPluginPkgOptions.Version = "2.0.0"
	agentDeployment.Info.ImportPluginPkgOptions.Platforms = []platfmt.Platform{
		platfmt.NewPlatform(criteria.OSWindows, criteria.CPUArchAmd64),
	}

	for _, deployment := range []*types.PackageDeployment{pluginDeployment, agentDeployment} {
		if err := h.CreatePackageDeployment(nCtx, deployment); err != nil {
			t.Fatalf("CreatePackageDeployment() error = %v", err)
		}
	}

	tests := []struct {
		name      string
		opts      []OptFn
		wantTotal int64
	}{
		{
			name:      "all package deployments",
			wantTotal: 2,
		},
		{
			name:      "by token",
			opts:      []OptFn{WithToken(pluginDeployment.Token)},
			wantTotal: 1,
		},
		{
			name:      "by upload ID",
			opts:      []OptFn{WithUploadID(agentDeployment.Info.UploadID)},
			wantTotal: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, total, err := h.ListPackageDeployment(nCtx, types.UnlimitedPage(), tt.opts...)
			if err != nil {
				t.Fatalf("ListPackageDeployment() error = %v", err)
			}
			if total != tt.wantTotal {
				t.Fatalf("ListPackageDeployment() total = %d, want %d", total, tt.wantTotal)
			}
			if int64(len(got)) != tt.wantTotal {
				t.Fatalf("ListPackageDeployment() len = %d, want %d", len(got), tt.wantTotal)
			}
		})
	}
}

func TestHandler_UpdatePackageDeploymentInfo(t *testing.T) {
	h := testClient(t)
	nCtx := testContext()
	deployment := testPackageDeployment()

	if err := h.CreatePackageDeployment(nCtx, deployment); err != nil {
		t.Fatalf("CreatePackageDeployment() error = %v", err)
	}

	want := &types.PackageDeploymentInfo{
		UploadID: uuid.NewString(),
		ImportPluginPkgOptions: types.PackageImportPluginPkgOptions{
			FileSourceType: types.FileSourceTypeDownload,
			FileSource:     "https://example.com/packages/gse_plugin-1.1.0.tgz",
			MD5:            "482f46fa4999054774412e3a8f9dfca0",
			PluginName:     "gse_plugin",
			PluginPkgName:  "gse_plugin.tgz",
			Version:        "1.1.0",
			Platforms: []platfmt.Platform{
				platfmt.NewPlatform(criteria.OSLinux, criteria.CPUArchArm64),
			},
		},
	}

	if err := h.UpdatePackageDeploymentInfo(nCtx, deployment.Token, want); err != nil {
		t.Fatalf("UpdatePackageDeploymentInfo() error = %v", err)
	}

	got, err := h.GetPackageDeploymentInfo(nCtx, deployment.Token)
	if err != nil {
		t.Fatalf("GetPackageDeploymentInfo() error = %v", err)
	}

	assertPackageDeploymentInfoEqual(t, got, want)
}

func assertPackageDeploymentInfoEqual(t *testing.T, got, want *types.PackageDeploymentInfo) {
	t.Helper()

	if got == nil {
		t.Fatal("got nil package deployment info")
	}
	if got.UploadID != want.UploadID {
		t.Fatalf("UploadID = %s, want %s", got.UploadID, want.UploadID)
	}
	if !reflect.DeepEqual(got.ImportPluginPkgOptions, want.ImportPluginPkgOptions) {
		t.Fatalf("ImportPluginPkgOptions = %#v, want %#v", got.ImportPluginPkgOptions, want.ImportPluginPkgOptions)
	}
}
