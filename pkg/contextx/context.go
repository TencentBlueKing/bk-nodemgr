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
	"errors"
	"time"
)

var _ IContext = &Context{}

// Context defines the context of the nodemgr.
type Context struct {
	ctx context.Context

	values map[string]any

	tenantID   string
	bkUsername string
	messageID  string
}

// Deadline implement IContext.
func (c Context) Deadline() (time.Time, bool) {
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

// TenantID get tenant-id from values.
func (c Context) TenantID() string {
	return c.tenantID
}

// CheckTenantID check tenant-id.
func (c Context) CheckTenantID() error {
	if c.tenantID == "" {
		return errors.New("tenant-id not found")
	}

	return nil
}

// BKUsername get bk-username from values.
func (c Context) BKUsername() string {
	return c.bkUsername
}

// CheckBKUsername valcheckidate bk-username.
func (c Context) CheckBKUsername() error {
	if c.bkUsername == "" {
		return errors.New("bk-username not found")
	}

	return nil
}

// MessageID get message-id from values.
func (c Context) MessageID() string {
	return c.messageID
}

// CheckMessageID check message-id.
func (c Context) CheckMessageID() error {
	if c.messageID == "" {
		return errors.New("message-id not found")
	}

	return nil
}

// New new a context.
func New(ctx context.Context, opts ...Opts) *Context {
	r := &Context{
		ctx:    ctx,
		values: make(map[string]any),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// From with context.
func From(ctx IContext, opts ...Opts) *Context {
	return New(ctx, append([]Opts{WithValues(ctx.Values())}, opts...)...)
}

// WithCancel this is the same as context.WithCancel.
func WithCancel(nCtx IContext) (IContext, context.CancelFunc) {
	ctxWithC, cancel := context.WithCancel(nCtx)

	return New(ctxWithC, WithValues(nCtx.Values())), cancel
}

// WithTimeout this is the same as context.WithTimeout.
func WithTimeout(nCtx IContext, timeout time.Duration) (IContext, context.CancelFunc) {
	ctxWithC, cancel := context.WithTimeout(nCtx, timeout)

	return New(ctxWithC, WithValues(nCtx.Values())), cancel
}

// Opts describes the context assignment options.
type Opts func(*Context)

// WithTenantID assign tenant-id.
func WithTenantID(tenantID string) Opts {
	return func(c *Context) {
		c.tenantID = tenantID
	}
}

// WithBKUsername assign bk-username.
func WithBKUsername(bkUsername string) Opts {
	return func(c *Context) {
		c.bkUsername = bkUsername
	}
}

// WithMessageID assign message-id.
func WithMessageID(messageID string) Opts {
	return func(c *Context) {
		c.messageID = messageID
	}
}

// WithValues assign values.
func WithValues(values map[string]any) Opts {
	return func(c *Context) {
		for k, v := range values {
			c.values[k] = v
		}
	}
}
