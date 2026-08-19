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

package combine

import (
	"context"
	"sync"
	"testing"
	"time"
)

func Test_Call(t *testing.T) {
	type args struct {
		ctx context.Context

		maxDataLimit     int
		maxLaunchTimeGap time.Duration
		dofunc           DoFunc[int, *struct{}]

		data []int

		wantErr   bool
		wantTimes int
	}
	times := make(map[string]int)
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test1",
			args: args{
				ctx:              context.Background(),
				maxDataLimit:     100,
				maxLaunchTimeGap: 1 * time.Second,
				dofunc: func(key string, data []int) (*struct{}, error) {
					times["test1"]++

					return nil, nil
				},
				data:      []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				wantErr:   false,
				wantTimes: 2,
			},
		},
		{
			name: "test2",
			args: args{
				ctx:              context.Background(),
				maxDataLimit:     3,
				maxLaunchTimeGap: 1 * time.Second,
				dofunc: func(key string, data []int) (*struct{}, error) {
					times["test2"]++

					return nil, nil
				},
				data:      []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				wantErr:   false,
				wantTimes: 4,
			},
		},
		{
			name: "invalid ctx",
			args: args{
				ctx:              nil,
				maxDataLimit:     3,
				maxLaunchTimeGap: 1 * time.Second,
				dofunc: func(key string, data []int) (*struct{}, error) {
					return nil, nil
				},
				data:      []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				wantErr:   true,
				wantTimes: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New(tt.args.maxDataLimit, tt.args.maxLaunchTimeGap, tt.args.dofunc)

			times[tt.name] = 0

			wg := sync.WaitGroup{}
			for _, d := range tt.args.data {
				wg.Add(1)
				go func(d int) {
					_, _, err := h.Call(tt.args.ctx, d)
					if (err != nil) != tt.args.wantErr {
						t.Errorf("Call() error = %v, wantErr = %v", err, tt.args.wantErr)
						return
					}
					wg.Done()
				}(d)
			}
			wg.Wait()

			if !tt.args.wantErr && times[tt.name] != tt.args.wantTimes {
				t.Errorf("Call() times = %v, wantTimes = %v", times[tt.name], tt.args.wantTimes)
				return
			}
		})
	}
}

type testAggreationData struct {
	key  string
	item int
}

func Test_CallWithAggregationKey(t *testing.T) {
	type args struct {
		ctx context.Context

		maxDataLimit     int
		maxLaunchTimeGap time.Duration
		dofunc           DoFunc[int, *struct{}]

		data []*testAggreationData

		wantErr   bool
		wantTimes int
	}
	times := make(map[string]int)
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test1",
			args: args{
				ctx:              context.Background(),
				maxDataLimit:     100,
				maxLaunchTimeGap: 1 * time.Second,
				dofunc: func(key string, data []int) (*struct{}, error) {
					times["test1"]++

					return nil, nil
				},
				data: []*testAggreationData{
					{"k1", 1},
					{"k1", 2},
					{"k2", 3},
					{"k2", 4},
					{"k2", 5},
					{"k2", 6},
					{"k3", 7},
				},
				wantErr:   false,
				wantTimes: 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New(tt.args.maxDataLimit, tt.args.maxLaunchTimeGap, tt.args.dofunc)

			times[tt.name] = 0

			wg := sync.WaitGroup{}
			for _, d := range tt.args.data {
				wg.Add(1)
				go func(d *testAggreationData) {
					_, _, err := h.CallWithAggregationKey(tt.args.ctx, d.key, d.item)
					if (err != nil) != tt.args.wantErr {
						t.Errorf("Call() error = %v, wantErr = %v", err, tt.args.wantErr)
						return
					}
					wg.Done()
				}(d)
			}
			wg.Wait()

			if !tt.args.wantErr && times[tt.name] != tt.args.wantTimes {
				t.Errorf("Call() times = %v, wantTimes = %v", times[tt.name], tt.args.wantTimes)
				return
			}
		})
	}
}

func Test_CallAndCheckIndex(t *testing.T) {
	type args struct {
		ctx context.Context

		maxDataLimit     int
		maxLaunchTimeGap time.Duration
		dofunc           DoFunc[int, int]

		dataList [][]int

		wantErr   bool
		wantTimes int
	}
	times := make(map[string]int)
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test1",
			args: args{
				ctx:              context.Background(),
				maxDataLimit:     100,
				maxLaunchTimeGap: 1 * time.Second,
				dofunc: func(key string, data []int) (int, error) {
					times["test1"]++

					return len(data), nil
				},
				dataList: [][]int{
					{1, 2, 3},
					{1, 2, 3, 4},
					{1, 2, 3, 4, 5},
					{1, 2, 3, 4, 5, 6},
				},
				wantErr:   false,
				wantTimes: 2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New(tt.args.maxDataLimit, tt.args.maxLaunchTimeGap, tt.args.dofunc)

			times[tt.name] = 0
			routes := make(map[int]map[int]int)
			var routesMutex sync.Mutex

			wg := sync.WaitGroup{}
			for _, d := range tt.args.dataList {
				wg.Add(1)
				go func(d []int) {
					total, index, err := h.Call(tt.args.ctx, d...)
					if (err != nil) != tt.args.wantErr {
						t.Errorf("Call() error = %v, wantErr = %v", err, tt.args.wantErr)
						return
					}

					routesMutex.Lock()
					if _, ok := routes[total]; !ok {
						routes[total] = make(map[int]int)
					}
					routes[total][index] = index + len(d)
					routesMutex.Unlock()

					wg.Done()
				}(d)
			}
			wg.Wait()

			if !tt.args.wantErr && times[tt.name] != tt.args.wantTimes {
				t.Errorf("Call() times = %v, wantTimes = %v", times[tt.name], tt.args.wantTimes)
				return
			}

			for total, r := range routes {
				idx := 0
				for {
					_, ok := r[idx]
					if !ok {
						t.Errorf("Call() total = %d, routes = %v, does not have index = %d", total, r, idx)
						return
					}

					if idx >= r[idx] {
						t.Errorf("Call() total = %d, routes = %v, got error index = %d", total, r, idx)
						return
					}

					idx = r[idx]

					// success.
					if idx == total {
						return
					}
				}
			}
		})
	}
}
