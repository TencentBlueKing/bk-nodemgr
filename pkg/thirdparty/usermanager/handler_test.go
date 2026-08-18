/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package usermanager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	restheader "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
	tenantpkg "github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
)

type testTraceService struct{}

func (testTraceService) TracerProvider() trace.TracerProvider {
	return noop.NewTracerProvider()
}

func (testTraceService) ServiceName() string {
	return "usermanager_test"
}

func (testTraceService) Shutdown(_ context.Context) error {
	return nil
}

func (testTraceService) TracerPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
}

func TestHandlerMultiTenantListALLTenantsUsesSystemTenant(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		require.Equal(t, http.MethodGet, req.Method)
		require.Equal(t, "/api/v3/open/tenants/", req.URL.Path)
		assert.Equal(t, tenantpkg.SystemTenantID, req.Header.Get(restheader.BKTenantIDKey))

		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(`{"data":[{"id":"tenant-a","name":"Tenant A","status":"enabled"}]}`))
	}))
	defer server.Close()

	h := newHTTPTestHandler(t, server.URL)
	tenants, err := h.ListALLTenants(contextx.New(context.Background(), contextx.WithTenantID("caller-tenant")))

	require.NoError(t, err)
	require.Len(t, tenants, 1)
	assert.Equal(t, "tenant-a", tenants[0].ID)
	assert.True(t, tenants[0].Enabled)
}

func TestHandlerMultiTenantListALLTenantsReturnsAPIGWError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(`{"code":1640301,"message":"App has no permission to the resource [reason=\"no permission, bk_app_code=bk-nodemgr\"]","data":null,"result":false,"code_name":"APP_NO_PERMISSION"}`))
	}))
	defer server.Close()

	h := newHTTPTestHandler(t, server.URL)
	_, err := h.ListALLTenants(contextx.New(context.Background(), contextx.WithTenantID("caller-tenant")))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list tenant: result(false), code(1640301)")
	assert.Contains(t, err.Error(), "App has no permission")
}

func newHTTPTestHandler(t *testing.T, endpoint string) *HandlerMultiTenant {
	t.Helper()

	httpClient, err := restclient.NewHTTPClient(&ssl.TLSConfig{InsecureSkipVerify: true})
	require.NoError(t, err)

	h, err := NewHandlerMultiTenant(&restclient.Capability{
		Name:                 "usermanager-test",
		HTTPClient:           httpClient,
		Discover:             restdiscovery.NewDiscovery("usermanager-test", []string{endpoint}),
		ToleranceLatencyTime: restclient.ToleranceLatencyTimeDefault,
		MetricOpts:           restclient.MetricOption{},
		TraceSvc:             testTraceService{},
	}, &Config{})
	require.NoError(t, err)

	return h
}
