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
	"errors"
	"time"
)

// PollingOpts the options for retrying.
type PollingOpts struct {
	// Timeout is the timeout for the retry.
	Timeout time.Duration

	// Interval is the interval between retries.
	Interval time.Duration
}

// Polling the polling retryer.
type Polling struct {
	opts PollingOpts
}

// PollingOptsDefault default retry options.
// nolint: mnd
func PollingOptsDefault() PollingOpts {
	return PollingOpts{
		Timeout:  30 * time.Second,
		Interval: 1 * time.Second,
	}
}

// NewPolling new a polling retryer.
func NewPolling(opts PollingOpts) *Polling {
	retrier := &Polling{
		opts: opts,
	}

	return retrier
}

// Do define the retrying logic.
// nolint: varnamelen
func (p *Polling) Do(ctx context.Context, fn func(attempt int) error) error {
	attempt := 0

	timer := time.NewTimer(p.opts.Timeout)
	defer timer.Stop()

	ticker := time.NewTicker(p.opts.Interval)
	defer ticker.Stop()

	err := fn(attempt)
	if err == nil {
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			return errors.New("reach max timeout")
		case <-ticker.C:
			attempt++
			err := fn(attempt)
			if err == nil {
				return nil
			}
		}
	}
}
