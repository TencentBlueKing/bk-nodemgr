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

package networkunit

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

const expectedTableNamePrefix = "networkunit"

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

func networkUnitFixture(tenantID string, networkAreaID int64, name string) *types.NetworkUnit {
	return &types.NetworkUnit{
		TenantID:      tenantID,
		NetworkAreaID: networkAreaID,
		Name:          name,
		AccessPoints:  []int64{networkAreaID + 10},
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

func TestHandler_RoutesTenantCollectionsAndPreservesReadIsolation(t *testing.T) {
	db, h := testHandler(t)
	tenantACtx := testContext(t, "tenant_a")
	tenantBCtx := testContext(t, "tenant_b")

	tenantAUnit := networkUnitFixture("tenant_a", 100, "tenant-a-unit")
	tenantBUnit := networkUnitFixture("tenant_b", 200, "tenant-b-unit")

	tenantAID, err := h.Create(tenantACtx, tenantAUnit)
	require.NoError(t, err)
	tenantBID, err := h.Create(tenantBCtx, tenantBUnit)
	require.NoError(t, err)

	assert.NotEqual(t, tenantAID, tenantBID)
	assert.NotEqual(t, tenantCollectionName("tenant_a"), tenantCollectionName("tenant_b"))
	assert.Equal(t, int64(1), countTenantDocuments(t, db, "tenant_a"))
	assert.Equal(t, int64(1), countTenantDocuments(t, db, "tenant_b"))

	tenantARecords, tenantATotal, err := h.List(tenantACtx, types.Page{})
	require.NoError(t, err)
	require.Equal(t, int64(1), tenantATotal)
	require.Len(t, tenantARecords, 1)
	assert.Equal(t, "tenant_a", tenantARecords[0].TenantID)
	assert.Equal(t, tenantAID, tenantARecords[0].ID)

	tenantBRecords, tenantBTotal, err := h.List(tenantBCtx, types.Page{})
	require.NoError(t, err)
	require.Equal(t, int64(1), tenantBTotal)
	require.Len(t, tenantBRecords, 1)
	assert.Equal(t, "tenant_b", tenantBRecords[0].TenantID)
	assert.Equal(t, tenantBID, tenantBRecords[0].ID)

	tenantARecord, err := h.Get(tenantACtx, tenantAID)
	require.NoError(t, err)
	assert.Equal(t, "tenant_a", tenantARecord.TenantID)

	_, err = h.Get(tenantACtx, tenantBID)
	require.Error(t, err)
}

func TestHandler_MutationsAndAggregationStayTenantScoped(t *testing.T) {
	db, h := testHandler(t)
	tenantACtx := testContext(t, "tenant_a")
	tenantBCtx := testContext(t, "tenant_b")

	tenantAFirst := networkUnitFixture("tenant_a", 300, "tenant-a-first")
	tenantASecond := networkUnitFixture("tenant_a", 300, "tenant-a-second")
	tenantBUnit := networkUnitFixture("tenant_b", 300, "tenant-b-unit")

	tenantAFirstID, err := h.Create(tenantACtx, tenantAFirst)
	require.NoError(t, err)
	tenantASecondID, err := h.Create(tenantACtx, tenantASecond)
	require.NoError(t, err)
	tenantBID, err := h.Create(tenantBCtx, tenantBUnit)
	require.NoError(t, err)
	assert.NotEqual(t, tenantAFirstID, tenantASecondID)
	assert.NotEqual(t, tenantAFirstID, tenantBID)
	assert.NotEqual(t, tenantASecondID, tenantBID)

	err = h.UpdateMany(tenantACtx, types.NetworkUnitUpdateFields{Name: true}, &types.NetworkUnit{
		TenantID: "tenant_a",
		ID:       tenantAFirstID,
		Name:     "tenant-a-updated",
	})
	require.NoError(t, err)

	tenantAUpdated, err := h.Get(tenantACtx, tenantAFirstID)
	require.NoError(t, err)
	assert.Equal(t, "tenant-a-updated", tenantAUpdated.Name)

	tenantBRecord, err := h.Get(tenantBCtx, tenantBID)
	require.NoError(t, err)
	assert.Equal(t, "tenant_b", tenantBRecord.TenantID)
	assert.Equal(t, "tenant-b-unit", tenantBRecord.Name)

	tenantADistribution, err := h.GetNetworkUnitDistributionByNetworkAreaID(tenantACtx)
	require.NoError(t, err)
	assert.Equal(t, map[int64]int64{300: 2}, tenantADistribution)

	tenantBDistribution, err := h.GetNetworkUnitDistributionByNetworkAreaID(tenantBCtx)
	require.NoError(t, err)
	assert.Equal(t, map[int64]int64{300: 1}, tenantBDistribution)

	err = h.DeleteMany(tenantACtx, tenantASecondID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), countAliveTenantDocuments(t, db, "tenant_a"))
	assert.Equal(t, int64(1), countAliveTenantDocuments(t, db, "tenant_b"))

	_, err = h.Get(tenantACtx, tenantASecondID)
	require.Error(t, err)

	tenantBRecord, err = h.Get(tenantBCtx, tenantBID)
	require.NoError(t, err)
	assert.Equal(t, "tenant_b", tenantBRecord.TenantID)
}

func TestHandler_CreateRejectsCrossTenantData(t *testing.T) {
	_, h := testHandler(t)
	tenantACtx := testContext(t, "tenant_a")

	_, err := h.Create(tenantACtx, networkUnitFixture("tenant_b", 400, "tenant-b-unit"))
	require.Error(t, err)
}
