//go:build integration

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

package redisratelimit

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ratelimit"
	"github.com/TencentBlueKing/bk-nodemgr/testsuite/support"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestRedisLimiter(t *testing.T) {
	client, prefix := support.RequireRedisClientWithKeyPrefix(t)
	// Reuse the isolated test target, but give each limiter factory its own
	// non-retrying client, as required for non-idempotent quota operations.
	options := *client.(*redis.Client).Options()
	options.MaxRetries = -1
	options.ContextTimeoutEnabled = true
	firstClient := redis.NewClient(&options)
	t.Cleanup(func() { require.NoError(t, firstClient.Close()) })
	secondOptions := options
	secondOptions.MaxRetries = -1
	secondClient := redis.NewClient(&secondOptions)
	t.Cleanup(func() { _ = secondClient.Close() })
	first, err := New(firstClient, prefix)
	require.NoError(t, err)
	second, err := New(secondClient, prefix)
	require.NoError(t, err)

	t.Run("shared concurrent quota and rule conflict", func(t *testing.T) {
		rule := ratelimit.Rule{Limit: 1, Interval: time.Hour, Burst: 5}
		a, err := first.NewLimiter("shared", rule)
		require.NoError(t, err)
		b, err := second.NewLimiter("shared", rule)
		require.NoError(t, err)
		var allowedCount atomic.Int64
		var wg sync.WaitGroup
		for i := range 30 {
			lim := a
			if i%2 != 0 {
				lim = b
			}
			wg.Go(func() {
				allowed, err := lim.Allow(t.Context())
				if err != nil {
					t.Error(err)
					return
				}
				if allowed {
					allowedCount.Add(1)
				}
			})
		}
		wg.Wait()
		require.EqualValues(t, rule.Burst, allowedCount.Load())
		again, err := first.NewLimiter("shared", rule)
		require.NoError(t, err)
		allowed, err := again.Allow(t.Context())
		require.NoError(t, err)
		require.False(t, allowed)

		rule.Burst++
		_, err = first.NewLimiter("shared", rule)
		require.ErrorIs(t, err, ratelimit.ErrRuleConflict)
		third, err := New(secondClient, prefix)
		require.NoError(t, err)
		conflict, err := third.NewLimiter("shared", rule)
		require.NoError(t, err)
		_, err = conflict.Allow(t.Context())
		require.ErrorIs(t, err, ratelimit.ErrRuleConflict)
	})

	t.Run("wait cancellation and refill", func(t *testing.T) {
		rule := ratelimit.Rule{Limit: 1, Interval: time.Second, Burst: 1}
		a, err := first.NewLimiter("wait", rule)
		require.NoError(t, err)
		b, err := second.NewLimiter("wait", rule)
		require.NoError(t, err)
		require.NoError(t, a.Wait(t.Context()))
		ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
		defer cancel()
		require.ErrorIs(t, b.Wait(ctx), context.DeadlineExceeded)
		ctx, cancel = context.WithTimeout(t.Context(), 3*time.Second)
		defer cancel()
		require.NoError(t, b.Wait(ctx))
		allowed, err := a.Allow(ctx)
		require.NoError(t, err)
		require.False(t, allowed)
	})

	t.Run("idle expiration preserves capacity", func(t *testing.T) {
		rule := ratelimit.Rule{Limit: 1, Interval: 200 * time.Millisecond, Burst: 2}
		lim, err := first.NewLimiter("expiry", rule)
		require.NoError(t, err)
		require.NoError(t, lim.Wait(t.Context()))
		require.NoError(t, lim.Wait(t.Context()))
		key := lim.(*limiter).key
		count, err := client.Exists(t.Context(), key).Result()
		require.NoError(t, err)
		require.EqualValues(t, 1, count)
		require.Eventually(t, func() bool {
			count, err := client.Exists(t.Context(), key).Result()
			if err != nil {
				return false
			}

			return count == 0
		}, 3*time.Second, 20*time.Millisecond)
		for range rule.Burst {
			allowed, err := lim.Allow(t.Context())
			require.NoError(t, err)
			require.True(t, allowed)
		}
		allowed, err := lim.Allow(t.Context())
		require.NoError(t, err)
		require.False(t, allowed)
	})

	t.Run("backend failure", func(t *testing.T) {
		lim, err := second.NewLimiter("failure", ratelimit.Rule{Limit: 1, Interval: time.Second, Burst: 1})
		require.NoError(t, err)
		require.NoError(t, secondClient.Close())
		require.Error(t, lim.Wait(t.Context()))
	})
}
