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
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

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
func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("key(%s) not found", key)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get redis cache. key(%s), err: %w", key, err)
	}

	return data, nil
}

// SetWithExpiration sets a value with an expiration time.
func (c *RedisCache) SetWithExpiration(ctx context.Context,
	key string, value []byte, ttl time.Duration) error {

	return c.client.Set(ctx, key, value, ttl).Err()
}

// Set stores a value with a key without expiration.
func (c *RedisCache) Set(ctx context.Context, key string, value []byte) error {
	return c.client.Set(ctx, key, value, c.defaultTTL).Err()
}

// SetNX sets a value with a key if the key exist will return error.
func (c *RedisCache) SetNX(ctx context.Context, key string, value []byte) (bool, error) {
	return c.client.SetNX(ctx, key, value, c.defaultTTL).Result()
}

// SetNXWithExpiration sets a value with an expiration time.
func (c *RedisCache) SetNXWithExpiration(ctx context.Context,
	key string, value []byte, expiration time.Duration) (bool, error) {

	return c.client.SetNX(ctx, key, value, expiration).Result()
}

// Exists checks if a key exists in the cache.
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.client.Exists(ctx, key).Result()
	return count > 0, err
}

// Delete removes a value by key.
func (c *RedisCache) Delete(ctx context.Context, key string) (bool, error) {
	count, err := c.client.Del(ctx, key).Result()
	return count > 0, err
}
