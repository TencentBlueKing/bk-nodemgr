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
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IPkgDownloadHandler define the interface of pkg download.
type IPkgDownloadHandler interface {
	// DownloadReleaseAgent download release agent.
	DownloadReleaseAgent(nCtx contextx.IContext, gen types.Generation, plat platfmt.Platform, version string) (*restserver.StreamResponse, error)

	// DownloadReleaseProxy download release proxy.
	DownloadReleaseProxy(nCtx contextx.IContext, gen types.Generation, plat platfmt.Platform, version string) (*restserver.StreamResponse, error)

	// DownloadReleasePlugin download release plugin.
	DownloadReleasePlugin(nCtx contextx.IContext, pluginPkgName string, plat platfmt.Platform, version string) (*restserver.StreamResponse, error)

	// DownloadReleaseCert download release cert.
	DownloadReleaseCert(nCtx contextx.IContext, gen types.Generation) (*restserver.StreamResponse, error)

	// DownloadReleaseBinTool download release bintool.
	DownloadReleaseBinTool(nCtx contextx.IContext, gen types.Generation) (*restserver.StreamResponse, error)

	// DownloadReleasePluginBinTool download release plugin bintool.
	DownloadReleasePluginBinTool(nCtx contextx.IContext, gen types.Generation, name string) (*restserver.StreamResponse, error)

	// DownloadInstaller download installer.
	DownloadInstaller(nCtx contextx.IContext, osType criteria.OSType, cpuArch criteria.CPUArch) (*restserver.StreamResponse, error)

	// DownloadRemoteFile downloads a verified remote file.
	DownloadRemoteFile(nCtx contextx.IContext, filename, downloadURL, expectedMD5 string) (*restserver.StreamResponse, error)
}

// DownloadReleaseAgent download release agent.
func (h *handler) DownloadReleaseAgent(nCtx contextx.IContext, gen types.Generation, plat platfmt.Platform, version string) (
	*restserver.StreamResponse, error) {

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
	resp, err := h.cli.downloadReleaseAgent(nCtx, tenantID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to download release agent: %w", err)
	}

	return resp, nil
}

// DownloadReleaseProxy download release proxy.
func (h *handler) DownloadReleaseProxy(nCtx contextx.IContext, gen types.Generation, plat platfmt.Platform, version string) (
	*restserver.StreamResponse, error) {

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
	resp, err := h.cli.downloadReleaseProxy(nCtx, tenantID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to download release proxy: %w", err)
	}

	return resp, nil
}

// DownloadReleasePlugin download release plugin.
func (h *handler) DownloadReleasePlugin(nCtx contextx.IContext, pluginPkgName string, plat platfmt.Platform, version string) (
	*restserver.StreamResponse, error) {

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
	resp, err := h.cli.downloadReleasePlugin(nCtx, tenantID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to download release plugin: %w", err)
	}

	return resp, nil
}

// DownloadReleaseCert download release cert.
func (h *handler) DownloadReleaseCert(nCtx contextx.IContext, gen types.Generation) (*restserver.StreamResponse, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.DownloadCertReq{
		Generation: int64(gen),
	}
	resp, err := h.cli.downloadReleaseCert(nCtx, tenantID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to download release cert: %w", err)
	}

	return resp, nil
}

// DownloadReleaseBinTool download release bintool.
func (h *handler) DownloadReleaseBinTool(nCtx contextx.IContext, gen types.Generation) (*restserver.StreamResponse, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.DownloadBinToolReq{
		Generation: int64(gen),
	}
	resp, err := h.cli.downloadReleaseBinTool(nCtx, tenantID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to download release bintool: %w", err)
	}

	return resp, nil
}

// DownloadReleasePluginBinTool download release plugin bintool.
func (h *handler) DownloadReleasePluginBinTool(nCtx contextx.IContext, gen types.Generation, name string) (*restserver.StreamResponse, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.DownloadPluginBinToolReq{
		Generation: int64(gen),
		Name:       name,
	}
	resp, err := h.cli.downloadReleasePluginBinTool(nCtx, tenantID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to download release plugin bintool: %w", err)
	}

	return resp, nil
}

// DownloadInstaller download installer.
func (h *handler) DownloadInstaller(nCtx contextx.IContext, osType criteria.OSType, cpuArch criteria.CPUArch) (*restserver.StreamResponse, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.DownloadInstallerReq{
		OsType:  string(osType),
		CpuArch: string(cpuArch),
	}
	resp, err := h.cli.downloadInstaller(nCtx, tenantID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to download installer: %w", err)
	}

	return resp, nil
}

// DownloadRemoteFile downloads a verified remote file.
func (h *handler) DownloadRemoteFile(nCtx contextx.IContext, filename, downloadURL, expectedMD5 string) (*restserver.StreamResponse, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	params := &protoFile.DownloadRemoteFileReq{
		Filename:    filename,
		DownloadUrl: downloadURL,
		Md5:         expectedMD5,
	}

	return h.cli.downloadRemoteFile(nCtx, nCtx.TenantID(), params)
}
