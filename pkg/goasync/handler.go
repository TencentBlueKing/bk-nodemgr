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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/panjf2000/ants/v2"
	"go.opentelemetry.io/otel/trace"
)

var _ IHandler = &Handler{}

// Handler defines a handler of goasync.
type Handler struct {
	_ struct{}

	pool   *ants.MultiPoolWithFuncGeneric[*Task]
	tracer trace.Tracer
}

// HandlerOption defines the options of handler.
type HandlerOption struct {
	Size                  int
	SizePerPool           int
	LoadBalancingStrategy ants.LoadBalancingStrategy
	Tracer                trace.Tracer
}

// Validate validates the handler option.
func (opt *HandlerOption) Validate() error {
	if opt.Size <= 0 {
		return fmt.Errorf("size should be greater than 0")
	}

	if opt.SizePerPool <= 0 {
		return fmt.Errorf("sizePerPool should be greater than 0")
	}

	if opt.Tracer == nil {
		return fmt.Errorf("tracer should not be nil")
	}

	return nil
}

const (
	spanName = "goasync"
)

// NewHandler returns a new handler.
func NewHandler(option HandlerOption) (*Handler, error) {
	h := &Handler{
		tracer: option.Tracer,
	}

	pool, err := ants.NewMultiPoolWithFuncGeneric[*Task](option.Size, option.SizePerPool, func(task *Task) {
		spanCtx, span := h.tracer.Start(task.nCtx, spanName,
			trace.WithSpanKind(trace.SpanKindInternal),
		)
		defer span.End()

		nCtx := contextx.FromContext(spanCtx)

		if err := task.runFn(nCtx); err != nil {
			logger.G.Biz(nCtx).WithErr(err).Error("failed to run async function")
		}
	}, option.LoadBalancingStrategy)
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
