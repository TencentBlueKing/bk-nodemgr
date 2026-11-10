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

package rediscache

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// MarkerStore provides shared cache invalidation markers backed by Redis INCR.
// It implements markcache.MarkerStore.
type MarkerStore struct {
	client redis.UniversalClient
}

// NewMarkerStore creates a Redis-backed marker store.
func NewMarkerStore(client redis.UniversalClient) *MarkerStore {
	return &MarkerStore{client: client}
}

// Get returns the current marker version; a missing key means version zero.
func (s *MarkerStore) Get(ctx context.Context, key string) (int64, error) {
	metric := metric().start(MetricOperationGet, len(key))
	var err error
	defer func() { metric.end(err, 0) }()

	var version int64
	version, err = s.client.Get(ctx, key).Int64()
	if errors.Is(err, redis.Nil) {
		// A missing key is a normal cache-miss, not a failure.
		err = nil
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get cache invalidation marker. key(%s): %w", key, err)
	}

	return version, nil
}

// Incr atomically increments the marker version.
func (s *MarkerStore) Incr(ctx context.Context, key string) error {
	metric := metric().start(MetricOperationIncr, len(key))
	var err error
	defer func() { metric.end(err, 0) }()

	if err = s.client.Incr(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to incr cache invalidation marker. key(%s): %w", key, err)
	}

	return nil
}
