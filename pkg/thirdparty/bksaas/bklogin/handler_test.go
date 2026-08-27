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

package bklogin

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

const testTenantModeEnv = "BKLOGIN_TEST_TENANT_MODE"

type testTraceService struct{}

func (testTraceService) TracerProvider() trace.TracerProvider {
	return noop.NewTracerProvider()
}

func (testTraceService) ServiceName() string {
	return "bklogin_test"
}

func (testTraceService) Shutdown(_ context.Context) error {
	return nil
}

func (testTraceService) TracerPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
}

func newTestHandler(t *testing.T, authType string, h http.HandlerFunc) *Handler {
	t.Helper()

	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	return newTestHandlerWithEndpoint(t, authType, srv.URL)
}

func newTestHandlerWithEndpoint(t *testing.T, authType string, endpoint string) *Handler {
	t.Helper()

	conf := &Config{
		LoginURL: "https://bklogin.example.com/login",
		AuthType: authType,
		VirtualUserConfig: apigwclient.VirtualUserConfig{
			AppConfig: apigwclient.NewAppConfig([]string{endpoint}, "bk-nodemgr", "app-secret"),
			AuthMode:  apigwclient.AuthModeUn,
			LoginName: "admin",
		},
	}

	clientCap := &restclient.Capability{
		Name:                 "bklogin",
		HTTPClient:           &http.Client{Timeout: 2 * time.Second},
		Discover:             restdiscovery.NewDiscovery("bklogin", []string{endpoint}),
		ToleranceLatencyTime: restclient.ToleranceLatencyTimeDefault,
		MetricOpts:           restclient.MetricOption{},
		TraceSvc:             testTraceService{},
	}

	hh, err := New(clientCap, conf)
	require.NoError(t, err)

	handler, ok := hh.(*Handler)
	require.True(t, ok, "unexpected handler type: %T", hh)

	return handler
}

func runTenantModeTest(t *testing.T, testName string, mode tenant.Mode) {
	t.Helper()

	executable, err := os.Executable()
	require.NoError(t, err)

	cmd := exec.Command(executable, "-test.run=^"+testName+"$", "-test.v")
	cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s", testTenantModeEnv, mode))
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
}

func configuredTenantMode(t *testing.T) (tenant.Mode, bool) {
	t.Helper()

	mode := tenant.Mode(os.Getenv(testTenantModeEnv))
	if mode == "" {
		return "", false
	}

	require.NoError(t, mode.Validate())
	tenant.SetMode(mode)

	return mode, true
}

