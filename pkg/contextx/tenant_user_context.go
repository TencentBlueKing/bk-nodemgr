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

var _ ITenantUserContext = &TenantUserContext{}

// TenantUserContext rest context.
type TenantUserContext struct {
	ctx        context.Context
	bkUsername string
	tenantID   string
}

// NewTenantUserContext new context.
func NewTenantUserContext(ctx context.Context, tenantID, bkUsername string) *TenantUserContext {
	return &TenantUserContext{
		ctx:        ctx,
		bkUsername: bkUsername,
		tenantID:   tenantID,
	}
}

// Deadline implement ITenantUserContext.
// nolint: nonamedreturns
func (ctx *TenantUserContext) Deadline() (deadline time.Time, ok bool) {
	return ctx.ctx.Deadline()
}

// Done implement ITenantUserContext.
func (ctx *TenantUserContext) Done() <-chan struct{} {
	return ctx.ctx.Done()
}

// Err implement ITenantUserContext.
func (ctx *TenantUserContext) Err() error {
	return ctx.ctx.Err()
}

// Value implement ITenantUserContext.
func (ctx *TenantUserContext) Value(key any) any {
	return ctx.ctx.Value(key)
}

// Values ... implement ITenantUserContext.
func (ctx *TenantUserContext) Values() map[string]any {
	return map[string]any{
		"bk_username": ctx.bkUsername,
		"tenant_id":   ctx.tenantID,
	}
}

// BKUsername implement ITenantUserContext.
func (ctx *TenantUserContext) BKUsername() string {
	return ctx.bkUsername
}

// TenantID implement ITenantUserContext.
func (ctx *TenantUserContext) TenantID() string {
	return ctx.tenantID
}
