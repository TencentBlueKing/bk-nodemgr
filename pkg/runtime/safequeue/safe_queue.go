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

// Package safequeue ...
package safequeue

import "sync"

// NewSafeQueue creates a new SafeQueue instance.
func NewSafeQueue[T any]() *SafeQueue[T] {
	return &SafeQueue[T]{
		items: make([]T, 0),
		mu:    sync.Mutex{},
	}
}

// SafeQueue is a concurrency-safe FIFO queue.
// Fix channel-based queue unable automatically grow when the number of items is unknown, and avoid blocking when the queue is full.
type SafeQueue[T any] struct {
	mu    sync.Mutex
	items []T
}

// Enqueue adds an item to the end of the queue.
func (queue *SafeQueue[T]) Enqueue(v T) {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	queue.items = append(queue.items, v)
}

// Dequeue removes and returns the item at the front of the queue. If the queue is empty, it returns the zero value of T and false.
func (queue *SafeQueue[T]) Dequeue() (T, bool) {
	queue.mu.Lock()
	defer queue.mu.Unlock()

	var zero T

	if len(queue.items) == 0 {
		return zero, false
	}

	v := queue.items[0]
	queue.items[0] = zero
	queue.items = queue.items[1:]

	return v, true
}

// Peek returns the item at the front of the queue without removing it. If the queue is empty, it returns the zero value of T and false.
func (queue *SafeQueue[T]) Peek() (T, bool) {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	if len(queue.items) == 0 {
		var zero T
		return zero, false
	}

	return queue.items[0], true
}

// Len returns the number of items in the queue.
func (queue *SafeQueue[T]) Len() int {
	queue.mu.Lock()
	defer queue.mu.Unlock()

	return len(queue.items)
}

// IsEmpty returns true if the queue is empty, false otherwise.
func (queue *SafeQueue[T]) IsEmpty() bool {
	return queue.Len() == 0
}
