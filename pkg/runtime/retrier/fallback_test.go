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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFallback_FirstSuccess(t *testing.T) {
	candidates := []string{"a", "b", "c"}

	f := NewFallback(candidates, nil, FallbackOpts{})

	callCount := 0
	calledCandidates := []string{}

	err := f.Do(context.Background(), func(candidate string) error {
		callCount++
		calledCandidates = append(calledCandidates, candidate)
		return nil // First one succeeds
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, callCount, "should only call once when first succeeds")
	assert.Equal(t, []string{"a"}, calledCandidates)
}

func TestFallback_SecondSuccess(t *testing.T) {
	candidates := []string{"a", "b", "c"}

	f := NewFallback(candidates, nil, FallbackOpts{})

	callCount := 0
	calledCandidates := []string{}

	err := f.Do(context.Background(), func(candidate string) error {
		callCount++
		calledCandidates = append(calledCandidates, candidate)
		if candidate == "a" {
			return errors.New("first failed")
		}
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 2, callCount, "should call twice when first fails")
	assert.Equal(t, []string{"a", "b"}, calledCandidates)
}

func TestFallback_AllFail(t *testing.T) {
	candidates := []string{"a", "b"}
	f := NewFallback(candidates, nil, FallbackOpts{})

	callCount := 0
	lastErr := errors.New("candidate-b failed")

	err := f.Do(context.Background(), func(candidate string) error {
		callCount++
		if candidate == "a" {
			return errors.New("candidate-a failed")
		}
		return lastErr
	})

	assert.Error(t, err)
	assert.Equal(t, 2, callCount, "should try all candidates")
	assert.Contains(t, err.Error(), "all candidates failed")
	assert.ErrorIs(t, err, lastErr)
}

func TestFallback_WithValidator(t *testing.T) {
	// Candidates: "invalid1", "invalid2", "valid"
	candidates := []string{"invalid1", "invalid2", "valid"}

	// Validator: only "valid" passes
	validator := func(s string) bool {
		return s == "valid"
	}

	f := NewFallback(candidates, validator, FallbackOpts{})

	callCount := 0
	calledCandidates := []string{}

	err := f.Do(context.Background(), func(candidate string) error {
		callCount++
		calledCandidates = append(calledCandidates, candidate)
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, callCount, "should only call valid candidate")
	assert.Equal(t, []string{"valid"}, calledCandidates)
}

func TestFallback_NoValidCandidates(t *testing.T) {
	candidates := []string{"a", "b"}

	// All candidates are invalid
	validator := func(s string) bool {
		return false
	}

	f := NewFallback(candidates, validator, FallbackOpts{})

	callCount := 0

	err := f.Do(context.Background(), func(candidate string) error {
		callCount++
		return nil
	})

	assert.Error(t, err)
	assert.Equal(t, 0, callCount, "should not call any candidate")
	assert.Contains(t, err.Error(), "no valid candidates found")
}

func TestFallback_EmptyCandidates(t *testing.T) {
	f := NewFallback([]string{}, nil, FallbackOpts{})

	err := f.Do(context.Background(), func(candidate string) error {
		return nil
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no candidates provided")
}

func TestFallback_Callbacks(t *testing.T) {
	// "skip" will be skipped, "fail" will fail, "success" will succeed
	candidates := []string{"skip", "fail", "success"}

	validator := func(s string) bool {
		return s != "skip"
	}

	var skipped []int
	var attempted []int
	var errored []int
	var succeeded []int

	f := NewFallback(candidates, validator, FallbackOpts{
		OnSkip: func(index int) {
			skipped = append(skipped, index)
		},
		OnAttempt: func(index int, total int) {
			attempted = append(attempted, index)
		},
		OnError: func(index int, total int, err error) {
			errored = append(errored, index)
		},
		OnSuccess: func(index int) {
			succeeded = append(succeeded, index)
		},
	})

	err := f.Do(context.Background(), func(candidate string) error {
		if candidate == "fail" {
			return errors.New("failed")
		}
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, []int{0}, skipped)
	assert.Equal(t, []int{1, 2}, attempted)
	assert.Equal(t, []int{1}, errored)
	assert.Equal(t, []int{2}, succeeded)
}

func TestFallback_ContextCancelled(t *testing.T) {
	candidates := []string{"a", "b", "c"}
	f := NewFallback(candidates, nil, FallbackOpts{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := f.Do(ctx, func(candidate string) error {
		return nil
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context cancelled")
}

func TestFallback_NilValidator(t *testing.T) {
	candidates := []string{"a"}

	// nil validator means all candidates are valid
	f := NewFallback(candidates, nil, FallbackOpts{})

	called := false

	err := f.Do(context.Background(), func(candidate string) error {
		called = true
		return nil
	})

	assert.NoError(t, err)
	assert.True(t, called)
}

// TestFallback_WithStructCandidates demonstrates using Fallback with struct candidates.
func TestFallback_WithStructCandidates(t *testing.T) {
	type RelayInfo struct {
		HostID int64
	}

	candidates := []*RelayInfo{
		nil,           // Invalid: nil
		{HostID: 0},   // Invalid: HostID = 0
		{HostID: 100}, // Valid
	}

	// Single validator function for all candidates
	validator := func(r *RelayInfo) bool {
		return r != nil && r.HostID != 0
	}

	f := NewFallback(candidates, validator, FallbackOpts{})

	var usedCandidate *RelayInfo

	err := f.Do(context.Background(), func(candidate *RelayInfo) error {
		usedCandidate = candidate
		return nil
	})

	assert.NoError(t, err)
	assert.NotNil(t, usedCandidate)
	assert.Equal(t, int64(100), usedCandidate.HostID)
}

// TestFallback_IntCandidates tests with integer candidates.
func TestFallback_IntCandidates(t *testing.T) {
	candidates := []int{1, 2, 3, 4, 5}

	// Only even numbers are valid
	validator := func(n int) bool {
		return n%2 == 0
	}

	f := NewFallback(candidates, validator, FallbackOpts{})

	var result int

	err := f.Do(context.Background(), func(candidate int) error {
		result = candidate
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 2, result, "should use first even number")
}
