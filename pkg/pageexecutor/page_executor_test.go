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
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// TestExecutor_Execute ...
func TestExecutor_Execute(t *testing.T) {
	type args[T any] struct {
		ctx contextx.IContext
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
	tests := []testCase[int]{
		{
			name: "over max page size",
			e:    NewPageExecutor[int](10, 30*time.Second),
			args: args[int]{
				ctx: contextx.Background(),
				req: types.Page{
					Offset: 0,
					Limit:  11,
					Sort:   "",
				},
				fn: func(ctx contextx.IContext, p types.Page) ([]int, error) {
					result := make([]int, 0)
					for i := p.Offset; i < p.Offset+p.Limit; i++ {
						result = append(result, i)
					}

					return result, nil
				},
			},
			want: &PageResult[int]{
				Items: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				Total: 11,
			},
			wantErr: false,
		},
		{
			name: "less than max page size",
			e:    NewPageExecutor[int](10, 30*time.Second),
			args: args[int]{
				ctx: contextx.Background(),
				req: types.Page{
					Offset: 0,
					Limit:  9,
					Sort:   "",
				},
				fn: func(ctx contextx.IContext, p types.Page) ([]int, error) {
					result := make([]int, 0)
					for i := p.Offset; i < p.Offset+p.Limit; i++ {
						result = append(result, i)
					}

					return result, nil
				},
			},
			want: &PageResult[int]{
				Items: []int{0, 1, 2, 3, 4, 5, 6, 7, 8},
				Total: 9,
			},
			wantErr: false,
		},
		{
			name: "execute timeout",
			e:    NewPageExecutor[int](10, 1*time.Second),
			args: args[int]{
				ctx: contextx.Background(),
				req: types.Page{
					Offset: 0,
					Limit:  11,
					Sort:   "",
				},
				fn: func(ctx contextx.IContext, p types.Page) ([]int, error) {
					time.Sleep(2 * time.Second)

					result := make([]int, 0)
					for i := p.Offset; i < p.Offset+p.Limit; i++ {
						result = append(result, i)
					}

					return result, nil
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "execute error",
			e:    NewPageExecutor[int](10, 30*time.Second),
			args: args[int]{
				ctx: contextx.Background(),
				req: types.Page{
					Offset: 0,
					Limit:  11,
					Sort:   "",
				},
				fn: func(ctx contextx.IContext, p types.Page) ([]int, error) {
					return nil, errors.New("test error")
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "first item",
			e:    NewPageExecutor[int](10, 30*time.Second),
			args: args[int]{
				ctx: contextx.Background(),
				req: types.Page{
					Offset: 0,
					Limit:  1,
					Sort:   "",
				},
				fn: func(ctx contextx.IContext, p types.Page) ([]int, error) {
					result := make([]int, 0)
					for i := p.Offset; i < p.Offset+p.Limit; i++ {
						result = append(result, i)
					}

					return result, nil
				},
			},
			want: &PageResult[int]{
				Items: []int{0},
				Total: 1,
			},
			wantErr: false,
		},
		{
			name: "pagination",
			e:    NewPageExecutor[int](10, 30*time.Second),
			args: args[int]{
				ctx: contextx.Background(),
				req: types.Page{
					Offset: 5,
					Limit:  5,
					Sort:   "",
				},
				fn: func(ctx contextx.IContext, p types.Page) ([]int, error) {
					result := make([]int, 0)
					for i := p.Offset; i < p.Offset+p.Limit; i++ {
						result = append(result, i)
					}

					return result, nil
				},
			},
			want: &PageResult[int]{
				Items: []int{5, 6, 7, 8, 9},
				Total: 5,
			},
			wantErr: false,
		},
		{
			name: "context canceled test",
			e:    NewPageExecutor[int](10, 30*time.Second),
			args: args[int]{
				ctx: func() contextx.IContext {
					ctx, cancel := contextx.WithCancel(contextx.Background())
					cancel()
					return ctx
				}(),
				req: types.Page{
					Offset: 0,
					Limit:  10,
					Sort:   "",
				},
				fn: func(ctx contextx.IContext, p types.Page) ([]int, error) {
					return nil, nil
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.e.Execute(tt.args.ctx, tt.args.req, tt.args.fn)
			if err != nil {
				t.Logf("Retry() got err = %v", err)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("Retry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Retry() got = %v, want %v", got, tt.want)
			}
		})
	}
}
