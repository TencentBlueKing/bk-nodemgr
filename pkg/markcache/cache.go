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

// Package markcache provides a cache decorator that adds shared invalidation
// markers on top of a process-local cache.
//
// Each entry is stamped with the value of its invalidation marker observed
// before the load; a read refreshes the entry once the shared marker has
// advanced beyond that stamp. Several cache keys may share one marker, so a
// single invalidate call refreshes a whole group of entries on every pod.
//
// Behavior boundaries:
//   - Invalidation is group-scoped: unrelated keys sharing the marker are
//     refreshed as well, which is intended for low-frequency writes.
//   - A write racing with an in-flight load self-heals: entries carry the
//     pre-load marker value, so the next read reloads once more.
//   - Marker store failures fail open to the underlying TTL behavior; entries
//     loaded during an outage carry version zero and refresh after recovery.
//   - This is read-your-writes for the caller that invalidates, not strong
//     consistency: a read concurrent with an in-flight write may still serve
//     pre-write data.
package markcache

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
)

// MarkerStore provides shared invalidation version markers.
// A missing marker key reads as version zero.
type MarkerStore interface {
	// Get returns the current marker version for the key.
	Get(ctx context.Context, key string) (int64, error)
	// Incr atomically increments the marker version for the key.
	Incr(ctx context.Context, key string) error
}

// dataCache is the subset of the underlying cache implementation that Cache relies on.
// *cache.MemoryCache satisfies it, providing TTL expiration and singleflight loads.
type dataCache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) (bool, error)
	GetOrLoad(ctx context.Context, key string, load func(context.Context) ([]byte, error)) ([]byte, error)
}

// Cache decorates a process-local cache with shared invalidation markers.
// A nil MarkerStore disables markers and keeps plain TTL-only behavior.
type Cache struct {
	data    dataCache
	markers MarkerStore
}

// New creates a Cache over the given data cache and marker store.
func New(data dataCache, markers MarkerStore) *Cache {
	return &Cache{data: data, markers: markers}
}

const versionHeaderLen = 8

// encodeValue wraps payload with a big-endian version header.
func encodeValue(version int64, payload []byte) []byte {
	data := make([]byte, versionHeaderLen+len(payload))
	binary.BigEndian.PutUint64(data, uint64(version))
	copy(data[versionHeaderLen:], payload)

	return data
}

// decodeValue splits the version header from the payload.
func decodeValue(data []byte) (int64, []byte, error) {
	if len(data) < versionHeaderLen {
		return 0, nil, errors.New("cache value is shorter than the version header")
	}

	return int64(binary.BigEndian.Uint64(data)), data[versionHeaderLen:], nil
}

// GetOrLoad returns the cached payload for key, loading and caching it on a miss or
// when the shared marker has advanced beyond the version the entry was loaded under.
// Concurrent misses for the same key share one load via the underlying cache.
func (mc *Cache) GetOrLoad(
	ctx context.Context, key, markerKey string, load func(context.Context) ([]byte, error),
) ([]byte, error) {

	if mc.markers == nil {
		// Markers are disabled; keep the TTL-only behavior.
		return mc.data.GetOrLoad(ctx, key, load)
	}

	marker, err := mc.markers.Get(ctx, markerKey)
	failOpen := err != nil
	if failOpen {
		logger.G.Sys().WithErr(err).With("key", markerKey).
			Error("failed to get cache invalidation marker, fallback to TTL-only cache")
	}

	if payload, ok := mc.freshEntry(ctx, key, marker, failOpen); ok {
		return payload, nil
	}

	// Entries loaded during a marker outage are stamped with version zero, so they
	// refresh once the marker store recovers and a write has bumped the marker.
	entryVersion := marker
	if failOpen {
		entryVersion = 0
	}

	data, err := mc.data.GetOrLoad(ctx, key, func(ctx context.Context) ([]byte, error) {
		payload, err := load(ctx)
		if err != nil {
			return nil, err
		}

		return encodeValue(entryVersion, payload), nil
	})
	if err != nil {
		return nil, err
	}

	// Values on this path are always written by the envelope loader above.
	_, payload, err := decodeValue(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cache value. key(%s): %w", key, err)
	}

	return payload, nil
}

// freshEntry returns the cached payload when a decodable entry exists and is fresh
// against the marker. Stale or malformed entries are dropped so the caller reloads.
func (mc *Cache) freshEntry(ctx context.Context, key string, marker int64, failOpen bool) ([]byte, bool) {
	data, err := mc.data.Get(ctx, key)
	if err != nil {
		// Cache miss.
		return nil, false
	}

	version, payload, decodeErr := decodeValue(data)
	if decodeErr == nil && (failOpen || version >= marker) {
		return payload, true
	}

	// Stale or malformed entry: drop it before reloading.
	if _, delErr := mc.data.Delete(ctx, key); delErr != nil {
		logger.G.Sys().WithErr(delErr).With("key", key).Warn("failed to delete stale cache entry")
	}

	return nil, false
}

// Invalidate bumps the shared markers so every pod refreshes the associated entries.
// Each key is retried once; the write triggering invalidation has already committed,
// so failures are reported to the caller instead of rolling anything back.
func (mc *Cache) Invalidate(ctx context.Context, markerKeys ...string) error {
	if mc.markers == nil {
		// Markers are disabled.
		return nil
	}

	var errs []error
	for _, key := range markerKeys {
		if err := mc.markers.Incr(ctx, key); err != nil {
			if retryErr := mc.markers.Incr(ctx, key); retryErr != nil {
				errs = append(errs, retryErr)
			}
		}
	}

	return errors.Join(errs...)
}
