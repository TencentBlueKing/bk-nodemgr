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

package networkarea

import (
	"fmt"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/testsuite/support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const expectedTableNamePrefix = "networkarea"

func testHandler(t *testing.T) (*mongo.Database, IHandler) {
	t.Helper()

	_, db := support.RequireMongoDatabase(t)
	return db, New(db)
}

func testContext(t *testing.T, tenantID string) contextx.IContext {
	t.Helper()

	return contextx.New(t.Context(), contextx.WithTenantID(tenantID))
}

func tenantCollectionName(tenantID string) string {
	return fmt.Sprintf("%s_%s", expectedTableNamePrefix, tenantID)
}

func networkAreaFixture(tenantID string, networkAreaID int64, name string) *types.NetworkArea {
	return &types.NetworkArea{
		TenantID: tenantID,
		ID:       networkAreaID,
		Name:     name,
	}
}

func countTenantDocuments(t *testing.T, db *mongo.Database, tenantID string) int64 {
	t.Helper()

	num, err := db.Collection(tenantCollectionName(tenantID)).CountDocuments(t.Context(), bson.D{})
	require.NoError(t, err)

	return num
}

func countAliveTenantDocuments(t *testing.T, db *mongo.Database, tenantID string) int64 {
	t.Helper()

	num, err := db.Collection(tenantCollectionName(tenantID)).CountDocuments(t.Context(), bson.D{
		{Key: base.FieldKeyIsDeleted, Value: false},
	})
	require.NoError(t, err)

	return num
}

func networkAreaIDs(networkAreas []*types.NetworkArea) []int64 {
	ids := make([]int64, len(networkAreas))
	for idx, networkArea := range networkAreas {
		ids[idx] = networkArea.ID
	}

	return ids
}

func TestHandler_RoutesTenantCollectionsAndPreservesReadIsolation(t *testing.T) {
	db, h := testHandler(t)
	tenantACtx := testContext(t, "tenant_a")
	tenantBCtx := testContext(t, "tenant_b")

	err := h.UpsertMany(tenantACtx,
		networkAreaFixture("tenant_a", types.DefaultNetworkAreaID, "tenant-a-default"),
		networkAreaFixture("tenant_a", 90001, "tenant-a-shared-id"),
		networkAreaFixture("tenant_a", 90002, "tenant-a-filter-name"),
	)
	require.NoError(t, err)
	err = h.UpsertMany(tenantBCtx,
		networkAreaFixture("tenant_b", types.DefaultNetworkAreaID, "tenant-b-default"),
		networkAreaFixture("tenant_b", 90001, "tenant-b-shared-id"),
	)
	require.NoError(t, err)

	assert.NotEqual(t, tenantCollectionName("tenant_a"), tenantCollectionName("tenant_b"))
	assert.Equal(t, int64(3), countTenantDocuments(t, db, "tenant_a"))
	assert.Equal(t, int64(2), countTenantDocuments(t, db, "tenant_b"))

	tenantARecords, tenantATotal, err := h.List(tenantACtx, types.Page{})
	require.NoError(t, err)
	require.Equal(t, int64(3), tenantATotal)
	require.Len(t, tenantARecords, 3)
	assert.Equal(t, []int64{90002, 90001, types.DefaultNetworkAreaID}, networkAreaIDs(tenantARecords))

	tenantBRecords, tenantBTotal, err := h.List(tenantBCtx, types.Page{})
	require.NoError(t, err)
	require.Equal(t, int64(2), tenantBTotal)
	require.Len(t, tenantBRecords, 2)
	assert.Equal(t, []int64{90001, types.DefaultNetworkAreaID}, networkAreaIDs(tenantBRecords))

	tenantARecord, err := h.Get(tenantACtx, 90001)
	require.NoError(t, err)
	assert.Equal(t, "tenant_a", tenantARecord.TenantID)
	assert.Equal(t, "tenant-a-shared-id", tenantARecord.Name)

	tenantBRecord, err := h.Get(tenantBCtx, 90001)
	require.NoError(t, err)
	assert.Equal(t, "tenant_b", tenantBRecord.TenantID)
	assert.Equal(t, "tenant-b-shared-id", tenantBRecord.Name)

	tenantADefault, err := h.Get(tenantACtx, types.DefaultNetworkAreaID)
	require.NoError(t, err)
	assert.Equal(t, "tenant-a-default", tenantADefault.Name)

	tenantBDefault, err := h.Get(tenantBCtx, types.DefaultNetworkAreaID)
	require.NoError(t, err)
	assert.Equal(t, "tenant-b-default", tenantBDefault.Name)
}

func TestHandler_ListAndCountFiltersStayTenantScoped(t *testing.T) {
	_, h := testHandler(t)
	tenantACtx := testContext(t, "tenant_a")
	tenantBCtx := testContext(t, "tenant_b")

	err := h.UpsertMany(tenantACtx,
		networkAreaFixture("tenant_a", 91001, "same-name"),
		networkAreaFixture("tenant_a", 91002, "same-name"),
		networkAreaFixture("tenant_a", 91003, "other-name"),
	)
	require.NoError(t, err)
	err = h.UpsertMany(tenantBCtx,
		networkAreaFixture("tenant_b", 91001, "same-name"),
	)
	require.NoError(t, err)

	total, err := h.Count(tenantACtx, WithFuzzyNetworkAreaName("same"))
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)

	records, total, err := h.List(tenantACtx, types.Page{Offset: 1, Limit: 1}, WithFuzzyNetworkAreaName("same"))
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, records, 1)
	assert.Equal(t, int64(91001), records[0].ID)

	tenantBTotal, err := h.Count(tenantBCtx, WithNetworkAreaID(91001, 91002))
	require.NoError(t, err)
	assert.Equal(t, int64(1), tenantBTotal)
}

