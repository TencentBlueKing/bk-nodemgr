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
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandler is interface for backend file handler.
type IHandler interface {
	// UploadOriginAgent upload origin agent.
	UploadOriginAgent(ctx context.Context, fileName string, file io.Reader) (
		*types.OriginPkgDetail, error)

	// UploadOriginServer upload origin server.
	UploadOriginServer(ctx context.Context, fileName string, file io.Reader) (
		*types.OriginPkgDetail, error)

	// UploadOriginCert upload origin cert.
	UploadOriginCert(ctx context.Context, fileName string, file io.Reader) (
		*types.OriginCertPkgDetail, error)

	// UploadOriginBinTool upload origin bintool.
	UploadOriginBinTool(ctx context.Context, fileName string, file io.Reader) (
		*types.OriginBinToolPkgDetail, error)

	// PublishReleaseAgent publish release agent.
	PublishReleaseAgent(ctx context.Context, uploadID string) error

	// PublishReleaseProxy publish release server.
	PublishReleaseProxy(ctx context.Context, uploadID string) error

	// PublishReleaseCert publish release cert.
	PublishReleaseCert(ctx context.Context, uploadID string) error

	// PublishReleaseBinTool publish release bintool.
	PublishReleaseBinTool(ctx context.Context, uploadID string) error

	// LaunchTransferRelease launch transfer release.
	LaunchTransferRelease(ctx context.Context,
		gen types.Generation,
		rt types.ReleaseType,
		plat platform.Platform,
		version string,
		dstDir string,
		dstHost *types.Host) (types.ISimpleTransferHandler, error)

	// LaunchTransferInstaller launch transfer installer.
	LaunchTransferInstaller(ctx context.Context,
		gen types.Generation,
		plat platform.Platform,
		dstDir string,
		dstHost *types.Host) (types.ISimpleTransferHandler, error)

	// QueryTransfer query transfer package.
	// return upload result, download result and error.
	QueryTransfer(
		ctx context.Context, taskID string) (*types.SimpleTransferResult, *types.SimpleTransferResult, error)
}

const (
	transferQueryTickTime             = 1 * time.Second
	transferQueryContinuesFailedTimes = 5
)

type handler struct {
	cli *cli
}

// New initialize a new nodeman backend handler.
func New(c *client.Capability, conf *Config) (IHandler, error) {
	cli, err := newClient(c, conf)
	if err != nil {
		return nil, err
	}

	return &handler{cli: cli}, nil
}

// UploadOriginAgent upload origin agent.
func (h *handler) UploadOriginAgent(ctx context.Context, fileName string, file io.Reader) (
	*types.OriginPkgDetail, error) {

	resp, err := h.cli.uploadOriginAgent(
		ctx, &protoFile.UploadOriginAgentReq{Generation: int64(types.Generation2)}, fileName, file)
	if err != nil {
		return nil, err
	}

	plats := make([]platform.Platform, 0)
	for _, plat := range resp.GetPlatforms() {
		plats = append(plats, platform.Platform{
			OS:   criteria.OSType(plat.GetOsType()),
			Arch: criteria.CPUArch(plat.GetCpuArch()),
		})
	}

	return &types.OriginPkgDetail{
		FileInfo: iface.FileInfo{
			Name: resp.GetName(),
			Size: resp.GetSize(),
			MD5:  resp.GetMd5(),
		},
		UploadID:    resp.GetUploadId(),
		Existed:     resp.GetExisted(),
		Version:     resp.GetVersion(),
		Platforms:   plats,
		ChangeLogEN: resp.GetChangelogEn(),
		ChangeLogZH: resp.GetChangelogZh(),
	}, nil
}

// UploadOriginServer upload origin server.
func (h *handler) UploadOriginServer(ctx context.Context, fileName string, file io.Reader) (
	*types.OriginPkgDetail, error) {

	resp, err := h.cli.uploadOriginServer(
		ctx, &protoFile.UploadOriginServerReq{Generation: int64(types.Generation2)}, fileName, file)
	if err != nil {
		return nil, err
	}

	plats := make([]platform.Platform, 0)
	for _, plat := range resp.GetPlatforms() {
		plats = append(plats, platform.Platform{
			OS:   criteria.OSType(plat.GetOsType()),
			Arch: criteria.CPUArch(plat.GetCpuArch()),
		})
	}

	return &types.OriginPkgDetail{
		FileInfo: iface.FileInfo{
			Name: resp.GetName(),
			Size: resp.GetSize(),
			MD5:  resp.GetMd5(),
		},
		UploadID:  resp.GetUploadId(),
		Existed:   resp.GetExisted(),
		Version:   resp.GetVersion(),
		Platforms: plats,
	}, nil
}

