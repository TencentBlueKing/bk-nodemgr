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

package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	gocache "github.com/patrickmn/go-cache"
)

var _ ICache = (*MemoryCache)(nil)

var errContextIsNil = errors.New("context is nil")

// MemoryCache is an in-memory cache implementation.
type MemoryCache struct {
	cache *gocache.Cache
}

// NewMemoryCache creates a MemoryCache with the given default expiration.
func NewMemoryCache(defaultExpiration time.Duration) *MemoryCache {
	return &MemoryCache{
		cache: gocache.New(defaultExpiration, 0),
	}
}

// Get retrieves a value by key.
func (mc *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	if ctx == nil {
		return nil, errContextIsNil
	}

	value, ok := mc.cache.Get(key)
	if !ok {
		return nil, fmt.Errorf("key not found. key(%s)", key)
	}

	data, ok := value.([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid cache value type. key(%s)", key)
	}

	return cloneBytes(data), nil
}

// Set stores a value with a key with default expiration.
func (mc *MemoryCache) Set(ctx context.Context, key string, value []byte) error {
	if ctx == nil {
		return errContextIsNil
	}

	mc.cache.Set(key, cloneBytes(value), gocache.DefaultExpiration)

	return nil
}

// SetNX sets a value with a key if the key exist will return false.
func (mc *MemoryCache) SetNX(ctx context.Context, key string, value []byte) (bool, error) {
	if ctx == nil {
		return false, errContextIsNil
	}

	return mc.add(key, value, gocache.DefaultExpiration), nil
}

// SetNXWithExpiration sets a value with an expiration time if the key exist will return false.
func (mc *MemoryCache) SetNXWithExpiration(
	ctx context.Context,
	key string,
	value []byte,
	expiration time.Duration,
) (bool, error) {

	if ctx == nil {
		return false, errContextIsNil
	}

	return mc.add(key, value, normalizeExpiration(expiration)), nil
}

// SetWithExpiration sets a value with an expiration time.
func (mc *MemoryCache) SetWithExpiration(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	if ctx == nil {
		return errContextIsNil
	}

	mc.cache.Set(key, cloneBytes(value), normalizeExpiration(expiration))

	return nil
}

// Exists checks if a key exists in the cache.
func (mc *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	if ctx == nil {
		return false, errContextIsNil
	}

	_, ok := mc.cache.Get(key)

	return ok, nil
}

// Delete removes a value by key.
func (mc *MemoryCache) Delete(ctx context.Context, key string) (bool, error) {
	if ctx == nil {
		return false, errContextIsNil
	}

	_, existed := mc.cache.Get(key)
	mc.cache.Delete(key)

	return existed, nil
}

func normalizeExpiration(expiration time.Duration) time.Duration {
	if expiration == 0 {
		return gocache.NoExpiration
	}

	return expiration
}

func (mc *MemoryCache) add(key string, value []byte, expiration time.Duration) bool {
	if mc.cache.Add(key, cloneBytes(value), expiration) != nil {
		return false
	}

	return true
}

func cloneBytes(value []byte) []byte {
	return append([]byte(nil), value...)
}
