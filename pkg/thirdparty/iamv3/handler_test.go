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

package iamv3

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	restheader "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
	runtimecache "github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/cache"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

type testTraceService struct{}

func (testTraceService) TracerProvider() trace.TracerProvider {
	return noop.NewTracerProvider()
}

func newTestContext(tenantID string) contextx.IContext {
	return contextx.New(context.Background(), contextx.WithTenantID(tenantID))
}

func (testTraceService) ServiceName() string {
	return "iamv3_test"
}

func (testTraceService) Shutdown(_ context.Context) error {
	return nil
}

func (testTraceService) TracerPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
}

func TestNew(t *testing.T) {
	validAPIGWUserConfig := apigwclient.UserConfig{
		AppConfig: apigwclient.NewAppConfig(
			[]string{"https://example.com/api/bk-iam/prod"},
			"test-app",
			"test-secret",
		),
		AuthMode:  apigwclient.AuthModeUn,
		LoginName: "admin",
	}

	httpClient, err := restclient.NewHTTPClient(&ssl.TLSConfig{
		InsecureSkipVerify: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	clientCap := &restclient.Capability{
		Name:                 "iam-v3",
		HTTPClient:           httpClient,
		Discover:             restdiscovery.NewDiscovery("iam-v3", []string{"https://example.com/api/bk-iam/prod"}),
		ToleranceLatencyTime: restclient.ToleranceLatencyTimeDefault,
		MetricOpts:           restclient.MetricOption{},
		TraceSvc:             testTraceService{},
	}

	tests := []struct {
		name    string
		cap     *restclient.Capability
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cap:  clientCap,
			config: &Config{
				APIGWUserConfig: validAPIGWUserConfig,
				SystemID:        "bk_nodemgr",
				CallbackPath:    "/api/v3/iam/callback",
			},
			wantErr: false,
		},
		{
			name: "missing system ID",
			cap:  clientCap,
			config: &Config{
				APIGWUserConfig: validAPIGWUserConfig,
				SystemID:        "",
				CallbackPath:    "/api/v3/iam/callback",
			},
			wantErr: true,
		},
		{
			name: "missing callback path",
			cap:  clientCap,
			config: &Config{
				APIGWUserConfig: validAPIGWUserConfig,
				SystemID:        "bk_nodemgr",
				CallbackPath:    "",
			},
			wantErr: true,
		},
		{
			name: "invalid APIGWUserConfig",
			cap:  clientCap,
			config: &Config{
				APIGWUserConfig: apigwclient.UserConfig{
					AppConfig: apigwclient.NewAppConfig(
						[]string{"https://example.com/api/bk-iam/prod"},
						"", // empty app code
						"test-secret",
					),
					AuthMode:  apigwclient.AuthModeUn,
					LoginName: "admin",
				},
				SystemID:     "bk_nodemgr",
				CallbackPath: "/api/v3/iam/callback",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := New(tt.cap, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && handler == nil {
				t.Errorf("New() returned nil handler for valid config")
			}
			if !tt.wantErr && handler.cli == nil {
				t.Errorf("New() handler.cli is nil for valid config")
			}
		})
	}
}

func TestListAuthorizedInstances(t *testing.T) {
	tests := []struct {
		name             string
		req              types.IAMAuthorizedInstancesRequest
		responseBody     string
		responseStatus   int
		wantIsAny        bool
		wantIDs          []string
		wantErrSubstring string
		wantCalls        int32
	}{
		{
			name: "validation failure short-circuits",
			req: types.IAMAuthorizedInstancesRequest{
				SystemID: "bk_nodemgr",
				Username: "admin",
			},
			wantErrSubstring: "action is required",
			wantCalls:        0,
		},
		{
			name: "nil policy returns empty ids",
			req: types.IAMAuthorizedInstancesRequest{
				SystemID:     "bk_nodemgr",
				Username:     "admin",
				ActionID:     "agent_view",
				ResourceType: "biz",
			},
			responseBody:   `{"code":0,"message":"ok","data":null}`,
			responseStatus: http.StatusOK,
			wantIDs:        []string{},
			wantCalls:      1,
		},
		{
			name: "any policy returns is any",
			req: types.IAMAuthorizedInstancesRequest{
				SystemID:     "bk_nodemgr",
				Username:     "admin",
				ActionID:     "agent_view",
				ResourceType: "biz",
			},
			responseBody:   `{"code":0,"message":"ok","data":{"op":"any"}}`,
			responseStatus: http.StatusOK,
			wantIsAny:      true,
			wantCalls:      1,
		},
		{
			name: "in policy returns ids",
			req: types.IAMAuthorizedInstancesRequest{
				SystemID:     "bk_nodemgr",
				Username:     "admin",
				ActionID:     "agent_view",
				ResourceType: "biz",
			},
			responseBody:   `{"code":0,"message":"ok","data":{"op":"in","field":"biz.id","value":["1","2"]}}`,
			responseStatus: http.StatusOK,
			wantIDs:        []string{"1", "2"},
			wantCalls:      1,
		},
		{
			name: "iam failure bubbles up",
			req: types.IAMAuthorizedInstancesRequest{
				SystemID:     "bk_nodemgr",
				Username:     "admin",
				ActionID:     "agent_view",
				ResourceType: "biz",
			},
			responseBody:     `{"code":1,"message":"failed","data":{}}`,
			responseStatus:   http.StatusOK,
			wantErrSubstring: "v2 policy query failed",
			wantCalls:        1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var queryCalls int32
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				if req.URL.Path != "/api/v2/policy/systems/bk_nodemgr/query/" {
					rw.WriteHeader(http.StatusNotFound)
					return
				}
				atomic.AddInt32(&queryCalls, 1)
				rw.Header().Set("Content-Type", "application/json")
				rw.WriteHeader(tc.responseStatus)
				_, _ = rw.Write([]byte(tc.responseBody))
			}))
			defer server.Close()

			h := newTestIAMHandler(t, server.URL)
			gotIsAny, gotResources, err := h.ListAuthorizedInstances(contextx.New(context.Background()), tc.req)

			if tc.wantErrSubstring != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErrSubstring) {
					t.Fatalf("ListAuthorizedInstances() error = %v, want substring %q", err, tc.wantErrSubstring)
				}
			} else {
				if err != nil {
					t.Fatalf("ListAuthorizedInstances() unexpected error: %v", err)
				}
				if gotIsAny != tc.wantIsAny {
					t.Fatalf("ListAuthorizedInstances().isAny = %v, want %v", gotIsAny, tc.wantIsAny)
				}
				gotIDs := make([]string, 0, len(gotResources))
				for _, resource := range gotResources {
					gotIDs = append(gotIDs, resource.ID)
					if resource.SystemID != tc.req.SystemID {
						t.Fatalf("ListAuthorizedInstances().resource.SystemID = %q, want %q", resource.SystemID, tc.req.SystemID)
					}
					if resource.Type != tc.req.ResourceType {
						t.Fatalf("ListAuthorizedInstances().resource.Type = %q, want %q", resource.Type, tc.req.ResourceType)
					}
				}
				if strings.Join(gotIDs, ",") != strings.Join(tc.wantIDs, ",") {
					t.Fatalf("ListAuthorizedInstances().resourceIDs = %v, want %v", gotIDs, tc.wantIDs)
				}
			}

			if calls := atomic.LoadInt32(&queryCalls); calls != tc.wantCalls {
				t.Fatalf("policy query called %d times, want %d", calls, tc.wantCalls)
			}
		})
	}
}

