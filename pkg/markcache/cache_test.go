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

package markcache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/cache"
	"github.com/stretchr/testify/require"
)

type fakeMarkerStore struct {
	versions  map[string]int64
	incrKeys  []string
	getErr    error
	failFirst map[string]bool
}

func newFakeMarkerStore() *fakeMarkerStore {
	return &fakeMarkerStore{versions: map[string]int64{}, failFirst: map[string]bool{}}
}

func (f *fakeMarkerStore) Get(_ context.Context, key string) (int64, error) {
	if f.getErr != nil {
		return 0, f.getErr
	}

	return f.versions[key], nil
}

func (f *fakeMarkerStore) Incr(_ context.Context, key string) error {
	f.incrKeys = append(f.incrKeys, key)
	if f.failFirst[key] {
		delete(f.failFirst, key)
		return errors.New("incr failed")
	}
	f.versions[key]++

	return nil
}

func newTestCache(store MarkerStore) *Cache {
	return New(cache.NewMemoryCache(time.Minute), store)
}

func TestValueEnvelope(t *testing.T) {
	payload := []byte("{\"name\":\"plugin\"}")
	data := encodeValue(42, payload)

	version, got, err := decodeValue(data)
	require.NoError(t, err)
	require.Equal(t, int64(42), version)
	require.Equal(t, payload, got)

	_, _, err = decodeValue([]byte("raw"))
	require.Error(t, err)
}

func TestGetOrLoadWithoutMarkersKeepsTTLBehavior(t *testing.T) {
	c := newTestCache(nil)

	loads := 0
	load := func(context.Context) ([]byte, error) {
		loads++
		return []byte("raw"), nil
	}

	data, err := c.GetOrLoad(t.Context(), "key", "marker", load)
	require.NoError(t, err)
	require.Equal(t, []byte("raw"), data)

	data, err = c.GetOrLoad(t.Context(), "key", "marker", load)
	require.NoError(t, err)
	require.Equal(t, []byte("raw"), data)
	require.Equal(t, 1, loads)
}

func TestGetOrLoadFreshHit(t *testing.T) {
	c := newTestCache(newFakeMarkerStore())

	loads := 0
	load := func(context.Context) ([]byte, error) {
		loads++
		return []byte("v1"), nil
	}

	for range 2 {
		data, err := c.GetOrLoad(t.Context(), "key", "marker", load)
		require.NoError(t, err)
		require.Equal(t, []byte("v1"), data)
	}
	require.Equal(t, 1, loads)
}

func TestGetOrLoadRefreshesAfterInvalidate(t *testing.T) {
	c := newTestCache(newFakeMarkerStore())

	loads := 0
	load := func(context.Context) ([]byte, error) {
		loads++
		return []byte{byte('0' + loads)}, nil
	}

	data, err := c.GetOrLoad(t.Context(), "key", "marker", load)
	require.NoError(t, err)
	require.Equal(t, []byte("1"), data)

	require.NoError(t, c.Invalidate(t.Context(), "marker"))

	data, err = c.GetOrLoad(t.Context(), "key", "marker", load)
	require.NoError(t, err)
	require.Equal(t, []byte("2"), data)

	data, err = c.GetOrLoad(t.Context(), "key", "marker", load)
	require.NoError(t, err)
	require.Equal(t, []byte("2"), data)
	require.Equal(t, 2, loads)
}

// TestGetOrLoadStampsPreLoadMarker covers a write racing with a load: the entry must
// keep the pre-load marker so the next read refreshes it again.
func TestGetOrLoadStampsPreLoadMarker(t *testing.T) {
	store := newFakeMarkerStore()
	c := newTestCache(store)
	store.versions["marker"] = 1

	loads := 0
	load := func(context.Context) ([]byte, error) {
		loads++
		// A write lands while this load is in flight.
		store.versions["marker"]++
		return []byte("v"), nil
	}

	_, err := c.GetOrLoad(t.Context(), "key", "marker", load)
	require.NoError(t, err)

	// The entry was stamped with version 1, but the marker is already 2.
	_, err = c.GetOrLoad(t.Context(), "key", "marker", load)
	require.NoError(t, err)
	require.Equal(t, 2, loads)
}

func TestGetOrLoadMarkerOutageFailOpen(t *testing.T) {
	store := newFakeMarkerStore()
	c := newTestCache(store)
	store.getErr = errors.New("marker store down")

	loads := 0
	load := func(context.Context) ([]byte, error) {
		loads++
		return []byte{byte('0' + loads)}, nil
	}

	data, err := c.GetOrLoad(t.Context(), "key", "marker", load)
	require.NoError(t, err)
	require.Equal(t, []byte("1"), data)

	// During the outage cached entries still serve reads.
	data, err = c.GetOrLoad(t.Context(), "key", "marker", load)
	require.NoError(t, err)
	require.Equal(t, []byte("1"), data)
	require.Equal(t, 1, loads)

	// After recovery, entries loaded during the outage carry version zero and refresh
	// as soon as any write has bumped the marker.
	store.getErr = nil
	require.NoError(t, c.Invalidate(t.Context(), "marker"))
	data, err = c.GetOrLoad(t.Context(), "key", "marker", load)
	require.NoError(t, err)
	require.Equal(t, []byte("2"), data)
	require.Equal(t, 2, loads)
}

func TestGetOrLoadMalformedEntryReloads(t *testing.T) {
	mem := cache.NewMemoryCache(time.Minute)
	c := New(mem, newFakeMarkerStore())
	require.NoError(t, mem.Set(t.Context(), "key", []byte("raw")))

	loads := 0
	load := func(context.Context) ([]byte, error) {
		loads++
		return []byte("fresh"), nil
	}

	data, err := c.GetOrLoad(t.Context(), "key", "marker", load)
	require.NoError(t, err)
	require.Equal(t, []byte("fresh"), data)
	require.Equal(t, 1, loads)
}

func TestInvalidateWithoutMarkers(t *testing.T) {
	c := newTestCache(nil)
	require.NoError(t, c.Invalidate(t.Context(), "marker"))
}

func TestInvalidateRetriesOnce(t *testing.T) {
	store := newFakeMarkerStore()
	c := newTestCache(store)
	store.failFirst["marker"] = true

	require.NoError(t, c.Invalidate(t.Context(), "marker"))
	require.Equal(t, []string{"marker", "marker"}, store.incrKeys)
	require.Equal(t, int64(1), store.versions["marker"])
}

func TestInvalidateReportsFailure(t *testing.T) {
	store := newFakeMarkerStore()
	c := newTestCache(store)
	store.failFirst["a"] = true // first attempt fails, retry succeeds

	require.NoError(t, c.Invalidate(t.Context(), "a"))
	require.Equal(t, int64(1), store.versions["a"])

	store.getErr = errors.New("marker store down")
	store.failFirst = map[string]bool{}
	// both attempts fail when the store is down: fake fails via getErr only on Get,
	// so force Incr failures with a store that always errors
	c2 := New(cache.NewMemoryCache(time.Minute), failingMarkerStore{})
	require.Error(t, c2.Invalidate(t.Context(), "a"))
}

type failingMarkerStore struct{}

func (failingMarkerStore) Get(context.Context, string) (int64, error) {
	return 0, errors.New("marker store down")
}

func (failingMarkerStore) Incr(context.Context, string) error {
	return errors.New("marker store down")
}
