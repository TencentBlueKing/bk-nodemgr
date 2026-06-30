/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package batchexecutor executes slice items in fixed-size batches.
package batchexecutor

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

const (
	defaultBatchSize = 500
	defaultTimeout   = 30 * time.Second
)

type option struct {
	batchSize int
	timeout   time.Duration
}

// Option configures batch execution.
type Option func(*option)

// Result describes collected batch execution results.
type Result[T any] struct {
	Items []T
	Total int
}

// ExecuteFn handles one batch of items.
type ExecuteFn[T any] func(nCtx contextx.IContext, items []T) error

// CollectFn handles one input batch and returns collected output items.
type CollectFn[In any, Out any] func(nCtx contextx.IContext, items []In) ([]Out, error)

// WithBatchSize configures the maximum item count per batch.
func WithBatchSize(batchSize int) Option {
	return func(opt *option) {
		opt.batchSize = batchSize
	}
}

// WithTimeout configures the timeout for the whole batch execution.
func WithTimeout(timeout time.Duration) Option {
	return func(opt *option) {
		opt.timeout = timeout
	}
}

// Execute executes items in fixed-size batches.
func Execute[T any](nCtx contextx.IContext, items []T, fn ExecuteFn[T], opts ...Option) error {
	opt := buildOption(opts...)

	nCtx, cancel := contextx.WithTimeout(nCtx, opt.timeout)
	defer cancel()

	return eachBatch(nCtx, items, opt.batchSize, func(batch []T) error {
		return fn(nCtx, batch)
	})
}

// Collect executes items in fixed-size batches and collects returned items.
func Collect[In any, Out any](
	nCtx contextx.IContext, items []In, fn CollectFn[In, Out], opts ...Option,
) (*Result[Out], error) {

	opt := buildOption(opts...)

	nCtx, cancel := contextx.WithTimeout(nCtx, opt.timeout)
	defer cancel()

	result := &Result[Out]{
		Items: make([]Out, 0, len(items)),
		Total: 0,
	}
	err := eachBatch(nCtx, items, opt.batchSize, func(batch []In) error {
		batchItems, err := fn(nCtx, batch)
		if err != nil {
			return err
		}

		result.Items = append(result.Items, batchItems...)

		return nil
	})
	if err != nil {
		return nil, err
	}

	result.Total = len(result.Items)

	return result, nil
}

func eachBatch[T any](
	nCtx contextx.IContext, items []T, batchSize int, fn func(batch []T) error,
) error {

	for start := 0; start < len(items); start += batchSize {
		select {
		case <-nCtx.Done():
			return nCtx.Err()
		default:
		}

		end := min(start+batchSize, len(items))
		if err := fn(items[start:end]); err != nil {
			return err
		}
	}

	return nil
}

func buildOption(opts ...Option) option {
	opt := option{
		batchSize: defaultBatchSize,
		timeout:   defaultTimeout,
	}
	for _, fn := range opts {
		fn(&opt)
	}
	opt.batchSize = normalizeBatchSize(opt.batchSize)
	opt.timeout = normalizeTimeout(opt.timeout)

	return opt
}

func normalizeBatchSize(batchSize int) int {
	if batchSize > 0 {
		return batchSize
	}

	return defaultBatchSize
}

func normalizeTimeout(timeout time.Duration) time.Duration {
	if timeout > 0 {
		return timeout
	}

	return defaultTimeout
}