func newTestIAMHandler(t *testing.T, endpoint string) *Handler {
	t.Helper()

	httpClient, err := restclient.NewHTTPClient(&ssl.TLSConfig{InsecureSkipVerify: true})
	if err != nil {
		t.Fatalf("failed to create HTTP client: %v", err)
	}

	capability := &restclient.Capability{
		Name:                 "iam-v3-test",
		HTTPClient:           httpClient,
		Discover:             restdiscovery.NewDiscovery("iam-v3", []string{endpoint}),
		ToleranceLatencyTime: restclient.ToleranceLatencyTimeDefault,
		MetricOpts:           restclient.MetricOption{},
		TraceSvc:             testTraceService{},
	}

	handler, err := New(capability, &Config{
		APIGWUserConfig: apigwclient.UserConfig{
			AppConfig: apigwclient.NewAppConfig([]string{endpoint}, "test-app", "test-secret"),
			AuthMode:  apigwclient.AuthModeUn,
			LoginName: "admin",
		},
		SystemID:     "bk_nodemgr",
		CallbackPath: "/api/v3/iam/callback",
	})
	if err != nil {
		t.Fatalf("failed to create IAM handler: %v", err)
	}

	return handler
}

func TestIsBasicAuthAllowed(t *testing.T) {
	tests := []struct {
		name             string
		username         string
		password         string
		tokenResponse    string
		tokenStatusCode  int
		wantErrSubstring string
		wantTokenCalls   int32
	}{
		{
			name:            "valid credentials",
			username:        "bk_iam",
			password:        "test-token",
			tokenResponse:   tokenResponse("test-token"),
			tokenStatusCode: http.StatusOK,
			wantTokenCalls:  1,
		},
		{
			name:             "invalid username short-circuits before token fetch",
			username:         "wrong-user",
			password:         "test-token",
			tokenResponse:    tokenResponse("test-token"),
			tokenStatusCode:  http.StatusOK,
			wantErrSubstring: "invalid credentials",
			wantTokenCalls:   0,
		},
		{
			name:             "invalid password",
			username:         "bk_iam",
			password:         "wrong-password",
			tokenResponse:    tokenResponse("test-token"),
			tokenStatusCode:  http.StatusOK,
			wantErrSubstring: "invalid credentials",
			wantTokenCalls:   1,
		},
		{
			name:             "token fetch failure",
			username:         "bk_iam",
			password:         "does-not-matter",
			tokenResponse:    `{"code":1,"message":"token service failed","data":{}}`,
			tokenStatusCode:  http.StatusOK,
			wantErrSubstring: "failed to get token",
			wantTokenCalls:   1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var tokenCalls int32
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				if req.Method != http.MethodGet {
					rw.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				if req.URL.Path != "/api/v1/model/systems/bk_nodemgr/token" {
					rw.WriteHeader(http.StatusNotFound)
					return
				}

				atomic.AddInt32(&tokenCalls, 1)
				rw.Header().Set("Content-Type", "application/json")
				rw.WriteHeader(tc.tokenStatusCode)
				_, _ = rw.Write([]byte(tc.tokenResponse))
			}))
			defer server.Close()

			h := newTestIAMHandler(t, server.URL)
			err := h.IsBasicAuthAllowed(newTestContext("default"), tc.username, tc.password)

			if tc.wantErrSubstring == "" {
				if err != nil {
					t.Fatalf("IsBasicAuthAllowed() unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("IsBasicAuthAllowed() expected error containing %q, got nil", tc.wantErrSubstring)
				}
				if !strings.Contains(err.Error(), tc.wantErrSubstring) {
					t.Fatalf("IsBasicAuthAllowed() error = %v, want substring %q", err, tc.wantErrSubstring)
				}
			}

			if got := atomic.LoadInt32(&tokenCalls); got != tc.wantTokenCalls {
				t.Fatalf("token endpoint called %d times, want %d", got, tc.wantTokenCalls)
			}
		})
	}
}

