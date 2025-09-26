/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package handler ...
package handler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/relayconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
)

const (
	actionNameEnsurePkg         = "ensure_pkg"
	reportRelayFileStateURL     = "/relay/report_file_state"
	reportRelayStorageResultURL = "/relay/report_storage_result"
)

// CheckPkgStats is a handler for the CheckPkgStats event.
func (h *handler) CheckPkgStats(nCtx contextx.IContext, payload []byte) {
	logger.G.Biz(nCtx).Info("handler check pkg state event")

	var event protoRelay.CheckPkgStateReq
	if err := json.Unmarshal(payload, &event); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to unmarshal check pkg stat event")

		return
	}

	fileStates := make([]fileState, 0)

	for _, fileInfo := range event.FileList {
		statePkg := relayconstant.RelayReportPkgInComplete
		if exists := h.fileManager.FileExists(nCtx, fileInfo.FileName, fileInfo.FileMD5); exists {
			statePkg = relayconstant.RelayReportPkgComplete
		}

		logger.G.Biz(nCtx).With("filename", fileInfo.FileName, "md5", fileInfo.FileMD5, "exists", statePkg).Info("check package status")

		fileStates = append(fileStates, fileState{
			FileName:   fileInfo.FileName,
			FileStatus: string(statePkg),
		})
	}

	req := reportRelayFileState{
		ActionName:    event.ActionName,
		OperInstID:    event.OperInstID,
		StorageTmpDir: h.storageTmpDir,
		FileState:     fileStates,
	}

	if err := h.reportRelayFileState(nCtx, h.client, req); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to report relay file state")
	}

	logger.G.Biz(nCtx).Info("check package status successfully")
}

// StoragePkg is a handler for the StoragePkg event.
func (h *handler) StoragePkg(nCtx contextx.IContext, payload []byte) {
	logger.G.Biz(nCtx).Info("handler storage pkg event")

	var event protoRelay.NotifyReceiveReq

	if err := json.Unmarshal(payload, &event); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to unmarshal storage pkg event")

		return
	}

	var errMsg string
	for _, pkgName := range event.PkgName {
		fileInfo, err := h.fileManager.StoreFile(nCtx, h.storageTmpDir, pkgName)
		if err != nil {
			logger.G.Biz(nCtx).
				AssignWhenLogging(&errMsg).
				WithErr(err).
				With("dest-dir", h.storageTmpDir, "pkgname", pkgName).
				Error("failed to store file")

			break
		}

		logger.G.Biz(nCtx).With("filename", fileInfo.Name, "md5", fileInfo.MD5, "size", fileInfo.Size).Info("storage package successfully")
	}

	req := reportRelayStorageResult{
		ActionName: event.ActionName,
		OperInstID: event.OperInstID,
		ErrMsg:     errMsg,
	}
	if err := h.reportStorageResult(nCtx, h.client, req); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to report relay storage result")
	}

	for _, pkgName := range event.PkgName {
		if err := h.safeRemove(h.storageTmpDir, pkgName); err != nil {
			logger.G.Biz(nCtx).WithErr(err).With("dest-dir", h.storageTmpDir, "pkgname", pkgName).Error("failed to remove file")
		}
	}

	logger.G.Biz(nCtx).Info("storage package event successfully")
}

func (h *handler) safeRemove(baseDir string, filename string) error {
	absPath := filepath.Join(baseDir, filename)
	if absPath == "" ||
		absPath == "/" ||
		strings.HasPrefix(absPath, "/dev/") ||
		strings.HasPrefix(absPath, "/sys/") ||
		strings.HasPrefix(absPath, "/proc/") {

		return fmt.Errorf("failed to remove all, got invalid path. path(%s)", absPath)
	}

	if err := os.RemoveAll(absPath); err != nil {
		return fmt.Errorf("failed to remove all. path(%s): %w", absPath, err)
	}

	return nil
}
