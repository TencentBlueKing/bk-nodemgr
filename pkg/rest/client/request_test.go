/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
)

type requestTestTraceService struct{}

func (requestTestTraceService) TracerProvider() trace.TracerProvider {
	return noop.NewTracerProvider()
}

func (requestTestTraceService) ServiceName() string {
	return "rest-client-request-test"
}

func (requestTestTraceService) Shutdown(_ context.Context) error {
	return nil
}

func (requestTestTraceService) TracerPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
}

type requestTestDiscovery struct {
	calls atomic.Int64
}

func (d *requestTestDiscovery) GetEndpoints() ([]string, error) {
	d.calls.Add(1)

	return nil, fmt.Errorf("discovery endpoint unavailable")
}

func TestMaskURLPreservesQueryByDefault(t *testing.T) {
	rawURL := "http://example.com/login/accounts/get_user/?bk_token=bkcrypt%2Bgabcdef%3D&foo=bar&token=short"

	got := maskURL(rawURL, nil)
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	query := parsed.Query()
	if query.Get("bk_token") != "bkcrypt+gabcdef=" {
		t.Fatalf("expected bk_token to be preserved by default, got %q", query.Get("bk_token"))
	}
	if query.Get("token") != "short" {
		t.Fatalf("expected token to be preserved by default, got %q", query.Get("token"))
	}
}

func TestMaskURLMasksConfiguredQuery(t *testing.T) {
	rawURL := "http://example.com/login/accounts/get_user/?bk_token=bkcrypt%2Bgabcdef%3D&foo=bar&token=short"
	urlQueryMasker := map[string]func(string) string{
		"bk_token": defaultHeaderMasker,
		"token":    defaultHeaderMasker,
	}

	got := maskURL(rawURL, urlQueryMasker)
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("failed to parse masked URL: %v", err)
	}

	query := parsed.Query()
	if query.Get("foo") != "bar" {
		t.Fatalf("expected non-configured query to be preserved, got %q", query.Get("foo"))
	}
	if query.Get("bk_token") != "bkc***ef=" {
		t.Fatalf("expected configured bk_token to use default masking, got %q", query.Get("bk_token"))
	}
	if query.Get("token") != "*****" {
		t.Fatalf("expected configured short token to be fully masked, got %q", query.Get("token"))
	}
}

func TestRequest_DoReturnsDiscoveryErrorBeforeTracing(t *testing.T) {
	discovery := new(requestTestDiscovery)
	client, err := NewClient(&Capability{
		Name:                 "rest-client-request-test",
		Discover:             discovery,
		ToleranceLatencyTime: ToleranceLatencyTimeDefault,
		MetricOpts:           MetricOption{},
	}, "/api/v1")
	require.NoError(t, err)

	result := client.Get().WithContext(contextx.Background()).SubResourcef("items").Do()

	require.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "discovery endpoint unavailable")
	assert.Equal(t, int64(1), discovery.calls.Load())
}

func TestRequest_WithURL(t *testing.T) {
	t.Run("uses target URL without discovery and preserves escaped path", func(t *testing.T) {
		discovery := new(requestTestDiscovery)
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			require.Equal(t, http.MethodGet, req.Method)
			assert.Equal(t, "/download/a%2Fb?from=url&from=builder&timeout=5s", req.URL.RequestURI())
			_, _ = rw.Write([]byte("ok"))
		}))
		defer server.Close()

		client := newRequestTestClient(t, discovery, "/api/v1")
		targetURL, err := url.Parse(server.URL + "/download/a%2Fb?from=url")
		require.NoError(t, err)

		result := client.Get().
			WithContext(contextx.Background()).
			WithURL(targetURL).
			WithParam("from", "builder").
			WithTimeout(5 * time.Second).
			Do()

		data, err := result.RawData()
		require.NoError(t, err)
		assert.Equal(t, "ok", string(data))
		assert.Equal(t, int64(0), discovery.calls.Load())
		assert.Contains(t, result.FullURL, "/download/a%2Fb")
		assert.NotContains(t, result.FullURL, "/api/v1")
	})

	t.Run("rejects sub resource and target URL combinations", func(t *testing.T) {
		discovery := new(requestTestDiscovery)
		client := newRequestTestClient(t, discovery, "/")
		targetURL, err := url.Parse("https://example.com/download/a%2Fb")
		require.NoError(t, err)

		result := client.Get().SubResourcef("items").WithURL(targetURL).Do()
		require.Error(t, result.Err)
		assert.Contains(t, result.Err.Error(), "target URL cannot be used with sub resource")
		assert.Equal(t, int64(0), discovery.calls.Load())

		result = client.Get().WithURL(targetURL).SubResourcef("items").Do()
		require.Error(t, result.Err)
		assert.Contains(t, result.Err.Error(), "target URL cannot be used with sub resource")
		assert.Equal(t, int64(0), discovery.calls.Load())
	})

	t.Run("rejects invalid target URL", func(t *testing.T) {
		discovery := new(requestTestDiscovery)
		client := newRequestTestClient(t, discovery, "/")

		tests := []struct {
			name    string
			target  *url.URL
			message string
		}{
			{
				name:    "nil URL",
				message: "target URL is nil",
			},
			{
				name:    "empty scheme",
				target:  &url.URL{Host: "example.com", Path: "/download"},
				message: "target URL scheme is empty",
			},
			{
				name:    "unsupported scheme",
				target:  &url.URL{Scheme: "ftp", Host: "example.com", Path: "/download"},
				message: "unsupported target URL scheme ftp",
			},
			{
				name:    "empty host",
				target:  &url.URL{Scheme: "https", Path: "/download"},
				message: "target URL host is empty",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := client.Get().WithURL(tt.target).Do()
				require.Error(t, result.Err)
				assert.Contains(t, result.Err.Error(), tt.message)
			})
		}

		assert.Equal(t, int64(0), discovery.calls.Load())
	})
}

func newRequestTestClient(t *testing.T, discovery *requestTestDiscovery, baseURL string) IClient {
	t.Helper()

	httpClient, err := NewHTTPClient(&ssl.TLSConfig{InsecureSkipVerify: true})
	require.NoError(t, err)

	client, err := NewClient(&Capability{
		Name:                 "rest-client-request-test",
		HTTPClient:           httpClient,
		Discover:             discovery,
		ToleranceLatencyTime: ToleranceLatencyTimeDefault,
		MetricOpts:           MetricOption{},
		TraceSvc:             requestTestTraceService{},
	}, baseURL)
	require.NoError(t, err)

	return client
}
