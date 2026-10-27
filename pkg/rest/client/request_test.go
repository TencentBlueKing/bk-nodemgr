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

package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
)

type requestTestTraceService struct {
	tracerProvider trace.TracerProvider
}

func (s requestTestTraceService) TracerProvider() trace.TracerProvider {
	if s.tracerProvider != nil {
		return s.tracerProvider
	}

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

func TestRequest_DoRecordsResponseHeadersOnSpan(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.Header().Set("X-Bkapi-Request-Id", "response-request-id")
		rw.Header().Add("X-Upstream", "first")
		rw.Header().Add("X-Upstream", "second")
		rw.Header().Set("Set-Cookie", "session=response-secret")
		_, _ = rw.Write([]byte("ok"))
	}))
	defer server.Close()

	httpClient, err := NewHTTPClient(&ssl.TLSConfig{InsecureSkipVerify: true})
	require.NoError(t, err)
	targetURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	tests := []struct {
		name        string
		targetURL   bool
		maskHeaders bool
	}{
		{name: "target URL", targetURL: true, maskHeaders: true},
		{name: "discovered endpoint", maskHeaders: true},
		{name: "unconfigured headers", targetURL: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := tracetest.NewSpanRecorder()
			provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
			defer func() { require.NoError(t, provider.Shutdown(context.Background())) }()

			var opts []Opt
			if tt.maskHeaders {
				opts = append(opts, WithHeaderMasker("Set-Cookie"))
			}
			client, err := NewClient(&Capability{
				Name:                 "rest-client-request-test",
				HTTPClient:           httpClient,
				Discover:             restdiscovery.NewDiscovery("test", []string{server.URL}),
				ToleranceLatencyTime: ToleranceLatencyTimeDefault,
				MetricOpts:           MetricOption{},
				TraceSvc:             requestTestTraceService{tracerProvider: provider},
			}, "/", opts...)
			require.NoError(t, err)

			request := client.Get().WithContext(contextx.Background())
			if tt.targetURL {
				request = request.WithURL(targetURL)
			}
			result := request.Do()
			require.NoError(t, result.Err)
			defer func() { require.NoError(t, result.Body.Close()) }()

			spans := recorder.Ended()
			require.Len(t, spans, 1)
			var headers string
			for _, attr := range spans[0].Attributes() {
				if attr.Key == attribute.Key("http.response.header") {
					headers = attr.Value.AsString()
				}
			}
			assert.Contains(t, headers, "X-Bkapi-Request-Id:[response-request-id]")
			assert.Contains(t, headers, "X-Upstream:[first second]")
			if tt.maskHeaders {
				assert.Contains(t, headers, "Set-Cookie:[ses***ret]")
				assert.NotContains(t, headers, "session=response-secret")
			} else {
				assert.Contains(t, headers, "Set-Cookie:[session=response-secret]")
			}
			assert.Equal(t, "session=response-secret", result.Header.Get("Set-Cookie"))
			_, err = result.RawData()
			require.NoError(t, err)
		})
	}
}

func TestResultIntoReturnsErrorForEmptyServerError(t *testing.T) {
	result := &Result{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(http.NoBody),
	}

	err := result.Into(new(struct{}))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "http request failed, status(500), body()")
}

func TestResultRawDataReturnsErrorForEmptyServerError(t *testing.T) {
	result := &Result{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(http.NoBody),
	}

	data, err := result.RawData()

	require.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "http request failed, status(500), body()")
}

func TestResultRawDataReturnsEmptyDataForEmptySuccess(t *testing.T) {
	result := &Result{
		StatusCode: http.StatusNoContent,
		Body:       io.NopCloser(http.NoBody),
	}

	data, err := result.RawData()

	require.NoError(t, err)
	assert.NotNil(t, data)
	assert.Empty(t, data)
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
