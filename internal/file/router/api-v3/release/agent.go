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

// ListReleaseAgent queries agent release metadata.
func (h *handler) ListReleaseAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleaseAgentListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query agent release list, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	items, total, err := h.manager.ListReleaseAgent(rCtx, page, req.ConvertConditionsToTypes()...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query agent release list")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to query agent release list: %w", err))
	}

	resp := new(protoFile.ReleaseAgentListResp)
	if err := resp.ConvertFromTypes(total, items); err != nil {
		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to convert agent release list: %w", err))
	}

	return resp.GetData(), nil
}

// CountReleaseAgent queries agent release metadata.
func (h *handler) CountReleaseAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleaseAgentCountReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query agent release count, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.manager.CountReleaseAgent(rCtx, req.ConvertConditionsToTypes()...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query agent release count")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to query agent release count: %w", err))
	}

	resp := new(protoFile.ReleaseAgentCountResp)
	resp.ConvertFromTypes(result)

	return resp.GetData(), nil
}

// GetReleaseAgent queries agent release metadata.
func (h *handler) GetReleaseAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleaseAgentGetReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query agent release get, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	key := req.GetIdentifier()
	result, err := h.manager.GetReleaseAgent(rCtx, key.Generation, key.Platform, key.Version)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query agent release get")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to query agent release get: %w", err))
	}

	resp := new(protoFile.ReleaseAgentGetResp)
	if err := resp.ConvertFromTypes(result); err != nil {
		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to convert agent release get: %w", err))
	}

	return resp.GetData(), nil
}

// DistinctReleaseAgent queries agent release metadata.
func (h *handler) DistinctReleaseAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleaseAgentDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query agent release distinct, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.manager.DistinctReleaseAgent(rCtx, req.ConvertDistinctFieldToTypes(), req.ConvertConditionsToTypes()...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query agent release distinct")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to query agent release distinct: %w", err))
	}

	resp := new(protoFile.ReleaseAgentDistinctResp)
	resp.ConvertFromTypes(result)

	return resp.GetData(), nil
}

// EnableReleaseAgent enables an agent release.
func (h *handler) EnableReleaseAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleaseAgentEnableReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to enable agent release, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := h.manager.EnableReleaseAgent(rCtx, req.GetIdentifier()); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to enable agent release")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to enable agent release: %w", err))
	}

	return new(protoFile.ReleaseAgentEnableResp).GetData(), nil
}

// DisableReleaseAgent disables an agent release.
func (h *handler) DisableReleaseAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleaseAgentDisableReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to disable agent release, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	if err := h.manager.DisableReleaseAgent(rCtx, req.GetIdentifier()); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to disable agent release")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to disable agent release: %w", err))
	}

	return new(protoFile.ReleaseAgentDisableResp).GetData(), nil
}

// SetAsDefaultReleaseAgent sets an agent release as default.
func (h *handler) SetAsDefaultReleaseAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleaseAgentSetAsDefaultReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set agent release as default, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := h.manager.SetAsDefaultReleaseAgent(rCtx, req.GetIdentifier()); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set agent release as default")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to set agent release as default: %w", err))
	}

	return new(protoFile.ReleaseAgentSetAsDefaultResp).GetData(), nil
}

// CancelAsDefaultReleaseAgent cancels an agent release default.
func (h *handler) CancelAsDefaultReleaseAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleaseAgentCancelAsDefaultReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to cancel agent release default, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	if err := h.manager.CancelAsDefaultReleaseAgent(rCtx, req.GetIdentifier()); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to cancel agent release default")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to cancel agent release default: %w", err))
	}

	return new(protoFile.ReleaseAgentCancelAsDefaultResp).GetData(), nil
}

// DeleteReleaseAgent deletes release metadata.
func (h *handler) DeleteReleaseAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleaseAgentDeleteReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete agent release, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	if err := h.manager.DeleteReleaseAgent(rCtx, req.GetIdentifier()); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete agent release")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to delete agent release: %w", err))
	}

	return new(protoFile.ReleaseAgentDeleteResp).GetData(), nil
}

// SetReleaseAgentLabelsMany updates labels for matching releases.
func (h *handler) SetReleaseAgentLabelsMany(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ReleaseAgentSetLabelsManyReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set agent release labels, failed to bind request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	labels, conditions := req.ConvertToTypes()
	if err := h.manager.SetReleaseAgentLabelsMany(rCtx, labels, conditions...); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set agent release labels")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to set agent release labels: %w", err))
	}

	return new(protoFile.ReleaseAgentSetLabelsManyResp).GetData(), nil
}
