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
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/cache"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"golang.org/x/sync/singleflight"
)

// IHandler is interface for file handler.
type IHandler interface {
	IPkgManager
	ITransfer
	ISharedReleaseSyncHandler
}

// ITransfer is interface for file transfer handler.
type ITransfer interface {
	// LaunchTransferNode launch transfer release.
	LaunchTransferNode(nCtx contextx.IContext,
		gen types.Generation,
		rt types.ReleaseType,
		plat platfmt.Platform,
		version string,
		dstDir string,
		dstHost *types.Host) (types.ISimpleTransferHandler, error)

	// LaunchTransferPlugin launch transfer release.
	LaunchTransferPlugin(
		nCtx contextx.IContext,
		name string,
		gen types.Generation,
		plat platfmt.Platform,
		version string,
		dstDir string,
		dstHost *types.Host,
	) (types.ISimpleTransferHandler, error)

	// LaunchTransferInstaller launch transfer installer.
	LaunchTransferInstaller(nCtx contextx.IContext,
		gen types.Generation,
		plat platfmt.Platform,
		dstDir string,
		dstHost *types.Host) (types.ISimpleTransferHandler, error)

	// QueryTransfer query transfer package.
	// return upload result, download result and error.
	QueryTransfer(
		nCtx contextx.IContext, taskID string) (*types.SimpleTransferResult, *types.SimpleTransferResult, error)
}

// IPkgManager is interface for file pkg manager.
type IPkgManager interface {
	IPkgUploadHandler
	IPkgPublishHandler
	IPkgDownloadHandler
	IPkgInfoHandler
	IPkgExportHandler
	IPkgReleaseHandler
	IPackageEventHandler
}

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

// IPkgUploadHandler defines the interface of pkg upload.
type IPkgUploadHandler interface {
	// UploadOriginAgent upload origin agent.
	UploadOriginAgent(nCtx contextx.IContext, fileName string, file io.Reader, gen types.Generation, overwrite bool) (*types.OriginPkgDetail, error)

	// UploadOriginServer upload origin server.
	UploadOriginServer(nCtx contextx.IContext, fileName string, file io.Reader, gen types.Generation, overwrite bool) (*types.OriginPkgDetail, error)

	// UploadOriginProxy upload origin proxy.
	UploadOriginProxy(nCtx contextx.IContext, fileName string, file io.Reader, gen types.Generation, overwrite bool) (*types.OriginPkgDetail, error)

	// UploadOriginCert upload origin cert.
	UploadOriginCert(nCtx contextx.IContext, fileName string, file io.Reader, overwrite bool) (*types.OriginCertPkgDetail, error)

	// UploadOriginBinTool upload origin bintool.
	UploadOriginBinTool(nCtx contextx.IContext, fileName string, file io.Reader, gen types.Generation, overwrite bool) (
		*types.OriginBinToolPkgDetail, error)

	// UploadOriginPluginV2 upload origin plugin v2.
	UploadOriginPluginV2(nCtx contextx.IContext, fileName string, file io.Reader, overwrite bool) (*types.OriginPluginV2PkgDetail, error)

	// UploadOriginExternalPluginV2 upload origin external plugin v2.
	UploadOriginExternalPluginV2(nCtx contextx.IContext, fileName string, file io.Reader, overwrite bool) (
		*types.OriginExternalPluginV2PkgDetail, error)

	// UploadOriginPluginV3 upload origin plugin v3.
	UploadOriginPluginV3(nCtx contextx.IContext, fileName string, file io.Reader, overwrite bool) (*types.OriginPluginV3PkgDetail, error)

	// UploadOriginPluginBinTool upload origin plugin bin tool.
	UploadOriginPluginBinTool(nCtx contextx.IContext, fileName string, file io.Reader, overwrite bool) (
		*types.OriginPluginBinToolPkgDetail, error)
}

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

// IPkgExportHandler defines package export operations.
type IPkgExportHandler interface {
	// ExportPrepareOriginPluginPackage prepares an export with its package name and version.
	ExportPrepareOriginPluginPackage(nCtx contextx.IContext, name, version string) (string, error)

	// ExportGetOriginPluginPackageDownloadAddress gets the download address of the origin plugin package.
	ExportGetOriginPluginPackageDownloadAddress(nCtx contextx.IContext, exportID string) (string, int64, error)
}

// ISharedReleaseSyncHandler synchronizes system shared releases for a tenant.
type ISharedReleaseSyncHandler interface {
	// SyncSharedReleases synchronizes system shared releases to the current tenant.
	SyncSharedReleases(nCtx contextx.IContext) error
}

const (
	transferQueryTickTime             = 1 * time.Second
	transferQueryContinuesFailedTimes = 5
)

const releaseCacheExpiration = time.Minute

type handler struct {
	cli *cli

	releaseCache     cache.ICache
	releaseCacheKeys cache.IKeyGenerator
	releaseGroup     singleflight.Group
}

// New initialize a new nodeman backend handler.
func New(c *restclient.Capability, conf *Config) (IHandler, error) {
	cli, err := newClient(c, conf)
	if err != nil {
		return nil, err
	}

	return &handler{
		cli:              cli,
		releaseCache:     cache.NewMemoryCache(releaseCacheExpiration),
		releaseCacheKeys: cache.NewDefaultKeyGenerator("file:release"),
	}, nil
}

