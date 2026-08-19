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

package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer/poller"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoCallback "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/callback"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/sshx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const proxyInstallerPollingTimeout = 10 * time.Minute

// InstallProxyBySSH installs proxy by SSH.
func (h *handler) InstallProxyBySSH(nCtx contextx.IContext, payload []byte) {
	logger.G.Biz(nCtx).Info("handler install proxy by ssh event")

	var (
		event  protoRelay.InstallProxyBySSHReq
		outStr string
		errMsg string
	)

	defer func() {
		if event.ActionName == "" && event.OperInstID == "" {
			return
		}

		if err := h.reportInstallResult(nCtx, event.ActionName, event.OperInstID, outStr, errMsg); err != nil {
			logger.G.Biz(nCtx).WithErr(err).Error("failed to report install proxy by ssh result")

			return
		}

		logger.G.Biz(nCtx).
			With("stdout", outStr, "ip", event.IP, "port", event.Port, "user", event.User).
			Info("done report install proxy by ssh result")
	}()

	if err := json.Unmarshal(payload, &event); err != nil {
		logger.G.Biz(nCtx).AssignWhenLogging(&errMsg).WithErr(err).Error("failed to unmarshal install proxy by ssh event")

		return
	}

	dataDir := path.Clean(path.Join(event.InstallerWorkDir, installer.OfflinePkgRelPathData))
	configDir := path.Clean(path.Join(event.InstallerWorkDir, installer.OfflinePkgRelPathConfig))
	transfers, err := h.buildProxyInstallFileTransfers(nCtx, &event, dataDir, configDir)
	if err != nil {
		logger.G.Biz(nCtx).
			AssignWhenLogging(&errMsg).
			WithErr(err).
			Error("failed to prepare proxy install files")

		return
	}

	client, prepareOutput, err := h.prepareProxyInstallTarget(nCtx, &event, configDir)
	outStr += prepareOutput
	if err != nil {
		logger.G.Biz(nCtx).AssignWhenLogging(&errMsg).WithErr(err).Error("failed to prepare proxy install target")

		return
	}
	defer func() { _ = client.Close() }()

	transferOutput, err := h.transferProxyInstallFiles(nCtx, client, &event, transfers)
	outStr += transferOutput
	if err != nil {
		logger.G.Biz(nCtx).
			AssignWhenLogging(&errMsg).
			WithErr(err).
			With("installer-work-dir", event.InstallerWorkDir).
			Error("failed to transfer proxy install files")

		return
	}

	executeOutput, err := h.executeProxyInstallCommand(nCtx, client, &event)
	outStr += executeOutput
	if err != nil {
		logger.G.Biz(nCtx).
			AssignWhenLogging(&errMsg).
			WithErr(err).
			With("installer-work-dir", event.InstallerWorkDir).
			Error("failed to execute proxy install command")

		return
	}

	h.startProxyInstallerPolling(contextx.WithoutCancel(nCtx), &event)

	logger.G.Biz(nCtx).
		With("stdout", executeOutput, "installer-work-dir", event.InstallerWorkDir).
		Info("install proxy by ssh successfully")
}

func (h *handler) prepareProxyInstallTarget(
	nCtx contextx.IContext,
	event *protoRelay.InstallProxyBySSHReq,
	configDir string,
) (*sshx.Client, string, error) {

	client, err := generateSSHClient(nCtx, event.IP, int(event.Port), event.User, event.Password, types.LoginMode(event.LoginMode))
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate ssh client: %w", err)
	}

	logger.G.Biz(nCtx).With("ip", event.IP, "port", event.Port, "user", event.User).Info("connect to host successfully")
	stdoutResult, stderrResult, err := client.RunCommand("mkdir -p " + configDir)
	if err != nil {
		_ = client.Close()

		return nil, "", fmt.Errorf("failed to make installer work dir %s: %w", configDir, err)
	}
	logger.G.Biz(nCtx).With("dir", configDir).Info("make installer work dir successfully")

	return client, buildLogOutput("mkdir", configDir, stdoutResult, stderrResult), nil
}

func (h *handler) buildProxyInstallFileTransfers(nCtx contextx.IContext, event *protoRelay.InstallProxyBySSHReq, dataDir string, configDir string) (
	[]proxyInstallFileTransfer, error) {

	configTransfers, err := h.proxyConfigTransfers(nCtx, event, configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch proxy config files: %w", err)
	}

	checklistTransfer, err := h.proxyChecklistTransfer(nCtx, event, dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch proxy checklist file: %w", err)
	}

	transfers := []proxyInstallFileTransfer{
		{filename: event.InstallerName, destPath: path.Clean(path.Join(event.InstallerWorkDir, event.InstallerName))},
		{filename: event.ReleaseName, destPath: path.Clean(path.Join(dataDir, event.ReleaseName))},
		checklistTransfer,
	}
	transfers = append(transfers, configTransfers...)

	return transfers, nil
}

