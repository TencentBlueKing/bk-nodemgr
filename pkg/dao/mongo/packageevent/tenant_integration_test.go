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

package packageevent

import (
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/testsuite/support"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestHandler_routesEventsToTenantCollections(t *testing.T) {
	_, database := support.RequireMongoDatabase(t)
	h := New(database)
	tenantACtx := contextx.New(t.Context(), contextx.WithTenantID("tenant-a"))
	tenantBCtx := contextx.New(t.Context(), contextx.WithTenantID("tenant-b"))
	operateTime := time.Date(2026, time.August, 10, 12, 0, 0, 0, time.UTC)

	require.NoError(t, h.CreateMany(tenantACtx, &types.PackageEvent{
		Name:        "agent-a",
		OperateTime: operateTime,
	}))
	require.NoError(t, h.CreateMany(tenantBCtx, &types.PackageEvent{
		Name:        "agent-b",
		OperateTime: operateTime,
	}))

	tenantAEvent := findStoredPackageEvent(t, database, tenantACtx)
	tenantBEvent := findStoredPackageEvent(t, database, tenantBCtx)
	require.Equal(t, "tenant-a", tenantAEvent.Data.TenantID)
	require.Equal(t, "tenant-b", tenantBEvent.Data.TenantID)
	require.NotEqual(t, tenantAEvent.Data.EventID, tenantBEvent.Data.EventID)

	tenantAEvents, tenantATotal, err := h.List(tenantACtx, types.UnlimitedPage())
	require.NoError(t, err)
	require.EqualValues(t, 1, tenantATotal)
	require.Len(t, tenantAEvents, 1)
	require.Equal(t, "agent-a", tenantAEvents[0].Name)
	require.Equal(t, "tenant-a", tenantAEvents[0].TenantID)
}

func findStoredPackageEvent(
	t *testing.T,
	database *mongo.Database,
	nCtx contextx.IContext,
) TablePackageEvent {
	t.Helper()

	var event TablePackageEvent
	err := database.Collection(TableName(nCtx.TenantID())).FindOne(nCtx, bson.D{}).Decode(&event)
	require.NoError(t, err)

	return event
}
