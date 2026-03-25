/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package workflow

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/nodepkg"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/goasync"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tarstream"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	// installerFileMode is the file mode for the installer binary inside the tar package.
	installerFileMode = 0o755

	// scriptFileMode is the file mode for the offline install script inside the tar package.
	scriptFileMode = 0o755

	// configFileMode is the file mode for config and metadata files inside the tar package.
	configFileMode = 0o644
)

// GetOfflinePackageDownload downloads the offline install tar.gz package for an operation.
// nolint: funlen, gocognit
func (h *handler) GetOfflinePackageDownload(rCtx restserver.IContext) (*restserver.StreamResponse, error) {
	req := new(protoApplication.NodeWorkflowOperationOfflinePackageDownloadReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get offline package, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	operationID := req.GetOperationId()
	if operationID == "" {
		logger.G.Biz(rCtx).Error("failed to get offline package, operation_id is required")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("operation_id is required"))
	}

	// Get offline install info from backend (configs, scripts, metadata, installer info).
	infoData, err := h.backendHandler.GetNodeWorkflowOperationOfflineInstallInfo(rCtx, operationID)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get offline package, failed to get offline install info")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	installerInfo := infoData.GetInstallerInfo()
	if installerInfo == nil {
		logger.G.Biz(rCtx).Error("failed to get offline package, installer info is nil")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, fmt.Errorf("installer info is nil"))
	}

	releaseInfo := infoData.GetReleaseInfo()
	if releaseInfo == nil {
		logger.G.Biz(rCtx).Error("failed to get offline package, release info is nil")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, fmt.Errorf("release info is nil"))
	}

	// Compute release package filename and platform for download and tar placement.
	gen := types.Generation(releaseInfo.GetGeneration())
	releasePlat := platfmt.NewPlatform(
		criteria.OSType(releaseInfo.GetOsType()), criteria.CPUArch(releaseInfo.GetCpuArch()))
	releasePkgFilename, err := nodepkg.FormatPkgFileName(gen, types.ReleaseTypeProxy, releasePlat, releaseInfo.GetVersion())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get offline package, failed to format release package filename")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	// Download the release package from the file service.
	releaseStream, err := h.fileHandler.DownloadReleaseProxy(rCtx, gen, releasePlat, releaseInfo.GetVersion())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get offline package, failed to download release package")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	// Download the installer binary from the file service.
	installerOs := criteria.OSType(installerInfo.GetOsType())
	installerArch := criteria.CPUArch(installerInfo.GetCpuArch())
	installerFileName, err := tool.FormatInstallerName(installerOs, installerArch)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get offline package, failed to format installer file name")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	installerStream, err := h.fileHandler.DownloadInstaller(rCtx, installerOs, installerArch)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get offline package, failed to download installer binary")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	pkgName := infoData.GetPackageName()

	// Stream the tar.gz through a pipe to avoid buffering the entire package in memory.
	pipeReader, pipeWriter := io.Pipe()

	if err := h.goAsyncPool.Run(rCtx, func(_ contextx.IContext) error {
		gzWriter := gzip.NewWriter(pipeWriter)
		tarWriter := tar.NewWriter(gzWriter)

		var buildErr error

		defer func() {
			_ = installerStream.Data.Close()
			_ = releaseStream.Data.Close()

			if buildErr != nil {
				_ = pipeWriter.CloseWithError(buildErr)

				return
			}

			if err := tarWriter.Close(); err != nil {
				_ = pipeWriter.CloseWithError(fmt.Errorf("failed to close tar writer: %w", err))

				return
			}

			if err := gzWriter.Close(); err != nil {
				_ = pipeWriter.CloseWithError(fmt.Errorf("failed to close gzip writer: %w", err))

				return
			}

			_ = pipeWriter.Close()
		}()

		// Add installer binary (size comes from Content-Length header returned by file service).
		if buildErr = tarstream.AddStreamFileToTar(tarWriter, pkgName, installerFileName,
			installerStream.Data, installerStream.Headers, installerFileMode); buildErr != nil {
			return buildErr
		}

		// Tar entry names are POSIX paths; use path.Join (not filepath.Join) so separators stay '/'.
		dataDirPrefix := path.Join(pkgName, installer.OfflinePkgRelPathData)
		configDirPrefix := path.Join(pkgName, installer.OfflinePkgRelPathConfig)

		// Add release package under data/ (required by install.sh + --skip_download).
		if buildErr = tarstream.AddStreamFileToTar(tarWriter, dataDirPrefix, releasePkgFilename,
			releaseStream.Data, releaseStream.Headers, configFileMode); buildErr != nil {
			return buildErr
		}

		// Add offline install script at bundle root.
		if buildErr = tarstream.AddTextFileToTar(tarWriter, pkgName, installer.OfflinePkgInstallScriptName,
			[]byte(infoData.GetInstallScript()), scriptFileMode); buildErr != nil {
			return buildErr
		}

		// Add metadata at bundle root.
		if buildErr = tarstream.AddTextFileToTar(tarWriter, pkgName, installer.OfflinePkgMetadataFileName,
			[]byte(infoData.GetMetadata()), configFileMode); buildErr != nil {
			return buildErr
		}

		// Add precheck JSON under data/.
		if buildErr = tarstream.AddTextFileToTar(tarWriter, dataDirPrefix, installer.OfflinePkgPrecheckFileName,
			[]byte(infoData.GetPrecheck()), configFileMode); buildErr != nil {
			return buildErr
		}

		// Add GSE config files under data/config/.
		for fileName, content := range infoData.GetConfigs() {
			if buildErr = tarstream.AddTextFileToTar(tarWriter, configDirPrefix, fileName,
				[]byte(content), configFileMode); buildErr != nil {
				return buildErr
			}
		}

		return nil
	}, goasync.WithName("offline_package_tar_stream")); err != nil {
		_ = installerStream.Data.Close()
		_ = releaseStream.Data.Close()
		_ = pipeWriter.CloseWithError(err)
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get offline package, failed to schedule tar stream task")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to schedule offline package stream: %w", err))
	}

	headers := http.Header{}
	headers.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.tar.gz", pkgName))
	headers.Set("Content-Type", "application/gzip")

	return &restserver.StreamResponse{
		Data:       pipeReader,
		StatusCode: http.StatusOK,
		Headers:    headers,
	}, nil
}

