/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package retrier ...
package retrier

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
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

	// Logger ...
	Logger logger.ILogger
}

// ExpoBackoffOptsDefault default retry options.
func ExpoBackoffOptsDefault() ExpoBackoffOpts {
	return ExpoBackoffOpts{
		MaxRetries:    3,
		BaseDelay:     time.Second,
		MaxDelay:      5 * time.Second,
		JitterPercent: 0.2,
		Logger:        logger.LoggerDefault{},
	}
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
// nolint: varnamelen
func (e *ExpoBackoff) calculateDelay(attempt int) time.Duration {
	// cal base delay.
	delay := float64(e.opts.BaseDelay) * math.Pow(2, float64(attempt))
	if delay > float64(e.opts.MaxDelay) {
		delay = float64(e.opts.MaxDelay)
	}

	// add jitter.
	jitter := delay * rand.Float64() * e.opts.JitterPercent
	finalDelay := time.Duration(delay + jitter)

	return finalDelay
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
		e.opts.Logger.Warnf("fn failed, attempt(%d/%d), retry-after(%vs): %v.",
			attempt+1, e.opts.MaxRetries, delay.Seconds(), err)

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
