/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/sshx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	proxyInstallerPollingInterval = 3 * time.Second
	proxyInstallerPollingTimeout  = 10 * time.Minute
)

type proxyInstallerStatusFile struct {
	OperInstID string `json:"oper_inst_id"`
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
}

type proxyInstallerDataFile struct {
	AgentID    string `json:"agent_id"`
	Token      string `json:"token"`
	OperInstID string `json:"oper_inst_id"`
}

type proxyInstallerPollingResult struct {
	Status  string
	AgentID string
}

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
	outStr += h.readProxyInstallerLogOutput(nCtx, client, event)
	if err != nil {
		return outStr, fmt.Errorf("failed to run proxy install command: %w", err)
	}

	return outStr, nil
}

func (h *handler) readProxyInstallerLogOutput(
	nCtx contextx.IContext,
	client *sshx.Client,
	event *protoRelay.InstallProxyBySSHReq,
) string {

	logGlobPath := path.Clean(path.Join(event.InstallerWorkDir, installer.OfflinePkgRelPathData, "logs", "installer_*.log"))
	cmd := fmt.Sprintf(`latest_log=$(ls -1t %s 2>/dev/null | head -1); if [ -n "$latest_log" ]; then cat "$latest_log"; fi`, logGlobPath)
	stdoutResult, stderrResult, err := client.RunCommand(cmd)
	if err != nil {
		logger.G.Biz(nCtx).
			WithErr(err).
			With("installer-log-path", logGlobPath).
			Warn("failed to read proxy installer log")

		return buildLogOutput("read", "installer log", "", fmt.Sprintf("failed to read installer log: %v", err))
	}

	return buildLogOutput("read", "installer log", stdoutResult, stderrResult)
}

func (h *handler) startProxyInstallerPolling(nCtx contextx.IContext, event *protoRelay.InstallProxyBySSHReq) {
	eventCopy := *event
	go func() {
		client, err := generateSSHClient(nCtx, eventCopy.IP, int(eventCopy.Port), eventCopy.User, eventCopy.Password,
			types.LoginMode(eventCopy.LoginMode))
		if err != nil {
			logger.G.Biz(nCtx).WithErr(err).Error("failed to generate proxy install polling ssh client")
			if err := h.reportProxyInstallerStatusField(nCtx, &eventCopy, string(installer.ProcessStateTimeout)); err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to report proxy installer timeout status")
			}

			return
		}
		defer func() { _ = client.Close() }()

		pollingResult, pollingOutput, err := h.waitProxyInstallerComplete(nCtx, client, &eventCopy)
		logger.G.Biz(nCtx).With("stdout", pollingOutput, "installer-work-dir", eventCopy.InstallerWorkDir).
			Info("proxy installer polling finished")
		if err != nil {
			logger.G.Biz(nCtx).WithErr(err).Error("failed to wait proxy installer complete")
			if err := h.reportProxyInstallerStatusField(nCtx, &eventCopy, string(installer.ProcessStateTimeout)); err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to report proxy installer timeout status")
			}

			return
		}

		if err := h.reportProxyInstallerStatus(nCtx, &eventCopy, pollingResult); err != nil {
			logger.G.Biz(nCtx).WithErr(err).Error("failed to report proxy installer status")
		}
	}()
}

func (h *handler) waitProxyInstallerComplete(
	nCtx contextx.IContext,
	client *sshx.Client,
	event *protoRelay.InstallProxyBySSHReq,
) (proxyInstallerPollingResult, string, error) {

	pollCtx, cancel := contextx.WithTimeout(nCtx, proxyInstallerPollingTimeout)
	defer cancel()

	statusPath := path.Clean(path.Join(event.InstallerWorkDir, installer.OfflinePkgRelPathData, installer.StatusFileName))
	dataPath := path.Clean(path.Join(event.InstallerWorkDir, installer.OfflinePkgRelPathData, installer.DataFileName))

	ticker := time.NewTicker(proxyInstallerPollingInterval)
	defer ticker.Stop()

	var outStr string
	for {
		select {
		case <-pollCtx.Done():
			return proxyInstallerPollingResult{}, outStr, fmt.Errorf("wait proxy installer status timed out: %w", pollCtx.Err())

		case <-ticker.C:
			result, output, completed, err := h.readProxyInstallerStatus(client, event, statusPath, dataPath)
			outStr += output
			if err != nil {
				return proxyInstallerPollingResult{}, outStr, err
			}
			if completed {
				return result, outStr, nil
			}
		}
	}
}

func (h *handler) readProxyInstallerStatus(
	client *sshx.Client,
	event *protoRelay.InstallProxyBySSHReq,
	statusPath string,
	dataPath string,
) (proxyInstallerPollingResult, string, bool, error) {

	statusContent, _, err := client.RunCommand(fmt.Sprintf("if [ -f %s ]; then cat %s; fi", statusPath, statusPath))
	if err != nil {
		return proxyInstallerPollingResult{}, "", false, fmt.Errorf("failed to read proxy installer status: %w", err)
	}

	statusContent = strings.TrimSpace(statusContent)
	if statusContent == "" {
		return proxyInstallerPollingResult{}, "", false, nil
	}

	var status proxyInstallerStatusFile
	statusReady := json.Unmarshal([]byte(statusContent), &status) == nil
	if !statusReady {
		return proxyInstallerPollingResult{}, "", false, nil
	}
	if status.OperInstID != event.OperInstID {
		return proxyInstallerPollingResult{}, "", false, nil
	}

	installerStatus := installer.ProcessState(status.Status)
	switch installerStatus {
	case installer.ProcessStateSuccess:
		agentID, err := h.readProxyInstallerAgentID(client, event, dataPath)
		if err != nil {
			return proxyInstallerPollingResult{}, "", false, err
		}

		return proxyInstallerPollingResult{Status: status.Status, AgentID: agentID},
			buildLogOutput("poll", installer.StatusFileName, statusContent, ""), true, nil

	case installer.ProcessStateFailed, installer.ProcessStateTimeout:
		return proxyInstallerPollingResult{Status: status.Status},
			buildLogOutput("poll", installer.StatusFileName, statusContent, ""), true, nil

	default:
		return proxyInstallerPollingResult{}, "", false, nil
	}
}

func (h *handler) readProxyInstallerAgentID(
	client *sshx.Client,
	event *protoRelay.InstallProxyBySSHReq,
	dataPath string,
) (string, error) {

	dataContent, _, err := client.RunCommand(fmt.Sprintf("cat %s 2>/dev/null", dataPath))
	if err != nil {
		return "", fmt.Errorf("failed to read proxy installer data: %w", err)
	}

	var data proxyInstallerDataFile
	if err := json.Unmarshal([]byte(strings.TrimSpace(dataContent)), &data); err != nil {
		return "", fmt.Errorf("failed to parse proxy installer data: %w", err)
	}
	if data.OperInstID != "" && data.OperInstID != event.OperInstID {
		return "", fmt.Errorf("proxy installer data belongs to another operation. oper-inst-id(%s)", data.OperInstID)
	}
	if data.AgentID == "" {
		return "", fmt.Errorf("proxy installer data agent_id is empty")
	}

	return data.AgentID, nil
}

func (h *handler) reportProxyInstallerStatus(
	nCtx contextx.IContext,
	event *protoRelay.InstallProxyBySSHReq,
	result proxyInstallerPollingResult,
) error {

	if err := h.reportProxyInstallerStatusField(nCtx, event, result.Status); err != nil {
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
