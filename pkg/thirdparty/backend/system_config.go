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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandlerSystemConfig provides read-only Backend configuration access.
type IHandlerSystemConfig interface {
	GetSystemConfig(nCtx contextx.IContext) (types.SystemConfig, error)
}

// GetSystemConfig gets the Backend startup configuration whitelist.
func (h *Handler) GetSystemConfig(nCtx contextx.IContext) (types.SystemConfig, error) {
	resp, err := h.cli.getSystemConfig(nCtx, new(protoBackend.GetSystemConfigReq))
	if err != nil {
		return types.SystemConfig{}, fmt.Errorf("failed to get system config: %w", err)
	}

	return resp.GetData().ConvertToTypes(), nil
}
