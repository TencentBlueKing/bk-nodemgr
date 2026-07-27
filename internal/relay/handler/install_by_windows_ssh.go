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
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf16"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/winpath"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/sshx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	windowsSSHProfileCygwin = "ssh_cygwin"
	windowsSSHProfileNative = "ssh_native"

	windowsNativeArchCommand    = "powershell -NoProfile -NonInteractive -ExecutionPolicy Bypass -Command \"$env:PROCESSOR_ARCHITECTURE\""
	windowsNativeProfileCommand = "powershell -NoProfile -NonInteractive -ExecutionPolicy Bypass -Command \"Write-Output native\""

	utf16BytesPerCodeUnit = 2
	utf16BitsPerByte      = 8
)

// InstallPagentByWindowsSSH installs the pagent by Windows SSH.
func (h *handler) InstallPagentByWindowsSSH(nCtx contextx.IContext, payload []byte) {
	logger.G.Biz(nCtx).Info("handler install pagent by windows ssh event")

	var (
		event  protoRelay.InstallPagentByWindowsSSHReq
		outStr string
		errMsg string
	)

	defer func() {
		if err := h.reportInstallResult(nCtx, event.ActionName, event.OperInstID, outStr, errMsg); err != nil {
			logger.G.Biz(nCtx).WithErr(err).Error("failed to report windows ssh install result")

			return
		}

		logger.G.Biz(nCtx).
			With("stdout", outStr, "ip", event.IP, "port", event.Port, "user", event.User).
			Info("done report install result by windows ssh")
	}()

	if err := json.Unmarshal(payload, &event); err != nil {
		errMsg = "failed to decode windows ssh install request"
		logger.G.Biz(nCtx).WithErr(err).Error(errMsg)

		return
	}

	client, err := generateWindowsSSHClient(
		nCtx,
		event.IP,
		int(event.Port),
		event.User,
		event.Password,
		types.LoginMode(event.LoginMode),
	)
	if err != nil {
		errMsg = "failed to create windows ssh client"
		logger.G.Biz(nCtx).WithErr(err).With("ip", event.IP, "port", event.Port, "user", event.User).Error(errMsg)

		return
	}
	defer func() { _ = client.Close() }()

	profile, profileOutput, err := detectWindowsSSHProfile(client)
	outStr += profileOutput
	if err != nil {
		errMsg = err.Error()
		logger.G.Biz(nCtx).WithErr(err).With("ip", event.IP, "port", event.Port, "user", event.User).Error(errMsg)

		return
	}

	installOutput, err := h.executeWindowsSSHInstall(nCtx, client, profile, &event)
	outStr += installOutput
	if err != nil {
		errMsg = err.Error()
		logger.G.Biz(nCtx).
			WithErr(err).
			With("ip", event.IP, "port", event.Port, "user", event.User, "profile", profile).
			Error("failed to install pagent by windows ssh")

		return
	}

	logger.G.Biz(nCtx).
		With("ip", event.IP, "port", event.Port, "user", event.User, "profile", profile).
		Info("install pagent by windows ssh successfully")
}

func generateWindowsSSHClient(
	ctx contextx.IContext,
	ip string,
	port int,
	user string,
	credential string,
	loginMode types.LoginMode,
) (*sshx.Client, error) {

	sshConf := &sshx.Config{
		Network: sshx.NetworkTCP,
		IP:      ip,
		Port:    port,
		User:    user,
		Ciphers: sshx.WindowsCompatibleCiphers(),
		MACs:    sshx.WindowsCompatibleMACs(),
	}

	switch loginMode {
	case types.LoginModePassword, types.LoginModePasswordVault:
		sshConf.AuthMethod = sshx.AuthMethodPassword
		sshConf.Password = credential
	case types.LoginModeKeyFile:
		sshConf.AuthMethod = sshx.AuthMethodPrivateKey
		sshConf.PrivateKey = []byte(credential)
	default:
		return nil, fmt.Errorf("unsupported login mode, mode(%s)", loginMode)
	}

	client, err := sshx.NewClient(ctx, sshConf, sshx.DefaultTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to windows ssh host: %w", err)
	}

	return client, nil
}

type windowsSSHCommandRunner interface {
	RunCommand(cmd string) (string, string, error)
}

func detectWindowsSSHProfile(runner windowsSSHCommandRunner) (string, string, error) {
	unameStdout, unameStderr, unameErr := runner.RunCommand("uname -s")
	output := buildLogOutput("detect", "windows ssh profile", unameStdout, unameStderr)
	if unameErr == nil && isCygwinUname(unameStdout) {
		return windowsSSHProfileCygwin, output, nil
	}

	nativeStdout, nativeStderr, nativeErr := runner.RunCommand(windowsNativeProfileCommand)
	output += buildLogOutput("validate", "windows ssh native profile", nativeStdout, nativeStderr)
	if nativeErr != nil {
		return "", output, errors.New("failed to detect windows ssh native profile")
	}

	return windowsSSHProfileNative, output, nil
}

func isCygwinUname(value string) bool {
	return strings.Contains(strings.ToUpper(strings.TrimSpace(value)), "CYGWIN")
}

func (h *handler) executeWindowsSSHInstall(
	nCtx contextx.IContext,
	client *sshx.Client,
	profile string,
	event *protoRelay.InstallPagentByWindowsSSHReq,
) (string, error) {

	workspaceOutput, err := ensureWindowsSSHWorkspace(client, profile, event.InstallerWorkDir)

	if err != nil {
		return workspaceOutput, err
	}

	if err := h.transferWindowsSSHManagedFile(nCtx, client, event.InstallerWorkDir, event.ToolsName); err != nil {
		return workspaceOutput, err
	}

	switch profile {
	case windowsSSHProfileNative:
		installOutput, installErr := executeNativeWindowsSSHInstall(client, event)
		return workspaceOutput + installOutput, installErr
	case windowsSSHProfileCygwin:
		installOutput, installErr := executeCygwinWindowsSSHInstall(client, event)
		return workspaceOutput + installOutput, installErr
	default:
		return workspaceOutput, fmt.Errorf("unsupported windows ssh profile, profile(%s)", profile)
	}
}

