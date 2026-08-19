//go:build integration

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
	platform := platfmt.NewPlatform(criteria.OSLinux, criteria.CPUArchAmd64)
	uploadID := uuid.NewString()
	return &types.PackageDeployment{
		Token: token,
		Info: &types.PackageDeploymentInfo{
			ImportPluginPkgOptions: types.PackageImportPluginPkgOptions{
				FileSourceType: types.FileSourceTypeDownload,
				FileSource:     "https://example.com/packages/gse_plugin-1.0.0.tgz",
				FileName:       "gse_plugin-1.0.0.tgz",
				MD5:            "4d96767dd3f2c09e19101d0da9ce251f",
			},
			Upload: types.PackageDeploymentUploadInfo{
				UploadID:  uploadID,
				Name:      "gse_plugin",
				Version:   "1.0.0",
				Platforms: []platfmt.Platform{platform},
			},
			Release: []types.Release{
				{
					Name:         "gse_plugin",
					Generation:   types.Generation2,
					Type:         types.ReleaseTypePlugin,
					Version:      "1.0.0",
					Platform:     platform,
					Labels:       []string{"plugin"},
					FileName:     "gse_plugin-1.0.0.tgz",
					MD5:          "4d96767dd3f2c09e19101d0da9ce251f",
					Enabled:      true,
					Operator:     "admin",
					AdditionInfo: map[string]any{"description": "test plugin"},
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
	agentPlatform := platfmt.NewPlatform(criteria.OSWindows, criteria.CPUArchAmd64)
	agentDeployment.Info.Upload.UploadID = uuid.NewString()
	agentDeployment.Info.Upload.Name = "gse_agent"
	agentDeployment.Info.Upload.Version = "2.0.0"
	agentDeployment.Info.Upload.Platforms = []platfmt.Platform{agentPlatform}
	agentDeployment.Info.Release[0].Name = "gse_agent"
	agentDeployment.Info.Release[0].Version = "2.0.0"
	agentDeployment.Info.Release[0].Platform = agentPlatform
	agentDeployment.Info.ImportPluginPkgOptions.FileSource = "https://example.com/packages/gse_agent-2.0.0.tgz"

	for _, deployment := range []*types.PackageDeployment{pluginDeployment, agentDeployment} {
		if err := h.CreatePackageDeployment(nCtx, deployment); err != nil {
			t.Fatalf("CreatePackageDeployment() error = %v", err)
		}
	}

	tests := []struct {
		name         string
		opts         []OptFn
		wantTotal    int64
		wantToken    string
		wantUploadID string
	}{
		{
			name:      "all package deployments",
			wantTotal: 2,
		},
		{
			name:      "by token",
			opts:      []OptFn{WithToken(pluginDeployment.Token)},
			wantTotal: 1,
			wantToken: pluginDeployment.Token,
		},
		{
			name:         "by upload ID",
			opts:         []OptFn{WithUploadID(agentDeployment.Info.Upload.UploadID)},
			wantTotal:    1,
			wantUploadID: agentDeployment.Info.Upload.UploadID,
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
			if tt.wantToken != "" {
				if got[0].Token != tt.wantToken {
					t.Fatalf("ListPackageDeployment() token = %s, want %s", got[0].Token, tt.wantToken)
				}
			}
			if tt.wantUploadID != "" {
				if got[0].Info == nil || got[0].Info.Upload.UploadID != tt.wantUploadID {
					t.Fatalf("ListPackageDeployment() upload id = %v, want %s", got[0].Info, tt.wantUploadID)
				}
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

	platform := platfmt.NewPlatform(criteria.OSLinux, criteria.CPUArchArm64)
	want := &types.PackageDeploymentInfo{
		ImportPluginPkgOptions: types.PackageImportPluginPkgOptions{
			FileSourceType: types.FileSourceTypeDownload,
			FileSource:     "https://example.com/packages/gse_plugin-1.1.0.tgz",
			FileName:       "gse_plugin-1.1.0.tgz",
			MD5:            "482f46fa4999054774412e3a8f9dfca0",
		},
		Upload: types.PackageDeploymentUploadInfo{
			UploadID:  uuid.NewString(),
			Name:      "gse_plugin",
			Version:   "1.1.0",
			Platforms: []platfmt.Platform{platform},
		},
		Release: []types.Release{
			{
				Name:       "gse_plugin",
				Generation: types.Generation2,
				Type:       types.ReleaseTypePlugin,
				Version:    "1.1.0",
				Platform:   platform,
				FileName:   "gse_plugin-1.1.0.tgz",
				MD5:        "482f46fa4999054774412e3a8f9dfca0",
				Enabled:    true,
				Operator:   "admin",
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

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("PackageDeploymentInfo = %#v, want %#v", got, want)
	}
}