func TestHandlerGetWebUserInfo(t *testing.T) {
	mode, ok := configuredTenantMode(t)
	if !ok {
		for _, mode := range []tenant.Mode{tenant.ModeSingle, tenant.ModeMultiple} {
			t.Run(string(mode), func(t *testing.T) {
				runTenantModeTest(t, "TestHandlerGetWebUserInfo", mode)
			})
		}
		return
	}

	nCtx := contextx.New(context.Background())

	switch mode {
	case tenant.ModeSingle:
		t.Run("bk_ticket_success", func(t *testing.T) {
			h := newTestHandler(t, CookieKeyBKTicket, func(rw http.ResponseWriter, req *http.Request) {
				if req.URL.Path != "/user/get_info/" {
					t.Fatalf("unexpected request path: %s", req.URL.Path)
				}

				_, _ = rw.Write([]byte(`{"ret":0,"msg":"ok","data":{"username":"ticket_user","avatar_url":"https://avatar"}}`))
			})

			info, err := h.GetWebUserInfo(nCtx, "ticket-value")
			if err != nil {
				t.Fatalf("GetWebUserInfo() error = %v", err)
			}
			if info.LoginName != "ticket_user" {
				t.Fatalf("GetWebUserInfo() username = %s, want %s", info.LoginName, "ticket_user")
			}
		})

		t.Run("bk_token_single_success", func(t *testing.T) {
			h := newTestHandler(t, CookieKeyBKToken, func(rw http.ResponseWriter, req *http.Request) {
				if req.URL.Path != "/accounts/get_user/" {
					t.Fatalf("unexpected request path: %s", req.URL.Path)
				}

				_, _ = rw.Write([]byte(`{"result":true,"code":"00","message":"ok","data":{"username":"token_user"}}`))
			})

			info, err := h.GetWebUserInfo(nCtx, "token-value")
			if err != nil {
				t.Fatalf("GetWebUserInfo() error = %v", err)
			}
			if info.LoginName != "token_user" {
				t.Fatalf("GetWebUserInfo() username = %s, want %s", info.LoginName, "token_user")
			}
		})

		t.Run("upstream_failed", func(t *testing.T) {
			h := newTestHandler(t, CookieKeyBKTicket, func(rw http.ResponseWriter, _ *http.Request) {
				_, _ = rw.Write([]byte(`{"ret":1,"msg":"invalid","data":{}}`))
			})

			_, err := h.GetWebUserInfo(nCtx, "ticket-value")
			require.Error(t, err)
			if !strings.Contains(err.Error(), "failed to get web user info by bk_ticket") {
				t.Fatalf("GetWebUserInfo() error = %v, want contain %q", err, "failed to get web user info by bk_ticket")
			}
		})

		t.Run("missing_username", func(t *testing.T) {
			h := newTestHandler(t, CookieKeyBKToken, func(rw http.ResponseWriter, _ *http.Request) {
				_, _ = rw.Write([]byte(`{"result":true,"code":"00","message":"ok","data":{}}`))
			})

			_, err := h.GetWebUserInfo(nCtx, "token-value")
			require.Error(t, err)
			if !strings.Contains(err.Error(), "username is empty") {
				t.Fatalf("GetWebUserInfo() error = %v, want contain %q", err, "username is empty")
			}
		})

		t.Run("unsupported_auth_type", func(t *testing.T) {
			h := &Handler{
				conf: &Config{
					LoginURL: "https://bklogin.example.com/login",
					AuthType: "unknown",
				},
			}

			_, err := h.GetWebUserInfo(nCtx, "token-value")
			require.Error(t, err)
			if !strings.Contains(err.Error(), "unsupported auth type") {
				t.Fatalf("GetWebUserInfo() error = %v, want contain %q", err, "unsupported auth type")
			}
		})

		t.Run("invalid_context", func(t *testing.T) {
			h := newTestHandler(t, CookieKeyBKToken, func(rw http.ResponseWriter, _ *http.Request) {
				_, _ = rw.Write([]byte(`{"result":true,"code":"00","message":"ok","data":{"username":"token_user"}}`))
			})

			_, err := h.GetWebUserInfo(nil, "token-value")
			require.Error(t, err)
			if !strings.Contains(err.Error(), "invalid context") {
				t.Fatalf("GetWebUserInfo() error = %v, want contain %q", err, "invalid context")
			}
		})
	case tenant.ModeMultiple:
		t.Run("bk_token_multiple_success", func(t *testing.T) {
			h := newTestHandler(t, CookieKeyBKToken, func(rw http.ResponseWriter, req *http.Request) {
				require.Equal(t, "/login/api/v3/open/bk-tokens/userinfo/", req.URL.Path)
				require.Equal(t, "token-value", req.URL.Query().Get(CookieKeyBKToken))

				_, err := rw.Write([]byte(`{
					"data": {
						"bk_username": "nteuuhzxlh0jcanw",
						"tenant_id": "system",
						"login_name": "admin",
						"display_name": "admin",
						"language": "zh-cn",
						"time_zone": "Asia/Shanghai"
					}
				}`))
				require.NoError(t, err)
			})

			info, err := h.GetWebUserInfo(nCtx, "token-value")
			require.NoError(t, err)
			require.Equal(t, "nteuuhzxlh0jcanw", info.BKUsername)
			require.Equal(t, "admin", info.LoginName)
			require.Equal(t, "Asia/Shanghai", info.TimeZone)
		})
	default:
		require.Failf(t, "unsupported tenant mode", "mode=%s", mode)
	}
}
