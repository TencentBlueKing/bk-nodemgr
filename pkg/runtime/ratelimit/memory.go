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
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	_ IHandler = (*MemoryHandler)(nil)
	_ ILimiter = (*tokenBucketLimiter)(nil)
)

// MemoryHandler creates process-local limiters. Each handler owns an independent quota scope.
// Entries are retained for the lifetime of the handler; use stable, bounded names.
type MemoryHandler struct {
	mu       sync.Mutex
	limiters map[string]memoryEntry
}

type memoryEntry struct {
	rule    Rule
	limiter ILimiter
}

// NewMemory creates an independent process-local limiter factory.
func NewMemory() *MemoryHandler {
	return &MemoryHandler{limiters: make(map[string]memoryEntry)}
}

// NewLimiter returns a limiter for the named quota without resetting its state.
func (handler *MemoryHandler) NewLimiter(name string, rule Rule) (ILimiter, error) {
	if name == "" {
		return nil, errors.New("limiter name is empty")
	}
	limiter, err := newMemoryLimiter(rule)
	if err != nil {
		return nil, err
	}

	handler.mu.Lock()
	defer handler.mu.Unlock()

	if entry, exists := handler.limiters[name]; exists {
		if entry.rule != rule {
			return nil, fmt.Errorf("limiter %q: %w", name, ErrRuleConflict)
		}

		return entry.limiter, nil
	}

	handler.limiters[name] = memoryEntry{rule: rule, limiter: limiter}

	return limiter, nil
}

func newMemoryLimiter(rule Rule) (ILimiter, error) {
	switch config := rule.(type) {
	case TokenBucketRule:
		if err := config.Validate(); err != nil {
			return nil, fmt.Errorf("invalid limiter rule: %w", err)
		}

		return &tokenBucketLimiter{
			rule:    config,
			credits: int64(config.Burst) * config.Interval.Microseconds(),
			last:    time.Now(),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported memory limiter rule %T", rule)
	}
}

type tokenBucketLimiter struct {
	rule    TokenBucketRule
	mu      sync.Mutex
	credits int64
	last    time.Time
}

func (limiter *tokenBucketLimiter) Allow(ctx context.Context) (bool, error) {
	delay, err := limiter.acquire(ctx)
	if err != nil {
		return false, err
	}

	return delay == 0, nil
}

func (limiter *tokenBucketLimiter) Wait(ctx context.Context) error {
	for {
		delay, err := limiter.acquire(ctx)
		if err != nil {
			return err
		}
		if delay == 0 {
			return nil
		}

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()

			return ctx.Err()
		case <-timer.C:
		}
	}
}

// acquire returns zero after consuming one token, or a positive delay without consuming it.
func (limiter *tokenBucketLimiter) acquire(ctx context.Context) (time.Duration, error) {
	if ctx == nil {
		return 0, errors.New("context is nil")
	}
	limiter.mu.Lock()
	defer limiter.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	now := time.Now()
	cost := limiter.rule.Interval.Microseconds()
	capacity := cost * int64(limiter.rule.Burst)
	refill := int64(limiter.rule.Limit)
	elapsed := now.Sub(limiter.last).Microseconds()

	// Cap elapsed before multiplying to avoid overflow during long idle periods.
	if elapsed >= (capacity-limiter.credits+refill-1)/refill {
		limiter.credits = capacity
		limiter.last = now
	} else {
		limiter.credits += elapsed * refill
		limiter.last = limiter.last.Add(time.Duration(elapsed) * time.Microsecond)
	}
	if limiter.credits < cost {
		// Round up so a waiter does not retry before the next full token is available.
		return time.Duration((cost-limiter.credits+refill-1)/refill) * time.Microsecond, nil
	}

	limiter.credits -= cost

	return 0, nil
}
