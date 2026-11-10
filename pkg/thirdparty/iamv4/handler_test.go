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

package iamv4

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestToResourcePreservesAttributes(t *testing.T) {
	attrs := map[string]interface{}{
		"_bk_iam_path_": "/networkarea,0/",
	}

	resource := toResource(types.IAMResource{ID: "0", Attributes: attrs})
	if resource.ID != "0" {
		t.Fatalf("toResource() ID = %q; want 0", resource.ID)
	}
	if !reflect.DeepEqual(resource.Attributes, attrs) {
		t.Fatalf("toResource() Attributes = %#v; want %#v", resource.Attributes, attrs)
	}
}

func TestFirstResourcePreservesAttributes(t *testing.T) {
	attrs := map[string]interface{}{
		"_bk_iam_path_": "/networkarea,0/",
	}

	resource := firstResource([]types.IAMResource{{ID: "0", Attributes: attrs}})
	if resource == nil {
		t.Fatal("firstResource() = nil; want resource")
	}
	if !reflect.DeepEqual(resource.Attributes, attrs) {
		t.Fatalf("firstResource() Attributes = %#v; want %#v", resource.Attributes, attrs)
	}
}

func TestFirstResourceEmpty(t *testing.T) {
	if resource := firstResource(nil); resource != nil {
		t.Fatalf("firstResource(nil) = %#v; want nil", resource)
	}
}

type testTraceService struct{}

func (testTraceService) TracerProvider() trace.TracerProvider {
	return noop.NewTracerProvider()
}

func (testTraceService) ServiceName() string {
	return "iamv4_test"
}

func (testTraceService) Shutdown(_ context.Context) error {
	return nil
}

func (testTraceService) TracerPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
}

func newTestContext(tenantID string) contextx.IContext {
	return contextx.New(context.Background(), contextx.WithTenantID(tenantID))
}

func newTestIAMHandler(t *testing.T, endpoint string) *Handler {
	t.Helper()

	httpClient, err := restclient.NewHTTPClient(&ssl.TLSConfig{InsecureSkipVerify: true})
	if err != nil {
		t.Fatalf("failed to create HTTP client: %v", err)
	}

	capability := &restclient.Capability{
		Name:                 "iam-v4-test",
		HTTPClient:           httpClient,
		Discover:             restdiscovery.NewDiscovery("iam-v4", []string{endpoint}),
		ToleranceLatencyTime: restclient.ToleranceLatencyTimeDefault,
		MetricOpts:           restclient.MetricOption{},
		TraceSvc:             testTraceService{},
	}

	handler, err := New(capability, &Config{
		VirtualUserConfig: apigwclient.VirtualUserConfig{
			AppConfig: apigwclient.NewAppConfig([]string{endpoint}, "test-app", "test-secret"),
			AuthMode:  apigwclient.AuthModeUn,
			LoginName: "admin",
		},
		SystemID:     "bk_nodemgr",
		CallbackPath: "/api/v3/iam/callback",
	})
	if err != nil {
		t.Fatalf("failed to create IAM v4 handler: %v", err)
	}

	return handler
}

func isDirectAuthRequest(req *http.Request) bool {
	return req.Method == http.MethodPost &&
		strings.Contains(req.URL.Path, "/rbac/authorization/systems/") &&
		strings.HasSuffix(req.URL.Path, "/auth/")
}

func writeDirectAuthResponse(rw http.ResponseWriter, allowed bool) {
	rw.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(rw, `{"code":0,"msg":"ok","data":{"allowed":%t}}`, allowed)
}

func newSystemAccessCheckRequest() types.IAMCheckRequest {
	return types.IAMCheckRequest{SystemID: "bk_nodemgr", Username: "admin", ActionID: "system_access"}
}

func TestIsAllowedWithCache_CachesAllowedResult(t *testing.T) {
	var authCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if !isDirectAuthRequest(req) {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		authCalls.Add(1)
		writeDirectAuthResponse(rw, true)
	}))
	defer server.Close()

	h := newTestIAMHandler(t, server.URL)
	ctx := newTestContext("tenant-a")

	for range 2 {
		allowed, err := h.IsAllowedWithCache(ctx, newSystemAccessCheckRequest(), time.Minute)
		if err != nil {
			t.Fatalf("IsAllowedWithCache() unexpected error: %v", err)
		}
		if !allowed {
			t.Fatal("IsAllowedWithCache() = false; want true")
		}
	}
	if got := authCalls.Load(); got != 1 {
		t.Fatalf("direct auth called %d times, want 1", got)
	}
}

func TestIsAllowedWithCache_DenyResultNotCached(t *testing.T) {
	var authCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if !isDirectAuthRequest(req) {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		authCalls.Add(1)
		writeDirectAuthResponse(rw, false)
	}))
	defer server.Close()

	h := newTestIAMHandler(t, server.URL)
	ctx := newTestContext("tenant-a")

	for range 2 {
		allowed, err := h.IsAllowedWithCache(ctx, newSystemAccessCheckRequest(), time.Minute)
		if err != nil {
			t.Fatalf("IsAllowedWithCache() unexpected error: %v", err)
		}
		if allowed {
			t.Fatal("IsAllowedWithCache() = true; want false")
		}
	}
	if got := authCalls.Load(); got != 2 {
		t.Fatalf("direct auth called %d times, want 2 (deny results must not be cached)", got)
	}
}

