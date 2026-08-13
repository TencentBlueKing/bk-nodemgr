/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package file

import (
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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandler is interface for file handler.
type IHandler interface {
	IPkgManager
	ITransfer
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
	PublishReleaseAgent(nCtx contextx.IContext, uploadID string) error

	// PublishReleaseProxy publish release server.
	PublishReleaseProxy(nCtx contextx.IContext, uploadID, uploadCategory string) error

	// PublishReleaseCert publish release cert.
	PublishReleaseCert(nCtx contextx.IContext, uploadID string) error

	// PublishReleaseBinTool publish release bintool.
	PublishReleaseBinTool(nCtx contextx.IContext, uploadID string) error

	// PublishReleasePluginV2 publish release plugin v2.
	PublishReleasePluginV2(nCtx contextx.IContext, uploadID string) error

	// PublishReleaseExternalPluginV2 publish release external plugin v2.
	PublishReleaseExternalPluginV2(nCtx contextx.IContext, uploadID string) error

	// PublishReleasePluginV3 publish release plugin v3.
	PublishReleasePluginV3(nCtx contextx.IContext, uploadID string) error

	// PublishReleasePluginBinTool publish release plugin bin tool.
	PublishReleasePluginBinTool(nCtx contextx.IContext, uploadID string) error
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

const (
	transferQueryTickTime             = 1 * time.Second
	transferQueryContinuesFailedTimes = 5
)

type handler struct {
	cli *cli
}

// New initialize a new nodeman backend handler.
func New(c *restclient.Capability, conf *Config) (IHandler, error) {
	cli, err := newClient(c, conf)
	if err != nil {
		return nil, err
	}

	return &handler{cli: cli}, nil
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
func (h *handler) PublishReleaseAgent(nCtx contextx.IContext, uploadID string) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.PublishReleaseAgentReq{UploadId: uploadID}
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
func (h *handler) PublishReleaseProxy(nCtx contextx.IContext, uploadID, uploadCategory string) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.PublishReleaseProxyReq{
		UploadId:            uploadID,
		UploadOriginPkgType: uploadCategory,
	}
	if _, err := h.cli.publishReleaseProxy(nCtx, tenantID, params); err != nil {
		return err
	}

	return nil
}

// PublishReleaseCert publish release cert.
func (h *handler) PublishReleaseCert(nCtx contextx.IContext, uploadID string) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.PublishReleaseCertReq{UploadId: uploadID}
	if _, err := h.cli.publishReleaseCert(nCtx, tenantID, params); err != nil {
		return err
	}

	return nil
}

// PublishReleaseBinTool publish release bintool.
func (h *handler) PublishReleaseBinTool(nCtx contextx.IContext, uploadID string) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.PublishReleaseBinToolReq{UploadId: uploadID}
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
func (h *handler) PublishReleasePluginV2(nCtx contextx.IContext, uploadID string) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.PublishReleasePluginV2Req{UploadId: uploadID}
	if _, err := h.cli.publishReleasePluginV2(nCtx, tenantID, params); err != nil {
		return err
	}

	return nil
}

// PublishReleaseExternalPluginV2 publish release external plugin v2.
func (h *handler) PublishReleaseExternalPluginV2(nCtx contextx.IContext, uploadID string) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.PublishReleaseExternalPluginV2Req{UploadId: uploadID}
	if _, err := h.cli.publishReleaseExternalPluginV2(nCtx, tenantID, params); err != nil {
		return err
	}

	return nil
}

// PublishReleasePluginV3 publish release plugin v3.
func (h *handler) PublishReleasePluginV3(nCtx contextx.IContext, uploadID string) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.PublishReleasePluginV3Req{UploadId: uploadID}
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
func (h *handler) PublishReleasePluginBinTool(nCtx contextx.IContext, uploadID string) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.PublishReleasePluginBinToolReq{UploadId: uploadID}
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
