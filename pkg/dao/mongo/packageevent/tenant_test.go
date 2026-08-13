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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestTableName_returnsTenantCollection(t *testing.T) {
	require.Equal(t, "packageevent_tenant-a", TableName("tenant-a"))
}

func TestPackageEventConversion_preservesTenantID(t *testing.T) {
	domainEvent := &types.PackageEvent{TenantID: "tenant-a"}

	storedEvent := convertPackageEventFromTypes(domainEvent)
	convertedEvent := convertPackageEventToTypes(storedEvent)

	require.Equal(t, "tenant-a", storedEvent.TenantID)
	require.Equal(t, "tenant-a", convertedEvent.TenantID)
}

func TestDAOIndexes_supportDefaultListSort(t *testing.T) {
	indexes := (&dao{}).GetIndexes()

	require.Len(t, indexes, 1)
	require.Equal(t, bson.D{
		{Key: FieldKeyOperateTime, Value: -1},
		{Key: FieldKeyEventID, Value: -1},
	}, indexes[0].Keys)
	require.Equal(t, bson.D{
		{Key: base.FieldKeyIsDeleted, Value: false},
	}, indexes[0].Options.PartialFilterExpression)
}

func TestHandler_Count_rejectsMissingTenant(t *testing.T) {
	h := new(Handler)

	_, err := h.Count(contextx.New(t.Context()))

	require.Error(t, err)
}

func TestHandler_CreateMany_rejectsMismatchedTenant(t *testing.T) {
	h := new(Handler)
	nCtx := contextx.New(t.Context(), contextx.WithTenantID("tenant-a"))

	err := h.CreateMany(nCtx, &types.PackageEvent{TenantID: "tenant-b"})

	require.Error(t, err)
}
