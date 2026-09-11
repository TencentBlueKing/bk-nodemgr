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

// ListReleasePlugin queries plugin release metadata.
func (h *handler) ListReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleasePluginListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query plugin release list, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	items, total, err := h.manager.ListReleasePlugin(rCtx, page, req.ConvertConditionsToTypes()...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query plugin release list")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to query plugin release list: %w", err))
	}

	resp := new(protoFile.ReleasePluginListResp)
	if err := resp.ConvertFromTypes(total, items); err != nil {
		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to convert plugin release list: %w", err))
	}

	return resp.GetData(), nil
}

// CountReleasePlugin queries plugin release metadata.
func (h *handler) CountReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleasePluginCountReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query plugin release count, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.manager.CountReleasePlugin(rCtx, req.ConvertConditionsToTypes()...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query plugin release count")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to query plugin release count: %w", err))
	}

	resp := new(protoFile.ReleasePluginCountResp)
	resp.ConvertFromTypes(result)

	return resp.GetData(), nil
}

// GetReleasePlugin queries plugin release metadata.
func (h *handler) GetReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleasePluginGetReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query plugin release get, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	key := req.GetIdentifier()
	result, err := h.manager.GetReleasePlugin(rCtx, key.Name, key.Generation, key.Platform, key.Version)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query plugin release get")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to query plugin release get: %w", err))
	}

	resp := new(protoFile.ReleasePluginGetResp)
	if err := resp.ConvertFromTypes(result); err != nil {
		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to convert plugin release get: %w", err))
	}

	return resp.GetData(), nil
}

// DistinctReleasePlugin queries plugin release metadata.
func (h *handler) DistinctReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleasePluginDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query plugin release distinct, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.manager.DistinctReleasePlugin(rCtx, req.ConvertDistinctFieldToTypes(), req.ConvertConditionsToTypes()...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query plugin release distinct")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to query plugin release distinct: %w", err))
	}

	resp := new(protoFile.ReleasePluginDistinctResp)
	resp.ConvertFromTypes(result)

	return resp.GetData(), nil
}

// ExistReleasePlugin queries plugin release metadata.
func (h *handler) ExistReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleasePluginExistReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query plugin release exist, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	total, err := h.manager.CountReleasePlugin(rCtx, req.ConvertConditionsToTypes()...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query plugin release exist")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to query plugin release exist: %w", err))
	}

	resp := new(protoFile.ReleasePluginExistResp)
	resp.ConvertFromTypes(total > 0)

	return resp.GetData(), nil
}

// GetReleasePluginDefaultVersion queries plugin release metadata.
func (h *handler) GetReleasePluginDefaultVersion(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleasePluginGetDefaultVersionReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query plugin release default_version, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	name, generation, plat := req.GetIdentifier()
	result, err := h.manager.GetReleasePluginDefaultVersion(rCtx, name, generation, plat)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query plugin release default_version")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to query plugin release default_version: %w", err))
	}

	resp := new(protoFile.ReleasePluginGetDefaultVersionResp)
	resp.ConvertFromTypes(result)

	return resp.GetData(), nil
}

// DistinctNameReleasePlugin queries plugin release metadata.
func (h *handler) DistinctNameReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleasePluginDistinctNameReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query plugin release distinct_name, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.manager.DistinctNameReleasePlugin(rCtx, req.ConvertConditionsToTypes()...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query plugin release distinct_name")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to query plugin release distinct_name: %w", err))
	}

	resp := new(protoFile.ReleasePluginDistinctNameResp)
	resp.ConvertFromTypes(result)

	return resp.GetData(), nil
}

// EnableReleasePlugin enables a plugin release.
func (h *handler) EnableReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleasePluginEnableReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to enable plugin release, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := h.manager.EnableReleasePlugin(rCtx, req.GetIdentifier()); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to enable plugin release")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to enable plugin release: %w", err))
	}

	return new(protoFile.ReleasePluginEnableResp).GetData(), nil
}

// DisableReleasePlugin disables a plugin release.
func (h *handler) DisableReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleasePluginDisableReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to disable plugin release, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	if err := h.manager.DisableReleasePlugin(rCtx, req.GetIdentifier()); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to disable plugin release")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to disable plugin release: %w", err))
	}

	return new(protoFile.ReleasePluginDisableResp).GetData(), nil
}

// SetAsDefaultReleasePlugin sets a plugin release as default.
func (h *handler) SetAsDefaultReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleasePluginSetAsDefaultReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set plugin release as default, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := h.manager.SetAsDefaultReleasePlugin(rCtx, req.GetIdentifier()); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set plugin release as default")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to set plugin release as default: %w", err))
	}

	return new(protoFile.ReleasePluginSetAsDefaultResp).GetData(), nil
}

// CancelAsDefaultReleasePlugin cancels a plugin release default.
func (h *handler) CancelAsDefaultReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleasePluginCancelAsDefaultReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to cancel plugin release default, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	if err := h.manager.CancelAsDefaultReleasePlugin(rCtx, req.GetIdentifier()); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to cancel plugin release default")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to cancel plugin release default: %w", err))
	}

	return new(protoFile.ReleasePluginCancelAsDefaultResp).GetData(), nil
}

// DeleteReleasePlugin deletes release metadata.
func (h *handler) DeleteReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleasePluginDeleteReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete plugin release, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	if err := h.manager.DeleteReleasePlugin(rCtx, req.GetIdentifier()); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete plugin release")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to delete plugin release: %w", err))
	}

	return new(protoFile.ReleasePluginDeleteResp).GetData(), nil
}

// SetHiddenReleasePlugin hides a plugin release.
func (h *handler) SetHiddenReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleasePluginSetHiddenReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set plugin release hidden, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	if err := h.manager.SetHiddenReleasePlugin(rCtx, req.GetIdentifier()); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set plugin release hidden")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to set plugin release hidden: %w", err))
	}

	return new(protoFile.ReleasePluginSetHiddenResp).GetData(), nil
}

// CancelHiddenReleasePlugin unhides a plugin release.
func (h *handler) CancelHiddenReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleasePluginCancelHiddenReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to cancel plugin release hidden, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	if err := h.manager.CancelHiddenReleasePlugin(rCtx, req.GetIdentifier()); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to cancel plugin release hidden")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to cancel plugin release hidden: %w", err))
	}

	return new(protoFile.ReleasePluginCancelHiddenResp).GetData(), nil
}