func TestIsBasicAuthAllowed_CachesSuccessfulToken(t *testing.T) {
	var tokenCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if !isTokenRequest(rw, req) {
			return
		}

		tokenCalls.Add(1)
		writeTokenResponse(rw, http.StatusOK, tokenResponse("test-token"))
	}))
	defer server.Close()

	h := newTestIAMHandler(t, server.URL)
	ctx := newTestContext("tenant-a")

	if err := h.IsBasicAuthAllowed(ctx, "bk_iam", "test-token"); err != nil {
		t.Fatalf("first IsBasicAuthAllowed() unexpected error: %v", err)
	}
	if err := h.IsBasicAuthAllowed(ctx, "bk_iam", "test-token"); err != nil {
		t.Fatalf("second IsBasicAuthAllowed() unexpected error: %v", err)
	}
	if got := tokenCalls.Load(); got != 1 {
		t.Fatalf("token endpoint called %d times, want 1", got)
	}
}

func TestIsBasicAuthAllowed_FetchFailureNotCached(t *testing.T) {
	var tokenCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if !isTokenRequest(rw, req) {
			return
		}

		call := tokenCalls.Add(1)
		if call == 1 {
			writeTokenResponse(rw, http.StatusOK, `{"code":1,"message":"token service failed","data":{}}`)
			return
		}
		writeTokenResponse(rw, http.StatusOK, tokenResponse("test-token"))
	}))
	defer server.Close()

	h := newTestIAMHandler(t, server.URL)
	ctx := newTestContext("tenant-a")

	err := h.IsBasicAuthAllowed(ctx, "bk_iam", "test-token")
	if err == nil || !strings.Contains(err.Error(), "failed to get token") {
		t.Fatalf("first IsBasicAuthAllowed() error = %v, want fetch failure", err)
	}
	if err := h.IsBasicAuthAllowed(ctx, "bk_iam", "test-token"); err != nil {
		t.Fatalf("second IsBasicAuthAllowed() unexpected error: %v", err)
	}
	if got := tokenCalls.Load(); got != 2 {
		t.Fatalf("token endpoint called %d times, want 2", got)
	}
}

