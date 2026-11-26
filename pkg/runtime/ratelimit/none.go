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
)

var (
	_ IHandler = NoneHandler{}
	_ ILimiter = NoneLimiter{}
)

// NoneHandler creates unrestricted limiters without allocating quota state.
// It accepts only an explicit NoneRule, never a finite rate rule.
type NoneHandler struct{}

// NewNone creates an unrestricted limiter factory.
func NewNone() NoneHandler {
	return NoneHandler{}
}

// NewLimiter creates an unrestricted limiter for a valid name and NoneRule.
func (handler NoneHandler) NewLimiter(name string, rule Rule) (ILimiter, error) {
	if name == "" {
		return nil, errors.New("limiter name is empty")
	}
	if _, ok := rule.(NoneRule); !ok {
		return nil, fmt.Errorf("unsupported none limiter rule %T", rule)
	}

	return NoneLimiter{}, nil
}

// NoneLimiter grants permission without rate limiting or maintaining state.
// The zero value is ready for use.
type NoneLimiter struct{}

// Allow grants permission immediately unless the context is invalid or cancelled.
func (limiter NoneLimiter) Allow(ctx context.Context) (bool, error) {
	if err := limiter.Wait(ctx); err != nil {
		return false, err
	}

	return true, nil
}

// Wait grants permission without waiting for quota.
func (limiter NoneLimiter) Wait(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is nil")
	}

	return ctx.Err()
}
