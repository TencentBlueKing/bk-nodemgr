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
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/cache"
	"github.com/stretchr/testify/require"
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
