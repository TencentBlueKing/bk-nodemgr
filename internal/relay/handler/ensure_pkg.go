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

// Package handler ...
package handler

import (
	"encoding/json"

	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/relayconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
)

const (
	reportRelayFileStateURL     = "/relay/report_file_state"
	reportRelayStorageResultURL = "/relay/report_storage_result"
	storageTmpDirMode           = 0750
)

// CheckPkgStats is a handler for the CheckPkgStats event.
func (h *handler) CheckPkgStats(nCtx contextx.IContext, payload []byte) {
	logger.G.Biz(nCtx).Info("handler check pkg state event")

	var event protoRelay.CheckPkgStateReq
	if err := json.Unmarshal(payload, &event); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to unmarshal check pkg stat event")

		return
	}

	// Each operation instance gets its own staging directory, reported back so the transfer
	// lands there. Concurrent installations of the same package then never share a path.
	stagingDir, err := h.storageFS.instanceDir(event.OperInstID)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("oper-inst-id", event.OperInstID).Error("failed to resolve staging dir")

		return
	}

	fileStates := make([]fileState, 0)
	needsTransfer := false

	for _, fileInfo := range event.FileList {
		statePkg := relayconstant.RelayReportPkgInComplete
		if exists := h.fileManager.FileExists(nCtx, fileInfo.FileName, fileInfo.FileMD5); exists {
			statePkg = relayconstant.RelayReportPkgComplete
		} else {
			needsTransfer = true
		}

		logger.G.Biz(nCtx).With("filename", fileInfo.FileName, "md5", fileInfo.FileMD5, "exists", statePkg).Info("check package status")

		fileStates = append(fileStates, fileState{
			FileName:   fileInfo.FileName,
			FileStatus: string(statePkg),
		})
	}

	// Only materialize the directory when something will actually be transferred into it.
	// The backend skips the store step entirely when every package is already cached, and it
	// is that step which removes the directory, so creating it eagerly would leave an empty
	// directory behind on every cache hit until the orphan gc reclaims it.
	if needsTransfer {
		if _, err := h.storageFS.ensureInstanceDir(event.OperInstID); err != nil {
			logger.G.Biz(nCtx).WithErr(err).With("oper-inst-id", event.OperInstID).Error("failed to ensure staging dir")

			return
		}
	}

	req := reportRelayFileState{
		ActionName:    event.ActionName,
		OperInstID:    event.OperInstID,
		StorageTmpDir: stagingDir,
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

	stagingDir, err := h.storageFS.instanceDir(event.OperInstID)

	var errMsg string
	if err != nil {
		// Report instead of returning silently: the backend is blocked waiting for this result
		// and would otherwise only learn about the failure when its timeout expires.
		logger.G.Biz(nCtx).
			AssignWhenLogging(&errMsg).
			WithErr(err).
			With("oper-inst-id", event.OperInstID).
			Error("failed to resolve staging dir")
	} else {
		errMsg = h.storeTransferredPkgs(nCtx, stagingDir, event.FileList)
	}

	req := reportRelayStorageResult{
		ActionName: event.ActionName,
		OperInstID: event.OperInstID,
		ErrMsg:     errMsg,
	}
	if err := h.reportStorageResult(nCtx, h.client, req); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to report relay storage result")
	}

	if stagingDir == "" {
		return
	}

	// The staging directory belongs to this operation instance alone, so dropping it wholesale
	// cannot disturb a concurrent installation.
	if err := h.storageFS.removeInstanceDir(event.OperInstID); err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("staging-dir", stagingDir).Error("failed to remove staging dir")
	}

	logger.G.Biz(nCtx).Info("storage package event successfully")
}

// storeTransferredPkgs promotes every transferred package into the cache and returns the
// error message to report back, empty when all packages were stored.
//
// A missing or malformed MD5 is refused instead of being stored unverified: the staging file
// is produced by an external transfer, so without the expected MD5 there is no way to tell a
// complete package from a truncated one, and caching a truncated one would serve a corrupt
// package to every node that follows.
func (h *handler) storeTransferredPkgs(
	nCtx contextx.IContext, stagingDir string, files []protoRelay.FileInfo) string {

	if len(files) == 0 {
		var errMsg string
		logger.G.Biz(nCtx).
			AssignWhenLogging(&errMsg).
			With("staging-dir", stagingDir).
			Error("storage pkg event carries no file list, refusing to store unverified packages")

		return errMsg
	}

	var errMsg string

	for _, file := range files {
		if err := validateWorkspaceFilename(file.FileName); err != nil {
			logger.G.Biz(nCtx).
				AssignWhenLogging(&errMsg).
				WithErr(err).
				With("staging-dir", stagingDir, "pkgname", file.FileName).
				Error("failed to validate package filename")

			return errMsg
		}

		if file.FileMD5 == "" {
			logger.G.Biz(nCtx).
				AssignWhenLogging(&errMsg).
				With("staging-dir", stagingDir, "pkgname", file.FileName).
				Error("package md5 is empty, refusing to store unverified package")

			return errMsg
		}

		fileInfo, err := h.fileManager.StoreFile(nCtx, stagingDir, file.FileName, file.FileMD5)
		if err != nil {
			logger.G.Biz(nCtx).
				AssignWhenLogging(&errMsg).
				WithErr(err).
				With("staging-dir", stagingDir, "pkgname", file.FileName).
				Error("failed to store file")

			return errMsg
		}

		logger.G.Biz(nCtx).
			With("filename", fileInfo.Name, "md5", fileInfo.MD5, "size", fileInfo.Size).
			Info("storage package successfully")
	}

	return errMsg
}
