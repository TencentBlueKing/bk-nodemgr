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
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
)

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

func (c *cli) syncSharedReleases(nCtx contextx.IContext, tenantID string) error {
	resp := new(protoFile.SyncSharedReleasesResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return err
	}

	err = c.client.Post().
		SubResourcef("/release/sync_shared").
		WithContext(nCtx).
		WithHeaders(header).
		Body(new(protoFile.SyncSharedReleasesReq)).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("failed to sync shared releases. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}
