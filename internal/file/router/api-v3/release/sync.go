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

package release

import (
	"fmt"

	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
)

// SyncSharedReleases synchronizes system shared releases to the authenticated tenant.
func (h *handler) SyncSharedReleases(rCtx restserver.IContext) (interface{}, error) {
	if rCtx.TenantID() == tenant.SystemTenantID {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("system tenant cannot sync shared releases"))
	}

	if err := h.manager.SyncSharedReleases(rCtx); err != nil {
		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to sync shared releases: %w", err))
	}

	resp := new(protoFile.SyncSharedReleasesResp)

	return resp.GetData(), nil
}
