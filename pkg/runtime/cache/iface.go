/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package cache provides a unified interface for various cache backends.
package cache

import (
	"context"
	"fmt"
	"time"
)

// ICache defines the unified cache interface.
type ICache interface {
	// Get retrieves a value by key.
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value with a key. using default ttl.
	Set(ctx context.Context, key string, value []byte) error

	// SetNX sets a value with a key if the key exist will return false.
	SetNX(ctx context.Context, key string, value []byte) (bool, error)

	// SetNXWithExpiration sets a value with an expiration time if the key exist will return false.
	SetNXWithExpiration(ctx context.Context, key string, value []byte, expiration time.Duration) (bool, error)

	// SetWithExpiration sets a value with an expiration time.
	SetWithExpiration(ctx context.Context, key string, value []byte, expiration time.Duration) error

	// Exists checks if a key exists in the cache.
	Exists(ctx context.Context, key string) (bool, error)

	// Delete removes a value by key.
	Delete(ctx context.Context, key string) (bool, error)
}

// IKeyGenerator defines the interface for generating cache keys.
type IKeyGenerator interface {
	Generate(components ...string) string
}

// NewDefaultKeyGenerator creates a default key generator.
func NewDefaultKeyGenerator(prefix string) IKeyGenerator {
	return &defaultKeyGenerator{prefix: prefix}
}

type defaultKeyGenerator struct {
	prefix string
}

// Generate generates a cache key.
func (g *defaultKeyGenerator) Generate(components ...string) string {
	key := g.prefix
	for _, comp := range components {
		key = fmt.Sprintf("%s:%s", key, comp)
	}

	return key
}