// UploadOriginAgent upload origin agent.
func (h *handler) UploadOriginAgent(nCtx contextx.IContext, fileName string, file io.Reader, gen types.Generation, overwrite bool) (
	*types.OriginPkgDetail, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	if gen != types.Generation2 {
		return nil, fmt.Errorf("param generateion(%d) invalid, only generation2 is supported for origin agent upload", gen)
	}

	tenantID := nCtx.TenantID()

	params := &protoFile.UploadOriginAgentReq{
		Generation: int64(gen),
		Overwrite:  overwrite,
	}
	resp, err := h.cli.uploadOriginAgent(nCtx, tenantID, params, fileName, file)
	if err != nil {
		return nil, err
	}

	return &types.OriginPkgDetail{
		FileInfo: fileiface.FileInfo{
			Name: resp.GetName(),
			Size: resp.GetSize(),
			MD5:  resp.GetMd5(),
		},
		UploadID:    resp.GetUploadId(),
		Existed:     resp.GetExisted(),
		Version:     resp.GetVersion(),
		Platforms:   resp.ConvertPlatformsToTypes(),
		ChangeLogEN: resp.GetChangelogEn(),
		ChangeLogZH: resp.GetChangelogZh(),
	}, nil
}

// UploadOriginServer upload origin server.
func (h *handler) UploadOriginServer(nCtx contextx.IContext, fileName string, file io.Reader, gen types.Generation, overwrite bool) (
	*types.OriginPkgDetail, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	if gen != types.Generation2 {
		return nil, fmt.Errorf("param generateion(%d) invalid, only generation2 is supported for origin server upload", gen)
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.UploadOriginServerReq{
		Generation: int64(gen),
		Overwrite:  overwrite,
	}
	resp, err := h.cli.uploadOriginServer(nCtx, tenantID, params, fileName, file)
	if err != nil {
		return nil, err
	}

	return &types.OriginPkgDetail{
		FileInfo: fileiface.FileInfo{
			Name: resp.GetName(),
			Size: resp.GetSize(),
			MD5:  resp.GetMd5(),
		},
		UploadID:  resp.GetUploadId(),
		Existed:   resp.GetExisted(),
		Version:   resp.GetVersion(),
		Platforms: resp.ConvertPlatformsToTypes(),
	}, nil
}

// UploadOriginProxy upload origin proxy.
func (h *handler) UploadOriginProxy(nCtx contextx.IContext, fileName string, file io.Reader, gen types.Generation, overwrite bool) (
	*types.OriginPkgDetail, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	if gen != types.Generation2 {
		return nil, fmt.Errorf("param generateion(%d) invalid, only generation2 is supported for origin proxy upload", gen)
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.UploadOriginProxyReq{
		Generation: int64(gen),
		Overwrite:  overwrite,
	}
	resp, err := h.cli.uploadOriginProxy(nCtx, tenantID, params, fileName, file)
	if err != nil {
		return nil, err
	}

	return &types.OriginPkgDetail{
		FileInfo: fileiface.FileInfo{
			Name: resp.GetName(),
			Size: resp.GetSize(),
			MD5:  resp.GetMd5(),
		},
		UploadID:    resp.GetUploadId(),
		Existed:     resp.GetExisted(),
		Version:     resp.GetVersion(),
		Platforms:   resp.ConvertPlatformsToTypes(),
		ChangeLogEN: resp.GetChangelogEn(),
		ChangeLogZH: resp.GetChangelogZh(),
	}, nil
}

// UploadOriginCert upload origin cert.
func (h *handler) UploadOriginCert(nCtx contextx.IContext, fileName string, file io.Reader, overwrite bool) (*types.OriginCertPkgDetail, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.UploadOriginCertReq{Overwrite: overwrite}
	resp, err := h.cli.uploadOriginCert(nCtx, tenantID, params, fileName, file)
	if err != nil {
		return nil, err
	}

	return &types.OriginCertPkgDetail{
		FileInfo: fileiface.FileInfo{
			Name: resp.GetName(),
			Size: resp.GetSize(),
			MD5:  resp.GetMd5(),
		},
		UploadID:  resp.GetUploadId(),
		Existed:   resp.GetExisted(),
		CertFiles: resp.GetCertFiles(),
	}, nil
}

// UploadOriginBinTool upload origin bintool.
func (h *handler) UploadOriginBinTool(nCtx contextx.IContext, fileName string, file io.Reader, gen types.Generation, overwrite bool) (
	*types.OriginBinToolPkgDetail, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.UploadOriginBinToolReq{
		Generation: int64(gen),
		Overwrite:  overwrite,
	}
	resp, err := h.cli.uploadOriginBinTool(nCtx, tenantID, params, fileName, file)
	if err != nil {
		return nil, err
	}

	return &types.OriginBinToolPkgDetail{
		FileInfo: fileiface.FileInfo{
			Name: resp.GetName(),
			Size: resp.GetSize(),
			MD5:  resp.GetMd5(),
		},
		UploadID:       resp.GetUploadId(),
		Existed:        resp.GetExisted(),
		AgentPlatforms: resp.ConvertAgentPlatformsToTypes(),
		ProxyPlatforms: resp.ConvertProxyPlatformsToTypes(),
	}, nil
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

// DownloadReleaseAgent download release agent.
func (h *handler) DownloadReleaseAgent(nCtx contextx.IContext,
	gen types.Generation, plat platfmt.Platform, version string) (*restserver.StreamResponse, error) {

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
func (h *handler) DownloadReleaseProxy(nCtx contextx.IContext,
	gen types.Generation, plat platfmt.Platform, version string) (*restserver.StreamResponse, error) {

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
func (h *handler) DownloadReleasePlugin(nCtx contextx.IContext,
	pluginPkgName string, plat platfmt.Platform, version string) (*restserver.StreamResponse, error) {

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
func (h *handler) DownloadInstaller(nCtx contextx.IContext, osType criteria.OSType, cpuArch criteria.CPUArch) (
	*restserver.StreamResponse, error) {

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

// ListReleaseAgent queries agent release metadata.
func (h *handler) ListReleaseAgent(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) (
	[]*types.ReleaseAgent, int64, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}

	req := new(protoFile.ReleaseAgentListReq)
	if err := req.ConvertFromTypes(page, conditions...); err != nil {
		return nil, 0, err
	}
	cacheKey, err := h.generateReleaseCacheKey("agent-list", nCtx.TenantID(), req)
	if err != nil {
		return nil, 0, err
	}

	data, err := h.queryReleaseCache(nCtx, cacheKey, func(nCtx contextx.IContext) ([]byte, error) {
		response, err := h.cli.listReleaseAgent(nCtx, nCtx.TenantID(), req)
		if err != nil {
			return nil, err
		}

		if _, _, err := response.ConvertToTypes(); err != nil {
			return nil, err
		}

		return json.Marshal(response)
	})
	if err != nil {
		return nil, 0, err
	}

	var response *protoFile.ReleaseAgentListResp_Data
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal agent release list: %w", err)
	}

	total, items, err := response.ConvertToTypes()
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// CountReleaseAgent queries agent release metadata.
func (h *handler) CountReleaseAgent(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return 0, err
	}

	req := new(protoFile.ReleaseAgentCountReq)
	if err := req.ConvertFromTypes(conditions...); err != nil {
		return 0, err
	}
	cacheKey, err := h.generateReleaseCacheKey("agent-count", nCtx.TenantID(), req)
	if err != nil {
		return 0, err
	}

	data, err := h.queryReleaseCache(nCtx, cacheKey, func(nCtx contextx.IContext) ([]byte, error) {
		response, err := h.cli.countReleaseAgent(nCtx, nCtx.TenantID(), req)
		if err != nil {
			return nil, err
		}

		return json.Marshal(response)
	})
	if err != nil {
		return 0, err
	}

	var response *protoFile.ReleaseAgentCountResp_Data
	if err := json.Unmarshal(data, &response); err != nil {
		return 0, fmt.Errorf("failed to unmarshal agent release count: %w", err)
	}

	return response.ConvertToTypes(), nil
}

// GetReleaseAgent queries agent release metadata.
func (h *handler) GetReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) (*types.ReleaseAgent, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	req := new(protoFile.ReleaseAgentGetReq)
	req.ConvertFromTypes(key)
	cacheKey, err := h.generateReleaseCacheKey("agent", nCtx.TenantID(), req)
	if err != nil {
		return nil, err
	}

	data, err := h.queryReleaseCache(nCtx, cacheKey, func(nCtx contextx.IContext) ([]byte, error) {
		response, err := h.cli.getReleaseAgent(nCtx, nCtx.TenantID(), req)
		if err != nil {
			return nil, err
		}

		if _, err := response.ConvertToTypes(); err != nil {
			return nil, err
		}

		return json.Marshal(response)
	})
	if err != nil {
		return nil, err
	}

	var response *protoFile.ReleaseAgentGetResp_Data
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal agent release: %w", err)
	}

	return response.ConvertToTypes()
}

// DistinctReleaseAgent queries agent release metadata.
func (h *handler) DistinctReleaseAgent(nCtx contextx.IContext, fields types.ReleaseDistinctField, conditions ...*types.ReleaseCondition) (
	*types.ReleaseDistinctResult, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	req := new(protoFile.ReleaseAgentDistinctReq)
	if err := req.ConvertFromTypes(fields, conditions...); err != nil {
		return nil, err
	}
	cacheKey, err := h.generateReleaseCacheKey("agent-distinct", nCtx.TenantID(), req)
	if err != nil {
		return nil, err
	}

	data, err := h.queryReleaseCache(nCtx, cacheKey, func(nCtx contextx.IContext) ([]byte, error) {
		response, err := h.cli.distinctReleaseAgent(nCtx, nCtx.TenantID(), req)
		if err != nil {
			return nil, err
		}

		return json.Marshal(response)
	})
	if err != nil {
		return nil, err
	}

	var response *protoFile.ReleaseAgentDistinctResp_Data
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal agent release distinct result: %w", err)
	}

	return response.ConvertToTypes(), nil
}

// ListReleaseProxy queries proxy release metadata.
func (h *handler) ListReleaseProxy(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) (
	[]*types.ReleaseProxy, int64, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}

	req := new(protoFile.ReleaseProxyListReq)
	if err := req.ConvertFromTypes(page, conditions...); err != nil {
		return nil, 0, err
	}
	cacheKey, err := h.generateReleaseCacheKey("proxy-list", nCtx.TenantID(), req)
	if err != nil {
		return nil, 0, err
	}

	data, err := h.queryReleaseCache(nCtx, cacheKey, func(nCtx contextx.IContext) ([]byte, error) {
		response, err := h.cli.listReleaseProxy(nCtx, nCtx.TenantID(), req)
		if err != nil {
			return nil, err
		}

		if _, _, err := response.ConvertToTypes(); err != nil {
			return nil, err
		}

		return json.Marshal(response)
	})
	if err != nil {
		return nil, 0, err
	}

	var response *protoFile.ReleaseProxyListResp_Data
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal proxy release list: %w", err)
	}

	total, items, err := response.ConvertToTypes()
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// CountReleaseProxy queries proxy release metadata.
func (h *handler) CountReleaseProxy(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return 0, err
	}

	req := new(protoFile.ReleaseProxyCountReq)
	if err := req.ConvertFromTypes(conditions...); err != nil {
		return 0, err
	}
	cacheKey, err := h.generateReleaseCacheKey("proxy-count", nCtx.TenantID(), req)
	if err != nil {
		return 0, err
	}

	data, err := h.queryReleaseCache(nCtx, cacheKey, func(nCtx contextx.IContext) ([]byte, error) {
		response, err := h.cli.countReleaseProxy(nCtx, nCtx.TenantID(), req)
		if err != nil {
			return nil, err
		}

		return json.Marshal(response)
	})
	if err != nil {
		return 0, err
	}

	var response *protoFile.ReleaseProxyCountResp_Data
	if err := json.Unmarshal(data, &response); err != nil {
		return 0, fmt.Errorf("failed to unmarshal proxy release count: %w", err)
	}

	return response.ConvertToTypes(), nil
}

