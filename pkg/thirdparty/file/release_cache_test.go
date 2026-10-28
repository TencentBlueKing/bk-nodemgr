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

package file

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/cache"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestHandlerQueryReleaseCacheSingleflight(t *testing.T) {
	h := &handler{releaseCache: cache.NewMemoryCache(time.Minute)}
	nCtx := contextx.Background()
	start := make(chan struct{})
	release := make(chan struct{})
	var calls atomic.Int32

	fetch := func(contextx.IContext) ([]byte, error) {
		calls.Add(1)
		close(start)
		<-release
		return []byte("release"), nil
	}

	results := make(chan []byte, 2)
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	for range 2 {
		go func() {
			defer wg.Done()
			data, err := h.queryReleaseCache(nCtx, "release-key", fetch)
			results <- data
			errs <- err
		}()
	}

	<-start
	close(release)
	wg.Wait()
	close(results)
	close(errs)

	if got := calls.Load(); got != 1 {
		t.Fatalf("expected one fetch call, got %d", got)
	}
	for err := range errs {
		if err != nil {
			t.Fatalf("query release cache failed: %v", err)
		}
	}
	for data := range results {
		if string(data) != "release" {
			t.Fatalf("expected cached release, got %q", data)
		}
	}

	data, err := h.queryReleaseCache(nCtx, "release-key", fetch)
	require.NoError(t, err)
	require.Equal(t, "release", string(data))
	require.Equal(t, int32(1), calls.Load())
}

func TestHandlerQueryReleaseCacheContextCancellation(t *testing.T) {
	h := &handler{releaseCache: cache.NewMemoryCache(time.Minute)}
	parent := contextx.Background()
	nCtx, cancel := contextx.WithCancel(parent)
	cancel()

	_, err := h.queryReleaseCache(nCtx, "release-key", func(contextx.IContext) ([]byte, error) {
		t.Fatal("fetch should not run for a cancelled context")
		return nil, nil
	})
	require.ErrorIs(t, err, context.Canceled)
}

func TestHandlerQueryReleaseCacheDoesNotCacheError(t *testing.T) {
	h := &handler{releaseCache: cache.NewMemoryCache(time.Minute)}
	nCtx := contextx.Background()
	var calls atomic.Int32

	fetch := func(contextx.IContext) ([]byte, error) {
		if calls.Add(1) == 1 {
			return nil, errors.New("temporary failure")
		}
		return []byte("release"), nil
	}

	_, err := h.queryReleaseCache(nCtx, "release-key", fetch)
	require.Error(t, err)

	data, err := h.queryReleaseCache(nCtx, "release-key", fetch)
	require.NoError(t, err)
	require.Equal(t, "release", string(data))
	if got := calls.Load(); got != 2 {
		t.Fatalf("expected two fetch calls, got %d", got)
	}
}

