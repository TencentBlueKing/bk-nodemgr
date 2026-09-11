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

// ListReleasePluginBinTool queries plugin_bintool release metadata.
func (h *handler) ListReleasePluginBinTool(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleasePluginBinToolListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query plugin_bintool release list, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	items, total, err := h.manager.ListReleasePluginBinTool(rCtx, page, req.ConvertConditionsToTypes()...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query plugin_bintool release list")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to query plugin_bintool release list: %w", err))
	}

	resp := new(protoFile.ReleasePluginBinToolListResp)
	if err := resp.ConvertFromTypes(total, items); err != nil {
		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to convert plugin_bintool release list: %w", err))
	}

	return resp.GetData(), nil
}

// DistinctNameReleasePluginBinTool queries plugin_bintool release metadata.
func (h *handler) DistinctNameReleasePluginBinTool(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleasePluginBinToolDistinctNameReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query plugin_bintool release distinct_name, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.manager.DistinctNameReleasePluginBinTool(rCtx, req.ConvertConditionsToTypes()...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query plugin_bintool release distinct_name")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to query plugin_bintool release distinct_name: %w", err))
	}

	resp := new(protoFile.ReleasePluginBinToolDistinctNameResp)
	resp.ConvertFromTypes(result)

	return resp.GetData(), nil
}

// DeleteReleasePluginBinTool deletes release metadata.
func (h *handler) DeleteReleasePluginBinTool(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleasePluginBinToolDeleteReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete pluginbintool release, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	if err := h.manager.DeleteReleasePluginBinTool(rCtx, req.GetIdentifier()); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete pluginbintool release")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to delete pluginbintool release: %w", err))
	}

	return new(protoFile.ReleasePluginBinToolDeleteResp).GetData(), nil
}
