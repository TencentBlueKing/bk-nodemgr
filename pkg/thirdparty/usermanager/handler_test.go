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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
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
		assert.Equal(t, tenant.SystemTenantID, req.Header.Get(restheader.BKTenantIDKey))

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

func TestHandlerMultiTenantGetBKUsernameByLoginName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		require.Equal(t, http.MethodGet, req.Method)
		require.Equal(t, "/api/v3/open/tenant/virtual-users/-/lookup/", req.URL.Path)
		assert.Equal(t, "tenant-a", req.Header.Get(restheader.BKTenantIDKey))
		assert.Equal(t, "bk-nodemgr", req.URL.Query().Get("lookups"))
		assert.Equal(t, lookupFieldLoginName, req.URL.Query().Get("lookup_field"))

		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(`{"data":[{"bk_username":"bk-nodemgr@tenant-a","login_name":"bk-nodemgr"}]}`))
	}))
	defer server.Close()

	h := newHTTPTestHandler(t, server.URL)
	bkUsername, err := h.GetBKUsernameByLoginName(
		contextx.New(context.Background(), contextx.WithTenantID("tenant-a")), "bk-nodemgr")

	require.NoError(t, err)
	assert.Equal(t, "bk-nodemgr@tenant-a", bkUsername)
}

func TestHandlerMultiTenantGetBKUsernameByLoginNameValidation(t *testing.T) {
	h := &HandlerMultiTenant{}

	_, err := h.GetBKUsernameByLoginName(nil, "bk-nodemgr")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context is nil")

	_, err = h.GetBKUsernameByLoginName(contextx.New(context.Background()), "bk-nodemgr")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tenant-id not found")

	_, err = h.GetBKUsernameByLoginName(contextx.New(context.Background(), contextx.WithTenantID("tenant-a")), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "login name is empty")
}

func TestHandlerSingleGetBKUsernameByLoginName(t *testing.T) {
	h := &HandlerSingle{}

	bkUsername, err := h.GetBKUsernameByLoginName(contextx.New(context.Background()), "bk-nodemgr")
	require.NoError(t, err)
	assert.Equal(t, "bk-nodemgr", bkUsername)

	_, err = h.GetBKUsernameByLoginName(contextx.New(context.Background()), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "login name is empty")
}

func TestHandlerMultiTenantGetBKUsernameByLoginNameResponseHandling(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    string
		wantErr string
	}{
		{
			name: "found exact login name",
			body: `{"data":[{"bk_username":"other@tenant-a","login_name":"other"},` +
				`{"bk_username":"bk-nodemgr@tenant-a","login_name":"bk-nodemgr"}]}`,
			want: "bk-nodemgr@tenant-a",
		},
		{
			name:    "not found",
			body:    `{"data":[]}`,
			wantErr: "virtual user not found",
		},
		{
			name:    "empty bk username",
			body:    `{"data":[{"login_name":"bk-nodemgr"}]}`,
			wantErr: "virtual user bk username is empty",
		},
		{
			name: "duplicate login name",
			body: `{"data":[{"bk_username":"bk-nodemgr@tenant-a","login_name":"bk-nodemgr"},` +
				`{"bk_username":"bk-nodemgr-2@tenant-a","login_name":"bk-nodemgr"}]}`,
			wantErr: "multiple virtual users found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.Header().Set("Content-Type", "application/json")
				_, _ = rw.Write([]byte(tt.body))
			}))
			defer server.Close()

			h := newHTTPTestHandler(t, server.URL)
			got, err := h.GetBKUsernameByLoginName(
				contextx.New(context.Background(), contextx.WithTenantID("tenant-a")), "bk-nodemgr")
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
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
