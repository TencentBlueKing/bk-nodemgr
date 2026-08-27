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

package access

import (
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

// ITenantVirtualUserResolver resolves a login name to the tenant-scoped virtual-user bk username.
type ITenantVirtualUserResolver interface {
	// GetBKUsernameByLoginName gets the tenant-scoped virtual-user bk username by login name.
	GetBKUsernameByLoginName(ctx contextx.IContext, loginName string) (string, error)
}

// nolint: gochecknoglobals // tenant virtual-user resolver is configured during service initialization.
var tenantVirtualUserResolver = struct {
	sync.RWMutex
	resolver ITenantVirtualUserResolver
}{
	resolver: NewNoopTenantVirtualUserResolver(),
}

// SetTenantVirtualUserResolver sets the tenant-scoped virtual-user resolver.
func SetTenantVirtualUserResolver(resolver ITenantVirtualUserResolver) {
	tenantVirtualUserResolver.Lock()
	defer tenantVirtualUserResolver.Unlock()

	tenantVirtualUserResolver.resolver = resolver
}

// GetBKUsernameByLoginName gets the tenant-scoped virtual-user bk username by login name.
func GetBKUsernameByLoginName(ctx contextx.IContext, loginName string) (string, error) {
	tenantVirtualUserResolver.RLock()
	resolver := tenantVirtualUserResolver.resolver
	tenantVirtualUserResolver.RUnlock()

	return resolver.GetBKUsernameByLoginName(ctx, loginName)
}

// NewNoopTenantVirtualUserResolver creates a tenant-scoped virtual-user resolver that keeps login name as bk username.
func NewNoopTenantVirtualUserResolver() ITenantVirtualUserResolver {
	return noopTenantVirtualUserResolver{}
}

type noopTenantVirtualUserResolver struct{}

// GetBKUsernameByLoginName gets the tenant-scoped virtual-user bk username by login name.
func (noopTenantVirtualUserResolver) GetBKUsernameByLoginName(_ contextx.IContext, loginName string) (string, error) {
	return loginName, nil
}
