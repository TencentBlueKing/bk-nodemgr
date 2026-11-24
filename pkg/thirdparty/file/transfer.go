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
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ITransfer is interface for file transfer handler.
type ITransfer interface {
	// LaunchTransferNode launch transfer release.
	LaunchTransferNode(
		nCtx contextx.IContext, gen types.Generation, rt types.ReleaseType, plat platfmt.Platform, version string, dstDir string, dstHost *types.Host) (
		types.ISimpleTransferHandler, error)

	// LaunchTransferPlugin launch transfer release.
	LaunchTransferPlugin(
		nCtx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform, version string, dstDir string, dstHost *types.Host) (
		types.ISimpleTransferHandler, error)

	// LaunchTransferInstaller launch transfer installer.
	LaunchTransferInstaller(nCtx contextx.IContext, gen types.Generation, plat platfmt.Platform, dstDir string, dstHost *types.Host) (
		types.ISimpleTransferHandler, error)

	// QueryTransfer query transfer package.
	// return upload result, download result and error.
	QueryTransfer(nCtx contextx.IContext, taskID string) (*types.SimpleTransferResult, *types.SimpleTransferResult, error)
}

const (
	transferQueryTickTime             = 1 * time.Second
	transferQueryContinuesFailedTimes = 5
)

// LaunchTransferNode launch transfer release.
func (h *handler) LaunchTransferNode(
	nCtx contextx.IContext, gen types.Generation, rt types.ReleaseType, plat platfmt.Platform, version string, dstDir string, dstHost *types.Host) (
	types.ISimpleTransferHandler, error) {

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
	nCtx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform, version string, dstDir string, dstHost *types.Host) (
	types.ISimpleTransferHandler, error) {

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
func (h *handler) LaunchTransferInstaller(nCtx contextx.IContext, gen types.Generation, plat platfmt.Platform, dstDir string, dstHost *types.Host) (
	types.ISimpleTransferHandler, error) {

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
func (h *handler) QueryTransfer(nCtx contextx.IContext, taskID string) (*types.SimpleTransferResult, *types.SimpleTransferResult, error) {
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
