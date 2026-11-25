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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
)

// IPkgPublishHandler defines the interface of pkg publish.
type IPkgPublishHandler interface {
	// PublishReleaseAgent publish release agent.
	PublishReleaseAgent(nCtx contextx.IContext, uploadID string, isShared bool) error

	// PublishReleaseProxy publish release server.
	PublishReleaseProxy(nCtx contextx.IContext, uploadID, uploadCategory string, isShared bool) error

	// PublishReleaseCert publish release cert.
	PublishReleaseCert(nCtx contextx.IContext, uploadID string, isShared bool) error

	// PublishReleaseBinTool publish release bintool.
	PublishReleaseBinTool(nCtx contextx.IContext, uploadID string, isShared bool) error

	// PublishReleasePluginV2 publish release plugin v2.
	PublishReleasePluginV2(nCtx contextx.IContext, uploadID string, isShared bool) error

	// PublishReleaseExternalPluginV2 publish release external plugin v2.
	PublishReleaseExternalPluginV2(nCtx contextx.IContext, uploadID string, isShared bool) error

	// PublishReleasePluginV3 publish release plugin v3.
	PublishReleasePluginV3(nCtx contextx.IContext, uploadID string, isShared bool) error

	// PublishReleasePluginBinTool publish release plugin bin tool.
	PublishReleasePluginBinTool(nCtx contextx.IContext, uploadID string, isShared bool) error
}

// PublishReleaseAgent publish release agent.
func (h *handler) PublishReleaseAgent(nCtx contextx.IContext, uploadID string, isShared bool) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.PublishReleaseAgentReq{
		UploadId: uploadID,
		IsShared: isShared,
	}
	if _, err := h.cli.publishReleaseAgent(nCtx, tenantID, params); err != nil {
		return err
	}

	return nil
}

// PublishReleaseProxy publish release proxy.
func (h *handler) PublishReleaseProxy(nCtx contextx.IContext, uploadID, uploadCategory string, isShared bool) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.PublishReleaseProxyReq{
		UploadId:            uploadID,
		UploadOriginPkgType: uploadCategory,
		IsShared:            isShared,
	}
	if _, err := h.cli.publishReleaseProxy(nCtx, tenantID, params); err != nil {
		return err
	}

	return nil
}

// PublishReleaseCert publish release cert.
func (h *handler) PublishReleaseCert(nCtx contextx.IContext, uploadID string, isShared bool) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.PublishReleaseCertReq{
		UploadId: uploadID,
		IsShared: isShared,
	}
	if _, err := h.cli.publishReleaseCert(nCtx, tenantID, params); err != nil {
		return err
	}

	return nil
}

// PublishReleaseBinTool publish release bintool.
func (h *handler) PublishReleaseBinTool(nCtx contextx.IContext, uploadID string, isShared bool) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.PublishReleaseBinToolReq{
		UploadId: uploadID,
		IsShared: isShared,
	}
	if _, err := h.cli.publishReleaseBinTool(nCtx, tenantID, params); err != nil {
		return err
	}

	return nil
}

// PublishReleasePluginV2 publish release plugin v2.
func (h *handler) PublishReleasePluginV2(nCtx contextx.IContext, uploadID string, isShared bool) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.PublishReleasePluginV2Req{
		UploadId: uploadID,
		IsShared: isShared,
	}
	if _, err := h.cli.publishReleasePluginV2(nCtx, tenantID, params); err != nil {
		return err
	}

	return nil
}

// PublishReleaseExternalPluginV2 publish release external plugin v2.
func (h *handler) PublishReleaseExternalPluginV2(nCtx contextx.IContext, uploadID string, isShared bool) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.PublishReleaseExternalPluginV2Req{
		UploadId: uploadID,
		IsShared: isShared,
	}
	if _, err := h.cli.publishReleaseExternalPluginV2(nCtx, tenantID, params); err != nil {
		return err
	}

	return nil
}

// PublishReleasePluginV3 publish release plugin v3.
func (h *handler) PublishReleasePluginV3(nCtx contextx.IContext, uploadID string, isShared bool) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.PublishReleasePluginV3Req{
		UploadId: uploadID,
		IsShared: isShared,
	}
	if _, err := h.cli.publishReleasePluginV3(nCtx, tenantID, params); err != nil {
		return err
	}

	return nil
}

// PublishReleasePluginBinTool publish release plugin bin tool.
func (h *handler) PublishReleasePluginBinTool(nCtx contextx.IContext, uploadID string, isShared bool) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.PublishReleasePluginBinToolReq{
		UploadId: uploadID,
		IsShared: isShared,
	}
	if _, err := h.cli.publishReleasePluginBinTool(nCtx, tenantID, params); err != nil {
		return err
	}

	return nil
}
