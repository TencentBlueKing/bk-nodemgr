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
	actionNameEnsurePkg         = "ensure_pkg"
	actionNameReportPrivateData = "report_private_data"

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
	fileManager file.IFileManager
	client      relayhandler.IClientMessager

	logger logger.Logger
}

// NewClientHandler creates a new file handler.
func NewClientHandler(fm file.IFileManager, client relayhandler.IClientMessager,
	logger logger.Logger) IHandler {

	return &handler{
		fileManager: fm,
		client:      client,
		logger:      logger,
	}
}

// CheckPkgStats is a handler for the CheckPkgStats event.
func (h *handler) CheckPkgStats(ctx context.Context, payload []byte) {
	var event protoRelay.CheckPkgStateReq
	if err := json.Unmarshal(payload, &event); err != nil {
		h.logger.Errorf("failed to unmarshal check pkg stat event: %v", err)
		return
	}
	signalPkg := protoRelay.ClientReportSignalPkgComplete
	if exists := h.fileManager.FileExists(ctx, event.PkgName, event.MD5); !exists {
		signalPkg = protoRelay.ClientReportSignalPkgUnComplete
		h.logger.Infof("CheckPkgStats. file(%s), md5(%s), exists(%v)", event.PkgName, event.MD5, exists)
	}

	req := &reportPrivateDataReq{
		ActionName: actionNameEnsurePkg,
		OperInstID: event.PkgName,
		Data: map[string]string{
			event.PkgName: string(signalPkg),
		},
	}

	if err := h.reportActionPrivateData(ctx, h.client, *req); err != nil {
		h.logger.Errorf("failed to report action private data: %v", err)
	}
}

// StoragePkg is a handler for the StoragePkg event.
func (h *handler) StoragePkg(ctx context.Context, payload []byte) {
	var event protoRelay.TransferPkgCompleteReq

	if err := json.Unmarshal(payload, &event); err != nil {
		h.logger.Errorf("failed to unmarshal storage pkg event: %v", err)
		return
	}

	fileInfo, err := h.fileManager.StoreFile(ctx, event.PackageDestDir, event.PkgName)
	if err != nil {
		h.logger.Errorf("failed to store file. dest-dir(%s), pkg-name(%s): %v",
			event.PackageDestDir, event.PkgName, err)

		return
	}

	h.logger.Infof("StoragePkg success. file-name(%s), size(%d), md5(%s)",
		fileInfo.Name, fileInfo.Size, fileInfo.MD5)
}

type reportPrivateDataReq struct {
	ActionName string            `json:"action_name"`
	OperInstID string            `json:"oper_inst_id"`
	Data       map[string]string `json:"data"`
}

func (h *handler) reportActionPrivateData(ctx context.Context,
	client relayhandler.IClientMessager, req reportPrivateDataReq) error {

	jsonData, err := json.Marshal(req)
	if err != nil {
		h.logger.Errorf("failed to marshal status request: %v", err)
		return fmt.Errorf("failed to marshal status request: %w", err)
	}

	errCh := client.ClientPushReq(ctx, actionNameReportPrivateData, jsonData)

	select {
	case err := <-errCh:
		if err != nil {
			h.logger.Errorf("report action private data failed. action-name(%s), instance-id(%s): %v",
				req.ActionName, req.OperInstID, req.Data)

			return fmt.Errorf("report action private data failed. action-name(%s), instance-id(%s): %w",
				req.ActionName, req.OperInstID, err)
		}
		h.logger.Infof("report action private data success. action-name(%s), instance-id(%s): %v",
			req.ActionName, req.OperInstID, req.Data)

		return nil
	case <-time.After(ReportPrivateDataTimeout):
		h.logger.Errorf("report action private data timed out. action-name(%s), instance-id(%s)",
			req.ActionName, req.OperInstID)

		return errors.New("report action private data timed out")
	}
}
