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

package safequeue

import (
	"sync"
	"testing"
)

// TestSafeQueue_Dequeue ...
func TestSafeQueue_Dequeue(t *testing.T) {
	type args struct {
		enqueue []int
	}
	tests := []struct {
		name    string
		args    args
		wantVal int
		wantOk  bool
	}{
		{
			name:    "dequeue from non-empty queue returns first element",
			args:    args{enqueue: []int{1, 2, 3}},
			wantVal: 1,
			wantOk:  true,
		},
		{
			name:    "dequeue from single-element queue",
			args:    args{enqueue: []int{42}},
			wantVal: 42,
			wantOk:  true,
		},
		{
			name:    "dequeue from empty queue returns zero value",
			args:    args{enqueue: nil},
			wantVal: 0,
			wantOk:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &SafeQueue[int]{}
			for _, v := range tt.args.enqueue {
				q.Enqueue(v)
			}
			got, ok := q.Dequeue()
			if ok != tt.wantOk {
				t.Errorf("Dequeue() ok = %v, want %v", ok, tt.wantOk)
			}
			if got != tt.wantVal {
				t.Errorf("Dequeue() val = %v, want %v", got, tt.wantVal)
			}
		})
	}
}

// TestSafeQueue_FIFOOrder verifies that elements are dequeued in FIFO order.
func TestSafeQueue_FIFOOrder(t *testing.T) {
	q := &SafeQueue[int]{}
	input := []int{10, 20, 30, 40, 50}
	for _, v := range input {
		q.Enqueue(v)
	}

	for i, want := range input {
		got, ok := q.Dequeue()
		if !ok {
			t.Fatalf("Dequeue() #%d: unexpected empty queue", i)
		}
		if got != want {
			t.Errorf("Dequeue() #%d = %v, want %v", i, got, want)
		}
	}

	if !q.IsEmpty() {
		t.Errorf("queue should be empty after dequeuing all elements, got len=%d", q.Len())
	}
}

// TestSafeQueue_Peek ...
func TestSafeQueue_Peek(t *testing.T) {
	type args struct {
		enqueue []int
	}
	tests := []struct {
		name    string
		args    args
		wantVal int
		wantOk  bool
		wantLen int
	}{
		{
			name:    "peek non-empty queue returns first element without removing",
			args:    args{enqueue: []int{1, 2}},
			wantVal: 1,
			wantOk:  true,
			wantLen: 2,
		},
		{
			name:    "peek empty queue returns zero value",
			args:    args{enqueue: nil},
			wantVal: 0,
			wantOk:  false,
			wantLen: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &SafeQueue[int]{}
			for _, v := range tt.args.enqueue {
				q.Enqueue(v)
			}
			got, ok := q.Peek()
			if ok != tt.wantOk {
				t.Errorf("Peek() ok = %v, want %v", ok, tt.wantOk)
			}
			if got != tt.wantVal {
				t.Errorf("Peek() val = %v, want %v", got, tt.wantVal)
			}
			if q.Len() != tt.wantLen {
				t.Errorf("Len() after Peek() = %d, want %d", q.Len(), tt.wantLen)
			}
		})
	}
}

// TestSafeQueue_LenAndIsEmpty ...
func TestSafeQueue_LenAndIsEmpty(t *testing.T) {
	tests := []struct {
		name        string
		enqueue     int
		dequeue     int
		wantLen     int
		wantIsEmpty bool
	}{
		{
			name:        "new queue is empty",
			enqueue:     0,
			dequeue:     0,
			wantLen:     0,
			wantIsEmpty: true,
		},
		{
			name:        "enqueue makes queue non-empty",
			enqueue:     3,
			dequeue:     0,
			wantLen:     3,
			wantIsEmpty: false,
		},
		{
			name:        "dequeue all makes queue empty again",
			enqueue:     2,
			dequeue:     2,
			wantLen:     0,
			wantIsEmpty: true,
		},
		{
			name:        "partial dequeue",
			enqueue:     5,
			dequeue:     3,
			wantLen:     2,
			wantIsEmpty: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &SafeQueue[int]{}
			for i := 0; i < tt.enqueue; i++ {
				q.Enqueue(i)
			}
			for i := 0; i < tt.dequeue; i++ {
				q.Dequeue()
			}
			if q.Len() != tt.wantLen {
				t.Errorf("Len() = %d, want %d", q.Len(), tt.wantLen)
			}
			if q.IsEmpty() != tt.wantIsEmpty {
				t.Errorf("IsEmpty() = %v, want %v", q.IsEmpty(), tt.wantIsEmpty)
			}
		})
	}
}

// TestSafeQueue_ConcurrentAccess verifies concurrent producers and consumers do not lose data.
func TestSafeQueue_ConcurrentAccess(t *testing.T) {
	q := &SafeQueue[int]{}
	const numGoroutines = 100
	const numOps = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	// concurrent producers
	for i := 0; i < numGoroutines; i++ {
		go func(base int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				q.Enqueue(base*numOps + j)
			}
		}(i)
	}

	// concurrent consumers
	consumed := make([]int, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				if _, ok := q.Dequeue(); ok {
					consumed[idx]++
				}
			}
		}(i)
	}

	wg.Wait()

	totalConsumed := 0
	for _, c := range consumed {
		totalConsumed += c
	}
	wantTotal := numGoroutines * numOps
	gotTotal := totalConsumed + q.Len()
	if gotTotal != wantTotal {
		t.Errorf("consumed(%d) + remaining(%d) = %d, want %d", totalConsumed, q.Len(), gotTotal, wantTotal)
	}
}
