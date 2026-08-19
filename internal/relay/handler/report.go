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
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/pkg/errors"
)

// reportHostInfo report host info to callback.
func (h *handler) reportHostInfo(nCtx contextx.IContext,
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
		logger.G.Biz(nCtx).WithErr(err).Error("failed to marshal status request")

		return err
	}

	logger.G.Biz(nCtx).With("action", actionName, "oper-inst-id", operInstID, "data", string(jsonData)).Info("report host info")
	errCh := h.client.ClientPushReq(nCtx, reportRelayDetectResultURL, jsonData)

	select {
	case err := <-errCh:
		if err != nil {
			logger.G.Biz(nCtx).WithErr(err).With("action", actionName, "oper-inst-id", operInstID).Error("failed to report host info")

			return fmt.Errorf("report host info failed. action-name(%s), instance-id(%s): %w",
				req.ActionName, req.OperInstID, err)
		}
		logger.G.Biz(nCtx).With("action", actionName, "oper-inst-id", operInstID).Info("success to report host info")

		return nil
	case <-time.After(ReportPrivateDataTimeout):
		logger.G.Biz(nCtx).With("action", actionName, "oper-inst-id", operInstID).Error("report host info timed out")

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
func (h *handler) reportRelayFileState(nCtx contextx.IContext,
	client relayhandler.IClientMessager, req reportRelayFileState) error {

	logger.G.Biz(nCtx).With("action", req.ActionName, "oper-inst-id", req.OperInstID).Info("report relay file state")

	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to marshal file state request")

		return err
	}

	errCh := client.ClientPushReq(nCtx, reportRelayFileStateURL, jsonData)

	select {
	case err := <-errCh:
		if err != nil {
			logger.G.Biz(nCtx).WithErr(err).With("action", req.ActionName, "oper-inst-id", req.OperInstID).Error("failed to report relay file state")

			return fmt.Errorf("report relay file state failed. action-name(%s), instance-id(%s): %w",
				req.ActionName, req.OperInstID, err)
		}
		logger.G.Biz(nCtx).With("action", req.ActionName, "oper-inst-id", req.OperInstID).Info("success to report relay file state")

		return nil
	case <-time.After(ReportPrivateDataTimeout):

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
func (h *handler) reportStorageResult(nCtx contextx.IContext,
	client relayhandler.IClientMessager, req reportRelayStorageResult) error {

	logger.G.Biz(nCtx).With("action", req.ActionName, "oper-inst-id", req.OperInstID).Info("report storage result")

	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to marshal storage result request")

		return err
	}

	errCh := client.ClientPushReq(nCtx, reportRelayStorageResultURL, jsonData)

	select {
	case err := <-errCh:
		if err != nil {
			logger.G.Biz(nCtx).WithErr(err).With("action", req.ActionName, "oper-inst-id", req.OperInstID).Error("failed to report storage result")

			return fmt.Errorf("report storage result failed. action-name(%s), instance-id(%s): %w",
				req.ActionName, req.OperInstID, err)
		}
		logger.G.Biz(nCtx).With("action", req.ActionName, "oper-inst-id", req.OperInstID).Info("success to report storage result")

		return nil
	case <-time.After(ReportPrivateDataTimeout):
		logger.G.Biz(nCtx).With("action", req.ActionName, "oper-inst-id", req.OperInstID).Error("report storage result timed out")

		return errors.New("report storage result timed out")
	}
}

type reportRelayStorageResult struct {
	ActionName string `json:"action_name"`
	OperInstID string `json:"oper_inst_id"`
	ErrMsg     string `json:"err_msg"`
}

// reportInstallResult report install result to callback.
func (h *handler) reportInstallResult(nCtx contextx.IContext,
	actionName, operInstID, outStr, errMsg string) error {

	logger.G.Biz(nCtx).With("action", actionName, "oper-inst-id", operInstID).Info("report install result")

	req := &reportInstallResult{
		ActionName: actionName,
		OperInstID: operInstID,
		StdOut:     outStr,
		ErrMsg:     errMsg,
	}
	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to marshal install result request")

		return err
	}

	errCh := h.client.ClientPushReq(nCtx, reportRelayInstallResultURL, jsonData)

	select {
	case err := <-errCh:
		if err != nil {
			logger.G.Biz(nCtx).WithErr(err).With("action", req.ActionName, "oper-inst-id", req.OperInstID).Error("failed to report install result")

			return fmt.Errorf("report install result failed. action-name(%s), instance-id(%s): %w",
				req.ActionName, req.OperInstID, err)
		}
		logger.G.Biz(nCtx).With("action", req.ActionName, "oper-inst-id", req.OperInstID).Info("success to report install result")

		return nil
	case <-time.After(ReportPrivateDataTimeout):
		logger.G.Biz(nCtx).With("action", req.ActionName, "oper-inst-id", req.OperInstID).Error("report install result timed out")

		return errors.New("report install result timed out")
	}
}

type reportInstallResult struct {
	ActionName string `json:"action_name"`
	OperInstID string `json:"oper_inst_id"`

	// outstr contains the stdout and stderr of the install process.
	StdOut string `json:"std_out"`

	ErrMsg string `json:"err_msg"`
}
