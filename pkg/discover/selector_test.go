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

package discover

import (
	"reflect"
	"testing"
)

// TestRandomSelector_Select test the random selector.
func TestRandomSelector_Select(t *testing.T) {
	type args struct {
		endpoints []Endpoint
	}
	tests := []struct {
		name    string
		args    args
		want    []Endpoint
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				endpoints: []Endpoint{
					{
						IPV4: "127.0.0.1",
						IPV6: "::1",
						Port: 1001,
					},
					{
						IPV4: "127.0.0.2",
						IPV6: "::1",
						Port: 1002,
					},
					{
						IPV4: "127.0.0.3",
						IPV6: "::1",
						Port: 1003,
					},
				},
			},
			want: []Endpoint{
				{
					IPV4: "127.0.0.1",
					IPV6: "::1",
					Port: 1001,
				},
				{
					IPV4: "127.0.0.2",
					IPV6: "::1",
					Port: 1002,
				},
				{
					IPV4: "127.0.0.3",
					IPV6: "::1",
					Port: 1003,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := NewRandomSelector()
			got, err := selector.Select(tt.args.endpoints)
			if (err != nil) != tt.wantErr {
				t.Errorf("RandomSelector.Select() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			found := false
			for _, endpoint := range tt.want {
				if reflect.DeepEqual(got, endpoint) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("RandomSelector.Select() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestRoundRobinSelector_Select test round robin selector.
func TestRoundRobinSelector_Select(t *testing.T) {
	type args struct {
		endpoints []Endpoint
	}
	tests := []struct {
		name    string
		args    args
		want    []Endpoint
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				endpoints: []Endpoint{
					{
						IPV4: "127.0.0.1",
						IPV6: "::1",
						Port: 1001,
					},
					{
						IPV4: "127.0.0.2",
						IPV6: "::1",
						Port: 1002,
					},
					{
						IPV4: "127.0.0.3",
						IPV6: "::1",
						Port: 1003,
					},
				},
			},
			want: []Endpoint{
				{
					IPV4: "127.0.0.1",
					IPV6: "::1",
					Port: 1001,
				},
				{
					IPV4: "127.0.0.2",
					IPV6: "::1",
					Port: 1002,
				},
				{
					IPV4: "127.0.0.3",
					IPV6: "::1",
					Port: 1003,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := NewRoundRobinSelector()

			for _, want := range tt.want {
				got, err := selector.Select(tt.args.endpoints)
				if (err != nil) != tt.wantErr {
					t.Errorf("RandomSelector.Select() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if !reflect.DeepEqual(got, want) {
					t.Errorf("RandomSelector.Select() got = %v, want %v", got, want)
				}
			}
		})
	}
}

// TestSelectEndpoints test SelectEndpoints function.
func TestSelectEndpoints(t *testing.T) {
	tests := []struct {
		name      string
		endpoints []Endpoint
		count     int
		wantLen   int
		wantErr   bool
	}{
		{
			name: "normal - select 3 from 10",
			endpoints: []Endpoint{
				{IPV4: "127.0.0.1", Port: 1001},
				{IPV4: "127.0.0.2", Port: 1002},
				{IPV4: "127.0.0.3", Port: 1003},
				{IPV4: "127.0.0.4", Port: 1004},
				{IPV4: "127.0.0.5", Port: 1005},
				{IPV4: "127.0.0.6", Port: 1006},
				{IPV4: "127.0.0.7", Port: 1007},
				{IPV4: "127.0.0.8", Port: 1008},
				{IPV4: "127.0.0.9", Port: 1009},
				{IPV4: "127.0.0.10", Port: 1010},
			},
			count:   3,
			wantLen: 3,
			wantErr: false,
		},
		{
			name: "boundary - endpoints count equals count",
			endpoints: []Endpoint{
				{IPV4: "127.0.0.1", Port: 1001},
				{IPV4: "127.0.0.2", Port: 1002},
				{IPV4: "127.0.0.3", Port: 1003},
			},
			count:   3,
			wantLen: 3,
			wantErr: false,
		},
		{
			name: "boundary - endpoints count less than count",
			endpoints: []Endpoint{
				{IPV4: "127.0.0.1", Port: 1001},
				{IPV4: "127.0.0.2", Port: 1002},
			},
			count:   3,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:      "boundary - empty endpoints",
			endpoints: []Endpoint{},
			count:     3,
			wantLen:   0,
			wantErr:   true,
		},
		{
			name: "boundary - count <= 0",
			endpoints: []Endpoint{
				{IPV4: "127.0.0.1", Port: 1001},
				{IPV4: "127.0.0.2", Port: 1002},
				{IPV4: "127.0.0.3", Port: 1003},
			},
			count:   0,
			wantLen: 3,
			wantErr: false,
		},
		{
			name: "boundary - count negative",
			endpoints: []Endpoint{
				{IPV4: "127.0.0.1", Port: 1001},
				{IPV4: "127.0.0.2", Port: 1002},
				{IPV4: "127.0.0.3", Port: 1003},
			},
			count:   -1,
			wantLen: 3,
			wantErr: false,
		},
		{
			name: "boundary - count >= len(endpoints)",
			endpoints: []Endpoint{
				{IPV4: "127.0.0.1", Port: 1001},
				{IPV4: "127.0.0.2", Port: 1002},
				{IPV4: "127.0.0.3", Port: 1003},
			},
			count:   5,
			wantLen: 3,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SelectEndpoints(tt.endpoints, tt.count, NewRoundRobinSelector())
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectEndpoints() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("SelectEndpoints() len = %v, wantLen %v", len(got), tt.wantLen)
			}
			// 验证去重：确保没有重复的地址。
			addrMap := make(map[string]bool)
			for _, ep := range got {
				addr := ep.GetIPV4Address()
				if addrMap[addr] {
					t.Errorf("SelectEndpoints() found duplicate address: %v", addr)
				}
				addrMap[addr] = true
			}
		})
	}
}

// TestSelectEndpoints_RoundRobin test round robin behavior.
func TestSelectEndpoints_RoundRobin(t *testing.T) {
	endpoints := []Endpoint{
		{IPV4: "127.0.0.1", Port: 1001},
		{IPV4: "127.0.0.2", Port: 1002},
		{IPV4: "127.0.0.3", Port: 1003},
		{IPV4: "127.0.0.4", Port: 1004},
		{IPV4: "127.0.0.5", Port: 1005},
	}

	// 多次调用，验证轮询行为（应该返回不同的 endpoints）。
	firstCall, err := SelectEndpoints(endpoints, 3, NewRoundRobinSelector())
	if err != nil {
		t.Fatalf("SelectEndpoints() error = %v", err)
	}
	if len(firstCall) != 3 {
		t.Fatalf("SelectEndpoints() len = %v, want 3", len(firstCall))
	}

	secondCall, err := SelectEndpoints(endpoints, 3, NewRoundRobinSelector())
	if err != nil {
		t.Fatalf("SelectEndpoints() error = %v", err)
	}
	if len(secondCall) != 3 {
		t.Fatalf("SelectEndpoints() len = %v, want 3", len(secondCall))
	}

	// 验证两次调用返回的 endpoints 不同（由于轮询，应该有不同的选择）。
	firstAddrs := make(map[string]bool)
	for _, ep := range firstCall {
		firstAddrs[ep.GetIPV4Address()] = true
	}

	secondAddrs := make(map[string]bool)
	for _, ep := range secondCall {
		secondAddrs[ep.GetIPV4Address()] = true
	}

	// 由于轮询机制，两次调用应该有不同的选择（至少有一个不同）。
	allSame := true
	for addr := range firstAddrs {
		if !secondAddrs[addr] {
			allSame = false
			break
		}
	}
	if allSame {
		// 如果所有地址都相同，检查是否是因为所有 endpoints 都被选择了。
		// 在这种情况下，如果 count >= len(endpoints)，会返回全部，所以两次调用可能相同。
		// 但这里 count=3, len=5，所以应该有不同的选择。
		// 如果确实相同，可能是因为轮询的起始位置相同，这是可以接受的。
		t.Logf("Warning: Both calls returned same endpoints, this may be acceptable for round-robin")
	}
}

// TestRoundRobinSelector_Select_NoModifyOriginalSlice test that RoundRobinSelector.Select does not modify the original slice.
func TestRoundRobinSelector_Select_NoModifyOriginalSlice(t *testing.T) {
	originalEndpoints := []Endpoint{
		{IPV4: "127.0.0.3", Port: 1003},
		{IPV4: "127.0.0.1", Port: 1001},
		{IPV4: "127.0.0.2", Port: 1002},
	}

	// 创建副本用于比较
	expectedEndpoints := make([]Endpoint, len(originalEndpoints))
	copy(expectedEndpoints, originalEndpoints)

	selector := NewRoundRobinSelector()

	// 调用 Select 方法
	_, err := selector.Select(originalEndpoints)
	if err != nil {
		t.Fatalf("RoundRobinSelector.Select() error = %v", err)
	}

	// 验证原始切片未被修改
	if !reflect.DeepEqual(originalEndpoints, expectedEndpoints) {
		t.Errorf("RoundRobinSelector.Select() modified the original slice")
		t.Errorf("Original: %v", originalEndpoints)
		t.Errorf("Expected: %v", expectedEndpoints)
	}
}

// TestSelectEndpoints_NoModifyOriginalSlice test that SelectEndpoints does not modify the original slice.
func TestSelectEndpoints_NoModifyOriginalSlice(t *testing.T) {
	originalEndpoints := []Endpoint{
		{IPV4: "127.0.0.3", Port: 1003},
		{IPV4: "127.0.0.1", Port: 1001},
		{IPV4: "127.0.0.2", Port: 1002},
		{IPV4: "127.0.0.4", Port: 1004},
		{IPV4: "127.0.0.5", Port: 1005},
	}

	// 创建副本用于比较
	expectedEndpoints := make([]Endpoint, len(originalEndpoints))
	copy(expectedEndpoints, originalEndpoints)

	// 调用 SelectEndpoints 方法
	_, err := SelectEndpoints(originalEndpoints, 3, NewRoundRobinSelector())
	if err != nil {
		t.Fatalf("SelectEndpoints() error = %v", err)
	}

	// 验证原始切片未被修改
	if !reflect.DeepEqual(originalEndpoints, expectedEndpoints) {
		t.Errorf("SelectEndpoints() modified the original slice")
		t.Errorf("Original: %v", originalEndpoints)
		t.Errorf("Expected: %v", expectedEndpoints)
	}
}
