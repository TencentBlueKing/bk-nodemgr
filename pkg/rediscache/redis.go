/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package rediscache provides a Redis-based cache implementation.
// nolint: nonamedreturns
package rediscache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/cache"
	"github.com/redis/go-redis/v9"
)

var _ cache.ICache = &RedisCache{}

// DefaultTimeout is a default RedisCache instance.
const DefaultTimeout = 12 * time.Hour

// RedisCache is a Redis-based cache implementation.
type RedisCache struct {
	client     *redis.Client
	defaultTTL time.Duration
}

var (
	errContextIsNil = errors.New("context is nil")
)

// NewRedisCache creates a new RedisCache instance.
func NewRedisCache(client *redis.Client, defaultTTL time.Duration) *RedisCache {
	return &RedisCache{
		client:     client,
		defaultTTL: defaultTTL,
	}
}

// Get retrieves a value by key.
func (rc *RedisCache) Get(ctx context.Context, key string) (data []byte, err error) {
	if ctx == nil {
		return nil, errContextIsNil
	}

	// record metric.
	metric := metric().start(MetricOperationGet, 0)
	defer metric.end(err, len(data))

	if data, err = rc.client.Get(ctx, key).Bytes(); err == redis.Nil {
		return nil, fmt.Errorf("key not found. key(%s)", key)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get redis cache. key(%s): %w", key, err)
	}

	return data, nil
}

// SetWithExpiration sets a value with an expiration time.
func (rc *RedisCache) SetWithExpiration(ctx context.Context, key string, value []byte, ttl time.Duration) (err error) {
	if ctx == nil {
		return errContextIsNil
	}

	// record metric.
	metric := metric().start(MetricOperationSet, len(value))
	defer metric.end(err, 0)

	if err = rc.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set redis cache. key(%s): %w", key, err)
	}

	return nil
}

// Set stores a value with a key with default expiration.
func (rc *RedisCache) Set(ctx context.Context, key string, value []byte) (err error) {
	if ctx == nil {
		return errContextIsNil
	}

	// record metric.
	metric := metric().start(MetricOperationSet, len(value))
	defer metric.end(err, 0)

	if err = rc.client.Set(ctx, key, value, rc.defaultTTL).Err(); err != nil {
		return fmt.Errorf("failed to set redis cache. key(%s): %w", key, err)
	}

	return nil
}

// SetNX sets a value with a key if the key exist will return false.
func (rc *RedisCache) SetNX(ctx context.Context, key string, value []byte) (result bool, err error) {
	if ctx == nil {
		return false, errContextIsNil
	}

	// record metric.
	metric := metric().start(MetricOperationSetNX, len(value))
	defer metric.end(err, 0)

	if result, err = rc.client.SetNX(ctx, key, value, rc.defaultTTL).Result(); err != nil {
		return false, fmt.Errorf("failed to set redis cache. key(%s): %w", key, err)
	}

	return result, nil
}

// SetNXWithExpiration sets a value with an expiration time if the key exist will return false.
func (rc *RedisCache) SetNXWithExpiration(ctx context.Context, key string, value []byte, expiration time.Duration) (result bool, err error) {
	if ctx == nil {
		return false, errContextIsNil
	}

	// record metric.
	metric := metric().start(MetricOperationSetNX, len(value))
	defer metric.end(err, 0)

	if result, err = rc.client.SetNX(ctx, key, value, expiration).Result(); err != nil {
		return false, fmt.Errorf("failed to set redis cache. key(%s): %w", key, err)
	}

	return result, nil
}

// Exists checks if a key exists in the cache.
func (rc *RedisCache) Exists(ctx context.Context, key string) (result bool, err error) {
	if ctx == nil {
		return false, errContextIsNil
	}

	// record metric.
	metric := metric().start(MetricOperationExists, 0)
	defer metric.end(err, 1)

	var count int64
	if count, err = rc.client.Exists(ctx, key).Result(); err != nil {
		return false, fmt.Errorf("failed to check redis cache. key(%s): %w", key, err)
	}

	return count > 0, nil
}

// Delete removes a value by key.
func (rc *RedisCache) Delete(ctx context.Context, key string) (result bool, err error) {
	if ctx == nil {
		return false, errContextIsNil
	}

	// record metric.
	metric := metric().start(MetricOperationDelete, 0)
	defer metric.end(err, 1)

	var count int64
	if count, err = rc.client.Del(ctx, key).Result(); err != nil {
		return false, fmt.Errorf("failed to delete redis cache. key(%s): %w", key, err)
	}

	return count > 0, nil
}