func TestIsBasicAuthAllowed_EmptyTokenNotCachedOrAccepted(t *testing.T) {
	var tokenCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if !isTokenRequest(rw, req) {
			return
		}

		tokenCalls.Add(1)
		writeTokenResponse(rw, http.StatusOK, tokenResponse(""))
	}))
	defer server.Close()

	h := newTestIAMHandler(t, server.URL)
	ctx := newTestContext("tenant-a")

	for range 2 {
		err := h.IsBasicAuthAllowed(ctx, "bk_iam", "")
		if err == nil || !strings.Contains(err.Error(), "invalid credentials") {
			t.Fatalf("IsBasicAuthAllowed() error = %v, want invalid credentials", err)
		}
	}
	if got := tokenCalls.Load(); got != 2 {
		t.Fatalf("token endpoint called %d times, want 2", got)
	}
}

func TestIsBasicAuthAllowed_CachedMismatchRefreshesAndAcceptsRotatedToken(t *testing.T) {
	var tokenCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if !isTokenRequest(rw, req) {
			return
		}

		call := tokenCalls.Add(1)
		if call == 1 {
			writeTokenResponse(rw, http.StatusOK, tokenResponse("old-token"))
			return
		}
		writeTokenResponse(rw, http.StatusOK, tokenResponse("new-token"))
	}))
	defer server.Close()

	h := newTestIAMHandler(t, server.URL)
	ctx := newTestContext("tenant-a")

	if err := h.IsBasicAuthAllowed(ctx, "bk_iam", "old-token"); err != nil {
		t.Fatalf("old-token IsBasicAuthAllowed() unexpected error: %v", err)
	}
	if err := h.IsBasicAuthAllowed(ctx, "bk_iam", "new-token"); err != nil {
		t.Fatalf("new-token IsBasicAuthAllowed() unexpected error: %v", err)
	}
	if err := h.IsBasicAuthAllowed(ctx, "bk_iam", "new-token"); err != nil {
		t.Fatalf("cached new-token IsBasicAuthAllowed() unexpected error: %v", err)
	}
	if got := tokenCalls.Load(); got != 2 {
		t.Fatalf("token endpoint called %d times, want 2", got)
	}
}

func TestIsBasicAuthAllowed_CachedMismatchRefreshesAndRejectsWrongPassword(t *testing.T) {
	var tokenCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if !isTokenRequest(rw, req) {
			return
		}

		call := tokenCalls.Add(1)
		if call == 1 {
			writeTokenResponse(rw, http.StatusOK, tokenResponse("old-token"))
			return
		}
		writeTokenResponse(rw, http.StatusOK, tokenResponse("new-token"))
	}))
	defer server.Close()

	h := newTestIAMHandler(t, server.URL)
	ctx := newTestContext("tenant-a")

	if err := h.IsBasicAuthAllowed(ctx, "bk_iam", "old-token"); err != nil {
		t.Fatalf("old-token IsBasicAuthAllowed() unexpected error: %v", err)
	}
	err := h.IsBasicAuthAllowed(ctx, "bk_iam", "wrong-password")
	if err == nil || !strings.Contains(err.Error(), "invalid credentials") {
		t.Fatalf("wrong-password IsBasicAuthAllowed() error = %v, want invalid credentials", err)
	}
	if got := tokenCalls.Load(); got != 2 {
		t.Fatalf("token endpoint called %d times, want 2", got)
	}
}

