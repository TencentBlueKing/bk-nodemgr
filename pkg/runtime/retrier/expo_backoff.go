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

// Package retrier ...
package retrier

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// ExpoBackoffOpts the options for retrying.
type ExpoBackoffOpts struct {
	// maximum number of retries.
	MaxRetries int

	// base delay when retrying.
	BaseDelay time.Duration

	// max delay when retrying.
	MaxDelay time.Duration

	// random jitter added to the retry delay.
	JitterPercent float64
}

// ExpoBackoffOptsDefault default retry options.
func ExpoBackoffOptsDefault() ExpoBackoffOpts {
	return ExpoBackoffOpts{
		MaxRetries:    3,
		BaseDelay:     time.Second,
		MaxDelay:      5 * time.Second,
		JitterPercent: 0.2,
	}
}

// ExpoBackoffDelayOpts options for exponential delay calculation.
type ExpoBackoffDelayOpts struct {
	// BaseDelay is the base delay before exponential scaling.
	BaseDelay time.Duration

	// MaxDelay is the maximum delay cap.
	MaxDelay time.Duration

	// JitterPercent is the random jitter percentage added to the delay.
	JitterPercent float64
}

// ExpoBackoffDelayOptsDefault returns default delay options.
func ExpoBackoffDelayOptsDefault() ExpoBackoffDelayOpts {
	return ExpoBackoffDelayOpts{
		BaseDelay:     time.Second,
		MaxDelay:      5 * time.Second,
		JitterPercent: 0.2,
	}
}

// CalculateExpoDelay calculates exponential delay with jitter for a given attempt.
// The delay formula is: min(BaseDelay * 2^attempt, MaxDelay) + jitter.
func CalculateExpoDelay(attempt int, opts ExpoBackoffDelayOpts) time.Duration {
	delay := float64(opts.BaseDelay) * math.Pow(2, float64(attempt))
	if delay > float64(opts.MaxDelay) {
		delay = float64(opts.MaxDelay)
	}

	jitter := delay * rand.Float64() * opts.JitterPercent // nolint: gosec
	return time.Duration(delay + jitter)
}

// ExpoBackoff the exponential backoff retryer.
type ExpoBackoff struct {
	opts ExpoBackoffOpts
}

// NewExpoBackoff new an exponential backoff retryer.
func NewExpoBackoff(opts ExpoBackoffOpts) *ExpoBackoff {
	retrier := &ExpoBackoff{
		opts: opts,
	}

	return retrier
}

// calculateDelay calculate the delay time
func (e *ExpoBackoff) calculateDelay(attempt int) time.Duration {
	return CalculateExpoDelay(attempt, ExpoBackoffDelayOpts{
		BaseDelay:     e.opts.BaseDelay,
		MaxDelay:      e.opts.MaxDelay,
		JitterPercent: e.opts.JitterPercent,
	})
}

// Do do the fn.
// nolint: varnamelen
func (e *ExpoBackoff) Do(ctx context.Context, fn func(attempt int) error) error {
	var err error

	for attempt := 0; attempt < e.opts.MaxRetries; attempt++ {
		// if context cancelled, return.
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("fn cancelled: %v", err)
		}

		err = fn(attempt)
		if err == nil {
			break
		}

		// last attempt, don't sleep and log.
		if attempt == e.opts.MaxRetries-1 {
			break
		}

		delay := e.calculateDelay(attempt)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	if err != nil {
		return fmt.Errorf("fn all failed, max-retries(%d), last-err(%v)", e.opts.MaxRetries, err)
	}

	return nil
}
