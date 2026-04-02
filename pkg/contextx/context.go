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

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var _ IContext = &Context{}

// Context defines the context of the nodemgr.
type Context struct {
	_ struct{}

	ctx context.Context

	info IContextInfo
}

// MessageID implement IContext.
func (c *Context) MessageID() string {
	return c.info.MessageID()
}

// CheckMessageID implement IContext.
func (c *Context) CheckMessageID() error {
	return c.info.CheckMessageID()
}

// Deadline implement IContext.
func (c *Context) Deadline() (time.Time, bool) {
	return c.ctx.Deadline()
}

// Done implement IContext.
func (c *Context) Done() <-chan struct{} {
	return c.ctx.Done()
}

// Err implement IContext.
func (c *Context) Err() error {
	return c.ctx.Err()
}

// Value implement IContext.
func (c *Context) Value(key any) any {
	value, ok := c.info.GetValue(key)
	if !ok {
		return c.ctx.Value(key)
	}

	return value
}

// Values implement IContext.
func (c *Context) Values() map[string]any {
	return c.info.Values()
}

// TenantID implement IContext.
func (c *Context) TenantID() string {
	return c.info.TenantID()
}

// CheckTenantID implement IContext.
func (c *Context) CheckTenantID() error {
	return c.info.CheckTenantID()
}

// BKUsername implement IContext.
func (c *Context) BKUsername() string {
	return c.info.BKUsername()
}

// CheckBKUsername implement IContext.
func (c *Context) CheckBKUsername() error {
	return c.info.CheckBKUsername()
}

// LoginName implement IContext.
func (c *Context) LoginName() string {
	return c.info.LoginName()
}

// CheckLoginName implement IContext.
func (c *Context) CheckLoginName() error {
	return c.info.CheckLoginName()
}

// New new a context.
func New(ctx context.Context, opts ...InfoOpts) *Context {
	info := Info{}
	if oldInfo := asInfo(ctx); oldInfo != nil {
		if cloned, ok := oldInfo.(*Info); ok {
			info = *cloned
		} else {
			info = Info{
				values:     oldInfo.Values(),
				tenantID:   oldInfo.TenantID(),
				bkUsername: oldInfo.BKUsername(),
				loginName:  oldInfo.LoginName(),
				messageID:  oldInfo.MessageID(),
			}
		}
	} else {
		info.values = make(map[string]any)
	}

	for _, opt := range opts {
		opt(&info)
	}

	return newWithInfo(ctx, &info)
}

func newWithInfo(ctx context.Context, info IContextInfo) *Context {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String(attributeKeyTenantID, info.TenantID()),
		attribute.String(attributeKeyBKUsername, info.BKUsername()),
		attribute.String(attributeKeyLoginName, info.LoginName()),
		attribute.String(attributeKeyMessageID, info.MessageID()),
	)

	nCtx := &Context{
		ctx:  context.WithValue(ctx, nCtxInfoKey, info),
		info: info,
	}

	return nCtx
}

// From with context.
func From(nCtx IContext, opts ...InfoOpts) *Context {
	newOpts := []InfoOpts{
		WithTenantID(nCtx.TenantID()),
		WithBKUsername(nCtx.BKUsername()),
		WithLoginName(nCtx.LoginName()),
		WithMessageID(nCtx.MessageID()),
		WithValues(nCtx.Values()),
	}

	// append opts to newOpts, if there has same optFn, the latter one will cover the front.
	newOpts = append(newOpts, opts...)

	c := New(nCtx, newOpts...)

	return c
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

// WithDeadline this is the same as context.WithDeadline.
func WithDeadline(nCtx IContext, deadline time.Time) (IContext, context.CancelFunc) {
	ctxWithC, cancel := context.WithDeadline(nCtx, deadline)

	return New(ctxWithC, WithValues(nCtx.Values())), cancel
}

// WithoutCancel this is the same as context.WithoutCancel.
func WithoutCancel(nCtx IContext) IContext {
	if nCtx == nil {
		return Background()
	}

	ctxWithoutC := context.WithoutCancel(nCtx)

	return FromContext(ctxWithoutC)
}

// InfoOpts describes the context assignment options.
type InfoOpts func(*Info)

// WithTenantID assign tenant-id.
func WithTenantID(tenantID string) InfoOpts {
	return func(info *Info) {
		info.tenantID = tenantID
	}
}

// WithBKUsername assign bk-username.
func WithBKUsername(bkUsername string) InfoOpts {
	return func(info *Info) {
		info.bkUsername = bkUsername
	}
}

// WithLoginName assign login-name.
func WithLoginName(loginName string) InfoOpts {
	return func(info *Info) {
		info.loginName = loginName
	}
}

// WithMessageID assign message-id.
func WithMessageID(messageID string) InfoOpts {
	return func(info *Info) {
		info.messageID = messageID
	}
}

// WithValues assign values.
func WithValues(values map[string]any) InfoOpts {
	return func(info *Info) {
		if len(values) == 0 {
			return
		}

		for k, v := range values {
			info.values[k] = v
		}
	}
}

// Background this is the same as context.Background.
func Background() *Context {
	return New(context.Background())
}

// private type could make sure the security of context.
type contextKeyType int

const (
	nCtxInfoKey contextKeyType = iota
)

// FromContext creates a new context from the given context.
func FromContext(ctx context.Context) *Context {
	if ctx == nil {
		// nolint: contextcheck
		return Background()
	}

	if info := asInfo(ctx); info != nil {
		return newWithInfo(ctx, info.clone())
	}

	return New(ctx)
}

func asInfo(ctx context.Context) IContextInfo {
	if ctx == nil {
		return nil
	}

	if v := ctx.Value(nCtxInfoKey); v != nil {
		if i, ok := v.(IContextInfo); ok {
			return i
		}
	}

	return nil
}