// offlineInstallResultData is the expected JSON structure of result_data (installer.data.json).
type offlineInstallResultData struct {
	AgentID    string `json:"agent_id"`
	Token      string `json:"token"`
	OperInstID string `json:"oper_inst_id"`
}

// offlineInstallMetadataLite extracts only the instance_id from the offline install metadata JSON.
type offlineInstallMetadataLite struct {
	InstanceID string `json:"instance_id"`
}

// SubmitOfflineInstallResult submits the offline install result for an operation.
// It validates the result_data format and oper_inst_id before forwarding to backend.
func (h *handler) SubmitOfflineInstallResult(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.NodeWorkflowOperationOfflineInstallResultSubmitReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to submit offline install result, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to submit offline install result, invalid request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	operationID := req.GetOperationId()
	resultDataRaw := req.GetResultData()

	// validate result_data is valid JSON with expected fields.
	var resultData offlineInstallResultData
	if err := json.Unmarshal([]byte(resultDataRaw), &resultData); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to submit offline install result, result_data is not valid JSON")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter,
			fmt.Errorf("result_data is not valid JSON: %w", err))
	}

	if resultData.AgentID == "" {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter,
			fmt.Errorf("result_data is missing required field: agent_id"))
	}

	if resultData.OperInstID == "" {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter,
			fmt.Errorf("result_data is missing required field: oper_inst_id"))
	}

	// fetch offline install info to get the expected instance_id for comparison.
	infoData, err := h.backendHandler.GetNodeWorkflowOperationOfflineInstallInfo(rCtx, operationID)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to submit offline install result, failed to get offline install info")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	var metadata offlineInstallMetadataLite
	if err := json.Unmarshal([]byte(infoData.GetMetadata()), &metadata); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to submit offline install result, failed to parse metadata")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed,
			fmt.Errorf("failed to parse offline install metadata: %w", err))
	}

	if metadata.InstanceID != "" && resultData.OperInstID != metadata.InstanceID {
		logger.G.Biz(rCtx).
			With("expected", metadata.InstanceID, "got", resultData.OperInstID).
			Error("failed to submit offline install result, oper_inst_id mismatch")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter,
			fmt.Errorf("oper_inst_id(%s) does not match expected instance(%s), "+
				"please verify you are submitting the result for the correct operation",
				resultData.OperInstID, metadata.InstanceID))
	}

	// all pre-checks passed, forward to backend.
	if err := h.backendHandler.SubmitNodeWorkflowOperationOfflineInstallResult(
		rCtx, operationID, resultDataRaw); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to submit offline install result")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	return nil, nil //nolint:nilnil
}
