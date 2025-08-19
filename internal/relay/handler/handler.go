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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/file"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/pkg/errors"
)

const (
	actionNameEnsurePkg     = "ensure_pkg"
	reportRelayFileStateURL = "/relay/report_file_state"

	// ReportPrivateDataTimeout defines the report private data timeout.
	ReportPrivateDataTimeout = 3 * time.Second
)

// IHandler defines a relay client handler.
type IHandler interface {
	CheckPkgStats(ctx context.Context, payload []byte)
	StoragePkg(ctx context.Context, payload []byte)
}

// handler is a relay client handler.
type handler struct {
	storageTmpDir string

	fileManager file.IFileManager
	client      relayhandler.IClientMessager

	logger logger.Logger
}

// NewClientHandler creates a new file handler.
func NewClientHandler(fm file.IFileManager, client relayhandler.IClientMessager,
	logger logger.Logger, storageTmpDir string) IHandler {

	return &handler{
		storageTmpDir: storageTmpDir,
		fileManager:   fm,
		client:        client,
		logger:        logger,
	}
}

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
		statePkg := protoRelay.ClientReportPkgInComplete
		if exists := h.fileManager.FileExists(ctx, fileInfo.FileName, fileInfo.FileMD5); exists {
			statePkg = protoRelay.ClientReportPkgComplete
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

	var event protoRelay.TransferPkgCompleteReq

	if err := json.Unmarshal(payload, &event); err != nil {
		h.logger.Errorf("failed to unmarshal storage pkg event: %v", err)
		return
	}

	for _, pkgName := range event.PkgName {
		fileInfo, err := h.fileManager.StoreFile(ctx, h.storageTmpDir, pkgName)
		if err != nil {
			h.logger.Errorf("failed to store file. dest-dir(%s), pkg-name(%s): %v",
				h.storageTmpDir, pkgName, err)

			continue
		}

		h.logger.Infof("storage package success. file-name(%s), size(%d), md5(%s)",
			fileInfo.Name, fileInfo.Size, fileInfo.MD5)
	}

	h.logger.Infof("storage package success")
}

type reportRelayFileState struct {
	ActionName    string      `json:"action_name"`
	OperInstID    string      `json:"oper_inst_id"`
	StorageTmpDir string      `json:"storage_tmp_dir"`
	FileState     []fileState `json:"file_state"`
}

type fileState struct {
	FileName   string `json:"file_name"`
	FileStatus string `json:"file_status"`
}

func (h *handler) reportRelayFileState(ctx context.Context,
	client relayhandler.IClientMessager, req reportRelayFileState) error {

	h.logger.Infof("report relay file state. action-name(%s), instance-id(%s)",
		req.ActionName, req.OperInstID)

	jsonData, err := json.Marshal(req)
	if err != nil {
		h.logger.Errorf("failed to marshal status request: %v", err)
		return fmt.Errorf("failed to marshal status request: %w", err)
	}

	errCh := client.ClientPushReq(ctx, reportRelayFileStateURL, jsonData)

	select {
	case err := <-errCh:
		if err != nil {
			h.logger.Errorf("report relay file state failed. action-name(%s), instance-id(%s): %v",
				req.ActionName, req.OperInstID, err)

			return fmt.Errorf("report relay file state failed. action-name(%s), instance-id(%s): %w",
				req.ActionName, req.OperInstID, err)
		}
		h.logger.Infof("report relay file state success. action-name(%s), instance-id(%s)",
			req.ActionName, req.OperInstID)

		return nil
	case <-time.After(ReportPrivateDataTimeout):
		h.logger.Errorf("report relay file state timed out. action-name(%s), instance-id(%s)",
			req.ActionName, req.OperInstID)

		return errors.New("report relay file state timed out")
	}
}
