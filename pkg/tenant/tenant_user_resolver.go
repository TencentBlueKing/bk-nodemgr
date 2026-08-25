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

package tenant

import (
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

// ITenantUserResolver resolves a login name to the tenant-scoped bk username.
type ITenantUserResolver interface {
	// GetBKUsernameByLoginName gets the tenant-scoped bk username by login name.
	GetBKUsernameByLoginName(ctx contextx.IContext, loginName string) (string, error)
}

var tenantUserResolver = struct {
	sync.RWMutex
	resolver ITenantUserResolver
}{
	resolver: NewNoopTenantUserResolver(),
}

// SetTenantUserResolver sets the tenant user resolver.
func SetTenantUserResolver(resolver ITenantUserResolver) {
	tenantUserResolver.Lock()
	defer tenantUserResolver.Unlock()

	tenantUserResolver.resolver = resolver
}

// GetBKUsernameByLoginName gets the tenant-scoped bk username by login name.
func GetBKUsernameByLoginName(ctx contextx.IContext, loginName string) (string, error) {
	tenantUserResolver.RLock()
	resolver := tenantUserResolver.resolver
	tenantUserResolver.RUnlock()

	return resolver.GetBKUsernameByLoginName(ctx, loginName)
}

// NewNoopTenantUserResolver creates a resolver that keeps login name as bk username.
func NewNoopTenantUserResolver() ITenantUserResolver {
	return noopTenantUserResolver{}
}

type noopTenantUserResolver struct{}

// GetBKUsernameByLoginName gets the tenant-scoped bk username by login name.
func (noopTenantUserResolver) GetBKUsernameByLoginName(_ contextx.IContext, loginName string) (string, error) {
	return loginName, nil
}
