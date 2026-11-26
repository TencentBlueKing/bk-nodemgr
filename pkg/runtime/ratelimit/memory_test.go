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

package ratelimit

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewLimiterValidation(t *testing.T) {
	rule := TokenBucketRule{Limit: 1, Interval: time.Second, Burst: 1}
	handler := NewMemory()
	for _, tc := range []struct {
		name string
		rule TokenBucketRule
	}{
		{"zero limit", TokenBucketRule{Interval: time.Second, Burst: 1}},
		{"zero interval", TokenBucketRule{Limit: 1, Burst: 1}},
		{"zero burst", TokenBucketRule{Limit: 1, Interval: time.Second}},
		{"unsupported precision", TokenBucketRule{Limit: 1, Interval: time.Nanosecond, Burst: 1}},
		{"unsupported capacity", TokenBucketRule{Limit: 1, Interval: time.Hour, Burst: 1000000}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := handler.NewLimiter("quota", tc.rule)
			require.Error(t, err)
		})
	}
	_, err := handler.NewLimiter("", rule)
	require.Error(t, err)
	for _, unsupported := range []Rule{nil, NoneRule{}, (*TokenBucketRule)(nil), &rule} {
		_, err := handler.NewLimiter("quota", unsupported)
		require.Error(t, err)
	}
}

func TestMemoryLimiterSharedQuota(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := t.Context()
		handler := NewMemory()
		rule := TokenBucketRule{Limit: 2, Interval: time.Second, Burst: 2}
		first, err := handler.NewLimiter("quota", rule)
		require.NoError(t, err)
		require.NoError(t, first.Wait(ctx))
		second, err := handler.NewLimiter("quota", rule)
		require.NoError(t, err)
		allowed, err := second.Allow(ctx)
		require.NoError(t, err)
		require.True(t, allowed)
		allowed, err = first.Allow(ctx)
		require.NoError(t, err)
		require.False(t, allowed)

		_, err = handler.NewLimiter("quota", TokenBucketRule{Limit: 1, Interval: time.Second, Burst: 2})
		require.ErrorIs(t, err, ErrRuleConflict)
		other, err := NewMemory().NewLimiter("quota", rule)
		require.NoError(t, err)
		allowed, err = other.Allow(ctx)
		require.NoError(t, err)
		require.True(t, allowed)

		time.Sleep(500*time.Millisecond - time.Microsecond)
		allowed, err = first.Allow(ctx)
		require.NoError(t, err)
		require.False(t, allowed)
		start := time.Now()
		require.NoError(t, second.Wait(ctx))
		require.Equal(t, time.Microsecond, time.Since(start))
		time.Sleep(time.Hour)
		for range rule.Burst {
			allowed, err = first.Allow(ctx)
			require.NoError(t, err)
			require.True(t, allowed)
		}
		allowed, err = first.Allow(ctx)
		require.NoError(t, err)
		require.False(t, allowed)
	})
}

func TestMemoryLimiterConcurrentQuota(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		limiter, err := NewMemory().NewLimiter("quota", TokenBucketRule{Limit: 1, Interval: time.Second, Burst: 5})
		require.NoError(t, err)
		var count atomic.Int64
		var wg sync.WaitGroup
		for range 50 {
			wg.Go(func() {
				allowed, err := limiter.Allow(t.Context())
				if err != nil {
					t.Error(err)
					return
				}
				if allowed {
					count.Add(1)
				}
			})
		}
		wg.Wait()
		require.EqualValues(t, 5, count.Load())
	})
}

func TestMemoryLimiterWaitCancellation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		limiter, err := NewMemory().NewLimiter("quota", TokenBucketRule{Limit: 1, Interval: time.Second, Burst: 1})
		require.NoError(t, err)
		require.NoError(t, limiter.Wait(t.Context()))
		ctx, cancel := context.WithTimeout(t.Context(), 200*time.Millisecond)
		defer cancel()
		require.ErrorIs(t, limiter.Wait(ctx), context.DeadlineExceeded)
		ctx, cancel = context.WithCancel(t.Context())
		result := make(chan error, 1)
		go func() { result <- limiter.Wait(ctx) }()
		synctest.Wait()
		cancel()
		require.ErrorIs(t, <-result, context.Canceled)
		time.Sleep(800 * time.Millisecond)
		_, err = limiter.Allow(ctx)
		require.ErrorIs(t, err, context.Canceled)
		allowed, err := limiter.Allow(t.Context())
		require.NoError(t, err)
		require.True(t, allowed)
	})
}