func (h *handler) transferProxyInstallFiles(
	nCtx contextx.IContext,
	client *sshx.Client,
	event *protoRelay.InstallProxyBySSHReq,
	transfers []proxyInstallFileTransfer,
) (string, error) {

	var outStr string
	for _, transfer := range transfers {
		if err := h.transferProxyInstallFile(nCtx, client, transfer); err != nil {
			return outStr, fmt.Errorf("failed to transfer proxy install file to %s: %w", transfer.destPath, err)
		}

		outStr += fmt.Sprintf("transfer file to %s successfully\n", transfer.destPath)
		logger.G.Biz(nCtx).
			With("filename", transfer.filename, "dest-path", transfer.destPath).
			Info("transfer proxy install file successfully")
	}

	stdoutResult, stderrResult, err := client.RunCommand("chmod +x " + path.Clean(path.Join(event.InstallerWorkDir, event.InstallerName)))
	if err != nil {
		return outStr, fmt.Errorf("failed to chmod installer: %w", err)
	}

	outStr += buildLogOutput("chmod", event.InstallerName, stdoutResult, stderrResult)

	return outStr, nil
}

func (h *handler) transferProxyInstallFile(
	nCtx contextx.IContext, client *sshx.Client, transfer proxyInstallFileTransfer) error {

	reader, err := h.openProxyInstallFile(nCtx, transfer)
	if err != nil {
		return err
	}

	if err := client.TransferFile(reader, transfer.destPath); err != nil {
		return fmt.Errorf("failed to transfer file to %s: %w", transfer.destPath, err)
	}

	return nil
}

func (h *handler) openProxyInstallFile(nCtx contextx.IContext, transfer proxyInstallFileTransfer) (io.ReadCloser, error) {
	if err := verifyProxyInstallFileTransfer(transfer); err != nil {
		return nil, fmt.Errorf("verify proxy install file transfer failed: %w", err)
	}

	if transfer.content != nil {
		return io.NopCloser(bytes.NewReader(transfer.content)), nil
	}

	file, err := h.fileManager.GetFile(nCtx, transfer.filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	reader, err := file.Content(nCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to get file content: %w", err)
	}

	return reader, nil
}

func (h *handler) executeProxyInstallCommand(nCtx contextx.IContext, client *sshx.Client, event *protoRelay.InstallProxyBySSHReq) (string, error) {
	if event.InstallerCmd == "" {
		return "", fmt.Errorf("proxy installer command is empty")
	}

	cmd := fmt.Sprintf("mkdir -p %s && cd %s && %s", event.InstallerWorkDir, event.InstallerWorkDir, event.InstallerCmd)
	logger.G.Biz(nCtx).
		With("installer-work-dir", event.InstallerWorkDir).
		Info("try to run proxy install command")

	stdoutResult, stderrResult, err := client.RunCommand(cmd)
	outStr := buildLogOutput("install", "installer command", stdoutResult, stderrResult)
	if err != nil {
		return outStr, fmt.Errorf("failed to run proxy install command: %w", err)
	}

	return outStr, nil
}

func (h *handler) startProxyInstallerPolling(nCtx contextx.IContext, event *protoRelay.InstallProxyBySSHReq) {
	eventCopy := *event
	go h.pollProxyInstaller(nCtx, &eventCopy)
}

func (h *handler) pollProxyInstaller(nCtx contextx.IContext, event *protoRelay.InstallProxyBySSHReq) {
	pollCtx, cancel := contextx.WithTimeout(nCtx, proxyInstallerPollingTimeout)
	defer cancel()

	result, err := h.waitProxyInstaller(pollCtx, event)
	logger.G.Biz(nCtx).With("installer-work-dir", event.InstallerWorkDir, "state", result.State).
		Info("proxy installer polling finished")
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to wait proxy installer complete")
		h.reportProxyInstallerTimeout(nCtx, event)

		return
	}
	if result.State == installer.ProcessStateSuccess && result.AgentID == "" {
		logger.G.Biz(nCtx).Error("proxy installer succeeded but agent_id is empty")
		h.reportProxyInstallerTimeout(nCtx, event)

		return
	}

	if result.Error != "" {
		h.reportProxyInstallerError(nCtx, event, result.Error)
	}
	if err := h.reportProxyInstallerStatus(nCtx, event, result); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to report proxy installer status")
	}
}

