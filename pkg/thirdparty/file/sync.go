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

package file

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
)

// ISharedReleaseSyncHandler synchronizes system shared releases for a tenant.
type ISharedReleaseSyncHandler interface {
	// SyncSharedReleases synchronizes system shared releases to the current tenant.
	SyncSharedReleases(nCtx contextx.IContext) error
}

// SyncSharedReleases synchronizes system shared releases to the current tenant.
func (h *handler) SyncSharedReleases(nCtx contextx.IContext) error {
	if nCtx == nil {
		return fmt.Errorf("context is nil")
	}
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}
	if nCtx.TenantID() == tenant.SystemTenantID {
		return fmt.Errorf("system tenant cannot sync shared releases")
	}

	return h.cli.syncSharedReleases(nCtx, nCtx.TenantID())
}
