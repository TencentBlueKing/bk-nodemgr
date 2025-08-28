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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/relayconstant"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
)

const (
	actionNameEnsurePkg         = "ensure_pkg"
	reportRelayFileStateURL     = "/relay/report_file_state"
	reportRelayStorageResultURL = "/relay/report_storage_result"
)

// CheckPkgStats is a handler for the CheckPkgStats event.
func (h *handler) CheckPkgStats(ctx context.Context, payload []byte) {
	h.logger.Infof("handler check pkg stat event.")

	var event protoRelay.CheckPkgStateReq
	if err := json.Unmarshal(payload, &event); err != nil {
		h.logger.Errorf("failed to unmarshal check pkg stat event: %v", err)
		return
	}

	fileStates := make([]fileState, 0)

	for _, fileInfo := range event.FileList {
		statePkg := relayconstant.RelayReportPkgInComplete
		if exists := h.fileManager.FileExists(ctx, fileInfo.FileName, fileInfo.FileMD5); exists {
			statePkg = relayconstant.RelayReportPkgComplete
		}

		h.logger.Infof("check package status. file-name(%s), md5(%s), exists(%v)",
			fileInfo.FileName, fileInfo.FileMD5, statePkg)

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

	if err := h.reportRelayFileState(ctx, h.client, req); err != nil {
		h.logger.Errorf("failed to report relay file state: %v", err)
	}

	h.logger.Infof("check package status success")
}

// StoragePkg is a handler for the StoragePkg event.
func (h *handler) StoragePkg(ctx context.Context, payload []byte) {
	h.logger.Infof("handler storage pkg event.")

	var event protoRelay.NotifyReceiveReq

	if err := json.Unmarshal(payload, &event); err != nil {
		h.logger.Errorf("failed to unmarshal storage pkg event: %v", err)
		return
	}

	var errMsg string
	for _, pkgName := range event.PkgName {
		fileInfo, err := h.fileManager.StoreFile(ctx, h.storageTmpDir, pkgName)
		if err != nil {
			h.logger.Errorf("failed to store file. dest-dir(%s), pkg-name(%s): %v",
				h.storageTmpDir, pkgName, err)
			errMsg = err.Error()

			break
		}

		h.logger.Infof("storage package success. file-name(%s), size(%d), md5(%s)",
			fileInfo.Name, fileInfo.Size, fileInfo.MD5)
	}

	req := reportRelayStorageResult{
		ActionName: event.ActionName,
		OperInstID: event.OperInstID,
		ErrMsg:     errMsg,
	}
	if err := h.reportStorageResultReq(ctx, h.client, req); err != nil {
		h.logger.Errorf("failed to report relay file state: %v", err)
	}

	for _, pkgName := range event.PkgName {
		if err := h.safeRemove(h.storageTmpDir, pkgName); err != nil {
			h.logger.Errorf("failed to remove file. dest-dir(%s), pkg-name(%s): %v",
				h.storageTmpDir, pkgName, err)
		}
	}

	h.logger.Infof("storage package event success")
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
