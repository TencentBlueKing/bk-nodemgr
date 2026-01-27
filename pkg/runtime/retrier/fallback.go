/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package retrier

import (
	"context"
	"fmt"
)

// FallbackOpts contains options for the Fallback retrier.
type FallbackOpts struct {
	// OnSkip is called when a candidate is skipped due to validation failure.
	// Parameters: index.
	OnSkip func(index int)

	// OnAttempt is called before each attempt.
	// Parameters: index, total count.
	OnAttempt func(index int, total int)

	// OnError is called when an attempt fails.
	// Parameters: index, total count, error.
	OnError func(index int, total int, err error)

	// OnSuccess is called when an attempt succeeds.
	// Parameters: index.
	OnSuccess func(index int)
}

// Fallback tries each candidate in sequence until one succeeds.
// It encapsulates the candidate list and provides type-safe access via generics.
// The fn receives the candidate directly, allowing clean and intuitive usage.
type Fallback[T any] struct {
	candidates []T
	validator  func(T) bool
	opts       FallbackOpts
}

// NewFallback creates a new Fallback retrier with the given candidates.
// candidates: the list of candidates to try in order.
// validator: optional validation function applied to each candidate (nil means all valid).
// opts: callback options for monitoring the fallback process.
func NewFallback[T any](candidates []T, validator func(T) bool, opts FallbackOpts) *Fallback[T] {
	return &Fallback[T]{
		candidates: candidates,
		validator:  validator,
		opts:       opts,
	}
}

// Do tries each candidate until one succeeds.
// The fn receives the candidate directly (not an index).
// Returns nil on success, or an error if all candidates fail.
func (f *Fallback[T]) Do(ctx context.Context, fn func(candidate T) error) error {
	total := len(f.candidates)
	if total == 0 {
		return fmt.Errorf("no candidates provided")
	}

	var lastErr error
	attempts := 0

	for i, candidate := range f.candidates {
		// Check context
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context cancelled: %w", err)
		}

		// Validate candidate if validator is provided
		if f.validator != nil && !f.validator(candidate) {
			if f.opts.OnSkip != nil {
				f.opts.OnSkip(i)
			}

			continue
		}

		attempts++

		// Notify attempt
		if f.opts.OnAttempt != nil {
			f.opts.OnAttempt(i, total)
		}

		// Try the candidate
		err := fn(candidate)
		if err != nil {
			lastErr = err
			if f.opts.OnError != nil {
				f.opts.OnError(i, total, err)
			}

			continue
		}

		// Success
		if f.opts.OnSuccess != nil {
			f.opts.OnSuccess(i)
		}

		return nil
	}

	// All candidates failed
	if lastErr != nil {
		return fmt.Errorf("all candidates failed. count(%d), attempts(%d): %w",
			total, attempts, lastErr)
	}

	return fmt.Errorf("no valid candidates found. count(%d)", total)
}
