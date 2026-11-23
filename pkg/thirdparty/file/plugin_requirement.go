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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// CreatePluginRequirement forwards a named plugin dependency to File without decoding its content.
func (h *handler) CreatePluginRequirement(
	nCtx contextx.IContext, key types.PluginRequirementFileKey, targetFileName, contentBase64 string,
	scope types.PluginRequirementFileScope,
) (*types.PluginRequirementFileRecord, error) {

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
func (h *handler) UpdatePluginRequirement(
	nCtx contextx.IContext, requirementFileID, name, targetFileName, contentBase64 string,
) (*types.PluginRequirementFileRecord, error) {

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
func (h *handler) UpdatePluginRequirementScope(
	nCtx contextx.IContext, id string, scope types.PluginRequirementFileScope,
) (*types.PluginRequirementFileRecord, error) {

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
func (h *handler) ListPluginRequirements(
	nCtx contextx.IContext, page types.Page, condition *types.PluginRequirementFileCondition,
) ([]*types.PluginRequirementFileRecord, int64, error) {

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
func (h *handler) GetPluginRequirement(
	nCtx contextx.IContext, id string,
) (*types.PluginRequirementFileRecord, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	resp, err := h.cli.getPluginRequirement(nCtx, &protoFile.PluginRequirementGetReq{RequirementFileId: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin requirement: %w", err)
	}

	return resp.GetData().ConvertFileToTypes(), nil
}

// LaunchTransferPluginRequirement transfers a stored file to a host.
func (h *handler) LaunchTransferPluginRequirement(
	nCtx contextx.IContext, requirementFileID, targetDir string, targetHost *types.Host,
) (types.ISimpleTransferHandler, error) {

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

func (c *cli) createPluginRequirement(
	nCtx contextx.IContext, req *protoFile.PluginRequirementCreateReq,
) (*protoFile.PluginRequirementCreateResp, error) {

	header, err := c.getCommonHeader(nCtx, nCtx.TenantID())
	if err != nil {
		return nil, fmt.Errorf("failed to generate file service headers: %w", err)
	}

	resp := new(protoFile.PluginRequirementCreateResp)
	if err := c.client.Post().
		SubResourcef("/plugin/requirement/create").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().
		Into(resp); err != nil {
		return nil, fmt.Errorf("failed to post plugin requirement create: %w", err)
	}
	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("plugin requirement create failed: code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}
	if resp.GetData().GetFile() == nil {
		return nil, fmt.Errorf("plugin requirement create returned empty file, request-id(%s)", resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) updatePluginRequirement(
	nCtx contextx.IContext, req *protoFile.PluginRequirementUpdateReq,
) (*protoFile.PluginRequirementUpdateResp, error) {

	header, err := c.getCommonHeader(nCtx, nCtx.TenantID())
	if err != nil {
		return nil, fmt.Errorf("failed to generate file service headers: %w", err)
	}

	resp := new(protoFile.PluginRequirementUpdateResp)
	if err := c.client.Post().
		SubResourcef("/plugin/requirement/update").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().
		Into(resp); err != nil {
		return nil, fmt.Errorf("failed to post plugin requirement update: %w", err)
	}
	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("plugin requirement update failed: code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}
	if resp.GetData().GetFile() == nil {
		return nil, fmt.Errorf("plugin requirement update returned empty file, request-id(%s)", resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) updatePluginRequirementScope(
	nCtx contextx.IContext, req *protoFile.PluginRequirementUpdateScopeReq,
) (*protoFile.PluginRequirementUpdateScopeResp, error) {

	header, err := c.getCommonHeader(nCtx, nCtx.TenantID())
	if err != nil {
		return nil, fmt.Errorf("failed to generate file service headers: %w", err)
	}

	resp := new(protoFile.PluginRequirementUpdateScopeResp)
	if err := c.client.Post().
		SubResourcef("/plugin/requirement/update_scope").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().
		Into(resp); err != nil {
		return nil, fmt.Errorf("failed to post plugin requirement update scope: %w", err)
	}
	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("plugin requirement update scope failed: code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}
	if resp.GetData().GetFile() == nil {
		return nil, fmt.Errorf("plugin requirement update scope returned empty file, request-id(%s)", resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listPluginRequirements(
	nCtx contextx.IContext, req *protoFile.PluginRequirementListReq,
) (*protoFile.PluginRequirementListResp, error) {

	header, err := c.getCommonHeader(nCtx, nCtx.TenantID())
	if err != nil {
		return nil, fmt.Errorf("failed to generate file service headers: %w", err)
	}

	resp := new(protoFile.PluginRequirementListResp)
	if err := c.client.Post().
		SubResourcef("/plugin/requirement/list").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().
		Into(resp); err != nil {
		return nil, fmt.Errorf("failed to post plugin requirement list: %w", err)
	}
	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("plugin requirement list failed: code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}
	if resp.GetData() == nil {
		return nil, fmt.Errorf("plugin requirement list returned empty data, request-id(%s)", resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) getPluginRequirement(
	nCtx contextx.IContext, req *protoFile.PluginRequirementGetReq,
) (*protoFile.PluginRequirementGetResp, error) {

	header, err := c.getCommonHeader(nCtx, nCtx.TenantID())
	if err != nil {
		return nil, fmt.Errorf("failed to generate file service headers: %w", err)
	}

	resp := new(protoFile.PluginRequirementGetResp)
	if err := c.client.Post().
		SubResourcef("/plugin/requirement/get").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().
		Into(resp); err != nil {
		return nil, fmt.Errorf("failed to post plugin requirement get: %w", err)
	}
	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("plugin requirement get failed: code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}
	if resp.GetData().GetFile() == nil {
		return nil, fmt.Errorf("plugin requirement get returned empty file, request-id(%s)", resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) transferPluginRequirement(
	nCtx contextx.IContext, req *protoFile.PluginRequirementTransferReq,
) (*protoFile.PluginRequirementTransferResp, error) {

	header, err := c.getCommonHeader(nCtx, nCtx.TenantID())
	if err != nil {
		return nil, fmt.Errorf("failed to generate file service headers: %w", err)
	}

	resp := new(protoFile.PluginRequirementTransferResp)
	if err := c.client.Post().
		SubResourcef("/plugin/requirement/transfer").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().
		Into(resp); err != nil {
		return nil, fmt.Errorf("failed to post plugin requirement transfer: %w", err)
	}
	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("plugin requirement transfer failed: code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}
	if resp.GetData() == nil {
		return nil, fmt.Errorf("plugin requirement transfer returned empty data, request-id(%s)", resp.GetRequestId())
	}
	if resp.GetData().GetTaskId() == "" {
		return nil, fmt.Errorf("plugin requirement transfer returned empty task id, request-id(%s)", resp.GetRequestId())
	}

	return resp, nil
}