// GetReleaseProxy queries proxy release metadata.
func (h *handler) GetReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) (*types.ReleaseProxy, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	req := new(protoFile.ReleaseProxyGetReq)
	req.ConvertFromTypes(key)
	cacheKey, err := h.generateReleaseCacheKey("proxy", nCtx.TenantID(), req)
	if err != nil {
		return nil, err
	}

	data, err := h.queryReleaseCache(nCtx, cacheKey, func(nCtx contextx.IContext) ([]byte, error) {
		response, err := h.cli.getReleaseProxy(nCtx, nCtx.TenantID(), req)
		if err != nil {
			return nil, err
		}

		if _, err := response.ConvertToTypes(); err != nil {
			return nil, err
		}

		return json.Marshal(response)
	})
	if err != nil {
		return nil, err
	}

	var response *protoFile.ReleaseProxyGetResp_Data
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal proxy release: %w", err)
	}

	return response.ConvertToTypes()
}

// DistinctReleaseProxy queries proxy release metadata.
func (h *handler) DistinctReleaseProxy(nCtx contextx.IContext, fields types.ReleaseDistinctField, conditions ...*types.ReleaseCondition) (
	*types.ReleaseDistinctResult, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	req := new(protoFile.ReleaseProxyDistinctReq)
	if err := req.ConvertFromTypes(fields, conditions...); err != nil {
		return nil, err
	}
	cacheKey, err := h.generateReleaseCacheKey("proxy-distinct", nCtx.TenantID(), req)
	if err != nil {
		return nil, err
	}

	data, err := h.queryReleaseCache(nCtx, cacheKey, func(nCtx contextx.IContext) ([]byte, error) {
		response, err := h.cli.distinctReleaseProxy(nCtx, nCtx.TenantID(), req)
		if err != nil {
			return nil, err
		}

		return json.Marshal(response)
	})
	if err != nil {
		return nil, err
	}

	var response *protoFile.ReleaseProxyDistinctResp_Data
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal proxy release distinct result: %w", err)
	}

	return response.ConvertToTypes(), nil
}