func ensureWindowsSSHWorkspace(client *sshx.Client, profile, workDir string) (string, error) {
	var command string
	var stageError string

	switch profile {
	case windowsSSHProfileNative:
		command = buildPowerShellCommand(fmt.Sprintf(
			"New-Item -ItemType Directory -Force -Path %s | Out-Null",
			powerShellSingleQuote(workDir),
		))
		stageError = "failed to prepare native windows workspace"
	case windowsSSHProfileCygwin:
		command = fmt.Sprintf(
			"mkdir -p %s",
			shellSingleQuote(winpath.ToSlash(winpath.Clean(workDir))),
		)
		stageError = "failed to prepare cygwin windows workspace"
	default:
		return "", fmt.Errorf("unsupported windows ssh profile, profile(%s)", profile)
	}

	stdout, stderr, err := runWindowsSSHCommand(client, command, stageError)

	return buildLogOutput("workspace", profile, stdout, stderr), err
}

func (h *handler) transferWindowsSSHManagedFile(
	nCtx contextx.IContext,
	client *sshx.Client,
	workDir string,
	fileName string,
) error {

	file, err := h.fileManager.GetFile(nCtx, fileName)

	if err != nil {
		return fmt.Errorf("failed to get windows ssh installer tool: %w", err)
	}

	reader, err := file.Content(nCtx)
	if err != nil {
		return fmt.Errorf("failed to read windows ssh installer tool: %w", err)
	}

	targetPath := winpath.Clean(winpath.Join(workDir, fileName))

	if err := client.TransferFile(reader, winpath.ToSlash(targetPath)); err != nil {
		return fmt.Errorf("failed to transfer windows ssh installer tool: %w", err)
	}

	return nil
}

func executeNativeWindowsSSHInstall(
	client *sshx.Client,
	event *protoRelay.InstallPagentByWindowsSSHReq,
) (string, error) {

	installBatPath := winpath.Clean(winpath.Join(event.InstallerWorkDir, event.InstallerBatName))

	if err := transferWindowsSSHScript(client, installBatPath, event.InstallerBatCmd); err != nil {
		return "", fmt.Errorf("failed to transfer native windows install script: %w", err)
	}

	command := buildNativeWindowsSSHInstallCommand(event.InstallerWorkDir, installBatPath)
	stdout, stderr, err := runWindowsSSHCommand(client, command, "failed to launch native windows installer")

	return buildLogOutput("launch", event.InstallerBatName, stdout, stderr), err
}

func executeCygwinWindowsSSHInstall(
	client *sshx.Client,
	event *protoRelay.InstallPagentByWindowsSSHReq,
) (string, error) {

	installShellPath := winpath.Clean(winpath.Join(event.InstallerWorkDir, event.InstallerShellName))

	if err := transferWindowsSSHScript(client, installShellPath, event.InstallerShellCmd); err != nil {
		return "", fmt.Errorf("failed to transfer cygwin windows install script: %w", err)
	}

	command := fmt.Sprintf(
		"cd %s && sh %s",
		shellSingleQuote(winpath.ToSlash(winpath.Clean(event.InstallerWorkDir))),
		shellSingleQuote(event.InstallerShellName),
	)
	stdout, stderr, err := runWindowsSSHCommand(client, command, "failed to launch cygwin windows installer")

	return buildLogOutput("launch", event.InstallerShellName, stdout, stderr), err
}

func transferWindowsSSHScript(client *sshx.Client, targetPath, script string) error {
	reader := io.NopCloser(strings.NewReader(script))
	if err := client.TransferFile(reader, winpath.ToSlash(targetPath)); err != nil {
		return err
	}

	return nil
}

func runWindowsSSHCommand(client *sshx.Client, command, stageError string) (string, string, error) {
	stdout, stderr, err := client.RunCommand(command)
	if err != nil {
		return stdout, stderr, errors.New(stageError)
	}

	return stdout, stderr, nil
}

func buildNativeWindowsSSHInstallCommand(workDir, installBatPath string) string {
	argumentList := fmt.Sprintf("/c \"%s\"", strings.ReplaceAll(installBatPath, "\"", "\"\""))

	return buildPowerShellCommand(fmt.Sprintf(
		"Start-Process -FilePath %s -ArgumentList %s -WorkingDirectory %s",
		powerShellSingleQuote("cmd.exe"),
		powerShellSingleQuote(argumentList),
		powerShellSingleQuote(workDir),
	))
}

func buildPowerShellCommand(script string) string {
	return "powershell.exe -NoProfile -NonInteractive -ExecutionPolicy Bypass -EncodedCommand " +
		encodePowerShellCommand(script)
}

func encodePowerShellCommand(script string) string {
	codeUnits := utf16.Encode([]rune(script))
	encoded := make([]byte, len(codeUnits)*utf16BytesPerCodeUnit)
	for index, codeUnit := range codeUnits {
		encoded[index*utf16BytesPerCodeUnit] = byte(codeUnit)
		encoded[index*utf16BytesPerCodeUnit+1] = byte(codeUnit >> utf16BitsPerByte)
	}

	return base64.StdEncoding.EncodeToString(encoded)
}

func powerShellSingleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func shellSingleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