func (h *handler) waitProxyInstaller(
	ctx contextx.IContext, event *protoRelay.InstallProxyBySSHReq,
) (poller.Result, error) {

	dataDir := path.Clean(path.Join(event.InstallerWorkDir, installer.OfflinePkgRelPathData))
	statusPath := path.Join(dataDir, installer.StatusFileName)
	dataPath := path.Join(dataDir, installer.DataFileName)
	logPath := path.Join(dataDir, "logs", "installer_*.log")
	result, err := poller.Wait(ctx, poller.Config{
		NewClient: func(ctx context.Context) (poller.FileClient, error) {
			return generateSSHClient(ctx, event.IP, int(event.Port), event.User, event.Password,
				types.LoginMode(event.LoginMode))
		},
		StatusFile:  statusPath,
		DataFile:    dataPath,
		LogGlobPath: logPath,
		InstanceID:  event.OperInstID,
		Interval:    3 * time.Second, // nolint: mnd
		Timeout:     proxyInstallerPollingTimeout,
		ReadTimeout: sshx.DefaultTimeout,
		OnLogs: func(ctx context.Context, logs []poller.LogEntry) error {
			if len(logs) == 0 {
				return nil
			}
			if err := h.reportProxyInstallerLogs(ctx, event, logs); err != nil {
				logger.G.Biz(contextx.FromContext(ctx)).WithErr(err).Warn("failed to report proxy installer logs")

				return err
			}

			return nil
		},
	})
	if err != nil {
		return poller.Result{}, err
	}

	return result, nil
}

func (h *handler) reportProxyInstallerTimeout(nCtx contextx.IContext, event *protoRelay.InstallProxyBySSHReq) {
	if err := h.reportProxyInstallerStatusField(nCtx, event, string(installer.ProcessStateTimeout)); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to report proxy installer timeout status")
	}
}

func (h *handler) reportProxyInstallerError(
	nCtx contextx.IContext,
	event *protoRelay.InstallProxyBySSHReq,
	errorDetail string,
) {

	if err := h.reportProxyInstallerLogs(nCtx, event, []poller.LogEntry{{
		Timestamp: time.Now().Unix(),
		Level:     "ERROR",
		Step:      "installer",
		Message:   errorDetail,
	}}); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to report proxy installer error detail")
	}
}

func (h *handler) reportProxyInstallerLogs(
	ctx context.Context,
	event *protoRelay.InstallProxyBySSHReq,
	logs []poller.LogEntry,
) error {

	if len(logs) == 0 {
		return nil
	}

	reportLogs := make([]*protoCallback.ReportLog, 0, len(logs))
	for _, entry := range logs {
		reportLogs = append(reportLogs, &protoCallback.ReportLog{
			Timestamp: entry.Timestamp,
			Level:     entry.Level,
			Step:      entry.Step,
			Log:       entry.Message,
		})
	}

	body, err := json.Marshal(&protoCallback.ReportLogReq{
		Token:      event.Token,
		Logs:       reportLogs,
		OperInstId: event.OperInstID,
	})
	if err != nil {
		return fmt.Errorf("marshal proxy installer log request failed: %w", err)
	}

	reportCtx, cancel := context.WithTimeout(ctx, ReportPrivateDataTimeout)
	defer cancel()
	callbackCtx := contextx.FromContext(reportCtx)

	if _, statusCode, err := h.client.RequestCallback(callbackCtx, http.MethodPost, proxyReportLogPath, "", body); err != nil {
		return fmt.Errorf("request proxy installer log callback failed: %w", err)
	} else if statusCode != http.StatusOK {
		return fmt.Errorf("proxy installer log callback returned status code %d", statusCode)
	}

	return nil
}

func (h *handler) reportProxyInstallerStatus(
	nCtx contextx.IContext,
	event *protoRelay.InstallProxyBySSHReq,
	result poller.Result,
) error {

	if err := h.reportProxyInstallerStatusField(nCtx, event, string(result.State)); err != nil {
		return err
	}
	if result.AgentID == "" {
		return nil
	}
	if err := h.reportProxyInstallerDataField(nCtx, event, result.AgentID); err != nil {
		return err
	}

	return nil
}

func (h *handler) reportProxyInstallerStatusField(
	nCtx contextx.IContext,
	event *protoRelay.InstallProxyBySSHReq,
	status string,
) error {

	req := struct {
		Token      string `json:"token"`
		Status     string `json:"status"`
		OperInstID string `json:"oper_inst_id"`
	}{
		Token:      event.Token,
		Status:     status,
		OperInstID: event.OperInstID,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal proxy installer status request failed: %w", err)
	}

	if _, statusCode, err := h.client.RequestCallback(nCtx, http.MethodPost, proxyReportStatusPath, "", body); err != nil {
		return fmt.Errorf("request proxy installer status callback failed: %w", err)
	} else if statusCode != http.StatusOK {
		return fmt.Errorf("proxy installer status callback returned status code %d", statusCode)
	}

	return nil
}

