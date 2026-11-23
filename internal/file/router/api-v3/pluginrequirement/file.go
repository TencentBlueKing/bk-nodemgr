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

package pluginrequirement

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// Create creates an independent plugin requirement file record.
func (h *handler) Create(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.PluginRequirementCreateReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create plugin requirement, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	record, err := h.manager.CreatePluginRequirement(
		rCtx, req.ConvertKeyToTypes(), req.GetTargetFileName(), req.GetContentBase64(), req.ConvertScopeToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create plugin requirement")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	resp := new(protoFile.PluginRequirementCreateResp)
	resp.ConvertResultFromTypes(record)

	return resp.GetData(), nil
}

// Update replaces the name and content of an existing requirement file.
func (h *handler) Update(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.PluginRequirementUpdateReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update plugin requirement, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	record, err := h.manager.UpdatePluginRequirement(
		rCtx, req.GetRequirementFileId(), req.GetName(), req.GetTargetFileName(), req.GetContentBase64())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update plugin requirement")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	resp := new(protoFile.PluginRequirementUpdateResp)
	resp.ConvertResultFromTypes(record)

	return resp.GetData(), nil
}

// UpdateScope replaces requirement file applicability without changing its content.
func (h *handler) UpdateScope(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.PluginRequirementUpdateScopeReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update plugin requirement scope, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	record, err := h.manager.UpdatePluginRequirementScope(rCtx, req.GetRequirementFileId(), req.ConvertScopeToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update plugin requirement scope")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	resp := new(protoFile.PluginRequirementUpdateScopeResp)
	resp.ConvertResultFromTypes(record)

	return resp.GetData(), nil
}

// List returns one page of requirement file metadata using exact filters.
func (h *handler) List(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.PluginRequirementListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin requirements, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin requirements, invalid page")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	records, total, err := h.manager.ListPluginRequirements(rCtx, page, req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin requirements")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	resp := new(protoFile.PluginRequirementListResp)
	resp.ConvertResultFromTypes(records, total)

	return resp.GetData(), nil
}

// Get returns requirement file metadata without its content.
func (h *handler) Get(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.PluginRequirementGetReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get plugin requirement, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	record, err := h.manager.GetPluginRequirement(rCtx, req.GetRequirementFileId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get plugin requirement")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	resp := new(protoFile.PluginRequirementGetResp)
	resp.ConvertResultFromTypes(record)

	return resp.GetData(), nil
}

// Download returns the requirement file stream for FileHandler to close.
func (h *handler) Download(rCtx restserver.IContext) (*restserver.FileResponse, error) {
	req := &protoFile.PluginRequirementDownloadReq{RequirementFileId: rCtx.GContext().Query("requirement_file_id")}
	if err := req.Validate(); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download plugin requirement, invalid request")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	file, err := h.manager.ReadPluginRequirement(rCtx, req.GetRequirementFileId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to read plugin requirement")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	content, err := file.Content(rCtx)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to open plugin requirement content")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	info := file.Info()

	return &restserver.FileResponse{
		Data:        content,
		Size:        info.Size,
		FileName:    info.Name,
		ContentType: restserver.MIMETypeBin,
	}, nil
}
