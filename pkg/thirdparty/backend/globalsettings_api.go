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

package backend

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
)

func (c *cli) listGlobalSettings(ctx contextx.IContext, req *protoBackend.ListGlobalSettingsReq) (
	*protoBackend.ListGlobalSettingsResp, error) {

	resp := new(protoBackend.ListGlobalSettingsResp)
	// Stored custom values may be sensitive; do not log the response body.
	err := c.client.Post().SubResourcef("/globalsettings/list").
		WithContext(ctx).WithHeaders(c.getHeader(ctx)).Body(req).
		EnableLogBody().Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to request global settings list: %w", err)
	}
	if resp.GetCode() != CodeOK {
		return nil, buildBackendResponseError("list global settings", resp, resp.GetError())
	}
	if resp.GetData() == nil {
		return nil, fmt.Errorf("list global settings returned empty data, request-id(%s)", resp.GetRequestId())
	}

	return resp, nil
}