// GetReleasePlugin queries plugin release metadata.
func (h *handler) GetReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) (*types.ReleasePlugin, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	req := new(protoFile.ReleasePluginGetReq)
	req.ConvertFromTypes(key)
	cacheKey, err := h.generateReleaseCacheKey("plugin", nCtx.TenantID(), key)
	if err != nil {
		return nil, err
	}

	data, err := h.queryReleaseCache(nCtx, cacheKey, func(nCtx contextx.IContext) ([]byte, error) {
		response, err := h.cli.getReleasePlugin(nCtx, nCtx.TenantID(), req)
		if err != nil {
			return nil, err
		}

		if _, err := response.ConvertToTypes(); err != nil {
			return nil, err
		}

		return json.Marshal(response)
	})
	if err != nil {
		return nil, err
	}

	result, err := decodeReleasePlugin(data)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetReleasePluginDefaultVersion queries plugin release metadata.
func (h *handler) GetReleasePluginDefaultVersion(nCtx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform) (string, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return "", err
	}

	req := new(protoFile.ReleasePluginGetDefaultVersionReq)
	req.ConvertFromTypes(name, gen, plat)
	cacheKey, err := h.generateReleaseCacheKey("plugin-default-version", nCtx.TenantID(), struct {
		Name       string
		Generation types.Generation
		Platform   platfmt.Platform
	}{Name: name, Generation: gen, Platform: plat})
	if err != nil {
		return "", err
	}

	data, err := h.queryReleaseCache(nCtx, cacheKey, func(nCtx contextx.IContext) ([]byte, error) {
		response, err := h.cli.getReleasePluginDefaultVersion(nCtx, nCtx.TenantID(), req)
		if err != nil {
			return nil, err
		}

		version := response.ConvertToTypes()

		return json.Marshal(version)
	})
	if err != nil {
		return "", err
	}

	var version string
	if err := json.Unmarshal(data, &version); err != nil {
		return "", fmt.Errorf("failed to unmarshal plugin default version: %w", err)
	}

	return version, nil
}

