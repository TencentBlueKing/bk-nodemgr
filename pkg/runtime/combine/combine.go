/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
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
type DoFunc func(key string, data []interface{}) (interface{}, error)

// Handler is the combined handler.
type Handler interface {
	// Call calls the function.
	Call(ctx context.Context, data ...interface{}) (interface{}, error)

	// CallWithAggregationKey calls the function with aggregation key.
	CallWithAggregationKey(ctx context.Context, key string, data ...interface{}) (interface{}, error)
}

// New creates a new combined handler.
func New(maxDataLimit int, maxLaunchTimeGap time.Duration, dofunc DoFunc) Handler {
	return &combinedHandlerGroup{
		handlers:         make(map[string]*combinedHandler),
		maxDataLimit:     maxDataLimit,
		maxLaunchTimeGap: maxLaunchTimeGap,
		dofunc:           dofunc,
	}
}

const (
	defaultAggregationKey = "__combined_aggregation__"
)

type combinedHandlerGroup struct {
	mu sync.Mutex

	handlers map[string]*combinedHandler

	maxDataLimit     int
	maxLaunchTimeGap time.Duration
	dofunc           DoFunc
}

// Call calls the function.
func (group *combinedHandlerGroup) Call(ctx context.Context, data ...interface{}) (interface{}, error) {
	return group.CallWithAggregationKey(ctx, defaultAggregationKey, data...)
}

// CallWithAggregationKey calls the function with aggregation key.
func (group *combinedHandlerGroup) CallWithAggregationKey(ctx context.Context, key string, data ...interface{}) (interface{}, error) {
	group.mu.Lock()
	handler, ok := group.handlers[key]
	if !ok {
		handler = &combinedHandler{
			key:              key,
			data:             make([]interface{}, 0),
			result:           make(map[int][]chan combinedResult),
			maxDataLimit:     group.maxDataLimit,
			maxLaunchTimeGap: group.maxLaunchTimeGap,
			dofunc:           group.dofunc,
		}

		group.handlers[key] = handler
	}
	group.mu.Unlock()

	return handler.call(ctx, data...)
}

type combinedResult struct {
	val interface{}
	err error
}

// CombinedHandler is the combined handler.
type combinedHandler struct {
	mu sync.Mutex

	key    string
	dofunc DoFunc

	// generation is to track the version of data.
	generation int
	data       []interface{}
	result     map[int][]chan combinedResult

	maxDataLimit int

	maxLaunchTimeGap time.Duration
	lastLaunchedTime time.Time
}

// call calls the function.
func (handler *combinedHandler) call(ctx context.Context, data ...interface{}) (interface{}, error) {
	if ctx == nil {
		return nil, errors.New("ctx is nil")
	}

	if handler.dofunc == nil {
		return nil, errors.New("dofunc is nil")
	}

	if len(data) == 0 {
		return nil, errors.New("data is empty")
	}

	handler.mu.Lock()
	ch := handler.add(data...)
	handler.check()
	handler.mu.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-ch:
		return result.val, result.err
	}
}

func (handler *combinedHandler) add(data ...interface{}) <-chan combinedResult {
	handler.data = append(handler.data, data...)
	if _, ok := handler.result[handler.generation]; !ok {
		handler.result[handler.generation] = make([]chan combinedResult, 0)

		// delay check for first added.
		go func() {
			time.Sleep(handler.maxLaunchTimeGap)

			handler.mu.Lock()
			handler.check()
			handler.mu.Unlock()
		}()
	}

	ch := make(chan combinedResult, 1)
	handler.result[handler.generation] = append(handler.result[handler.generation], ch)

	return ch
}

func (handler *combinedHandler) check() {
	if time.Since(handler.lastLaunchedTime) < handler.maxLaunchTimeGap && len(handler.data) < handler.maxDataLimit {
		return
	}

	data := handler.data
	generation := handler.generation

	handler.data = make([]interface{}, 0)
	handler.generation++

	handler.lastLaunchedTime = time.Now()

	go func() {
		result, err := handler.dofunc(handler.key, data)
		handler.mu.Lock()
		for _, ch := range handler.result[generation] {
			ch <- combinedResult{
				val: result,
				err: err,
			}
		}
		delete(handler.result, generation)
		handler.mu.Unlock()
	}()
}
