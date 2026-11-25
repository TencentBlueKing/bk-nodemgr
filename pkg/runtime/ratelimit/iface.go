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

// Package ratelimit defines portable token bucket limiters and their in-memory implementation.
package ratelimit

import (
	"context"
	"errors"
	"time"
)

// ErrRuleConflict indicates that an existing quota uses a different rule.
var ErrRuleConflict = errors.New("rate limiter rule conflict")

// IHandler creates limiters within a shared quota scope. Creating a limiter
// with the same name and rule does not reset its quota.
// Remote implementations may detect cross-process rule conflicts on first use.
type IHandler interface {
	NewLimiter(name string, rule Rule) (ILimiter, error)
}

// ILimiter acquires one token per successful operation. It is safe for concurrent use.
// An ambiguous remote failure may consume quota without returning permission.
type ILimiter interface {
	// Allow attempts to acquire a token without waiting for quota recovery.
	// A false result with no error indicates normal rate limiting.
	Allow(ctx context.Context) (bool, error)

	// Wait acquires a token, waiting for quota recovery until ctx is cancelled
	// or its deadline expires. It does not guarantee FIFO ordering.
	Wait(ctx context.Context) error
}

// Rule defines a token bucket that refills Limit tokens per Interval, up to Burst.
// New buckets start full. Rules cannot change during the lifetime of a limiter.
type Rule struct {
	Limit    int
	Interval time.Duration
	Burst    int
}

// maxCredit bounds integer arithmetic to a range exactly representable in Redis Lua.
// One token costs Interval.Microseconds() credits; each microsecond adds Limit credits.
const maxCredit = 1 << 40

// Validate checks that the rule is representable at microsecond precision in both backends.
func (rule Rule) Validate() error {
	if rule.Limit <= 0 || rule.Burst <= 0 || rule.Interval <= 0 {
		return errors.New("limit, interval and burst must be positive")
	}
	if rule.Interval%time.Microsecond != 0 {
		return errors.New("interval must be a whole number of microseconds")
	}
	if int64(rule.Limit) > maxCredit || int64(rule.Burst) > maxCredit/rule.Interval.Microseconds() {
		return errors.New("rate limiter rule exceeds supported credit range")
	}

	return nil
}