func TestIsAllowedWithCache_IAMErrorNotCached(t *testing.T) {
	var authCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if !isDirectAuthRequest(req) {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		call := authCalls.Add(1)
		if call == 1 {
			rw.Header().Set("Content-Type", "application/json")
			fmt.Fprint(rw, `{"code":1901401,"msg":"iam unavailable","data":{"allowed":false}}`)
			return
		}
		writeDirectAuthResponse(rw, true)
	}))
	defer server.Close()

	h := newTestIAMHandler(t, server.URL)
	ctx := newTestContext("tenant-a")

	if _, err := h.IsAllowedWithCache(ctx, newSystemAccessCheckRequest(), time.Minute); err == nil {
		t.Fatal("first IsAllowedWithCache() expected IAM error, got nil")
	}
	allowed, err := h.IsAllowedWithCache(ctx, newSystemAccessCheckRequest(), time.Minute)
	if err != nil {
		t.Fatalf("second IsAllowedWithCache() unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("second IsAllowedWithCache() = false; want true")
	}
	if got := authCalls.Load(); got != 2 {
		t.Fatalf("direct auth called %d times, want 2 (errors must not be cached)", got)
	}
}

func TestIsAllowedWithCache_TenantIsolation(t *testing.T) {
	var authCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if !isDirectAuthRequest(req) {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		authCalls.Add(1)
		writeDirectAuthResponse(rw, true)
	}))
	defer server.Close()

	h := newTestIAMHandler(t, server.URL)

	for _, tenantID := range []string{"tenant-a", "tenant-b"} {
		allowed, err := h.IsAllowedWithCache(newTestContext(tenantID), newSystemAccessCheckRequest(), time.Minute)
		if err != nil {
			t.Fatalf("IsAllowedWithCache() unexpected error: %v", err)
		}
		if !allowed {
			t.Fatal("IsAllowedWithCache() = false; want true")
		}
	}
	if got := authCalls.Load(); got != 2 {
		t.Fatalf("direct auth called %d times, want 2 (cache key must isolate tenants)", got)
	}
}

func TestIsAllowedWithCache_ExpiresAfterTTL(t *testing.T) {
	var authCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if !isDirectAuthRequest(req) {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		authCalls.Add(1)
		writeDirectAuthResponse(rw, true)
	}))
	defer server.Close()

	h := newTestIAMHandler(t, server.URL)
	ctx := newTestContext("tenant-a")

	if _, err := h.IsAllowedWithCache(ctx, newSystemAccessCheckRequest(), 20*time.Millisecond); err != nil {
		t.Fatalf("first IsAllowedWithCache() unexpected error: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	if _, err := h.IsAllowedWithCache(ctx, newSystemAccessCheckRequest(), 20*time.Millisecond); err != nil {
		t.Fatalf("second IsAllowedWithCache() unexpected error: %v", err)
	}
	if got := authCalls.Load(); got != 2 {
		t.Fatalf("direct auth called %d times, want 2 (expired entries must re-auth)", got)
	}
}

func TestAuthCacheKeyDeterministicWithMultiAttributes(t *testing.T) {
	h := &Handler{}
	attrs := map[string]interface{}{
		"_bk_iam_path_": "/biz,1/",
		"zone":          "z1",
		"region":        "r1",
		"vendor":        "v1",
	}
	req := types.IAMCheckRequest{
		SystemID: "bk_nodemgr", Username: "admin", ActionID: "host_view",
		Resources: []types.IAMResource{{ID: "host-1", Attributes: attrs}},
	}
	ctx := newTestContext("tenant-a")

	baseKey, err := h.authCacheKey(ctx, req)
	if err != nil {
		t.Fatalf("authCacheKey() unexpected error: %v", err)
	}
	for range 10 {
		key, err := h.authCacheKey(ctx, req)
		if err != nil {
			t.Fatalf("authCacheKey() unexpected error: %v", err)
		}
		if key != baseKey {
			t.Fatalf("authCacheKey() = %q, want deterministic key %q for multi-attribute resource", key, baseKey)
		}
	}
}

func TestAuthCacheKeyIsolation(t *testing.T) {
	h := &Handler{}
	base := newSystemAccessCheckRequest()

	baseKey, err := h.authCacheKey(newTestContext("tenant-a"), base)
	if err != nil {
		t.Fatalf("authCacheKey() unexpected error: %v", err)
	}

	variants := map[string]struct {
		tenantID string
		req      types.IAMCheckRequest
	}{
		"tenant": {"tenant-b", base},
		"user":   {"tenant-a", types.IAMCheckRequest{SystemID: "bk_nodemgr", Username: "bob", ActionID: "system_access"}},
		"system": {"tenant-a", types.IAMCheckRequest{SystemID: "bk_other", Username: "admin", ActionID: "system_access"}},
		"action": {"tenant-a", types.IAMCheckRequest{SystemID: "bk_nodemgr", Username: "admin", ActionID: "host_view"}},
		"resource": {"tenant-a", types.IAMCheckRequest{
			SystemID: "bk_nodemgr", Username: "admin", ActionID: "system_access",
			Resources: []types.IAMResource{{ID: "biz-1", Attributes: map[string]interface{}{"_bk_iam_path_": "/biz,1/"}}},
		}},
	}

	for name, variant := range variants {
		key, err := h.authCacheKey(newTestContext(variant.tenantID), variant.req)
		if err != nil {
			t.Fatalf("authCacheKey() unexpected error: %v", err)
		}
		if key == baseKey {
			t.Fatalf("authCacheKey() %s variant shares key with base request; keys must isolate requests", name)
		}
	}
}
