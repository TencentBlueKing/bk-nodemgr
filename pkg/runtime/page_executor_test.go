/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package runtime ...
package runtime

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// TestExecutor_Execute ...
func TestExecutor_Execute(t *testing.T) {
	type args[T any] struct {
		ctx context.Context
		req types.Page
		fn  PageExecutorFn[T]
	}
	type testCase[T any] struct {
		name    string
		e       *PageExecutor[T]
		args    args[T]
		want    *PageResult[T]
		wantErr bool
	}

	successFn := func(ctx context.Context, p types.Page) ([]int, error) {
		res := make([]int, 0, p.Limit)
		for i := p.Offset; i < p.Offset+p.Limit; i++ {
			res = append(res, i)
		}
		return res, nil
	}

	timeoutFn := func(ctx context.Context, p types.Page) ([]int, error) {
		time.Sleep(2 * time.Second)
		return successFn(ctx, p)
	}

	canceledCtx := func() context.Context {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		return ctx
	}

	tests := []testCase[int]{
		{
			name: "over max page size",
			e:    NewPageExecutor[int](10, 30*time.Second),
			args: args[int]{
				ctx: context.Background(),
				req: types.Page{Offset: 0, Limit: 11},
				fn:  successFn,
			},
			want: &PageResult[int]{Items: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, Total: 11},
		},
		{
			name: "less than max page size",
			e:    NewPageExecutor[int](10, 30*time.Second),
			args: args[int]{
				ctx: context.Background(),
				req: types.Page{Offset: 0, Limit: 9},
				fn:  successFn,
			},
			want: &PageResult[int]{Items: []int{0, 1, 2, 3, 4, 5, 6, 7, 8}, Total: 9},
		},
		{
			name: "execute timeout",
			e:    NewPageExecutor[int](10, 1*time.Second),
			args: args[int]{
				ctx: context.Background(),
				req: types.Page{Offset: 0, Limit: 11},
				fn:  timeoutFn,
			},
			wantErr: true,
		},
		{
			name: "execute error",
			e:    NewPageExecutor[int](10, 30*time.Second),
			args: args[int]{
				ctx: context.Background(),
				req: types.Page{Offset: 0, Limit: 11},
				fn:  func(context.Context, types.Page) ([]int, error) { return nil, errors.New("test error") },
			},
			wantErr: true,
		},
		{
			name: "first item",
			e:    NewPageExecutor[int](10, 30*time.Second),
			args: args[int]{
				ctx: context.Background(),
				req: types.Page{Offset: 0, Limit: 1},
				fn:  successFn,
			},
			want: &PageResult[int]{Items: []int{0}, Total: 1},
		},
		{
			name: "pagination",
			e:    NewPageExecutor[int](10, 30*time.Second),
			args: args[int]{
				ctx: context.Background(),
				req: types.Page{Offset: 5, Limit: 5},
				fn:  successFn,
			},
			want: &PageResult[int]{Items: []int{5, 6, 7, 8, 9}, Total: 5},
		},
		{
			name: "context canceled",
			e:    NewPageExecutor[int](10, 30*time.Second),
			args: args[int]{
				ctx: canceledCtx(),
				req: types.Page{Offset: 0, Limit: 10},
				fn:  successFn,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.e.Execute(tt.args.ctx, tt.args.req, tt.args.fn)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Execute() got = %v, want %v", got, tt.want)
			}
		})
	}
}
