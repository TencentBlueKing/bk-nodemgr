/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package iamv3

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

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

type testTraceService struct{}

func (testTraceService) TracerProvider() trace.TracerProvider {
	return noop.NewTracerProvider()
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
		AuthMode:   apigwclient.AuthModeUn,
		BKUsername: "admin",
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
					AuthMode:   apigwclient.AuthModeUn,
					BKUsername: "admin",
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
			AppConfig:  apigwclient.NewAppConfig([]string{endpoint}, "test-app", "test-secret"),
			AuthMode:   apigwclient.AuthModeUn,
			BKUsername: "admin",
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
			tokenResponse:   `{"code":0,"message":"ok","data":{"token":"test-token"}}`,
			tokenStatusCode: http.StatusOK,
			wantTokenCalls:  1,
		},
		{
			name:             "invalid username short-circuits before token fetch",
			username:         "wrong-user",
			password:         "test-token",
			tokenResponse:    `{"code":0,"message":"ok","data":{"token":"test-token"}}`,
			tokenStatusCode:  http.StatusOK,
			wantErrSubstring: "invalid credentials",
			wantTokenCalls:   0,
		},
		{
			name:             "invalid password",
			username:         "bk_iam",
			password:         "wrong-password",
			tokenResponse:    `{"code":0,"message":"ok","data":{"token":"test-token"}}`,
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
			err := h.IsBasicAuthAllowed(contextx.New(context.Background()), tc.username, tc.password)

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
