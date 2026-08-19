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

package accesspoint

import (
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

func testHandler(t *testing.T) (*mongo.Database, IHandler) {
	t.Helper()

	_, db := support.RequireMongoDatabase(t)
	return db, New(db)
}

func testContext(t *testing.T, tenantID string) contextx.IContext {
	t.Helper()

	return contextx.New(t.Context(), contextx.WithTenantID(tenantID))
}

func accessPointFixture(tenantID string, accessPointID int64, name string) *AccessPoint {
	return &AccessPoint{
		TenantID:        tenantID,
		AccessPointID:   accessPointID,
		AccessPointName: name,
		NetworkAreaID:   types.DefaultNetworkAreaID,
		Endpoints: &Endpoints{
			Cluster: []string{tenantID + "-cluster"},
			File:    []string{tenantID + "-file"},
			Data:    []string{tenantID + "-data"},
		},
	}
}

func typeAccessPointFixture(tenantID string, accessPointID int64, name string) *types.AccessPoint {
	return &types.AccessPoint{
		TenantID:      tenantID,
		ID:            accessPointID,
		Name:          name,
		NetworkAreaID: types.DefaultNetworkAreaID,
		Endpoints: types.Endpoints{
			Cluster: []string{tenantID + "-cluster"},
			File:    []string{tenantID + "-file"},
			Data:    []string{tenantID + "-data"},
		},
	}
}

func insertAccessPointDocuments(t *testing.T, db *mongo.Database, accessPoints ...*AccessPoint) {
	t.Helper()

	for _, accessPoint := range accessPoints {
		_, err := db.Collection(TableName(accessPoint.TenantID)).InsertOne(t.Context(), &TableAccessPoint{
			BasicInfo: base.BasicInfo{},
			Data:      accessPoint,
		})
		require.NoError(t, err)
	}
}

func accessPointIDs(accessPoints []*types.AccessPoint) []int64 {
	ids := make([]int64, len(accessPoints))
	for idx, accessPoint := range accessPoints {
		ids[idx] = accessPoint.ID
	}

	return ids
}

func countAliveDocuments(t *testing.T, db *mongo.Database, tenantID string) int64 {
	t.Helper()

	num, err := db.Collection(TableName(tenantID)).CountDocuments(t.Context(), bson.D{
		{Key: base.FieldKeyIsDeleted, Value: false},
	})
	require.NoError(t, err)

	return num
}

func TestHandler_ReadsStayTenantScopedForDefaultNetworkArea(t *testing.T) {
	db, h := testHandler(t)
	tenantACtx := testContext(t, "tenant_a")
	tenantBCtx := testContext(t, "tenant_b")

	insertAccessPointDocuments(t, db,
		accessPointFixture("tenant_a", 90001, "tenant-a-default"),
		accessPointFixture("tenant_a", 90002, "tenant-a-extra"),
		accessPointFixture("tenant_b", 90001, "tenant-b-default"),
	)

	tenantATotal, err := h.Count(tenantACtx, WithNetworkAreaID(types.DefaultNetworkAreaID))
	require.NoError(t, err)
	assert.Equal(t, int64(2), tenantATotal)

	tenantBTotal, err := h.Count(tenantBCtx, WithNetworkAreaID(types.DefaultNetworkAreaID))
	require.NoError(t, err)
	assert.Equal(t, int64(1), tenantBTotal)

	tenantARecords, total, err := h.List(tenantACtx, types.Page{}, WithNetworkAreaID(types.DefaultNetworkAreaID))
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	assert.ElementsMatch(t, []int64{90001, 90002}, accessPointIDs(tenantARecords))

	tenantARecord, err := h.Get(tenantACtx, 90001)
	require.NoError(t, err)
	assert.Equal(t, "tenant_a", tenantARecord.TenantID)
	assert.Equal(t, "tenant-a-default", tenantARecord.Name)

	tenantBRecord, err := h.Get(tenantBCtx, 90001)
	require.NoError(t, err)
	assert.Equal(t, "tenant_b", tenantBRecord.TenantID)
	assert.Equal(t, "tenant-b-default", tenantBRecord.Name)
}

func TestHandler_MutationsStayTenantScopedForDefaultNetworkArea(t *testing.T) {
	db, h := testHandler(t)
	tenantACtx := testContext(t, "tenant_a")
	tenantBCtx := testContext(t, "tenant_b")

	insertAccessPointDocuments(t, db,
		accessPointFixture("tenant_a", 91001, "tenant-a-original"),
		accessPointFixture("tenant_b", 91001, "tenant-b-original"),
	)

	err := h.UpdateMany(tenantACtx, typeAccessPointFixture("tenant_a", 91001, "tenant-a-updated"))
	require.NoError(t, err)

	tenantARecord, err := h.Get(tenantACtx, 91001)
	require.NoError(t, err)
	assert.Equal(t, "tenant-a-updated", tenantARecord.Name)

	tenantBRecord, err := h.Get(tenantBCtx, 91001)
	require.NoError(t, err)
	assert.Equal(t, "tenant-b-original", tenantBRecord.Name)

	err = h.DeleteMany(tenantACtx, 91001)
	require.NoError(t, err)
	assert.Equal(t, int64(0), countAliveDocuments(t, db, "tenant_a"))
	assert.Equal(t, int64(1), countAliveDocuments(t, db, "tenant_b"))

	_, err = h.Get(tenantACtx, 91001)
	require.Error(t, err)

	tenantBRecord, err = h.Get(tenantBCtx, 91001)
	require.NoError(t, err)
	assert.Equal(t, "tenant-b-original", tenantBRecord.Name)
}

func TestHandler_CreateRejectsCrossTenantAccessPoint(t *testing.T) {
	_, h := testHandler(t)
	tenantACtx := testContext(t, "tenant_a")

	_, err := h.Create(tenantACtx, typeAccessPointFixture("tenant_b", 0, "tenant-b-default"))
	require.Error(t, err)
}

func TestHandler_CreateRoutesToTenantCollectionAndUsesGlobalCounter(t *testing.T) {
	db, h := testHandler(t)
	tenantACtx := testContext(t, "tenant_a")
	tenantBCtx := testContext(t, "tenant_b")

	tenantAID, err := h.Create(tenantACtx, typeAccessPointFixture("tenant_a", 0, "tenant-a-created"))
	require.NoError(t, err)

	tenantBID, err := h.Create(tenantBCtx, typeAccessPointFixture("tenant_b", 0, "tenant-b-created"))
	require.NoError(t, err)
	require.NotEqual(t, tenantAID, tenantBID)

	assert.Equal(t, int64(1), countAliveDocuments(t, db, "tenant_a"))
	assert.Equal(t, int64(1), countAliveDocuments(t, db, "tenant_b"))

	tenantARecord, err := h.Get(tenantACtx, tenantAID)
	require.NoError(t, err)
	assert.Equal(t, "tenant_a", tenantARecord.TenantID)
	assert.Equal(t, "tenant-a-created", tenantARecord.Name)

	_, err = h.Get(tenantBCtx, tenantAID)
	require.Error(t, err)
}
