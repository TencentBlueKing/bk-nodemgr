/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package rediscache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/cache"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRedis(t *testing.T) cache.ICache {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "addr",
		Password: "",
		DB:       0,
	})

	// Ping to ensure connection works
	_, err := redisClient.Ping(context.Background()).Result()
	require.NoError(t, err, "Redis connection failed")

	return NewRedisCache(redisClient, 0)
}

func TestRedisCache_GetSet(t *testing.T) {
	rc := testRedis(t)
	ctx := context.Background()
	key := "test_key"
	value := []byte("test_value")

	t.Run("Set and Get", func(t *testing.T) {
		err := rc.Set(ctx, key, value)
		require.NoError(t, err)

		result, err := rc.Get(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, value, result)
	})

	t.Run("Non-existent key", func(t *testing.T) {
		_, err := rc.Get(ctx, "non_existent_key")
		require.NoError(t, err)
	})

	t.Run("Different data types", func(t *testing.T) {
		testCases := []struct {
			name     string
			value    interface{}
			expected string
		}{
			{"string", "string_value", "string_value"},
			{"int", 42, "42"},
			{"float", 3.14, "3.14"},
			{"bool", true, "1"},
			{"bytes", []byte("byte_slice"), "byte_slice"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				testKey := fmt.Sprintf("type_test_%s", tc.name)

				err := rc.Set(ctx, testKey, tc.value.([]byte))
				require.NoError(t, err)

				result, err := rc.Get(ctx, testKey)
				require.NoError(t, err)
				assert.Equal(t, tc.expected, string(result))
			})
		}
	})
}

func TestRedisCache_Exists(t *testing.T) {
	rc := testRedis(t)
	ctx := context.Background()

	t.Run("Existing key", func(t *testing.T) {
		key := "exists_key"
		err := rc.Set(ctx, key, []byte("value"))
		require.NoError(t, err)

		exists, err := rc.Exists(ctx, key)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Non-existing key", func(t *testing.T) {
		exists, err := rc.Exists(ctx, "non_existent_exists_key")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Expired key", func(t *testing.T) {
		key := "expired_exists_key"
		err := rc.SetWithExpiration(ctx, key, []byte("value"), 1*time.Second)
		require.NoError(t, err)

		exists, err := rc.Exists(ctx, key)
		require.NoError(t, err)
		assert.True(t, exists)

		time.Sleep(2 * time.Second)

		exists, err = rc.Exists(ctx, key)
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestRedisCache_Delete(t *testing.T) {
	rc := testRedis(t)
	ctx := context.Background()

	t.Run("Delete existing key", func(t *testing.T) {
		key := "delete_key"
		err := rc.Set(ctx, key, []byte("value"))
		require.NoError(t, err)

		exists, _ := rc.Exists(ctx, key)
		require.True(t, exists)

		deleted, err := rc.Delete(ctx, key)
		require.NoError(t, err)
		assert.True(t, deleted)

		exists, err = rc.Exists(ctx, key)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Delete non-existing key", func(t *testing.T) {
		deleted, err := rc.Delete(ctx, "non_existent_delete_key")
		require.NoError(t, err)
		assert.False(t, deleted)
	})

	t.Run("Delete multiple keys", func(t *testing.T) {
		keys := []string{"multi_key1", "multi_key2", "multi_key3"}

		for _, key := range keys {
			err := rc.Set(ctx, key, []byte("value"))
			require.NoError(t, err)
		}

		deleted, err := rc.Delete(ctx, keys[1])
		require.NoError(t, err)
		assert.True(t, deleted)

		exists1, _ := rc.Exists(ctx, keys[0])
		exists2, _ := rc.Exists(ctx, keys[1])
		exists3, _ := rc.Exists(ctx, keys[2])

		assert.True(t, exists1)
		assert.False(t, exists2)
		assert.True(t, exists3)
	})
}
