/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package backendadmin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

func TestHandlerInitTenantUsesHook(t *testing.T) {
	h := &Handler{cli: &cli{}}
	h.cli.initTenantFn = func(nCtx contextx.IContext, tenantID string) error {
		assert.Equal(t, "auth-tenant", nCtx.TenantID())
		assert.Equal(t, "target-tenant", tenantID)
		return nil
	}

	err := h.InitTenant(
		contextx.New(context.Background(), contextx.WithTenantID("auth-tenant")),
		"target-tenant",
	)

	require.NoError(t, err)
}

func TestHandlerInitTenantViaHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		require.Equal(t, http.MethodPost, req.Method)
		require.Equal(t, "/admin/tenant/init", req.URL.Path)
		assert.Equal(t, "auth-tenant", req.Header.Get("X-Bk-Tenant-Id"))
		require.NotEmpty(t, req.Header.Get("X-Bknodemgr-Authorization"))

		var body struct {
			TenantID string `json:"tenant_id"`
		}
		require.NoError(t, json.NewDecoder(req.Body).Decode(&body))
		assert.Equal(t, "target-tenant", body.TenantID)

		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(`{"code":0,"message":"OK","request_id":"rid","data":{}}`))
	}))
	defer server.Close()

	h := newHTTPTestHandler(t, server.URL)
	err := h.InitTenant(
		contextx.New(context.Background(), contextx.WithTenantID("auth-tenant"), contextx.WithLoginName("admin")),
		"target-tenant",
	)

	require.NoError(t, err)
}

func TestHandlerInitTenantReturnsResponseError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(`{"code":42,"message":"tenant exists","request_id":"rid-42"}`))
	}))
	defer server.Close()

	h := newHTTPTestHandler(t, server.URL)
	err := h.InitTenant(
		contextx.New(context.Background(), contextx.WithTenantID("auth-tenant"), contextx.WithLoginName("admin")),
		"target-tenant",
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "code(42)")
	assert.Contains(t, err.Error(), "message(tenant exists)")
	assert.Contains(t, err.Error(), "request-id(rid-42)")
}
