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

// Package combine provides a way to combine multiple data and launch function.
package combine

import (
	"context"
	"errors"
	"sync"
	"time"
)

// DoFunc is the function to be launched.
type DoFunc[T, V any] func(key string, data []T) (V, error)

// IHandler is the combined handler.
type IHandler[T, V any] interface {
	// Call calls the function.
	// return the result, param begin index and error.
	Call(ctx context.Context, data ...T) (V, int, error)

	// CallWithAggregationKey calls the function with aggregation key.
	// return the result, param begin index and error.
	CallWithAggregationKey(ctx context.Context, key string, data ...T) (V, int, error)
}

// New creates a new combined handler.
func New[T, V any](maxDataLimit int, maxLaunchTimeGap time.Duration, dofunc DoFunc[T, V]) IHandler[T, V] {
	return &combinedHandlerGroup[T, V]{
		handlers:         make(map[string]*combinedHandler[T, V]),
		maxDataLimit:     maxDataLimit,
		maxLaunchTimeGap: maxLaunchTimeGap,
		dofunc:           dofunc,
	}
}

const (
	// defaultAggregationKey the default aggregation key.
	defaultAggregationKey = "__combined_aggregation__"

	// maxGeneration the max generation.
	maxGeneration int = 1e6

	// handlerMaxIdleTime the max idle time of handler.
	handlerMaxIdleTime = 10 * time.Minute
)

type combinedHandlerGroup[T, V any] struct {
	mu sync.Mutex

	handlers map[string]*combinedHandler[T, V]

	maxDataLimit     int
	maxLaunchTimeGap time.Duration
	dofunc           DoFunc[T, V]
}

// Call calls the function.
func (group *combinedHandlerGroup[T, V]) Call(ctx context.Context, data ...T) (V, int, error) {
	return group.CallWithAggregationKey(ctx, defaultAggregationKey, data...)
}

// CallWithAggregationKey calls the function with aggregation key.
func (group *combinedHandlerGroup[T, V]) CallWithAggregationKey(ctx context.Context, key string, data ...T) (V, int, error) {
	group.mu.Lock()
	handler, ok := group.handlers[key]
	if !ok {
		handler = &combinedHandler[T, V]{
			key:              key,
			data:             make([]T, 0),
			resultChs:        make(map[int][]chan combinedResult[V]),
			maxDataLimit:     group.maxDataLimit,
			maxLaunchTimeGap: group.maxLaunchTimeGap,
			dofunc:           group.dofunc,
			lastCallTime:     time.Now(),
		}

		group.handlers[key] = handler

		// trigger check handlers.
		go group.checkHandlers()
	}
	group.mu.Unlock()

	return handler.call(ctx, data...)
}

func (group *combinedHandlerGroup[T, V]) checkHandlers() {
	group.mu.Lock()
	defer group.mu.Unlock()

	keys := make([]string, 0)
	for _, handler := range group.handlers {
		if handler.idle() {
			keys = append(keys, handler.key)
		}
	}
	for _, key := range keys {
		delete(group.handlers, key)
	}
}

type combinedResult[V any] struct {
	val V
	err error
}

// CombinedHandler is the combined handler.
type combinedHandler[T, V any] struct {
	mu sync.Mutex

	key    string
	dofunc DoFunc[T, V]

	// generation is to track the version of data.
	generation int
	data       []T
	resultChs  map[int][]chan combinedResult[V]

	maxDataLimit int

	maxLaunchTimeGap time.Duration
	lastLaunchedTime time.Time

	lastCallTime time.Time
}

// call calls the function.
func (handler *combinedHandler[T, V]) call(ctx context.Context, data ...T) (V, int, error) {
	var zeroV V
	if ctx == nil {
		return zeroV, -1, errors.New("ctx is nil")
	}

	if handler.dofunc == nil {
		return zeroV, -1, errors.New("dofunc is nil")
	}

	if len(data) == 0 {
		return zeroV, -1, errors.New("data is empty")
	}

	handler.mu.Lock()
	ch, index := handler.add(data...)
	handler.check()
	handler.mu.Unlock()

	select {
	case <-ctx.Done():
		return zeroV, index, ctx.Err()
	case result := <-ch:
		return result.val, index, result.err
	}
}

func (handler *combinedHandler[T, V]) add(data ...T) (<-chan combinedResult[V], int) {
	handler.lastCallTime = time.Now()

	index := len(handler.data)
	handler.data = append(handler.data, data...)
	if _, ok := handler.resultChs[handler.generation]; !ok {
		handler.resultChs[handler.generation] = make([]chan combinedResult[V], 0)

		// delay check for first added.
		go func() {
			time.Sleep(handler.maxLaunchTimeGap)

			handler.mu.Lock()
			handler.check()
			handler.mu.Unlock()
		}()
	}

	ch := make(chan combinedResult[V], 1)
	handler.resultChs[handler.generation] = append(handler.resultChs[handler.generation], ch)

	return ch, index
}

func (handler *combinedHandler[T, V]) check() {
	// empty data should not trigger any function.
	// trigger with limited data and the max gap time.
	if len(handler.data) == 0 || time.Since(handler.lastLaunchedTime) < handler.maxLaunchTimeGap && len(handler.data) < handler.maxDataLimit {
		return
	}

	data := handler.data
	generation := handler.generation

	handler.data = make([]T, 0)
	handler.generation = handler.generation%maxGeneration + 1

	handler.lastLaunchedTime = time.Now()

	go func() {
		result, err := handler.dofunc(handler.key, data)
		handler.mu.Lock()
		for _, ch := range handler.resultChs[generation] {
			ch <- combinedResult[V]{
				val: result,
				err: err,
			}
			close(ch)
		}
		delete(handler.resultChs, generation)
		handler.mu.Unlock()
	}()
}

func (handler *combinedHandler[T, V]) idle() bool {
	handler.mu.Lock()
	defer handler.mu.Unlock()

	return time.Since(handler.lastCallTime) > handlerMaxIdleTime && len(handler.data) == 0 && len(handler.resultChs) == 0
}