func TestIsBasicAuthAllowed_CacheIsTenantIsolated(t *testing.T) {
	var tokenCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if !isTokenRequest(rw, req) {
			return
		}

		tokenCalls.Add(1)
		tenantToken := req.Header.Get(restheader.BKTenantIDKey) + "-token"
		writeTokenResponse(rw, http.StatusOK, tokenResponse(tenantToken))
	}))
	defer server.Close()

	h := newTestIAMHandler(t, server.URL)
	tenantA := newTestContext("tenant-a")
	tenantB := newTestContext("tenant-b")

	if err := h.IsBasicAuthAllowed(tenantA, "bk_iam", "tenant-a-token"); err != nil {
		t.Fatalf("tenant-a first IsBasicAuthAllowed() unexpected error: %v", err)
	}
	if err := h.IsBasicAuthAllowed(tenantA, "bk_iam", "tenant-a-token"); err != nil {
		t.Fatalf("tenant-a second IsBasicAuthAllowed() unexpected error: %v", err)
	}
	if err := h.IsBasicAuthAllowed(tenantB, "bk_iam", "tenant-b-token"); err != nil {
		t.Fatalf("tenant-b first IsBasicAuthAllowed() unexpected error: %v", err)
	}
	if err := h.IsBasicAuthAllowed(tenantB, "bk_iam", "tenant-b-token"); err != nil {
		t.Fatalf("tenant-b second IsBasicAuthAllowed() unexpected error: %v", err)
	}
	if got := tokenCalls.Load(); got != 2 {
		t.Fatalf("token endpoint called %d times, want 2", got)
	}
}

func TestGetToken_RefreshesAfterCacheExpiration(t *testing.T) {
	var tokenCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if !isTokenRequest(rw, req) {
			return
		}

		call := tokenCalls.Add(1)
		writeTokenResponse(rw, http.StatusOK, tokenResponse(fmt.Sprintf("token-%d", call)))
	}))
	defer server.Close()

	h := newTestIAMHandler(t, server.URL)
	h.tokenCache = runtimecache.NewMemoryCache(20 * time.Millisecond)
	ctx := newTestContext("tenant-a")

	token, err := h.GetToken(ctx)
	if err != nil {
		t.Fatalf("first GetToken() unexpected error: %v", err)
	}
	if token != "token-1" {
		t.Fatalf("first GetToken() returned unexpected token")
	}
	waitForTokenCalls(t, &tokenCalls, 1, 200*time.Millisecond)
	time.Sleep(30 * time.Millisecond)

	token, err = h.GetToken(ctx)
	if err != nil {
		t.Fatalf("second GetToken() unexpected error: %v", err)
	}
	if token != "token-2" {
		t.Fatalf("second GetToken() returned unexpected token")
	}
	if got := tokenCalls.Load(); got != 2 {
		t.Fatalf("token endpoint called %d times, want 2", got)
	}
}

func tokenResponse(token string) string {
	return fmt.Sprintf(`{"code":0,"message":"ok","data":{"token":%q}}`, token)
}

func writeTokenResponse(rw http.ResponseWriter, statusCode int, body string) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(statusCode)
	_, _ = rw.Write([]byte(body))
}

func isTokenRequest(rw http.ResponseWriter, req *http.Request) bool {
	if req.Method != http.MethodGet {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return false
	}
	if req.URL.Path != "/api/v1/model/systems/bk_nodemgr/token" {
		rw.WriteHeader(http.StatusNotFound)
		return false
	}

	return true
}

func waitForTokenCalls(t *testing.T, calls *atomic.Int32, want int32, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if calls.Load() == want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}

	t.Fatalf("token endpoint called %d times, want %d", calls.Load(), want)
}