func TestHandlerReleaseMetadataReadsCacheResults(t *testing.T) {
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("tenant-a"))
	page := types.SingleItemPage()
	fields := types.ReleaseDistinctField{Name: true}
	tests := []struct {
		name  string
		data  string
		query func(*testing.T, IHandler) (any, error)
		want  any
	}{
		{"agent/list", `{}`, func(t *testing.T, h IHandler) (any, error) {
			items, total, err := h.ListReleaseAgent(nCtx, page)
			require.Empty(t, items)
			return total, err
		}, nil},
		{"agent/count", `{}`, func(_ *testing.T, h IHandler) (any, error) {
			return h.CountReleaseAgent(nCtx)
		}, nil},
		{"agent/get", `{}`, func(_ *testing.T, h IHandler) (any, error) {
			return h.GetReleaseAgent(nCtx, types.ReleaseAgentKey{})
		}, nil},
		{"agent/distinct", `{"result":{"name":["agent"]}}`, func(_ *testing.T, h IHandler) (any, error) {
			return h.DistinctReleaseAgent(nCtx, fields)
		}, &types.ReleaseDistinctResult{Name: []string{"agent"}}},
		{"proxy/list", `{}`, func(t *testing.T, h IHandler) (any, error) {
			items, total, err := h.ListReleaseProxy(nCtx, page)
			require.Empty(t, items)
			return total, err
		}, nil},
		{"proxy/count", `{}`, func(_ *testing.T, h IHandler) (any, error) {
			return h.CountReleaseProxy(nCtx)
		}, nil},
		{"proxy/get", `{}`, func(_ *testing.T, h IHandler) (any, error) {
			return h.GetReleaseProxy(nCtx, types.ReleaseProxyKey{})
		}, nil},
		{"proxy/distinct", `{"result":{"name":["proxy"]}}`, func(_ *testing.T, h IHandler) (any, error) {
			return h.DistinctReleaseProxy(nCtx, fields)
		}, &types.ReleaseDistinctResult{Name: []string{"proxy"}}},
		{"plugin/list", `{}`, func(t *testing.T, h IHandler) (any, error) {
			items, total, err := h.ListReleasePlugin(nCtx, page)
			require.Empty(t, items)
			return total, err
		}, nil},
		{"plugin/count", `{}`, func(_ *testing.T, h IHandler) (any, error) {
			return h.CountReleasePlugin(nCtx)
		}, nil},
		{"plugin/get", `{}`, func(_ *testing.T, h IHandler) (any, error) {
			return h.GetReleasePlugin(nCtx, types.ReleasePluginKey{})
		}, nil},
		{"plugin/distinct", `{"result":{"name":["plugin"]}}`, func(_ *testing.T, h IHandler) (any, error) {
			return h.DistinctReleasePlugin(nCtx, fields)
		}, &types.ReleaseDistinctResult{Name: []string{"plugin"}}},
		{"plugin/exist", `{}`, func(_ *testing.T, h IHandler) (any, error) {
			return h.ExistReleasePlugin(nCtx, types.ReleasePluginKey{})
		}, nil},
		{"plugin/default_version", `{"version":"1.2.3"}`, func(_ *testing.T, h IHandler) (any, error) {
			return h.GetReleasePluginDefaultVersion(nCtx, "plugin", types.Generation2, platfmt.Platform{})
		}, "1.2.3"},
		{"plugin/distinct_name", `{"names":["plugin"]}`, func(_ *testing.T, h IHandler) (any, error) {
			return h.DistinctNameReleasePlugin(nCtx)
		}, []string{"plugin"}},
		{"cert/list", `{}`, func(t *testing.T, h IHandler) (any, error) {
			items, total, err := h.ListReleaseCert(nCtx, page)
			require.Empty(t, items)
			return total, err
		}, nil},
		{"bintool/list", `{}`, func(t *testing.T, h IHandler) (any, error) {
			items, total, err := h.ListReleaseBinTool(nCtx, page)
			require.Empty(t, items)
			return total, err
		}, nil},
		{"plugin_bintool/list", `{}`, func(t *testing.T, h IHandler) (any, error) {
			items, total, err := h.ListReleasePluginBinTool(nCtx, page)
			require.Empty(t, items)
			return total, err
		}, nil},
		{"plugin_bintool/distinct_name", `{"names":["tool"]}`, func(_ *testing.T, h IHandler) (any, error) {
			return h.DistinctNameReleasePluginBinTool(nCtx)
		}, []string{"tool"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, calls := newReleaseCacheHTTPHandler(t, func(req *http.Request, _ int32) string {
				if req.Method != http.MethodPost || req.URL.Path != "/api/v3/release/"+tt.name {
					t.Errorf("unexpected request: %s %s", req.Method, req.URL.Path)
				}
				return tt.data
			})
			for range 2 {
				got, err := tt.query(t, h)
				require.NoError(t, err)
				if tt.want == nil {
					require.Empty(t, got)
					continue
				}
				require.Equal(t, tt.want, got)
			}
			require.Equal(t, int32(1), calls.Load())
		})
	}
}

