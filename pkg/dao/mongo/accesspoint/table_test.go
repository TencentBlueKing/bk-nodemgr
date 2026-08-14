/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package accesspoint

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestTableName(t *testing.T) {
	assert.Equal(t, "accesspoint_system", TableName("system"))
}

func TestNewDaoRoutesToTenantCollection(t *testing.T) {
	client, err := mongo.Connect(t.Context(), options.Client().ApplyURI("mongodb://127.0.0.1:27017"))
	require.NoError(t, err)

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		assert.NoError(t, client.Disconnect(ctx))
	})

	db := client.Database("accesspoint_route_test")
	tenantADao := newDao("tenant_a", db)
	tenantBDao := newDao("tenant_b", db)

	assert.Equal(t, "accesspoint_tenant_a", tenantADao.GetTableName())
	assert.Equal(t, "accesspoint_tenant_a", tenantADao.GetClient().Name())
	assert.Equal(t, "accesspoint_tenant_b", tenantBDao.GetTableName())
	assert.Equal(t, "accesspoint_tenant_b", tenantBDao.GetClient().Name())
}
