/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
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
		dofunc           DoFunc

		data []interface{}

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
				dofunc: func(key string, data []interface{}) (interface{}, error) {
					times["test1"]++

					return nil, nil
				},
				data:      []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
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
				dofunc: func(key string, data []interface{}) (interface{}, error) {
					times["test2"]++

					return nil, nil
				},
				data:      []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
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
				dofunc: func(key string, data []interface{}) (interface{}, error) {
					return nil, nil
				},
				data:      []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
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
				go func(d interface{}) {
					_, err := h.Call(tt.args.ctx, d)
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

func Test_CallWithAggregationKey(t *testing.T) {
	type args struct {
		ctx context.Context

		maxDataLimit     int
		maxLaunchTimeGap time.Duration
		dofunc           DoFunc

		data []struct {
			key  string
			item interface{}
		}

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
				dofunc: func(key string, data []interface{}) (interface{}, error) {
					times["test1"]++

					return nil, nil
				},
				data: []struct {
					key  string
					item interface{}
				}{
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
				go func(d struct {
					key  string
					item interface{}
				}) {
					_, err := h.CallWithAggregationKey(tt.args.ctx, d.key, d.item)
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
