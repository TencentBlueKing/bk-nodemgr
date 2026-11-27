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
	"errors"
	"time"
)

// TokenBucketRule refills Limit tokens per Interval, up to Burst.
// New buckets start full.
type TokenBucketRule struct {
	Limit    int
	Interval time.Duration
	Burst    int
}

// Algorithm identifies the token bucket algorithm.
func (rule TokenBucketRule) Algorithm() Algorithm {
	return AlgorithmTokenBucket
}

// maxCredit bounds integer arithmetic to a range exactly representable in Redis Lua.
// One token costs Interval.Microseconds() credits; each microsecond adds Limit credits.
const maxCredit = 1 << 40

// Validate checks that the rule is representable at microsecond precision in both backends.
func (rule TokenBucketRule) Validate() error {
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

// NoneRule explicitly disables rate limiting.
type NoneRule struct{}

// Algorithm identifies the unrestricted algorithm.
func (rule NoneRule) Algorithm() Algorithm {
	return AlgorithmNone
}

// Validate accepts the parameterless rule.
func (rule NoneRule) Validate() error {
	return nil
}
