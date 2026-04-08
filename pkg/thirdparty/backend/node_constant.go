/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package backend

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandlerNodeConstant defines the node constant Handler.
// nolint: interfacebloat
type IHandlerNodeConstant interface {
	// GetDeployConstant get deploy constant by generation and os type.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param generation the deploy constant generation.
	// @param osType the os type.
	// @return the custom deploy config and error.
	GetDeployConstant(nCtx contextx.IContext, generation types.Generation, osType criteria.OSType) (*types.CustomDeployConfig, error)
}

// GetDeployConstant get deploy constant by generation and os type.
func (h *Handler) GetDeployConstant(nCtx contextx.IContext, generation types.Generation, osType criteria.OSType) (*types.CustomDeployConfig, error) {
	req := &protoBackend.NodeConstantDeployGetReq{
		Generation: int64(generation),
		OsType:     string(osType),
	}

	resp, err := h.cli.getDeployConstant(nCtx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertConstantToTypes(), nil
}
