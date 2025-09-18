/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package pageexecutor ...
package pageexecutor

import (
	"context"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// PageResult ...
type PageResult[T any] struct {
	Items []T
	Total int
}

// PageExecutor page executor.
type PageExecutor[T any] struct {
	maxPageSize int
	timeout     time.Duration
}

// NewPageExecutor ...
func NewPageExecutor[T any](maxPageSize int, timeout time.Duration) *PageExecutor[T] {
	if maxPageSize <= 0 {
		maxPageSize = 500
	}

	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return &PageExecutor[T]{
		maxPageSize: maxPageSize,
		timeout:     timeout,
	}
}

// PageExecutorFn ...
type PageExecutorFn[T any] func(ctx context.Context, p types.Page) ([]T, error)

// Execute page query.
func (e *PageExecutor[T]) Execute(ctx context.Context, p types.Page, fn PageExecutorFn[T]) (*PageResult[T], error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	offset := p.Offset
	end := p.Offset + p.Limit
	result := &PageResult[T]{
		Items: make([]T, 0, min(p.Limit, e.maxPageSize)),
		Total: 0,
	}
	for offset < end {
		select {
		case <-ctx.Done():
			{
				return nil, ctx.Err()
			}
		default:
		}

		currentLimit := min(end-offset, e.maxPageSize)
		pageReq := types.Page{
			Offset: offset,
			Limit:  currentLimit,
			Sort:   p.Sort,
		}

		items, err := fn(ctx, pageReq)
		if err != nil {
			return nil, err
		}

		result.Items = append(result.Items, items...)

		if len(items) < currentLimit {
			break
		}

		offset += currentLimit
	}

	result.Total = len(result.Items)

	return result, nil
}
