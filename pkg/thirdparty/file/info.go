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
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IPkgInfoHandler defines the interface of pkg info query.
type IPkgInfoHandler interface {
	// InfoReleaseAgent queries file info for the release agent package.
	InfoReleaseAgent(nCtx contextx.IContext, gen types.Generation, plat platfmt.Platform, version string) (*fileiface.FileInfo, error)

	// InfoReleaseProxy queries file info for the release proxy package.
	InfoReleaseProxy(nCtx contextx.IContext, gen types.Generation, plat platfmt.Platform, version string) (*fileiface.FileInfo, error)

	// InfoReleasePlugin queries file info for the release plugin package.
	InfoReleasePlugin(nCtx contextx.IContext, pluginPkgName string, plat platfmt.Platform, version string) (*fileiface.FileInfo, error)

	// InfoReleaseCert queries file info for the release cert package.
	InfoReleaseCert(nCtx contextx.IContext, gen types.Generation) (*fileiface.FileInfo, error)

	// InfoReleaseBinTool queries file info for the release bintool package.
	InfoReleaseBinTool(nCtx contextx.IContext, gen types.Generation) (*fileiface.FileInfo, error)

	// InfoReleasePluginBinTool queries file info for the release plugin bintool package.
	InfoReleasePluginBinTool(nCtx contextx.IContext, gen types.Generation, name string) (*fileiface.FileInfo, error)

	// InfoInstaller queries file info for the installer.
	InfoInstaller(nCtx contextx.IContext, osType criteria.OSType, cpuArch criteria.CPUArch) (*fileiface.FileInfo, error)
}

// InfoReleaseAgent queries file info for the release agent package.
func (h *handler) InfoReleaseAgent(nCtx contextx.IContext, gen types.Generation, plat platfmt.Platform, version string) (*fileiface.FileInfo, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.DownloadAgentReq{
		OsType:     string(plat.OS),
		CpuArch:    string(plat.Arch),
		Version:    version,
		Generation: int64(gen),
	}

	data, err := h.cli.infoReleaseAgent(nCtx, tenantID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get release agent info: %w", err)
	}

	return &fileiface.FileInfo{Name: data.GetName(), Size: data.GetSize(), MD5: data.GetMd5()}, nil
}

// InfoReleaseProxy queries file info for the release proxy package.
func (h *handler) InfoReleaseProxy(nCtx contextx.IContext, gen types.Generation, plat platfmt.Platform, version string) (*fileiface.FileInfo, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.DownloadProxyReq{
		OsType:     string(plat.OS),
		CpuArch:    string(plat.Arch),
		Version:    version,
		Generation: int64(gen),
	}

	data, err := h.cli.infoReleaseProxy(nCtx, tenantID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get release proxy info: %w", err)
	}

	return &fileiface.FileInfo{Name: data.GetName(), Size: data.GetSize(), MD5: data.GetMd5()}, nil
}

// InfoReleasePlugin queries file info for the release plugin package.
func (h *handler) InfoReleasePlugin(nCtx contextx.IContext, pluginPkgName string, plat platfmt.Platform, version string) (
	*fileiface.FileInfo, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.DownloadPluginReq{
		OsType:        string(plat.OS),
		CpuArch:       string(plat.Arch),
		Version:       version,
		PluginPkgName: pluginPkgName,
	}

	data, err := h.cli.infoReleasePlugin(nCtx, tenantID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get release plugin info: %w", err)
	}

	return &fileiface.FileInfo{Name: data.GetName(), Size: data.GetSize(), MD5: data.GetMd5()}, nil
}

// InfoReleaseCert queries file info for the release cert package.
func (h *handler) InfoReleaseCert(nCtx contextx.IContext, gen types.Generation) (*fileiface.FileInfo, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.DownloadCertReq{Generation: int64(gen)}

	data, err := h.cli.infoReleaseCert(nCtx, tenantID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get release cert info: %w", err)
	}

	return &fileiface.FileInfo{Name: data.GetName(), Size: data.GetSize(), MD5: data.GetMd5()}, nil
}

// InfoReleaseBinTool queries file info for the release bintool package.
func (h *handler) InfoReleaseBinTool(nCtx contextx.IContext, gen types.Generation) (*fileiface.FileInfo, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.DownloadBinToolReq{Generation: int64(gen)}

	data, err := h.cli.infoReleaseBinTool(nCtx, tenantID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get release bintool info: %w", err)
	}

	return &fileiface.FileInfo{Name: data.GetName(), Size: data.GetSize(), MD5: data.GetMd5()}, nil
}

// InfoReleasePluginBinTool queries file info for the release plugin bintool package.
func (h *handler) InfoReleasePluginBinTool(nCtx contextx.IContext, gen types.Generation, name string) (*fileiface.FileInfo, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.DownloadPluginBinToolReq{Generation: int64(gen), Name: name}

	data, err := h.cli.infoReleasePluginBinTool(nCtx, tenantID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get release plugin bintool info: %w", err)
	}

	return &fileiface.FileInfo{Name: data.GetName(), Size: data.GetSize(), MD5: data.GetMd5()}, nil
}

// InfoInstaller queries file info for the installer.
func (h *handler) InfoInstaller(nCtx contextx.IContext, osType criteria.OSType, cpuArch criteria.CPUArch) (*fileiface.FileInfo, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.DownloadInstallerReq{
		OsType:  string(osType),
		CpuArch: string(cpuArch),
	}

	data, err := h.cli.infoInstaller(nCtx, tenantID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get installer info: %w", err)
	}

	return &fileiface.FileInfo{Name: data.GetName(), Size: data.GetSize(), MD5: data.GetMd5()}, nil
}
