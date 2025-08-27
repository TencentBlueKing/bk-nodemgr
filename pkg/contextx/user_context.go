/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package contextx

import (
	"context"
	"time"
)

var _ IUserContext = &UserContext{}

// UserContext rest context.
type UserContext struct {
	ctx        context.Context
	bkUsername string
}

// NewUserContext new context.
func NewUserContext(ctx context.Context, bkUsername string) *UserContext {
	return &UserContext{
		ctx:        ctx,
		bkUsername: bkUsername,
	}
}

// Deadline implement IContext.
// nolint: nonamedreturns
func (ctx *UserContext) Deadline() (deadline time.Time, ok bool) {
	return ctx.ctx.Deadline()
}

// Done implement IContext.
func (ctx *UserContext) Done() <-chan struct{} {
	return ctx.ctx.Done()
}

// Err implement IContext.
func (ctx *UserContext) Err() error {
	return ctx.ctx.Err()
}

// Value implement IContext.
func (ctx *UserContext) Value(key any) any {
	return ctx.ctx.Value(key)
}

// Values implement IContext.
func (ctx *UserContext) Values() map[string]any {
	return map[string]any{
		"bk_username": ctx.bkUsername,
	}
}

// BKUsername implement IContext.
func (ctx *UserContext) BKUsername() string {
	return ctx.bkUsername
}