// ListReleasePlugin queries plugin release metadata.
func (h *handler) ListReleasePlugin(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) (
	[]*types.ReleasePlugin, int64, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}

	req := new(protoFile.ReleasePluginListReq)
	if err := req.ConvertFromTypes(page, conditions...); err != nil {
		return nil, 0, err
	}

	cacheKey, err := h.generateReleaseCacheKey("plugin-list", nCtx.TenantID(), req)
	if err != nil {
		return nil, 0, err
	}

	data, err := h.queryReleaseCache(nCtx, cacheKey, func(nCtx contextx.IContext) ([]byte, error) {
		response, err := h.cli.listReleasePlugin(nCtx, nCtx.TenantID(), req)
		if err != nil {
			return nil, err
		}

		if _, _, err := response.ConvertToTypes(); err != nil {
			return nil, err
		}

		return json.Marshal(response)
	})
	if err != nil {
		return nil, 0, err
	}

	var response *protoFile.ReleasePluginListResp_Data
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal plugin release list: %w", err)
	}

	total, items, err := response.ConvertToTypes()
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// CountReleasePlugin queries plugin release metadata.
func (h *handler) CountReleasePlugin(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return 0, err
	}

	req := new(protoFile.ReleasePluginCountReq)
	if err := req.ConvertFromTypes(conditions...); err != nil {
		return 0, err
	}

	cacheKey, err := h.generateReleaseCacheKey("plugin-count", nCtx.TenantID(), req)
	if err != nil {
		return 0, err
	}

	data, err := h.queryReleaseCache(nCtx, cacheKey, func(nCtx contextx.IContext) ([]byte, error) {
		response, err := h.cli.countReleasePlugin(nCtx, nCtx.TenantID(), req)
		if err != nil {
			return nil, err
		}

		return json.Marshal(response)
	})
	if err != nil {
		return 0, err
	}

	var response *protoFile.ReleasePluginCountResp_Data
	if err := json.Unmarshal(data, &response); err != nil {
		return 0, fmt.Errorf("failed to unmarshal plugin release count: %w", err)
	}

	return response.ConvertToTypes(), nil
}

// DistinctReleasePlugin queries plugin release metadata.
func (h *handler) DistinctReleasePlugin(nCtx contextx.IContext, fields types.ReleaseDistinctField, conditions ...*types.ReleaseCondition) (
	*types.ReleaseDistinctResult, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	req := new(protoFile.ReleasePluginDistinctReq)
	if err := req.ConvertFromTypes(fields, conditions...); err != nil {
		return nil, err
	}

	cacheKey, err := h.generateReleaseCacheKey("plugin-distinct", nCtx.TenantID(), req)
	if err != nil {
		return nil, err
	}

	data, err := h.queryReleaseCache(nCtx, cacheKey, func(nCtx contextx.IContext) ([]byte, error) {
		response, err := h.cli.distinctReleasePlugin(nCtx, nCtx.TenantID(), req)
		if err != nil {
			return nil, err
		}

		return json.Marshal(response)
	})
	if err != nil {
		return nil, err
	}

	var response *protoFile.ReleasePluginDistinctResp_Data
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plugin release distinct result: %w", err)
	}

	return response.ConvertToTypes(), nil
}

// ExistReleasePlugin queries plugin release metadata.
func (h *handler) ExistReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) (bool, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return false, err
	}

	req := new(protoFile.ReleasePluginExistReq)
	req.ConvertFromTypes(key)
	cacheKey, err := h.generateReleaseCacheKey("plugin-exist", nCtx.TenantID(), req)
	if err != nil {
		return false, err
	}

	data, err := h.queryReleaseCache(nCtx, cacheKey, func(nCtx contextx.IContext) ([]byte, error) {
		response, err := h.cli.existReleasePlugin(nCtx, nCtx.TenantID(), req)
		if err != nil {
			return nil, err
		}

		return json.Marshal(response)
	})
	if err != nil {
		return false, err
	}

	var response *protoFile.ReleasePluginExistResp_Data
	if err := json.Unmarshal(data, &response); err != nil {
		return false, fmt.Errorf("failed to unmarshal plugin release existence: %w", err)
	}

	return response.ConvertToTypes(), nil
}

// DistinctNameReleasePlugin queries plugin release metadata.
func (h *handler) DistinctNameReleasePlugin(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) ([]string, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	req := new(protoFile.ReleasePluginDistinctNameReq)
	if err := req.ConvertFromTypes(conditions...); err != nil {
		return nil, err
	}

	cacheKey, err := h.generateReleaseCacheKey("plugin-distinct-name", nCtx.TenantID(), req)
	if err != nil {
		return nil, err
	}

	data, err := h.queryReleaseCache(nCtx, cacheKey, func(nCtx contextx.IContext) ([]byte, error) {
		response, err := h.cli.distinctNameReleasePlugin(nCtx, nCtx.TenantID(), req)
		if err != nil {
			return nil, err
		}

		return json.Marshal(response)
	})
	if err != nil {
		return nil, err
	}

	var response *protoFile.ReleasePluginDistinctNameResp_Data
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plugin release distinct names: %w", err)
	}

	return response.ConvertToTypes(), nil
}

