/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package goasync

import (
	"bytes"
	"fmt"
	"runtime/debug"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/panjf2000/ants/v2"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var _ IHandler = &Handler{}

// Handler defines a handler of goasync.
type Handler struct {
	_ struct{}

	pool *ants.MultiPoolWithFuncGeneric[*Task]
}

// LoadBalancingStrategy defines the strategy of goroutine pool.
type LoadBalancingStrategy string

const (
	// LoadBalancingStrategyRoundRobin defines round robin strategy.
	LoadBalancingStrategyRoundRobin LoadBalancingStrategy = "round_robin"

	// LoadBalancingStrategyLeastFirst defines least first strategy.
	LoadBalancingStrategyLeastFirst LoadBalancingStrategy = "least_first"
)

// HandlerOption defines the options of handler.
type HandlerOption struct {
	// the number of goroutine pool.
	PoolNum int
	// the size of each goroutine pool.
	PerPoolSize           int
	LoadBalancingStrategy LoadBalancingStrategy
}

// Validate validates the handler option.
func (opt *HandlerOption) Validate() error {
	if opt.PoolNum <= 0 {
		return fmt.Errorf("size should be greater than 0")
	}

	if opt.PerPoolSize <= 0 {
		return fmt.Errorf("sizePerPool should be greater than 0")
	}

	return nil
}

const (
	scopeNamePrefix = "goasync_"
	spanName        = "goasync"
)

// NewHandler returns a new handler.
func NewHandler(option HandlerOption) (*Handler, error) {
	if err := option.Validate(); err != nil {
		return nil, fmt.Errorf("invalid handler option: %w", err)
	}

	h := &Handler{}

	var loadBalancingStrategy ants.LoadBalancingStrategy
	switch option.LoadBalancingStrategy {
	case LoadBalancingStrategyRoundRobin:
		loadBalancingStrategy = ants.RoundRobin
	case LoadBalancingStrategyLeastFirst:
		loadBalancingStrategy = ants.LeastTasks
	default:
		return nil, fmt.Errorf("invalid load balancing strategy: %s", option.LoadBalancingStrategy)
	}

	pool, err := ants.NewMultiPoolWithFuncGeneric[*Task](option.PoolNum, option.PerPoolSize, func(task *Task) {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()

				// The first line of the stack trace is of the form "goroutine N [status]:",
				// but by the time the panic reaches here the goroutine may no longer exist,
				// and its status will have changed. Trim out the misleading line.
				if line := bytes.IndexByte(stack[:], '\n'); line >= 0 {
					stack = stack[line+1:]
				}

				logger.G.Biz(task.nCtx).
					With("recover", r, "stack", string(stack)).
					Error("goasync task execution panic")
			}
		}()

		parentSpan := trace.SpanFromContext(task.nCtx)
		tracer := parentSpan.TracerProvider().Tracer(fmt.Sprintf("%s%s", scopeNamePrefix, task.name))

		var taskErr error
		spanCtx, span := tracer.Start(task.nCtx, fmt.Sprintf("%s %s", spanName, task.name),
			trace.WithSpanKind(trace.SpanKindInternal),
		)
		defer func() {
			if taskErr != nil {
				span.SetStatus(codes.Error, taskErr.Error())
				span.RecordError(taskErr)
			} else {
				span.SetStatus(codes.Ok, "")
			}
			span.End()
		}()

		nCtx := contextx.FromContext(spanCtx)

		if task.runFn == nil {
			logger.G.Biz(nCtx).Error("goasync task run function is nil")
			return
		}

		if taskErr = task.runFn(nCtx); taskErr != nil {
			logger.G.Biz(nCtx).WithErr(taskErr).Error("failed to run async function")
		}
	}, loadBalancingStrategy)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	h.pool = pool

	return h, nil
}

// Run run a async task.
func (h *Handler) Run(nCtx contextx.IContext, runFn RunFn, opts ...RunOptions) error {
	task := &Task{
		nCtx:  contextx.WithoutCancel(nCtx),
		runFn: runFn,
	}

	for _, opt := range opts {
		opt(task)
	}

	if err := h.pool.Invoke(task); err != nil {
		return fmt.Errorf("failed to invoke task: %w", err)
	}

	return nil
}
