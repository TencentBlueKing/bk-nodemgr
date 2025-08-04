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

// NewRedisCache creates a new RedisCache instance.
func NewRedisCache(client *redis.Client, defaultTTL time.Duration) *RedisCache {
	return &RedisCache{
		client:     client,
		defaultTTL: defaultTTL,
	}
}

// Get retrieves a value by key.
func (rc *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	if ctx == nil {
		return nil, errors.New("context is nil")
	}

	data, err := rc.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("key not found. key(%s)", key)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get redis cache. key(%s), err: %w", key, err)
	}

	return data, nil
}

// SetWithExpiration sets a value with an expiration time.
func (rc *RedisCache) SetWithExpiration(ctx context.Context,
	key string, value []byte, ttl time.Duration) error {

	if ctx == nil {
		return errors.New("context is nil")
	}

	err := rc.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set redis cache. key(%s), err: %w", key, err)
	}

	return nil
}

// Set stores a value with a key with default expiration.
func (rc *RedisCache) Set(ctx context.Context, key string, value []byte) error {
	if ctx == nil {
		return errors.New("context is nil")
	}

	err := rc.client.Set(ctx, key, value, rc.defaultTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to set redis cache. key(%s), err: %w", key, err)
	}

	return nil
}

// SetNX sets a value with a key if the key exist will return false.
func (rc *RedisCache) SetNX(ctx context.Context, key string, value []byte) (bool, error) {
	if ctx == nil {
		return false, errors.New("context is nil")
	}
	exists, err := rc.client.SetNX(ctx, key, value, rc.defaultTTL).Result()
	if err != nil {
		return false, fmt.Errorf("failed to set redis cache. key(%s), err: %w", key, err)
	}

	return exists, nil
}

// SetNXWithExpiration sets a value with an expiration time if the key exist will return false.
func (rc *RedisCache) SetNXWithExpiration(ctx context.Context,
	key string, value []byte, expiration time.Duration) (bool, error) {

	if ctx == nil {
		return false, errors.New("context is nil")
	}

	exists, err := rc.client.SetNX(ctx, key, value, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to set redis cache. key(%s), err: %w", key, err)
	}

	return exists, nil
}

// Exists checks if a key exists in the cache.
func (rc *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	if ctx == nil {
		return false, errors.New("context is nil")
	}
	count, err := rc.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check redis cache. key(%s), err: %w", key, err)
	}

	return count > 0, err
}

// Delete removes a value by key.
func (rc *RedisCache) Delete(ctx context.Context, key string) (bool, error) {
	if ctx == nil {
		return false, errors.New("context is nil")
	}

	count, err := rc.client.Del(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to delete redis cache. key(%s), err: %w", key, err)
	}

	return count > 0, err
}
