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

// Package ratelimit defines portable rate limiters and built-in implementations.
package ratelimit

import (
	"context"
	"errors"
)

// ErrRuleConflict indicates that an existing quota uses a different rule.
var ErrRuleConflict = errors.New("rate limiter rule conflict")

// IHandler creates limiters within a shared quota scope. Creating a limiter
// with the same name and rule does not reset its quota.
// Remote implementations may detect cross-process rule conflicts on first use.
type IHandler interface {
	NewLimiter(name string, rule Rule) (ILimiter, error)
}

// ILimiter grants permission according to its configured rule. It is safe for concurrent use.
// A successful limited operation consumes one unit of quota; unrestricted rules consume none.
// An ambiguous remote failure may consume quota without returning permission.
type ILimiter interface {
	// Allow attempts to acquire permission without waiting for quota recovery.
	// A false result with no error indicates normal rate limiting.
	Allow(ctx context.Context) (bool, error)

	// Wait acquires permission, waiting for quota recovery until ctx is cancelled
	// or its deadline expires. It does not guarantee FIFO ordering.
	Wait(ctx context.Context) error
}

// Algorithm identifies a rate limiting policy, independently of its storage backend.
type Algorithm string

const (
	// AlgorithmNone allows requests without imposing a rate limit.
	AlgorithmNone Algorithm = "none"
	// AlgorithmTokenBucket refills a bounded pool of tokens continuously.
	AlgorithmTokenBucket Algorithm = "token_bucket"
)

// Rule describes an algorithm-specific, immutable configuration.
// Built-in handlers accept supported, comparable concrete rule values, not pointers.
// Unsupported rules must return an error rather than fall back to another algorithm.
type Rule interface {
	Algorithm() Algorithm
	Validate() error
}
