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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// ListReleaseBinTool queries bintool release metadata.
func (h *handler) ListReleaseBinTool(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleaseBinToolListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query bintool release list, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	items, total, err := h.manager.ListReleaseBinTool(rCtx, page, req.ConvertConditionsToTypes()...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query bintool release list")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to query bintool release list: %w", err))
	}

	resp := new(protoFile.ReleaseBinToolListResp)
	if err := resp.ConvertFromTypes(total, items); err != nil {
		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to convert bintool release list: %w", err))
	}

	return resp.GetData(), nil
}

// DeleteReleaseBinTool deletes release metadata.
func (h *handler) DeleteReleaseBinTool(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleaseBinToolDeleteReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete bintool release, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	if err := h.manager.DeleteReleaseBinTool(rCtx, req.GetIdentifier()); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete bintool release")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to delete bintool release: %w", err))
	}

	return new(protoFile.ReleaseBinToolDeleteResp).GetData(), nil
}