// UploadOriginCert upload origin cert.
func (h *handler) UploadOriginCert(ctx context.Context, fileName string, file io.Reader) (
	*types.OriginCertPkgDetail, error) {

	resp, err := h.cli.uploadOriginCert(ctx, &protoFile.UploadOriginCertReq{}, fileName, file)
	if err != nil {
		return nil, err
	}

	return &types.OriginCertPkgDetail{
		FileInfo: iface.FileInfo{
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
func (h *handler) UploadOriginBinTool(ctx context.Context, fileName string, file io.Reader) (
	*types.OriginBinToolPkgDetail, error) {

	resp, err := h.cli.uploadOriginBinTool(ctx, &protoFile.UploadOriginBinToolReq{}, fileName, file)
	if err != nil {
		return nil, err
	}

	agentPlats := make([]platform.Platform, 0)
	for _, plat := range resp.GetAgentPlatforms() {
		agentPlats = append(agentPlats, platform.Platform{
			OS:   criteria.OSType(plat.GetOsType()),
			Arch: criteria.CPUArch(plat.GetCpuArch()),
		})
	}
	proxyPlats := make([]platform.Platform, 0)
	for _, plat := range resp.GetProxyPlatforms() {
		proxyPlats = append(proxyPlats, platform.Platform{
			OS:   criteria.OSType(plat.GetOsType()),
			Arch: criteria.CPUArch(plat.GetCpuArch()),
		})
	}

	return &types.OriginBinToolPkgDetail{
		FileInfo: iface.FileInfo{
			Name: resp.GetName(),
			Size: resp.GetSize(),
			MD5:  resp.GetMd5(),
		},
		UploadID:       resp.GetUploadId(),
		Existed:        resp.GetExisted(),
		AgentPlatforms: agentPlats,
		ProxyPlatforms: proxyPlats,
	}, nil
}

// PublishReleaseAgent publish release agent.
func (h *handler) PublishReleaseAgent(ctx context.Context, uploadID string) error {
	_, err := h.cli.publishReleaseAgent(ctx, &protoFile.PublishReleaseAgentReq{
		UploadId: uploadID,
	})
	if err != nil {
		return err
	}

	return nil
}

// PublishReleaseProxy publish release proxy.
func (h *handler) PublishReleaseProxy(ctx context.Context, uploadID string) error {
	_, err := h.cli.publishReleaseProxy(ctx, &protoFile.PublishReleaseProxyReq{
		UploadId: uploadID,
	})
	if err != nil {
		return err
	}

	return nil
}

// PublishReleaseCert publish release cert.
func (h *handler) PublishReleaseCert(ctx context.Context, uploadID string) error {
	_, err := h.cli.publishReleaseCert(ctx, &protoFile.PublishReleaseCertReq{
		UploadId: uploadID,
	})
	if err != nil {
		return err
	}

	return nil
}

// PublishReleaseBinTool publish release bintool.
func (h *handler) PublishReleaseBinTool(ctx context.Context, uploadID string) error {
	_, err := h.cli.publishReleaseBinTool(ctx, &protoFile.PublishReleaseBinToolReq{
		UploadId: uploadID,
	})
	if err != nil {
		return err
	}

	return nil
}

// LaunchTransferRelease launch transfer release.
func (h *handler) LaunchTransferRelease(ctx context.Context,
	gen types.Generation,
	rt types.ReleaseType,
	plat platform.Platform,
	version string,
	dstDir string,
	dstHost *types.Host) (types.ISimpleTransferHandler, error) {

	resp, err := h.cli.launchTransferRelease(ctx, &protoFile.TransferLaunchReleaseReq{
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
		fileInfo: iface.FileInfo{
			Name: data.GetReleaseFileName(),
			Size: data.GetReleaseFileSize(),
			MD5:  data.GetReleaseFileMd5(),
		},
		handler: h,
	}, nil
}

// LaunchTransferInstaller launch transfer installer.
func (h *handler) LaunchTransferInstaller(ctx context.Context,
	gen types.Generation,
	plat platform.Platform,
	dstDir string,
	dstHost *types.Host) (types.ISimpleTransferHandler, error) {

	resp, err := h.cli.launchTransferInstaller(ctx, &protoFile.TransferLaunchInstallerReq{
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
		fileInfo: iface.FileInfo{
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
	ctx context.Context, taskID string) (*types.SimpleTransferResult, *types.SimpleTransferResult, error) {

	resp, err := h.cli.queryTransfer(ctx, &protoFile.TransferQueryReq{
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
	fileInfo iface.FileInfo
	handler  *handler
}

// GetTaskID get task id.
func (handler *simpleTransferHandler) GetTaskID() string {
	return handler.taskID
}

// GetFileInfo get file info.
func (handler *simpleTransferHandler) GetFileInfo() iface.FileInfo {
	return handler.fileInfo
}

// WaitUntilDone wait until done.
func (handler *simpleTransferHandler) WaitUntilDone(ctx context.Context) (*types.SimpleTransferResult, error) {
	ticker := time.NewTicker(transferQueryTickTime)
	failedCnt := 0
	for {
		select {
		case <-ctx.Done():
			return nil, errors.New("context done")
		case <-ticker.C:
			src, dst, err := handler.handler.QueryTransfer(ctx, handler.taskID)
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
