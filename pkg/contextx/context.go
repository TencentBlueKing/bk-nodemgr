/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package contextx defines the context of the nodemgr.
package contextx

import (
	"context"
	"time"
)

// IContext this is the context of the nodemgr.
type IContext interface {
	context.Context
	Values() map[string]any
}

var _ IContext = &Context{}

// Context defines the context of the nodemgr.
type Context struct {
	ctx    context.Context
	values map[string]any
}

// Deadline implement IContext.
func (c Context) Deadline() (deadline time.Time, ok bool) {
	return c.ctx.Deadline()
}

// Done implement IContext.
func (c Context) Done() <-chan struct{} {
	return c.ctx.Done()
}

// Err implement IContext.
func (c Context) Err() error {
	return c.ctx.Err()
}

// Value implement IContext.
func (c Context) Value(key any) any {
	strKey, ok := key.(string)
	if ok {
		if val, exists := c.values[strKey]; exists {
			return val
		}
	}

	return c.ctx.Value(key)
}

// Values implement IContext.
func (c Context) Values() map[string]any {
	return c.values
}

// NewContext new a context.
func NewContext(ctx context.Context, values map[string]any) *Context {
	return &Context{
		ctx:    ctx,
		values: values,
	}
}

// WithValue this is the same as context.WithValue.
func WithValue(ctx IContext, key string, val any) IContext {
	ctxWithV := context.WithValue(ctx, key, val)
	newCtx := NewContext(ctxWithV, ctx.Values())
	newCtx.values[key] = val

	return newCtx
}

// WithCancel this is the same as context.WithCancel.
func WithCancel(ctx IContext) (IContext, context.CancelFunc) {
	ctxWithC, cancel := context.WithCancel(ctx)

	return NewContext(ctxWithC, ctx.Values()), cancel
}

// WithTimeout this is the same as context.WithTimeout.
func WithTimeout(ctx IContext, timeout time.Duration) (IContext, context.CancelFunc) {
	ctxWithC, cancel := context.WithTimeout(ctx, timeout)

	return NewContext(ctxWithC, ctx.Values()), cancel
}