func TestHandlerReleaseMetadataCacheSeparatesQueries(t *testing.T) {
	h, calls := newReleaseCacheHTTPHandler(t, func(_ *http.Request, _ int32) string { return `{}` })
	tenantA := contextx.New(context.Background(), contextx.WithTenantID("tenant-a"))
	tenantB := contextx.New(context.Background(), contextx.WithTenantID("tenant-b"))
	page := types.SingleItemPage()
	condition := &types.ReleaseCondition{ExactInclude: &types.ReleaseExactFields{Name: []string{"plugin"}}}
	for range 2 {
		_, _, err := h.ListReleasePlugin(tenantA, page)
		require.NoError(t, err)
		_, _, err = h.ListReleasePlugin(tenantB, page)
		require.NoError(t, err)
		_, _, err = h.ListReleaseAgent(tenantA, page)
		require.NoError(t, err)
		_, _, err = h.ListReleasePlugin(tenantA, types.Page{Offset: 1, Limit: 1})
		require.NoError(t, err)
		_, _, err = h.ListReleasePlugin(tenantA, page, condition)
		require.NoError(t, err)
		_, err = h.DistinctReleasePlugin(tenantA, types.ReleaseDistinctField{Name: true})
		require.NoError(t, err)
		_, err = h.DistinctReleasePlugin(tenantA, types.ReleaseDistinctField{Version: true})
		require.NoError(t, err)
	}
	require.Equal(t, int32(7), calls.Load())
}

func TestHandlerReleaseMetadataCacheReturnsIndependentValues(t *testing.T) {
	h, calls := newReleaseCacheHTTPHandler(t, func(_ *http.Request, _ int32) string {
		return `{"total":1,"items":[{"name":"plugin","labels":["stable"],"addition_info":{"custom":"original"}}]}`
	})
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("tenant-a"))
	items, total, err := h.ListReleasePlugin(nCtx, types.SingleItemPage())
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	items[0].Name = "changed"
	items[0].Labels[0] = "changed"
	items[0].AdditionInfo["custom"] = "changed"
	items[0] = nil

	items, total, err = h.ListReleasePlugin(nCtx, types.SingleItemPage())
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	require.Equal(t, "plugin", items[0].Name)
	require.Equal(t, []string{"stable"}, items[0].Labels)
	require.Equal(t, "original", items[0].AdditionInfo["custom"])
	require.Equal(t, int32(1), calls.Load())
}

func TestHandlerReleaseMetadataCacheDoesNotCacheConversionError(t *testing.T) {
	h, calls := newReleaseCacheHTTPHandler(t, func(_ *http.Request, call int32) string {
		if call == 1 {
			return `{"item":{"name":"plugin","updated_at":{"seconds":253402300800}}}`
		}
		return `{"item":{"name":"plugin","updated_at":{"seconds":1700000000}}}`
	})
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("tenant-a"))
	key := types.ReleasePluginKey{Name: "plugin"}
	_, err := h.GetReleasePlugin(nCtx, key)
	require.ErrorContains(t, err, "invalid updated_at")
	for range 2 {
		item, err := h.GetReleasePlugin(nCtx, key)
		require.NoError(t, err)
		require.NotNil(t, item)
		require.Equal(t, "plugin", item.Name)
		require.Equal(t, int64(1700000000), item.UpdatedAt.Unix())
	}
	require.Equal(t, int32(2), calls.Load())
}

type releaseCacheTraceService struct{}

func (releaseCacheTraceService) TracerProvider() trace.TracerProvider {
	return noop.NewTracerProvider()
}
func (releaseCacheTraceService) ServiceName() string            { return "file-release-cache-test" }
func (releaseCacheTraceService) Shutdown(context.Context) error { return nil }
func (releaseCacheTraceService) TracerPropagator() propagation.TextMapPropagator {
	return propagation.TraceContext{}
}

func newReleaseCacheHTTPHandler(t *testing.T, respond func(*http.Request, int32) string) (IHandler, *atomic.Int32) {
	t.Helper()
	calls := new(atomic.Int32)
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		data := respond(req, calls.Add(1))
		rw.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(rw, fmt.Sprintf(`{"code":0,"data":%s}`, data)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	h, err := New(&restclient.Capability{
		Name: "file-release-cache-test", HTTPClient: server.Client(),
		Discover: restdiscovery.NewDiscovery("file-release-cache-test", []string{server.URL}),
		TraceSvc: releaseCacheTraceService{},
	}, &Config{RestJwtSecret: "test-secret", RestJwtTokenExpiration: time.Minute})
	require.NoError(t, err)
	return h, calls
}
