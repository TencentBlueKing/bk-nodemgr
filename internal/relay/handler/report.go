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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/pkg/errors"
)

// reportHostInfo report host info to callback.
func (h *handler) reportHostInfo(ctx context.Context,
	actionName, operInstID string,
	osType criteria.OSType, cpuArch criteria.CPUArch,
	connectedDir, msg string) error {

	req := &reportHostInfo{
		ActionName:   actionName,
		OperInstID:   operInstID,
		OsType:       string(osType),
		CPUArch:      string(cpuArch),
		ConnectedDir: connectedDir,
		ErrMsg:       msg,
	}
	jsonData, err := json.Marshal(req)
	if err != nil {
		h.logger.Errorf("failed to marshal status request: %v", err)
		return fmt.Errorf("failed to marshal status request: %w", err)
	}

	h.logger.Infof("report host info. action-name(%s), instance-id(%s), data(%s)",
		actionName, operInstID, string(jsonData))
	errCh := h.client.ClientPushReq(ctx, reportRelayDetectResultURL, jsonData)

	select {
	case err := <-errCh:
		if err != nil {
			h.logger.Errorf("report host info failed. action-name(%s), instance-id(%s): %v",
				req.ActionName, req.OperInstID, err)

			return fmt.Errorf("report host info failed. action-name(%s), instance-id(%s): %w",
				req.ActionName, req.OperInstID, err)
		}
		h.logger.Infof("report host info success. action-name(%s), instance-id(%s)",
			req.ActionName, req.OperInstID)

		return nil
	case <-time.After(ReportPrivateDataTimeout):
		h.logger.Errorf("report host info timed out. action-name(%s), instance-id(%s)",
			req.ActionName, req.OperInstID)

		return errors.New("report host info timed out")
	}
}

type reportHostInfo struct {
	ActionName string `json:"action_name"`
	OperInstID string `json:"oper_inst_id"`

	OsType       string `json:"os_type"`
	CPUArch      string `json:"cpu_arch"`
	ConnectedDir string `json:"connected_dir"`

	ErrMsg string `json:"err_msg"`
}

// reportRelayFileState report relay file state to callback.
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

// reportStorageResultReq report storage result to callback.
func (h *handler) reportStorageResultReq(ctx context.Context,
	client relayhandler.IClientMessager, req reportRelayStorageResult) error {

	h.logger.Infof("report storage result. action-name(%s), instance-id(%s)",
		req.ActionName, req.OperInstID)

	jsonData, err := json.Marshal(req)
	if err != nil {
		h.logger.Errorf("failed to marshal status request: %v", err)
		return fmt.Errorf("failed to marshal status request: %w", err)
	}

	errCh := client.ClientPushReq(ctx, reportRelayStorageResultURL, jsonData)

	select {
	case err := <-errCh:
		if err != nil {
			h.logger.Errorf("report storage result failed. action-name(%s), instance-id(%s): %v",
				req.ActionName, req.OperInstID, err)

			return fmt.Errorf("report storage result failed. action-name(%s), instance-id(%s): %w",
				req.ActionName, req.OperInstID, err)
		}
		h.logger.Infof("report storage result success. action-name(%s), instance-id(%s)",
			req.ActionName, req.OperInstID)

		return nil
	case <-time.After(ReportPrivateDataTimeout):
		h.logger.Errorf("report storage result timed out. action-name(%s), instance-id(%s)",
			req.ActionName, req.OperInstID)

		return errors.New("report storage result timed out")
	}
}

type reportRelayStorageResult struct {
	ActionName string `json:"action_name"`
	OperInstID string `json:"oper_inst_id"`
	ErrMsg     string `json:"err_msg"`
}

// reportInstallResult report install result to callback.
func (h *handler) reportInstallResult(ctx context.Context,
	actionName, operInstID, outStr, errMsg string) error {

	h.logger.Infof("report install info. action-name(%s), instance-id(%s)",
		actionName, operInstID)

	req := &reportInstallResult{
		ActionName: actionName,
		OperInstID: operInstID,
		StdOut:     outStr,
		ErrMsg:     errMsg,
	}
	jsonData, err := json.Marshal(req)
	if err != nil {
		h.logger.Errorf("failed to marshal install result request: %v", err)
		return fmt.Errorf("failed to marshal install result request: %w", err)
	}

	errCh := h.client.ClientPushReq(ctx, reportRelayInstallResultURL, jsonData)

	select {
	case err := <-errCh:
		if err != nil {
			h.logger.Errorf("report install result failed. action-name(%s), instance-id(%s): %v",
				req.ActionName, req.OperInstID, err)

			return fmt.Errorf("report install result failed. action-name(%s), instance-id(%s): %w",
				req.ActionName, req.OperInstID, err)
		}
		h.logger.Infof("report install result success. action-name(%s), instance-id(%s)",
			req.ActionName, req.OperInstID)

		return nil
	case <-time.After(ReportPrivateDataTimeout):
		h.logger.Errorf("report install result timed out. action-name(%s), instance-id(%s)",
			req.ActionName, req.OperInstID)

		return errors.New("report install result imed out")
	}
}

type reportInstallResult struct {
	ActionName string `json:"action_name"`
	OperInstID string `json:"oper_inst_id"`

	StdOut string `json:"std_out"`

	ErrMsg string `json:"err_msg"`
}
