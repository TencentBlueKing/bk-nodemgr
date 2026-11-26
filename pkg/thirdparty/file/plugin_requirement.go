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

package file

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IPluginRequirementHandler manages and transfers plugin requirement files.
type IPluginRequirementHandler interface {
	// CreatePluginRequirement stores a named plugin dependency with its filename and applicability.
	CreatePluginRequirement(
		nCtx contextx.IContext, key types.PluginRequirementFileKey, targetFileName, contentBase64 string, scope types.PluginRequirementFileScope) (
		*types.PluginRequirementFileRecord, error)

	// UpdatePluginRequirement replaces content and names without changing plugin ownership or applicability.
	UpdatePluginRequirement(nCtx contextx.IContext, requirementFileID, name, targetFileName, contentBase64 string) (
		*types.PluginRequirementFileRecord, error)

	// UpdatePluginRequirementScope replaces a stored file's applicability without replacing its content.
	UpdatePluginRequirementScope(nCtx contextx.IContext, id string, scope types.PluginRequirementFileScope) (
		*types.PluginRequirementFileRecord, error)

	// ListPluginRequirements retrieves one page of file records using exact filters.
	ListPluginRequirements(nCtx contextx.IContext, page types.Page, condition *types.PluginRequirementFileCondition) (
		[]*types.PluginRequirementFileRecord, int64, error)

	// GetPluginRequirement retrieves a stored file record by ID.
	GetPluginRequirement(nCtx contextx.IContext, id string) (*types.PluginRequirementFileRecord, error)

	// DownloadPluginRequirement streams a stored file by ID.
	DownloadPluginRequirement(nCtx contextx.IContext, id string) (*restserver.StreamResponse, error)

	// LaunchTransferPluginRequirement transfers a stored file to a host.
	LaunchTransferPluginRequirement(nCtx contextx.IContext, requirementFileID, targetDir string, targetHost *types.Host) (
		types.ISimpleTransferHandler, error)
}

// CreatePluginRequirement forwards a named plugin dependency to File without decoding its content.
func (h *handler) CreatePluginRequirement(
	nCtx contextx.IContext, key types.PluginRequirementFileKey, targetFileName, contentBase64 string, scope types.PluginRequirementFileScope) (
	*types.PluginRequirementFileRecord, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	req := &protoFile.PluginRequirementCreateReq{
		TargetFileName: targetFileName,
		ContentBase64:  contentBase64,
	}
	req.ConvertKeyFromTypes(key)
	req.ConvertScopeFromTypes(scope)
	resp, err := h.cli.createPluginRequirement(nCtx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin requirement: %w", err)
	}

	return resp.GetData().ConvertFileToTypes(), nil
}

// UpdatePluginRequirement replaces a stored file by ID.
func (h *handler) UpdatePluginRequirement(nCtx contextx.IContext, requirementFileID, name, targetFileName, contentBase64 string) (
	*types.PluginRequirementFileRecord, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	resp, err := h.cli.updatePluginRequirement(nCtx, &protoFile.PluginRequirementUpdateReq{
		RequirementFileId: requirementFileID,
		Name:              name,
		TargetFileName:    targetFileName,
		ContentBase64:     contentBase64,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update plugin requirement: %w", err)
	}

	return resp.GetData().ConvertFileToTypes(), nil
}

// UpdatePluginRequirementScope replaces a stored file's applicability without replacing its content.
func (h *handler) UpdatePluginRequirementScope(nCtx contextx.IContext, id string, scope types.PluginRequirementFileScope) (
	*types.PluginRequirementFileRecord, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	req := &protoFile.PluginRequirementUpdateScopeReq{RequirementFileId: id}
	req.ConvertScopeFromTypes(scope)
	resp, err := h.cli.updatePluginRequirementScope(nCtx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update plugin requirement scope: %w", err)
	}

	return resp.GetData().ConvertFileToTypes(), nil
}

// ListPluginRequirements retrieves one page of file records using exact filters.
func (h *handler) ListPluginRequirements(nCtx contextx.IContext, page types.Page, condition *types.PluginRequirementFileCondition) (
	[]*types.PluginRequirementFileRecord, int64, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}

	req := &protoFile.PluginRequirementListReq{
		Page:       new(protoFile.Page),
		Conditions: protoFile.ConvertPluginRequirementFileConditionsFromTypes(condition),
	}
	req.GetPage().ConvertFromTypes(page)
	resp, err := h.cli.listPluginRequirements(nCtx, req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list plugin requirements: %w", err)
	}

	return resp.GetData().ConvertFilesToTypes(), resp.GetData().GetTotal(), nil
}

// GetPluginRequirement retrieves a stored file record by ID.
func (h *handler) GetPluginRequirement(nCtx contextx.IContext, id string) (*types.PluginRequirementFileRecord, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	resp, err := h.cli.getPluginRequirement(nCtx, &protoFile.PluginRequirementGetReq{RequirementFileId: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin requirement: %w", err)
	}

	return resp.GetData().ConvertFileToTypes(), nil
}

// DownloadPluginRequirement streams a stored file by ID.
func (h *handler) DownloadPluginRequirement(nCtx contextx.IContext, id string) (*restserver.StreamResponse, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	resp, err := h.cli.downloadPluginRequirement(nCtx, &protoFile.PluginRequirementDownloadReq{RequirementFileId: id})
	if err != nil {
		return nil, fmt.Errorf("failed to download plugin requirement: %w", err)
	}

	return resp, nil
}

// LaunchTransferPluginRequirement transfers a stored file to a host.
func (h *handler) LaunchTransferPluginRequirement(nCtx contextx.IContext, requirementFileID, targetDir string, targetHost *types.Host) (
	types.ISimpleTransferHandler, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	if targetHost == nil {
		return nil, fmt.Errorf("target host is nil")
	}

	resp, err := h.cli.transferPluginRequirement(nCtx, &protoFile.PluginRequirementTransferReq{
		RequirementFileId: requirementFileID,
		TargetDir:         targetDir,
		TargetHostId:      targetHost.HostID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to transfer plugin requirement: %w", err)
	}

	data := resp.GetData()

	return &simpleTransferHandler{
		taskID: data.GetTaskId(),
		fileInfo: fileiface.FileInfo{
			Name: data.GetTargetFileName(),
			MD5:  data.GetFileMd5(),
			Size: data.GetFileSize(),
		},
		handler: h,
	}, nil
}
