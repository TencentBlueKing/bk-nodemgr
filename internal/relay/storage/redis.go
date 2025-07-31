/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package storage provides the message storage implementation for relay operations.
package storage

import (
	"context"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/rediscache"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/cache"
)

const (
	pendingAckPrefix  = "relay:acked:"
	processedPrefix   = "relay:processed:"
	defaultexpiration = 12 * time.Hour
)

// RedisStore is a Redis-based implementation of a message store for relay operations.
type RedisStore struct {
	cache      cache.Cache
	expiration time.Duration
}

// NewRedisStore creates a new RedisStore instance.
func NewRedisStore(cache *rediscache.RedisCache) MessageStore {
	return &RedisStore{
		cache:      cache,
		expiration: defaultexpiration,
	}
}

// MarkedAcked marks a message ID as acked.
func (s *RedisStore) MarkedAcked(ctx context.Context, mid string) error {
	key := pendingAckPrefix + mid
	return s.cache.SetWithExpiration(ctx, key, "1", s.expiration)
}

// IsAcked checks if a message ID is in the pending list.
func (s *RedisStore) IsAcked(ctx context.Context, mid string) (bool, error) {
	key := pendingAckPrefix + mid
	return s.cache.Exists(ctx, key)
}

// MarkProcessed marks a message ID as processed.
func (s *RedisStore) MarkProcessed(ctx context.Context, mid string) error {
	key := processedPrefix + mid
	return s.cache.SetWithExpiration(ctx, key, "1", s.expiration)
}

// IsProcessed checks if a message ID has been processed.
func (s *RedisStore) IsProcessed(ctx context.Context, mid string) (bool, error) {
	key := processedPrefix + mid
	return s.cache.Exists(ctx, key)
}

// CleanupExpired just checks the connection.
func (s *RedisStore) CleanupExpired(ctx context.Context) error {
	// Just check connection health
	return s.cache.CleanupExpired(ctx)
}
