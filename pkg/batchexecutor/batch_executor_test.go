/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package batchexecutor ...
package batchexecutor

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

// TestExecute ...
func TestExecute(t *testing.T) {
	type args struct {
		ctx   contextx.IContext
		items []int
		fn    ExecuteFn[int]
		opts  []Option
	}
	tests := []struct {
		name        string
		args        args
		wantBatches [][]int
		wantErr     bool
	}{
		{
			name: "split items by batch size",
			args: args{
				ctx:   contextx.Background(),
				items: []int{1, 2, 3, 4, 5},
				opts:  []Option{WithBatchSize(2), WithTimeout(30 * time.Second)},
			},
			wantBatches: [][]int{{1, 2}, {3, 4}, {5}},
			wantErr:     false,
		},
		{
			name: "empty items",
			args: args{
				ctx:   contextx.Background(),
				items: nil,
				opts:  []Option{WithBatchSize(2), WithTimeout(30 * time.Second)},
			},
			wantBatches: [][]int{},
			wantErr:     false,
		},
		{
			name: "execute error",
			args: args{
				ctx:   contextx.Background(),
				items: []int{1, 2, 3},
				opts:  []Option{WithBatchSize(2), WithTimeout(30 * time.Second)},
				fn: func(contextx.IContext, []int) error {
					return errors.New("test error")
				},
			},
			wantBatches: [][]int{},
			wantErr:     true,
		},
		{
			name: "context canceled",
			args: args{
				ctx: func() contextx.IContext {
					ctx, cancel := contextx.WithCancel(contextx.Background())
					cancel()
					return ctx
				}(),
				items: []int{1},
				opts:  []Option{WithBatchSize(2), WithTimeout(30 * time.Second)},
			},
			wantBatches: [][]int{},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBatches := make([][]int, 0)
			fn := tt.args.fn
			if fn == nil {
				fn = func(_ contextx.IContext, items []int) error {
					gotBatches = append(gotBatches, append([]int(nil), items...))
					return nil
				}
			}

			err := Execute(tt.args.ctx, tt.args.items, fn, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotBatches, tt.wantBatches) {
				t.Errorf("Execute() batches = %v, want %v", gotBatches, tt.wantBatches)
			}
		})
	}
}

// TestCollect ...
func TestCollect(t *testing.T) {
	type args struct {
		ctx   contextx.IContext
		items []int
		fn    CollectFn[int, string]
		opts  []Option
	}
	tests := []struct {
		name    string
		args    args
		want    *Result[string]
		wantErr bool
	}{
		{
			name: "collect batch results",
			args: args{
				ctx:   contextx.Background(),
				items: []int{1, 2, 3},
				opts:  []Option{WithBatchSize(2), WithTimeout(30 * time.Second)},
				fn: func(_ contextx.IContext, items []int) ([]string, error) {
					result := make([]string, 0, len(items))
					for _, item := range items {
						result = append(result, string(rune('0'+item)))
					}

					return result, nil
				},
			},
			want: &Result[string]{
				Items: []string{"1", "2", "3"},
				Total: 3,
			},
			wantErr: false,
		},
		{
			name: "collect error",
			args: args{
				ctx:   contextx.Background(),
				items: []int{1},
				opts:  []Option{WithBatchSize(2), WithTimeout(30 * time.Second)},
				fn: func(contextx.IContext, []int) ([]string, error) {
					return nil, errors.New("test error")
				},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Collect(tt.args.ctx, tt.args.items, tt.args.fn, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Collect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Collect() got = %v, want %v", got, tt.want)
			}
		})
	}
}
