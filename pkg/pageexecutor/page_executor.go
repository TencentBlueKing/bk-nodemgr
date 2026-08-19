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

// Package pageexecutor ...
package pageexecutor

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
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
type PageExecutorFn[T any] func(nCtx contextx.IContext, p types.Page) ([]T, error)

// Execute page query.
func (e *PageExecutor[T]) Execute(nCtx contextx.IContext, p types.Page, fn PageExecutorFn[T]) (*PageResult[T], error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	nCtx, cancel := contextx.WithTimeout(nCtx, e.timeout)
	defer cancel()

	offset := p.Offset
	end := p.Offset + p.Limit
	result := &PageResult[T]{
		Items: make([]T, 0, min(p.Limit, e.maxPageSize)),
		Total: 0,
	}
	for offset < end {
		select {
		case <-nCtx.Done():
			{
				return nil, nCtx.Err()
			}
		default:
		}

		currentLimit := min(end-offset, e.maxPageSize)
		pageReq := types.Page{
			Offset: offset,
			Limit:  currentLimit,
			Sort:   p.Sort,
		}

		items, err := fn(nCtx, pageReq)
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