func (h *handler) reportProxyInstallerDataField(
	nCtx contextx.IContext,
	event *protoRelay.InstallProxyBySSHReq,
	agentID string,
) error {

	req := struct {
		Token      string `json:"token"`
		AgentID    string `json:"agent_id"`
		OperInstID string `json:"oper_inst_id"`
	}{
		Token:      event.Token,
		AgentID:    agentID,
		OperInstID: event.OperInstID,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal proxy installer data request failed: %w", err)
	}

	if _, statusCode, err := h.client.RequestCallback(nCtx, http.MethodPost, proxyReportDataPath, "", body); err != nil {
		return fmt.Errorf("request proxy installer data callback failed: %w", err)
	} else if statusCode != http.StatusOK {
		return fmt.Errorf("proxy installer data callback returned status code %d", statusCode)
	}

	return nil
}

type proxyInstallFileTransfer struct {
	filename string
	destPath string
	content  []byte
}

type proxyInstallConfigCallback struct {
	filename string
	urlPath  string
}

func (h *handler) fetchProxyInstallArtifact(nCtx contextx.IContext, token string, urlPath string) ([]byte, error) {
	if token == "" {
		return nil, fmt.Errorf("token is empty")
	}

	req := struct {
		Token string `json:"token"`
	}{Token: token}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal callback request failed: %w", err)
	}

	content, statusCode, err := h.client.RequestCallback(nCtx, http.MethodPost, urlPath, "", body)
	if err != nil {
		return nil, fmt.Errorf("request callback failed: %w", err)
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("callback returned status code %d: %s", statusCode, string(content))
	}

	return content, nil
}

const (
	proxyGetAgentConfigPath     = "/api/v3/callback/workflow/node_install/get_agent_config"
	proxyGetFileProxyConfigPath = "/api/v3/callback/workflow/node_install/get_file_proxy_config"
	proxyGetDataProxyConfigPath = "/api/v3/callback/workflow/node_install/get_data_proxy_config"
	proxyGetCheckListPath       = "/api/v3/callback/workflow/node_install/get_check_list"
	proxyReportLogPath          = "/api/v3/callback/workflow/node_install/report_log"
	proxyReportStatusPath       = "/api/v3/callback/workflow/node_install/report_status"
	proxyReportDataPath         = "/api/v3/callback/workflow/node_install/report_data"
)

func (h *handler) proxyConfigTransfers(
	nCtx contextx.IContext, event *protoRelay.InstallProxyBySSHReq, configDir string) ([]proxyInstallFileTransfer, error) {

	callbacks := []proxyInstallConfigCallback{
		{filename: installer.OfflineGseAgentConfFileName, urlPath: proxyGetAgentConfigPath},
		{filename: installer.OfflineGseFileProxyConfFileName, urlPath: proxyGetFileProxyConfigPath},
		{filename: installer.OfflineGseDataProxyConfFileName, urlPath: proxyGetDataProxyConfigPath},
	}
	transfers := make([]proxyInstallFileTransfer, 0, len(callbacks))
	for _, callback := range callbacks {
		content, err := h.fetchProxyInstallArtifact(nCtx, event.Token, callback.urlPath)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch %s: %w", callback.filename, err)
		}

		transfers = append(transfers, proxyInstallFileTransfer{
			filename: callback.filename,
			destPath: path.Clean(path.Join(configDir, callback.filename)),
			content:  content,
		})
	}

	return transfers, nil
}

func (h *handler) proxyChecklistTransfer(
	nCtx contextx.IContext, event *protoRelay.InstallProxyBySSHReq, dataDir string) (proxyInstallFileTransfer, error) {

	content, err := h.fetchProxyInstallArtifact(nCtx, event.Token, proxyGetCheckListPath)
	if err != nil {
		return proxyInstallFileTransfer{}, fmt.Errorf("failed to fetch checklist: %w", err)
	}

	return proxyInstallFileTransfer{
		filename: installer.OfflinePkgPrecheckFileName,
		destPath: path.Clean(path.Join(dataDir, installer.OfflinePkgPrecheckFileName)),
		content:  content,
	}, nil
}

func verifyProxyInstallFileTransfer(transfer proxyInstallFileTransfer) error {
	if conv.IsEmpty(transfer.filename) {
		return fmt.Errorf("file name is empty")
	}

	if err := validateWorkspaceFilename(transfer.filename); err != nil {
		return fmt.Errorf("invalid file name: %w", err)
	}

	if conv.IsEmpty(transfer.destPath) {
		return fmt.Errorf("file dest path is empty, file-name(%s)", transfer.filename)
	}

	return nil
}