// ListReleaseCert queries cert release metadata.
func (h *handler) ListReleaseCert(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) (
	[]*types.ReleaseCert, int64, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}

	req := new(protoFile.ReleaseCertListReq)
	if err := req.ConvertFromTypes(page, conditions...); err != nil {
		return nil, 0, err
	}

	cacheKey, err := h.generateReleaseCacheKey("cert-list", nCtx.TenantID(), req)
	if err != nil {
		return nil, 0, err
	}

	data, err := h.queryReleaseCache(nCtx, cacheKey, func(nCtx contextx.IContext) ([]byte, error) {
		response, err := h.cli.listReleaseCert(nCtx, nCtx.TenantID(), req)
		if err != nil {
			return nil, err
		}

		if _, _, err := response.ConvertToTypes(); err != nil {
			return nil, err
		}

		return json.Marshal(response)
	})
	if err != nil {
		return nil, 0, err
	}

	var response *protoFile.ReleaseCertListResp_Data
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal cert release list: %w", err)
	}

	total, items, err := response.ConvertToTypes()
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// ListReleaseBinTool queries bintool release metadata.
func (h *handler) ListReleaseBinTool(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) (
	[]*types.ReleaseBinTool, int64, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}

	req := new(protoFile.ReleaseBinToolListReq)
	if err := req.ConvertFromTypes(page, conditions...); err != nil {
		return nil, 0, err
	}

	cacheKey, err := h.generateReleaseCacheKey("bintool-list", nCtx.TenantID(), req)
	if err != nil {
		return nil, 0, err
	}

	data, err := h.queryReleaseCache(nCtx, cacheKey, func(nCtx contextx.IContext) ([]byte, error) {
		response, err := h.cli.listReleaseBinTool(nCtx, nCtx.TenantID(), req)
		if err != nil {
			return nil, err
		}

		if _, _, err := response.ConvertToTypes(); err != nil {
			return nil, err
		}

		return json.Marshal(response)
	})
	if err != nil {
		return nil, 0, err
	}

	var response *protoFile.ReleaseBinToolListResp_Data
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal bintool release list: %w", err)
	}

	total, items, err := response.ConvertToTypes()
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// ListReleasePluginBinTool queries plugin_bintool release metadata.
func (h *handler) ListReleasePluginBinTool(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) (
	[]*types.ReleasePluginBinTool, int64, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}

	req := new(protoFile.ReleasePluginBinToolListReq)
	if err := req.ConvertFromTypes(page, conditions...); err != nil {
		return nil, 0, err
	}

	cacheKey, err := h.generateReleaseCacheKey("plugin-bintool-list", nCtx.TenantID(), req)
	if err != nil {
		return nil, 0, err
	}

	data, err := h.queryReleaseCache(nCtx, cacheKey, func(nCtx contextx.IContext) ([]byte, error) {
		response, err := h.cli.listReleasePluginBinTool(nCtx, nCtx.TenantID(), req)
		if err != nil {
			return nil, err
		}

		if _, _, err := response.ConvertToTypes(); err != nil {
			return nil, err
		}

		return json.Marshal(response)
	})
	if err != nil {
		return nil, 0, err
	}

	var response *protoFile.ReleasePluginBinToolListResp_Data
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal plugin bintool release list: %w", err)
	}

	total, items, err := response.ConvertToTypes()
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// DistinctNameReleasePluginBinTool queries plugin_bintool release metadata.
func (h *handler) DistinctNameReleasePluginBinTool(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) ([]string, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	req := new(protoFile.ReleasePluginBinToolDistinctNameReq)
	if err := req.ConvertFromTypes(conditions...); err != nil {
		return nil, err
	}

	cacheKey, err := h.generateReleaseCacheKey("plugin-bintool-distinct-name", nCtx.TenantID(), req)
	if err != nil {
		return nil, err
	}

	data, err := h.queryReleaseCache(nCtx, cacheKey, func(nCtx contextx.IContext) ([]byte, error) {
		response, err := h.cli.distinctNameReleasePluginBinTool(nCtx, nCtx.TenantID(), req)
		if err != nil {
			return nil, err
		}

		return json.Marshal(response)
	})
	if err != nil {
		return nil, err
	}

	var response *protoFile.ReleasePluginBinToolDistinctNameResp_Data
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plugin bintool release distinct names: %w", err)
	}

	return response.ConvertToTypes(), nil
}

func decodeReleasePlugin(data []byte) (*types.ReleasePlugin, error) {
	response := new(protoFile.ReleasePluginGetResp_Data)
	if err := json.Unmarshal(data, response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plugin release: %w", err)
	}

	return response.ConvertToTypes()
}

func (h *handler) generateReleaseCacheKey(prefix, tenantID string, query any) (string, error) {
	data, err := json.Marshal(query)
	if err != nil {
		return "", fmt.Errorf("failed to marshal release cache key: %w", err)
	}

	return h.releaseCacheKeys.Generate(prefix, tenantID, base64.RawURLEncoding.EncodeToString(data)), nil
}

func (h *handler) queryReleaseCache(
	nCtx contextx.IContext, key string, fetch func(contextx.IContext) ([]byte, error),
) ([]byte, error) {

	if err := nCtx.Err(); err != nil {
		return nil, err
	}

	if data, found := h.getReleaseCache(nCtx, key); found {
		return data, nil
	}

	resultCh := h.releaseGroup.DoChan(key, func() (any, error) {
		if data, found := h.getReleaseCache(nCtx, key); found {
			return data, nil
		}

		data, err := fetch(nCtx)
		if err != nil {
			return nil, err
		}
		if err := h.releaseCache.SetWithExpiration(nCtx, key, data, releaseCacheExpiration); err != nil {
			return nil, fmt.Errorf("failed to cache release data: %w", err)
		}

		return data, nil
	})

	select {
	case <-nCtx.Done():
		return nil, nCtx.Err()
	case result := <-resultCh:
		if result.Err != nil {
			return nil, result.Err
		}

		data, ok := result.Val.([]byte)
		if !ok {
			return nil, fmt.Errorf("invalid release cache result type")
		}

		return data, nil
	}
}

