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

package iamv3

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

const (
	// DefaultCacheExpiration is the default expiration time for cache entries.
	DefaultCacheExpiration = 5 * time.Minute
	// DefaultCacheCleanupInterval is the default cleanup interval for expired cache entries.
	DefaultCacheCleanupInterval = 10 * time.Minute
)

// Cache provides local in-memory caching for IAM permission check results.
type Cache struct {
	cache *gocache.Cache
}

// NewCache creates a new Cache instance with default settings.
func NewCache() *Cache {
	return &Cache{
		cache: gocache.New(DefaultCacheExpiration, DefaultCacheCleanupInterval),
	}
}

// NewCacheWithConfig creates a new Cache instance with custom settings.
func NewCacheWithConfig(defaultExpiration, cleanupInterval time.Duration) *Cache {
	return &Cache{
		cache: gocache.New(defaultExpiration, cleanupInterval),
	}
}

// Get retrieves a cached permission check result.
// Returns the cached value and whether it was found.
func (c *Cache) Get(key string) (bool, bool) {
	value, found := c.cache.Get(key)
	if !found {
		return false, false
	}

	result, ok := value.(bool)
	if !ok {
		return false, false
	}

	return result, true
}

// Set stores a permission check result in the cache with the specified TTL.
func (c *Cache) Set(key string, value bool, ttl time.Duration) {
	c.cache.Set(key, value, ttl)
}

// Delete removes an entry from the cache.
func (c *Cache) Delete(key string) {
	c.cache.Delete(key)
}

// Flush removes all entries from the cache.
func (c *Cache) Flush() {
	c.cache.Flush()
}

// ItemCount returns the number of items in the cache.
func (c *Cache) ItemCount() int {
	return c.cache.ItemCount()
}
