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

package topo

import (
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	hostdao "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/host"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/testsuite/support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testStorage(t *testing.T) (*Storage, hostdao.IHandler) {
	t.Helper()

	client, db := support.RequireMongoDatabase(t)
	s, err := NewStorage(client, db.Name())
	require.NoError(t, err)

	return s, hostdao.New(db)
}

func testContext(t *testing.T, tenantID string) contextx.IContext {
	t.Helper()

	return contextx.New(t.Context(), contextx.WithTenantID(tenantID))
}

func directAreaHost(tenantID string, hostID int64, ipv4, ipv6 string) *types.Host {
	return &types.Host{
		HostID:   hostID,
		TenantID: tenantID,
		Static: &types.HostStatic{
			NetworkAreaID: types.DefaultNetworkAreaID,
			InnerIPList:   []string{ipv4},
			InnerIPV6List: []string{ipv6},
		},
	}
}

func TestStorage_GetDirectNetworkAreaHostByAnyInnerIPStaysTenantScoped(t *testing.T) {
	s, hostHandler := testStorage(t)
	tenantACtx := testContext(t, "tenant_a")
	tenantBCtx := testContext(t, "tenant_b")

	err := hostHandler.CreateMany(tenantACtx,
		directAreaHost("tenant_a", 99001, "10.0.0.1", "fd00::1"),
		directAreaHost("tenant_a", 99002, "10.0.0.2", "fd00::2"),
	)
	require.NoError(t, err)
	err = hostHandler.CreateMany(tenantBCtx,
		directAreaHost("tenant_b", 99001, "10.0.0.1", "fd00::1"),
		directAreaHost("tenant_b", 99003, "10.0.0.3", "fd00::3"),
	)
	require.NoError(t, err)

	tenantAIPv4Host, err := s.GetDirectNetworkAreaHostByAnyInnerIP(tenantACtx, "10.0.0.1", "")
	require.NoError(t, err)
	assert.Equal(t, "tenant_a", tenantAIPv4Host.TenantID)
	assert.Equal(t, int64(99001), tenantAIPv4Host.HostID)

	tenantBIPv4Host, err := s.GetDirectNetworkAreaHostByAnyInnerIP(tenantBCtx, "10.0.0.1", "")
	require.NoError(t, err)
	assert.Equal(t, "tenant_b", tenantBIPv4Host.TenantID)
	assert.Equal(t, int64(99001), tenantBIPv4Host.HostID)

	tenantAIPv6Host, err := s.GetDirectNetworkAreaHostByAnyInnerIP(tenantACtx, "", "fd00::2")
	require.NoError(t, err)
	assert.Equal(t, "tenant_a", tenantAIPv6Host.TenantID)
	assert.Equal(t, int64(99002), tenantAIPv6Host.HostID)

	tenantBIPv6Host, err := s.GetDirectNetworkAreaHostByAnyInnerIP(tenantBCtx, "", "fd00::3")
	require.NoError(t, err)
	assert.Equal(t, "tenant_b", tenantBIPv6Host.TenantID)
	assert.Equal(t, int64(99003), tenantBIPv6Host.HostID)
}
