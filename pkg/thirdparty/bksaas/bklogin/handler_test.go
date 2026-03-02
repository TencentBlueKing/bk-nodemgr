/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package bklogin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

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

	conf := &Config{
		LoginURL: "https://bklogin.example.com/login",
		AuthType: authType,
	}

	clientCap := &restclient.Capability{
		Name:                 "bklogin",
		HTTPClient:           &http.Client{Timeout: 2 * time.Second},
		Discover:             restdiscovery.NewDiscovery("bklogin", []string{srv.URL}),
		ToleranceLatencyTime: restclient.ToleranceLatencyTimeDefault,
		MetricOpts:           restclient.MetricOption{},
		TraceSvc:             testTraceService{},
	}

	hh, err := New(clientCap, conf)
	if err != nil {
		t.Fatalf("failed to create bklogin handler: %v", err)
	}

	handler, ok := hh.(*Handler)
	if !ok {
		t.Fatalf("unexpected handler type: %T", hh)
	}

	return handler
}

func TestHandlerGetWebUserInfo(t *testing.T) {
	nCtx := contextx.New(context.Background())

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
		if info.Username != "ticket_user" {
			t.Fatalf("GetWebUserInfo() username = %s, want %s", info.Username, "ticket_user")
		}
	})

	t.Run("bk_token_success", func(t *testing.T) {
		h := newTestHandler(t, CookieKeyBKToken, func(rw http.ResponseWriter, req *http.Request) {
			if req.URL.Path != "/accounts/get_user/" {
				t.Fatalf("unexpected request path: %s", req.URL.Path)
			}

			_, _ = rw.Write([]byte(`{"ret":0,"msg":"ok","data":{"username":"token_user"}}`))
		})

		info, err := h.GetWebUserInfo(nCtx, "token-value")
		if err != nil {
			t.Fatalf("GetWebUserInfo() error = %v", err)
		}
		if info.Username != "token_user" {
			t.Fatalf("GetWebUserInfo() username = %s, want %s", info.Username, "token_user")
		}
	})

	t.Run("upstream_failed", func(t *testing.T) {
		h := newTestHandler(t, CookieKeyBKTicket, func(rw http.ResponseWriter, _ *http.Request) {
			_, _ = rw.Write([]byte(`{"ret":1,"msg":"invalid","data":{}}`))
		})

		_, err := h.GetWebUserInfo(nCtx, "ticket-value")
		if err == nil {
			t.Fatal("GetWebUserInfo() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to get web user info by bk_ticket") {
			t.Fatalf("GetWebUserInfo() error = %v, want contain %q", err, "failed to get web user info by bk_ticket")
		}
	})

	t.Run("missing_username", func(t *testing.T) {
		h := newTestHandler(t, CookieKeyBKToken, func(rw http.ResponseWriter, _ *http.Request) {
			_, _ = rw.Write([]byte(`{"ret":0,"msg":"ok","data":{}}`))
		})

		_, err := h.GetWebUserInfo(nCtx, "token-value")
		if err == nil {
			t.Fatal("GetWebUserInfo() expected error, got nil")
		}
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
		if err == nil {
			t.Fatal("GetWebUserInfo() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "unsupported auth type") {
			t.Fatalf("GetWebUserInfo() error = %v, want contain %q", err, "unsupported auth type")
		}
	})

	t.Run("invalid_context", func(t *testing.T) {
		h := newTestHandler(t, CookieKeyBKToken, func(rw http.ResponseWriter, _ *http.Request) {
			_, _ = rw.Write([]byte(`{"ret":0,"msg":"ok","data":{"username":"token_user"}}`))
		})

		_, err := h.GetWebUserInfo(nil, "token-value")
		if err == nil {
			t.Fatal("GetWebUserInfo() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "invalid context") {
			t.Fatalf("GetWebUserInfo() error = %v, want contain %q", err, "invalid context")
		}
	})
}