func TestHandler_MutationsStayTenantScoped(t *testing.T) {
	db, h := testHandler(t)
	tenantACtx := testContext(t, "tenant_a")
	tenantBCtx := testContext(t, "tenant_b")

	err := h.UpsertMany(tenantACtx,
		networkAreaFixture("tenant_a", types.DefaultNetworkAreaID, "tenant-a-default"),
		networkAreaFixture("tenant_a", 92001, "tenant-a-original"),
	)
	require.NoError(t, err)
	err = h.UpsertMany(tenantBCtx,
		networkAreaFixture("tenant_b", types.DefaultNetworkAreaID, "tenant-b-default"),
		networkAreaFixture("tenant_b", 92001, "tenant-b-original"),
	)
	require.NoError(t, err)

	err = h.UpdateMany(tenantACtx, networkAreaFixture("tenant_a", 92001, "tenant-a-updated"))
	require.NoError(t, err)

	tenantARecord, err := h.Get(tenantACtx, 92001)
	require.NoError(t, err)
	assert.Equal(t, "tenant-a-updated", tenantARecord.Name)

	tenantBRecord, err := h.Get(tenantBCtx, 92001)
	require.NoError(t, err)
	assert.Equal(t, "tenant-b-original", tenantBRecord.Name)

	err = h.DeleteMany(tenantACtx, types.DefaultNetworkAreaID, 92001)
	require.NoError(t, err)
	assert.Equal(t, int64(0), countAliveTenantDocuments(t, db, "tenant_a"))
	assert.Equal(t, int64(2), countAliveTenantDocuments(t, db, "tenant_b"))

	_, err = h.Get(tenantACtx, types.DefaultNetworkAreaID)
	require.Error(t, err)

	tenantBDefault, err := h.Get(tenantBCtx, types.DefaultNetworkAreaID)
	require.NoError(t, err)
	assert.Equal(t, "tenant-b-default", tenantBDefault.Name)
}

func TestHandler_RejectsCrossTenantMutations(t *testing.T) {
	_, h := testHandler(t)
	tenantACtx := testContext(t, "tenant_a")

	err := h.UpsertMany(tenantACtx, networkAreaFixture("tenant_b", 93001, "tenant-b-area"))
	require.Error(t, err)

	err = h.UpdateMany(tenantACtx, networkAreaFixture("tenant_b", 93001, "tenant-b-area"))
	require.Error(t, err)
}