func (h *handler) getReleaseCache(nCtx contextx.IContext, key string) ([]byte, bool) {
	data, err := h.releaseCache.Get(nCtx, key)
	if err != nil {
		return nil, false
	}

	return data, true
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

// LaunchTransferNode launch transfer release.
func (h *handler) LaunchTransferNode(nCtx contextx.IContext,
	gen types.Generation,
	rt types.ReleaseType,
	plat platfmt.Platform,
	version string,
	dstDir string,
	dstHost *types.Host) (types.ISimpleTransferHandler, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()

	resp, err := h.cli.launchTransferNode(nCtx, tenantID, &protoFile.TransferLaunchNodeReq{
		Generation:   int64(gen),
		ReleaseType:  string(rt),
		Platform:     protoFile.ConvertPlatformFromTypes(plat),
		Version:      version,
		TargetDir:    dstDir,
		TargetHostId: dstHost.HostID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to launch transfer release: %w", err)
	}

	data := resp.GetData()

	return &simpleTransferHandler{
		taskID: data.GetTaskId(),
		fileInfo: fileiface.FileInfo{
			Name: data.GetReleaseFileName(),
			Size: data.GetReleaseFileSize(),
			MD5:  data.GetReleaseFileMd5(),
		},
		handler: h,
	}, nil
}

// LaunchTransferPlugin launch transfer release.
func (h *handler) LaunchTransferPlugin(
	nCtx contextx.IContext,
	name string,
	gen types.Generation,
	plat platfmt.Platform,
	version string,
	dstDir string,
	dstHost *types.Host,
) (types.ISimpleTransferHandler, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	resp, err := h.cli.launchTransferPlugin(nCtx, nCtx.TenantID(), &protoFile.TransferLaunchPluginReq{
		Name:         name,
		Generation:   int64(gen),
		Platform:     protoFile.ConvertPlatformFromTypes(plat),
		Version:      version,
		TargetDir:    dstDir,
		TargetHostId: dstHost.HostID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to launch transfer release: %w", err)
	}

	data := resp.GetData()

	return &simpleTransferHandler{
		taskID: data.GetTaskId(),
		fileInfo: fileiface.FileInfo{
			Name: data.GetReleaseFileName(),
			Size: data.GetReleaseFileSize(),
			MD5:  data.GetReleaseFileMd5(),
		},
		handler: h,
	}, nil
}

// LaunchTransferInstaller launch transfer installer.
func (h *handler) LaunchTransferInstaller(nCtx contextx.IContext,
	gen types.Generation,
	plat platfmt.Platform,
	dstDir string,
	dstHost *types.Host) (types.ISimpleTransferHandler, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()

	resp, err := h.cli.launchTransferInstaller(nCtx, tenantID, &protoFile.TransferLaunchInstallerReq{
		Generation:   int64(gen),
		Platform:     protoFile.ConvertPlatformFromTypes(plat),
		TargetDir:    dstDir,
		TargetHostId: dstHost.HostID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to launch transfer installer: %w", err)
	}

	data := resp.GetData()

	return &simpleTransferHandler{
		taskID: data.GetTaskId(),
		fileInfo: fileiface.FileInfo{
			Name: data.GetInstallerFileName(),
			Size: data.GetInstallerFileSize(),
			MD5:  data.GetInstallerFileMd5(),
		},
		handler: h,
	}, nil
}

// QueryTransfer query transfer package.
// return upload result, download result and error.
func (h *handler) QueryTransfer(
	nCtx contextx.IContext, taskID string) (*types.SimpleTransferResult, *types.SimpleTransferResult, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, nil, err
	}

	tenantID := nCtx.TenantID()

	resp, err := h.cli.queryTransfer(nCtx, tenantID, &protoFile.TransferQueryReq{
		TaskId: taskID,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query transfer: %w", err)
	}

	data := resp.GetData()

	return protoFile.ConvertSimpleTransferToTypes(data.GetUpload()),
		protoFile.ConvertSimpleTransferToTypes(data.GetDownload()),
		nil
}

type simpleTransferHandler struct {
	taskID   string
	fileInfo fileiface.FileInfo
	handler  *handler
}

// GetTaskID get task id.
func (handler *simpleTransferHandler) GetTaskID() string {
	return handler.taskID
}

// GetFileInfo get file info.
func (handler *simpleTransferHandler) GetFileInfo() fileiface.FileInfo {
	return handler.fileInfo
}

// WaitUntilDone wait until done.
func (handler *simpleTransferHandler) WaitUntilDone(nCtx contextx.IContext) (*types.SimpleTransferResult, error) {
	ticker := time.NewTicker(transferQueryTickTime)
	failedCnt := 0
	for {
		select {
		case <-nCtx.Done():
			return nil, errors.New("context done")
		case <-ticker.C:
			src, dst, err := handler.handler.QueryTransfer(nCtx, handler.taskID)
			if err != nil {
				failedCnt++
				if failedCnt > transferQueryContinuesFailedTimes {
					return nil, fmt.Errorf("failed to query transfer. task-id(%s): %w", handler.taskID, err)
				}

				continue
			}
			failedCnt = 0

			if dst.Terminated {
				return dst, nil
			}

			if src.Terminated && src.ErrorCode != 0 {
				// if upload failed, set the upload error info into simple result.
				dst.ErrorCode = src.ErrorCode
				dst.ErrorMessage = src.ErrorMessage

				return dst, nil
			}
		}
	}
}

// UploadOriginPluginV2 upload origin plugin v2.
func (h *handler) UploadOriginPluginV2(nCtx contextx.IContext, fileName string, file io.Reader, overwrite bool) (
	*types.OriginPluginV2PkgDetail, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.UploadOriginPluginV2Req{Overwrite: overwrite}
	data, err := h.cli.uploadOriginPluginV2(nCtx, tenantID, params, fileName, file)
	if err != nil {
		return nil, err
	}

	return &types.OriginPluginV2PkgDetail{
		FileInfo: fileiface.FileInfo{
			Name: data.GetName(),
			Size: data.GetSize(),
			MD5:  data.GetMd5(),
		},
		UploadID:      data.GetUploadId(),
		Existed:       data.GetExisted(),
		PluginPkgName: data.GetName(),
		Version:       data.GetVersion(),
		Description:   data.GetDescription(),
		DescriptionEn: data.GetDescriptionEn(),
		Scenario:      data.GetScenario(),
		ScenarioEn:    data.GetScenarioEn(),
		ConfigFile:    data.GetConfigFile(),
		ConfigFormat:  data.GetConfigFormat(),
		LaunchNode:    data.GetLaunchNode(),
		Platforms:     data.ConvertPlatformsToTypes(),
	}, nil
}

// UploadOriginExternalPluginV2 upload origin external plugin v2.
func (h *handler) UploadOriginExternalPluginV2(nCtx contextx.IContext, fileName string, file io.Reader, overwrite bool) (
	*types.OriginExternalPluginV2PkgDetail, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.UploadOriginExternalPluginV2Req{Overwrite: overwrite}
	data, err := h.cli.uploadOriginExternalPluginV2(nCtx, tenantID, params, fileName, file)
	if err != nil {
		return nil, err
	}

	return &types.OriginExternalPluginV2PkgDetail{
		FileInfo: fileiface.FileInfo{
			Name: data.GetName(),
			Size: data.GetSize(),
			MD5:  data.GetMd5(),
		},
		UploadID:      data.GetUploadId(),
		Existed:       data.GetExisted(),
		PluginPkgName: data.GetName(),
		Version:       data.GetVersion(),
		Description:   data.GetDescription(),
		DescriptionEn: data.GetDescriptionEn(),
		Scenario:      data.GetScenario(),
		ScenarioEn:    data.GetScenarioEn(),
		ConfigFile:    data.GetConfigFile(),
		ConfigFormat:  data.GetConfigFormat(),
		LaunchNode:    data.GetLaunchNode(),
		Platforms:     data.ConvertPlatformsToTypes(),
	}, nil
}

// UploadOriginPluginV3 upload origin plugin v3.
func (h *handler) UploadOriginPluginV3(nCtx contextx.IContext, fileName string, file io.Reader, overwrite bool) (
	*types.OriginPluginV3PkgDetail, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.UploadOriginPluginV3Req{Overwrite: overwrite}
	data, err := h.cli.uploadOriginPluginV3(nCtx, tenantID, params, fileName, file)
	if err != nil {
		return nil, err
	}

	return &types.OriginPluginV3PkgDetail{
		FileInfo: fileiface.FileInfo{
			Name: data.GetName(),
			Size: data.GetSize(),
			MD5:  data.GetMd5(),
		},
		UploadID:         data.GetUploadId(),
		Existed:          data.GetExisted(),
		PluginPkgName:    data.GetPluginPkgName(),
		Version:          data.GetVersion(),
		Description:      data.GetDescription(),
		DescriptionEn:    data.GetDescriptionEn(),
		Scenario:         data.GetScenario(),
		ScenarioEn:       data.GetScenarioEn(),
		LaunchNode:       data.GetLaunchNode(),
		TemplateRenderer: types.TemplateRendererType(data.GetTemplateRenderer()),
		Platforms:        data.ConvertPlatformsToTypes(),
	}, nil
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

// UploadOriginPluginBinTool upload origin plugin bin tool.
func (h *handler) UploadOriginPluginBinTool(nCtx contextx.IContext, fileName string, file io.Reader, overwrite bool) (
	*types.OriginPluginBinToolPkgDetail, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.UploadOriginPluginBinToolReq{
		Overwrite: overwrite,
	}
	data, err := h.cli.uploadOriginPluginBinTool(nCtx, tenantID, params, fileName, file)
	if err != nil {
		return nil, err
	}

	platsV2 := make([]platfmt.Platform, 0)
	for _, plat := range data.GetV2().GetPlatforms() {
		platsV2 = append(platsV2, platfmt.Platform{
			OS:   criteria.OSType(plat.GetOsType()),
			Arch: criteria.CPUArch(plat.GetCpuArch()),
		})
	}

	platsV3 := make([]platfmt.Platform, 0)
	for _, plat := range data.GetV3().GetPlatforms() {
		platsV3 = append(platsV3, platfmt.Platform{
			OS:   criteria.OSType(plat.GetOsType()),
			Arch: criteria.CPUArch(plat.GetCpuArch()),
		})
	}

	return &types.OriginPluginBinToolPkgDetail{
		FileInfo: fileiface.FileInfo{
			Name: data.GetName(),
			Size: data.GetSize(),
			MD5:  data.GetMd5(),
		},
		UploadID: data.GetUploadId(),
		Existed:  data.GetExisted(),
		V2: types.OriginPluginBinToolPkgV2Info{
			Platforms: platsV2,
		},
		V3: types.OriginPluginBinToolPkgV3Info{
			Platforms: platsV3,
		},
	}, nil
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
func (h *handler) InfoReleasePlugin(
	nCtx contextx.IContext, pluginPkgName string, plat platfmt.Platform, version string,
) (*fileiface.FileInfo, error) {

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
