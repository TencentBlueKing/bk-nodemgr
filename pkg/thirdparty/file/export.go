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
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
)

// IPkgExportHandler defines package export operations.
type IPkgExportHandler interface {
	// ExportPrepareOriginPluginPackage prepares an export with its package name and version.
	ExportPrepareOriginPluginPackage(nCtx contextx.IContext, name, version string) (string, error)

	// ExportGetOriginPluginPackageDownloadAddress gets the download address of the origin plugin package.
	ExportGetOriginPluginPackageDownloadAddress(nCtx contextx.IContext, exportID string) (string, int64, error)
}

// ExportPrepareOriginPluginPackage prepares an origin plugin package export with its package name and version.
func (h *handler) ExportPrepareOriginPluginPackage(nCtx contextx.IContext, name, version string) (string, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return "", err
	}

	req := &protoFile.ExportPrepareOriginPluginPackageReq{
		PluginPkgName:    name,
		PluginPkgVersion: version,
	}
	data, err := h.cli.exportPrepareOriginPluginPackage(nCtx, nCtx.TenantID(), req)
	if err != nil {
		return "", fmt.Errorf("failed to prepare origin plugin package export: %w", err)
	}

	return data.GetExportId(), nil
}

// ExportGetOriginPluginPackageDownloadAddress gets the download address of the origin plugin package.
func (h *handler) ExportGetOriginPluginPackageDownloadAddress(nCtx contextx.IContext, exportID string) (string, int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return "", 0, err
	}

	tenantID := nCtx.TenantID()
	req := &protoFile.ExportGetOriginPluginPackageDownloadAddressReq{ExportId: exportID}
	data, err := h.cli.exportGetOriginPluginPackageDownloadAddress(nCtx, tenantID, req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get origin plugin package export download address: %w", err)
	}

	return data.GetDownloadUrl(), data.GetDownloadUrlExpiredAt(), nil
}
