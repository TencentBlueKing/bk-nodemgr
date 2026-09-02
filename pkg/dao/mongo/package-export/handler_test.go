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

package packageexport

import (
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/testsuite/support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

func testClient(t *testing.T) IHandler {
	t.Helper()

	_, db := support.RequireMongoDatabase(t)
	return New(db)
}

func testContext(t *testing.T, tenantID string) contextx.IContext {
	t.Helper()

	return contextx.New(t.Context(), contextx.WithTenantID(tenantID))
}

func testPackageExport(exportID string) *types.PackageExport {
	return &types.PackageExport{
		ExportID:     exportID,
		WorkflowID:   "workflow-" + exportID,
		StorageKey:   "package-export/" + exportID,
		DownloadName: exportID + ".tgz",
		MD5:          "4d96767dd3f2c09e19101d0da9ce251f",
		Size:         1024,
		Operator:     "admin",
	}
}

func TestHandler_CreateAndGet(t *testing.T) {
	h := testClient(t)
	nCtx := testContext(t, "tenant-a")
	want := testPackageExport("export-a")

	require.NoError(t, h.Create(nCtx, want))
	require.Equal(t, nCtx.TenantID(), want.TenantID)

	got, err := h.Get(nCtx, want.ExportID)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestHandler_CreateRejectsDuplicateExportID(t *testing.T) {
	h := testClient(t)
	nCtx := testContext(t, "tenant-a")
	exportData := testPackageExport("duplicate-export")

	require.NoError(t, h.Create(nCtx, exportData))

	err := h.Create(nCtx, testPackageExport(exportData.ExportID))
	require.Error(t, err)
	assert.True(t, mongo.IsDuplicateKeyError(err))
}

func TestHandler_List(t *testing.T) {
	h := testClient(t)
	fixtures := []struct {
		tenantID   string
		exportData *types.PackageExport
	}{
		{
			tenantID:   "tenant-a",
			exportData: testPackageExport("export-a"),
		},
		{
			tenantID: "tenant-a",
			exportData: &types.PackageExport{
				ExportID:     "export-b",
				WorkflowID:   "workflow-shared",
				StorageKey:   "package-export/export-b",
				DownloadName: "export-b.tgz",
				MD5:          "482f46fa4999054774412e3a8f9dfca0",
				Size:         2048,
				Operator:     "system",
			},
		},
		{
			tenantID:   "tenant-b",
			exportData: testPackageExport("export-c"),
		},
	}

	for _, fixture := range fixtures {
		nCtx := testContext(t, fixture.tenantID)
		require.NoError(t, h.Create(nCtx, fixture.exportData))
	}

	tests := []struct {
		name      string
		page      types.Page
		opts      []OptFn
		wantTotal int64
		wantIDs   []string
	}{
		{
			name:      "all package exports",
			page:      types.UnlimitedPage(),
			wantTotal: 3,
			wantIDs:   []string{"export-a", "export-b", "export-c"},
		},
		{
			name:      "by export ID",
			page:      types.UnlimitedPage(),
			opts:      []OptFn{WithExportID("export-a")},
			wantTotal: 1,
			wantIDs:   []string{"export-a"},
		},
		{
			name:      "by workflow ID",
			page:      types.UnlimitedPage(),
			opts:      []OptFn{WithWorkflowID("workflow-shared")},
			wantTotal: 1,
			wantIDs:   []string{"export-b"},
		},
		{
			name:      "by tenant ID",
			page:      types.UnlimitedPage(),
			opts:      []OptFn{WithTenantID("tenant-b")},
			wantTotal: 1,
			wantIDs:   []string{"export-c"},
		},
		{
			name:      "by operator",
			page:      types.UnlimitedPage(),
			opts:      []OptFn{WithOperator("admin")},
			wantTotal: 2,
			wantIDs:   []string{"export-a", "export-c"},
		},
		{
			name: "with sorting and pagination",
			page: types.Page{
				Offset: 1,
				Limit:  1,
				Sort:   types.WithFieldAsc(FieldKeyExportID),
			},
			wantTotal: 3,
			wantIDs:   []string{"export-b"},
		},
	}

	nCtx := testContext(t, "tenant-a")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, total, err := h.List(nCtx, tt.page, tt.opts...)
			require.NoError(t, err)
			assert.Equal(t, tt.wantTotal, total)

			gotIDs := make([]string, 0, len(got))
			for _, exportData := range got {
				gotIDs = append(gotIDs, exportData.ExportID)
			}
			assert.ElementsMatch(t, tt.wantIDs, gotIDs)
		})
	}
}

func TestHandler_Delete(t *testing.T) {
	h := testClient(t)
	nCtx := testContext(t, "tenant-a")
	exportData := testPackageExport("export-a")

	require.NoError(t, h.Create(nCtx, exportData))
	require.NoError(t, h.Delete(nCtx, exportData.ExportID))

	got, err := h.Get(nCtx, exportData.ExportID)
	require.Error(t, err)
	assert.Nil(t, got)

	items, total, err := h.List(nCtx, types.UnlimitedPage(), WithExportID(exportData.ExportID))
	require.NoError(t, err)
	assert.Zero(t, total)
	assert.Empty(t, items)
}
