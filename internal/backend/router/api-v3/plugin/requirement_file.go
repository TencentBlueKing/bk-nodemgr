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

// nolint: revive
package plugin

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// CreateRequirementFile stores a plugin's dependency file through File.
func (h *handler) CreateRequirementFile(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PluginRequirementCreateReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create plugin requirement, invalid request")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := h.authorizeRequirementPlugin(rCtx, auth.ActionPluginOperate, req.GetPluginName()); err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			With("plugin-name", req.GetPluginName()).
			Error("failed to create plugin requirement, permission denied")

		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, err)
	}

	record, err := h.requirementFiles.CreatePluginRequirement(
		rCtx, req.ConvertKeyToTypes(), req.GetTargetFileName(), req.GetContentBase64(), req.ConvertScopeToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create plugin requirement")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.PluginRequirementCreateResp)
	resp.ConvertResultFromTypes(record)

	return resp.GetData(), nil
}

// UpdateRequirementFile replaces a dependency's names and content.
func (h *handler) UpdateRequirementFile(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PluginRequirementUpdateReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update plugin requirement, invalid request")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	record, err := h.requirementFiles.GetPluginRequirement(rCtx, req.GetRequirementFileId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			With("requirement-file-id", req.GetRequirementFileId()).
			Error("failed to update plugin requirement, failed to get requirement file")

		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}
	if err := h.authorizeRequirementPlugin(rCtx, auth.ActionPluginOperate, record.PluginName); err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			With("requirement-file-id", req.GetRequirementFileId()).
			Error("failed to update plugin requirement, permission denied")

		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, err)
	}

	record, err = h.requirementFiles.UpdatePluginRequirement(
		rCtx, req.GetRequirementFileId(), req.GetName(), req.GetTargetFileName(), req.GetContentBase64())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update plugin requirement")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.PluginRequirementUpdateResp)
	resp.ConvertResultFromTypes(record)

	return resp.GetData(), nil
}

// UpdateRequirementFileScope replaces a dependency's applicability.
func (h *handler) UpdateRequirementFileScope(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PluginRequirementUpdateScopeReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update plugin requirement scope, invalid request")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	record, err := h.requirementFiles.GetPluginRequirement(rCtx, req.GetRequirementFileId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			With("requirement-file-id", req.GetRequirementFileId()).
			Error("failed to update plugin requirement scope, failed to get requirement file")

		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}
	if err := h.authorizeRequirementPlugin(rCtx, auth.ActionPluginOperate, record.PluginName); err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			With("requirement-file-id", req.GetRequirementFileId()).
			Error("failed to update plugin requirement scope, permission denied")

		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, err)
	}

	record, err = h.requirementFiles.UpdatePluginRequirementScope(
		rCtx, req.GetRequirementFileId(), req.ConvertScopeToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update plugin requirement scope")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.PluginRequirementUpdateScopeResp)
	resp.ConvertResultFromTypes(record)

	return resp.GetData(), nil
}

// GetRequirementFile returns a dependency's metadata.
func (h *handler) GetRequirementFile(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PluginRequirementGetReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get plugin requirement, invalid request")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	record, err := h.requirementFiles.GetPluginRequirement(rCtx, req.GetRequirementFileId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			With("requirement-file-id", req.GetRequirementFileId()).
			Error("failed to get plugin requirement, failed to get requirement file")

		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}
	if err := h.authorizeRequirementPlugin(rCtx, auth.ActionPluginView, record.PluginName); err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			With("requirement-file-id", req.GetRequirementFileId()).
			Error("failed to get plugin requirement, permission denied")

		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, err)
	}

	resp := new(protoBackend.PluginRequirementGetResp)
	resp.ConvertResultFromTypes(record)

	return resp.GetData(), nil
}

// ListRequirementFiles queries a page within the caller's plugin view scope.
func (h *handler) ListRequirementFiles(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PluginRequirementListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin requirements, invalid request")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin requirements, invalid page")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	condition := req.ConvertConditionsToTypes()
	pluginNames, scopeIsAny, err := h.authorizedPluginNames(rCtx, auth.ActionPluginView, condition.ExactInclude.PluginName)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin requirements, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, err)
	}

	resp := new(protoBackend.PluginRequirementListResp)
	if !scopeIsAny && len(pluginNames) == 0 {
		resp.ConvertResultFromTypes(nil, 0)
		return resp.GetData(), nil
	}

	condition.ExactInclude.PluginName = pluginNames
	records, total, err := h.requirementFiles.ListPluginRequirements(rCtx, page, condition)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin requirements")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp.ConvertResultFromTypes(records, total)

	return resp.GetData(), nil
}

// DownloadRequirementFile streams a dependency's contents through Backend.
func (h *handler) DownloadRequirementFile(rCtx restserver.IContext) (*restserver.StreamResponse, error) {
	req := &protoBackend.PluginRequirementDownloadReq{RequirementFileId: rCtx.GContext().Query("requirement_file_id")}
	if err := req.Validate(); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download plugin requirement, invalid request")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	record, err := h.requirementFiles.GetPluginRequirement(rCtx, req.GetRequirementFileId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			With("requirement-file-id", req.GetRequirementFileId()).
			Error("failed to download plugin requirement, failed to get requirement file")

		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}
	if err := h.authorizeRequirementPlugin(rCtx, auth.ActionPluginView, record.PluginName); err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			With("requirement-file-id", req.GetRequirementFileId()).
			Error("failed to download plugin requirement, permission denied")

		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, err)
	}

	stream, err := h.requirementFiles.DownloadPluginRequirement(rCtx, req.GetRequirementFileId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download plugin requirement")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	return stream, nil
}
