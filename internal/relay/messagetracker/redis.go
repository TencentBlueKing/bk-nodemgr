/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package messagetracker provides the message storage implementation for relay operations.
package messagetracker

import (
	"context"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/rediscache"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/cache"
)

const (
	pendingAckPrefix  = "bknm:relay:acked:"
	processedPrefix   = "bknm:relay:processed:"
	defaultexpiration = 12 * time.Hour
	markedSigal       = "1"
)

// RedisTracker is a Redis-based implementation of a message store for relay operations.
type RedisTracker struct {
	cache      cache.ICache
	expiration time.Duration
}

// NewRedisTracker creates a new RedisTracker instance.
func NewRedisTracker(cache *rediscache.RedisCache) IMessageTracker {
	return &RedisTracker{
		cache:      cache,
		expiration: defaultexpiration,
	}
}

// MarkAcked marks a message ID as acked.
func (rt *RedisTracker) MarkAcked(ctx context.Context, mid string) error {
	key := pendingAckPrefix + mid
	return rt.cache.SetWithExpiration(ctx, key, []byte("1"), rt.expiration)
}

// IsAcked checks if a message ID is in the pending list.
func (rt *RedisTracker) IsAcked(ctx context.Context, mid string) (bool, error) {
	key := pendingAckPrefix + mid
	return rt.cache.Exists(ctx, key)
}

// TryMarkProcessed marks a message ID as processed.if key exists, return false.
func (rt *RedisTracker) TryMarkProcessed(ctx context.Context, mid string) (bool, error) {
	key := processedPrefix + mid
	return rt.cache.SetNXWithExpiration(ctx, key, []byte(markedSigal), rt.expiration)
}