// TestCacheKeyEquivalence proves that a CheckRequest converted via toWireRequest
// produces the same CacheKey as an equivalent wire Request built directly.
// This guarantees cache key stability across the business→wire conversion boundary.
func TestCacheKeyEquivalence(t *testing.T) {
	checkReq := types.IAMCheckRequest{
		SystemID: "bk_nodemgr",
		Username: "admin",
		ActionID: "host_view",
		Resources: []types.IAMResource{
			{
				SystemID:   "bk_cmdb",
				Type:       "host",
				ID:         "host-1",
				Attributes: map[string]interface{}{"os": "linux"},
			},
		},
	}

	// Convert via private helper under test
	wireReq := toWireRequest(checkReq)
	convertedKey, err := wireReq.CacheKey()
	if err != nil {
		t.Fatalf("CacheKey() on converted request failed: %v", err)
	}

	// Build equivalent wire Request directly (Subject.Type is hardcoded "user")
	directReq := Request{
		System: "bk_nodemgr",
		Subject: Subject{
			Type: "user",
			ID:   "admin",
		},
		Action: Action{ID: "host_view"},
		Resources: Resources{
			{
				System:    "bk_cmdb",
				Type:      "host",
				ID:        "host-1",
				Attribute: map[string]interface{}{"os": "linux"},
			},
		},
	}
	directKey, err := directReq.CacheKey()
	if err != nil {
		t.Fatalf("CacheKey() on direct request failed: %v", err)
	}

	if convertedKey != directKey {
		t.Errorf("cache key mismatch: converted=%q, direct=%q", convertedKey, directKey)
	}
	if convertedKey == "" {
		t.Error("cache key must not be empty")
	}
}

