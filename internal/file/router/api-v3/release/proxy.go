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

// ListReleaseProxy queries proxy release metadata.
func (h *handler) ListReleaseProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleaseProxyListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query proxy release list, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	items, total, err := h.manager.ListReleaseProxy(rCtx, page, req.ConvertConditionsToTypes()...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query proxy release list")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to query proxy release list: %w", err))
	}

	resp := new(protoFile.ReleaseProxyListResp)
	if err := resp.ConvertFromTypes(total, items); err != nil {
		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to convert proxy release list: %w", err))
	}

	return resp.GetData(), nil
}

// CountReleaseProxy queries proxy release metadata.
func (h *handler) CountReleaseProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleaseProxyCountReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query proxy release count, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.manager.CountReleaseProxy(rCtx, req.ConvertConditionsToTypes()...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query proxy release count")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to query proxy release count: %w", err))
	}

	resp := new(protoFile.ReleaseProxyCountResp)
	resp.ConvertFromTypes(result)

	return resp.GetData(), nil
}

// GetReleaseProxy queries proxy release metadata.
func (h *handler) GetReleaseProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleaseProxyGetReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query proxy release get, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	key := req.GetIdentifier()
	result, err := h.manager.GetReleaseProxy(rCtx, key.Generation, key.Platform, key.Version)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query proxy release get")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to query proxy release get: %w", err))
	}

	resp := new(protoFile.ReleaseProxyGetResp)
	if err := resp.ConvertFromTypes(result); err != nil {
		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to convert proxy release get: %w", err))
	}

	return resp.GetData(), nil
}

// DistinctReleaseProxy queries proxy release metadata.
func (h *handler) DistinctReleaseProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleaseProxyDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query proxy release distinct, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.manager.DistinctReleaseProxy(rCtx, req.ConvertDistinctFieldToTypes(), req.ConvertConditionsToTypes()...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query proxy release distinct")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to query proxy release distinct: %w", err))
	}

	resp := new(protoFile.ReleaseProxyDistinctResp)
	resp.ConvertFromTypes(result)

	return resp.GetData(), nil
}

// EnableReleaseProxy enables a proxy release.
func (h *handler) EnableReleaseProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleaseProxyEnableReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to enable proxy release, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := h.manager.EnableReleaseProxy(rCtx, req.GetIdentifier()); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to enable proxy release")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to enable proxy release: %w", err))
	}

	return new(protoFile.ReleaseProxyEnableResp).GetData(), nil
}

// DisableReleaseProxy disables a proxy release.
func (h *handler) DisableReleaseProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleaseProxyDisableReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to disable proxy release, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	if err := h.manager.DisableReleaseProxy(rCtx, req.GetIdentifier()); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to disable proxy release")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to disable proxy release: %w", err))
	}

	return new(protoFile.ReleaseProxyDisableResp).GetData(), nil
}

// SetAsDefaultReleaseProxy sets a proxy release as default.
func (h *handler) SetAsDefaultReleaseProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleaseProxySetAsDefaultReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set proxy release as default, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := h.manager.SetAsDefaultReleaseProxy(rCtx, req.GetIdentifier()); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set proxy release as default")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to set proxy release as default: %w", err))
	}

	return new(protoFile.ReleaseProxySetAsDefaultResp).GetData(), nil
}

// CancelAsDefaultReleaseProxy cancels a proxy release default.
func (h *handler) CancelAsDefaultReleaseProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleaseProxyCancelAsDefaultReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to cancel proxy release default, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	if err := h.manager.CancelAsDefaultReleaseProxy(rCtx, req.GetIdentifier()); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to cancel proxy release default")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to cancel proxy release default: %w", err))
	}

	return new(protoFile.ReleaseProxyCancelAsDefaultResp).GetData(), nil
}

// DeleteReleaseProxy deletes release metadata.
func (h *handler) DeleteReleaseProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleaseProxyDeleteReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete proxy release, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	if err := h.manager.DeleteReleaseProxy(rCtx, req.GetIdentifier()); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete proxy release")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to delete proxy release: %w", err))
	}

	return new(protoFile.ReleaseProxyDeleteResp).GetData(), nil
}

// SetReleaseProxyLabelsMany updates labels for matching releases.
func (h *handler) SetReleaseProxyLabelsMany(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleaseProxySetLabelsManyReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set proxy release labels, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	labels, conditions := req.ConvertToTypes()
	if err := h.manager.SetReleaseProxyLabelsMany(rCtx, labels, conditions...); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set proxy release labels")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to set proxy release labels: %w", err))
	}

	return new(protoFile.ReleaseProxySetLabelsManyResp).GetData(), nil
}