// TestBatchResourceIDKey verifies that buildResourceID produces correct map keys
// for various resource set shapes, matching the expected BatchIsAllowed key format.
func TestBatchResourceIDKey(t *testing.T) {
	tests := []struct {
		name      string
		resources Resources
		want      string
	}{
		{
			name:      "empty resources returns empty string",
			resources: Resources{},
			want:      "",
		},
		{
			name: "single resource returns resource ID only",
			resources: Resources{
				{System: "bk_cmdb", Type: "host", ID: "host-42", Attribute: nil},
			},
			want: "host-42",
		},
		{
			name: "multiple resources returns type,id pairs joined by slash",
			resources: Resources{
				{System: "bk_cmdb", Type: "host", ID: "host-1", Attribute: nil},
				{System: "bk_cmdb", Type: "module", ID: "mod-2", Attribute: nil},
			},
			want: "host,host-1/module,mod-2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildResourceID(tt.resources)
			if got != tt.want {
				t.Errorf("buildResourceID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetApplyURL_PreservesInstances(t *testing.T) {
	var gotApplication Application
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if req.URL.Path != "/api/v1/open/application/" {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		if err := json.NewDecoder(req.Body).Decode(&gotApplication); err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			_, _ = rw.Write([]byte(`{"code":1,"message":"bad request","data":{}}`))
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(`{"code":0,"message":"ok","data":{"url":"https://example.com/apply"}}`))
	}))
	defer server.Close()

	h := newTestIAMHandler(t, server.URL)
	url, err := h.GetApplyURL(contextx.New(context.Background()), types.IAMApplyRequest{
		SystemID: "bk_nodemgr",
		Actions: []types.IAMApplyAction{
			{
				ID: "host_view",
				RelatedResourceTypes: []types.IAMApplyResourceType{
					{
						SystemID: "bk_cmdb",
						Type:     "host",
						Instances: []types.IAMApplyResourceInstance{
							{
								{Type: "biz", ID: "biz-1"},
								{Type: "host", ID: "host-1"},
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("GetApplyURL() unexpected error: %v", err)
	}
	if url != "https://example.com/apply" {
		t.Fatalf("GetApplyURL() = %q, want %q", url, "https://example.com/apply")
	}

	if gotApplication.SystemID != "bk_nodemgr" {
		t.Fatalf("application.system_id = %q, want %q", gotApplication.SystemID, "bk_nodemgr")
	}
	if len(gotApplication.Actions) != 1 {
		t.Fatalf("application.actions length = %d, want 1", len(gotApplication.Actions))
	}
	if len(gotApplication.Actions[0].RelatedResourceTypes) != 1 {
		t.Fatalf("related_resource_types length = %d, want 1", len(gotApplication.Actions[0].RelatedResourceTypes))
	}

	rt := gotApplication.Actions[0].RelatedResourceTypes[0]
	if len(rt.Instances) != 1 {
		t.Fatalf("instances length = %d, want 1", len(rt.Instances))
	}
	if len(rt.Instances[0]) != 2 {
		t.Fatalf("instance path length = %d, want 2", len(rt.Instances[0]))
	}
	if rt.Instances[0][0].Type != "biz" || rt.Instances[0][0].ID != "biz-1" {
		t.Fatalf("first instance node = %+v, want biz/biz-1", rt.Instances[0][0])
	}
	if rt.Instances[0][1].Type != "host" || rt.Instances[0][1].ID != "host-1" {
		t.Fatalf("second instance node = %+v, want host/host-1", rt.Instances[0][1])
	}
}

// TestGetApplyURL_MultipleInstancesSameType verifies that multiple denied
// resources of the same type are sent as separate instances in one
// RelatedResourceType entry.
func TestGetApplyURL_MultipleInstancesSameType(t *testing.T) {
	var gotApplication Application
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/api/v1/open/application/" {
			rw.WriteHeader(http.StatusNotFound)
			return
		}
		_ = json.NewDecoder(req.Body).Decode(&gotApplication)
		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(`{"code":0,"message":"ok","data":{"url":"https://example.com/apply"}}`))
	}))
	defer server.Close()

	h := newTestIAMHandler(t, server.URL)
	_, err := h.GetApplyURL(contextx.New(context.Background()), types.IAMApplyRequest{
		SystemID: "bk_nodemgr",
		Actions: []types.IAMApplyAction{
			{
				ID: "host_view",
				RelatedResourceTypes: []types.IAMApplyResourceType{
					{
						SystemID: "bk_cmdb",
						Type:     "biz",
						Instances: []types.IAMApplyResourceInstance{
							{{Type: "biz", ID: "biz-1"}},
							{{Type: "biz", ID: "biz-2"}},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("GetApplyURL() unexpected error: %v", err)
	}

	if len(gotApplication.Actions) != 1 {
		t.Fatalf("actions length = %d, want 1", len(gotApplication.Actions))
	}
	rt := gotApplication.Actions[0].RelatedResourceTypes[0]
	if len(rt.Instances) != 2 {
		t.Fatalf("instances length = %d, want 2", len(rt.Instances))
	}
	if rt.Instances[0][0].ID != "biz-1" || rt.Instances[1][0].ID != "biz-2" {
		t.Fatalf("instance IDs mismatch: got %v %v", rt.Instances[0], rt.Instances[1])
	}
}

// TestGetApplyURL_EmptyInstances verifies that an IAMApplyResourceType with no
// instances (e.g. action-level permissions) produces an empty instances array
// in the wire payload and does not fail validation.
func TestGetApplyURL_EmptyInstances(t *testing.T) {
	var gotApplication Application
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/api/v1/open/application/" {
			rw.WriteHeader(http.StatusNotFound)
			return
		}
		_ = json.NewDecoder(req.Body).Decode(&gotApplication)
		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(`{"code":0,"message":"ok","data":{"url":"https://example.com/apply"}}`))
	}))
	defer server.Close()

	h := newTestIAMHandler(t, server.URL)
	_, err := h.GetApplyURL(contextx.New(context.Background()), types.IAMApplyRequest{
		SystemID: "bk_nodemgr",
		Actions: []types.IAMApplyAction{
			{
				ID: "networkarea_create",
				RelatedResourceTypes: []types.IAMApplyResourceType{
					{
						SystemID:  "bk_cmdb",
						Type:      "biz",
						Instances: []types.IAMApplyResourceInstance{},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("GetApplyURL() unexpected error: %v", err)
	}

	rt := gotApplication.Actions[0].RelatedResourceTypes[0]
	if len(rt.Instances) != 0 {
		t.Fatalf("expected empty instances, got %d", len(rt.Instances))
	}
}
