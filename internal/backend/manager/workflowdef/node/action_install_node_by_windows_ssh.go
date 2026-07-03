/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package node

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf16"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filecache"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/winpath"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/sshx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameInstallNodeByWindowsSSH defines the action name.
	ActionNameInstallNodeByWindowsSSH = "install_node_by_windows_ssh"

	installShellName = "install.sh"
)

const (
	utf16BytesPerCodeUnit = 2
	utf16BitsPerByte      = 8
)

// NewActionInstallNodeByWindowsSSH get a new action.
func NewActionInstallNodeByWindowsSSH(capability *Capability) action.Definition {
	return &actionInstallNodeByWindowsSSH{
		fileHandler:           capability.FileHandler,
		fileCache:             capability.FileCache,
		storageHostCredit:     capability.StorageHostCredit,
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		provider:              capability.DiscoverProvider,
		passwordVault:         capability.HostPasswordVault,
		storageActionInstance: capability.StorageWorkflow,
	}
}

// ActParamInstallAgentByWindowsSSH defines the action parameter.
type ActParamInstallAgentByWindowsSSH struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionInstallNodeByWindowsSSH struct {
	fileHandler file.IHandler
	fileCache   filecache.IFileCache

	storageHostCredit     credit.IStorageHostCredit
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	provider              discover.Provider
	passwordVault         creditvault.IHostPasswordVault
	storageActionInstance workflow.IStorageActionInstance
}

// Name returns the name of the action.
func (act *actionInstallNodeByWindowsSSH) Name() string {
	return ActionNameInstallNodeByWindowsSSH
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionInstallNodeByWindowsSSH) DisplayNameZh() string {
	return "通过 Windows SSH 安装节点"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionInstallNodeByWindowsSSH) DisplayNameEn() string {
	return "Install Node via Windows SSH"
}

// Version returns the version of the action.
func (act *actionInstallNodeByWindowsSSH) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionInstallNodeByWindowsSSH) Description() string {
	return "Use Windows SSH to connect to the target machine, transfer files through sftp, and execute the installation command"
}

// Timeout returns the timeout of the action.
func (act *actionInstallNodeByWindowsSSH) Timeout() time.Duration {
	return 5 * time.Minute // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actionInstallNodeByWindowsSSH) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionInstallNodeByWindowsSSH) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionInstallNodeByWindowsSSH) DelayFn() func() {
	return func() {
		time.Sleep(5 * time.Second) // nolint: mnd
	}
}

// Do this func define what the action will do.
// nolint: perfsprint,funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionInstallNodeByWindowsSSH) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamInstallAgentByWindowsSSH)
	err = conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	std := nodeUtils.NewNodeActionStandarder(act.storageNodeDeployment, act.storageHost)
	if err = std.Initialize(ctx, param.NodeActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	if err := std.SaveBlockingActionName(ActionNameWaitInstallerComplete); err != nil {
		return fmt.Errorf("failed to save blocking action name: %w", err)
	}

	client, err := act.newWindowsSSHClient(std)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	profile, err := detectWindowsSSHProfile(client)
	if err != nil {
		return fmt.Errorf("failed to detect windows ssh profile: %w", err)
	}

	launchCmd, err := act.prepareInstallCommand(std, client, profile)
	if err != nil {
		return err
	}

	if err = saveWaitInstallerPrivateData(
		std.Context(), act.storageActionInstance, std.InstanceData().OperationInstanceID,
		false, true,
	); err != nil {
		return err
	}

	stdout, stderr, err := client.RunCommand(launchCmd)
	if err != nil {
		return fmt.Errorf("failed to start install node, stdout(%s), stderr(%s): %w", stdout, stderr, err)
	}

	std.InstanceData().Log().
		Zh("已启动节点安装, stdout(%s), stderr(%s)",
			strings.Split(strings.TrimSpace(stdout), "\n"), strings.Split(strings.TrimSpace(stderr), "\n")).
		En("started node install, stdout(%s), stderr(%s)",
			strings.Split(strings.TrimSpace(stdout), "\n"), strings.Split(strings.TrimSpace(stderr), "\n")).
		Info()

	return nil
}

func (act *actionInstallNodeByWindowsSSH) newWindowsSSHClient(std *nodeUtils.NodeActionStandarder) (*sshx.Client, error) {
	credit := nodeUtils.NewCreditHandler(act.storageHostCredit, act.passwordVault)
	cMethod, cKey, err := credit.GetSSHCredit(std)
	if err != nil {
		return nil, fmt.Errorf("failed to get ssh credit: %w", err)
	}

	client, err := sshx.NewClient(std.Context(), &sshx.Config{
		Network:    sshx.NetworkTCP,
		IP:         std.DeployInfo().Host.Dynamic.LoginIP,
		Port:       int(std.DeployInfo().Host.Dynamic.LoginPort),
		User:       std.DeployInfo().Host.Dynamic.LoginUser,
		AuthMethod: cMethod,
		Password: func() string {
			if cMethod == sshx.AuthMethodPassword {
				return cKey
			}

			return ""
		}(),
		PrivateKey: func() []byte {
			if cMethod == sshx.AuthMethodPrivateKey {
				return []byte(cKey)
			}

			return nil
		}(),
		Ciphers: sshx.WindowsCompatibleCiphers(),
		MACs:    sshx.WindowsCompatibleMACs(),
	}, sshx.DefaultTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new ssh client: %w", err)
	}

	return client, nil
}

func (act *actionInstallNodeByWindowsSSH) ensureWorkspace(std *nodeUtils.NodeActionStandarder, client *sshx.Client) error {
	workDir := std.DeployInfo().InstallerRuntime.WorkDir
	command := buildPowerShellCommand(fmt.Sprintf(
		"New-Item -ItemType Directory -Force -Path %s | Out-Null",
		powerShellSingleQuote(workDir),
	))

	stdout, stderr, err := client.RunCommand(command)
	if err != nil {
		return fmt.Errorf("failed to run command. command(%s), stdout(%s), stderr(%s): %w", command, stdout, stderr, err)
	}

	std.InstanceData().Log().
		Zh("确保安装器工作目录存在, stdout(%s), stderr(%s)",
			strings.Split(strings.TrimSpace(stdout), "\n"), strings.Split(strings.TrimSpace(stderr), "\n")).
		En("make sure the installer workspace exists, stdout(%s), stderr(%s)",
			strings.Split(strings.TrimSpace(stdout), "\n"), strings.Split(strings.TrimSpace(stderr), "\n")).
		Info()

	return nil
}

func (act *actionInstallNodeByWindowsSSH) ensureInstallerTool(std *nodeUtils.NodeActionStandarder, client *sshx.Client) (string, error) {
	toolName, err := tool.FormatInstallerName(std.DeployInfo().Host.Dynamic.NodeOsType, std.DeployInfo().Host.Dynamic.NodeCPUArch)
	if err != nil {
		return "", fmt.Errorf("failed to format installer tool name: %w", err)
	}

	content, err := act.openInstallerReader(std)
	if err != nil {
		return "", err
	}

	installerPath := winpath.Clean(winpath.Join(std.DeployInfo().InstallerRuntime.WorkDir, toolName))
	if err = client.TransferFile(content, windowsSSHTransferPath(installerPath)); err != nil {
		return "", fmt.Errorf("failed to transfer installer tool to host: %w", err)
	}

	std.InstanceData().Log().
		Zh("已传输安装器到主机, 路径(%s)", installerPath).
		En("transferred installer to host, path(%s)", installerPath).
		Info()

	return installerPath, nil
}

// openInstallerReader returns a reader for the installer binary, using the local
// file cache when available to avoid redundant downloads across workflow retries.
func (act *actionInstallNodeByWindowsSSH) openInstallerReader(std *nodeUtils.NodeActionStandarder) (io.ReadCloser, error) {
	osType := std.DeployInfo().Host.Dynamic.NodeOsType
	cpuArch := std.DeployInfo().Host.Dynamic.NodeCPUArch

	if act.fileCache == nil {
		toolFile, err := act.fileHandler.DownloadInstaller(std.Context(), osType, cpuArch)
		if err != nil {
			return nil, fmt.Errorf("failed to download installer from file service: %w", err)
		}

		return toolFile.Data, nil
	}

	fileInfo, err := act.fileHandler.InfoInstaller(std.Context(), osType, cpuArch)
	if err != nil {
		return nil, fmt.Errorf("failed to get installer info from file service: %w", err)
	}

	cachedFile, _, err := act.fileCache.GetOrFetch(std.Context(), fileInfo.Name, fileInfo.MD5,
		func(nCtx contextx.IContext) (io.ReadCloser, error) {
			resp, dlErr := act.fileHandler.DownloadInstaller(nCtx, osType, cpuArch)
			if dlErr != nil {
				return nil, dlErr
			}

			return resp.Data, nil
		})
	if err != nil {
		return nil, fmt.Errorf("failed to get installer from file cache: %w", err)
	}

	reader, err := cachedFile.Content(std.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to open cached installer content: %w", err)
	}

	return reader, nil
}

func (act *actionInstallNodeByWindowsSSH) prepareInstallCommand(
	std *nodeUtils.NodeActionStandarder, client *sshx.Client, profile string,
) (string, error) {

	switch profile {
	case windowsSSHProfileNative:
		if err := act.ensureWorkspace(std, client); err != nil {
			return "", fmt.Errorf("failed to ensure workspace through windows ssh: %w", err)
		}

		installerPath, err := act.ensureInstallerTool(std, client)
		if err != nil {
			return "", fmt.Errorf("failed to ensure installer tool through windows ssh: %w", err)
		}

		return act.prepareNativeInstallCommand(std, client, installerPath)
	case windowsSSHProfileCygwin:
		if err := act.ensureCygwinWorkspace(std, client); err != nil {
			return "", fmt.Errorf("failed to ensure cygwin workspace through windows ssh: %w", err)
		}

		installerPath, err := act.ensureInstallerTool(std, client)
		if err != nil {
			return "", fmt.Errorf("failed to ensure installer tool through windows ssh: %w", err)
		}

		return act.prepareCygwinInstallCommand(std, client, installerPath)
	default:
		return "", fmt.Errorf("unsupported windows ssh profile: %s", profile)
	}
}

func (act *actionInstallNodeByWindowsSSH) ensureCygwinWorkspace(
	std *nodeUtils.NodeActionStandarder, client *sshx.Client,
) error {

	workDir := winpath.ToSlash(winpath.Clean(std.DeployInfo().InstallerRuntime.WorkDir))
	command := fmt.Sprintf("mkdir -p %s", shellDoubleQuote(workDir))

	stdout, stderr, err := client.RunCommand(command)
	if err != nil {
		return fmt.Errorf("failed to run command. command(%s), stdout(%s), stderr(%s): %w", command, stdout, stderr, err)
	}

	std.InstanceData().Log().
		Zh("确保 Cygwin 安装器工作目录存在, stdout(%s), stderr(%s)",
			strings.Split(strings.TrimSpace(stdout), "\n"), strings.Split(strings.TrimSpace(stderr), "\n")).
		En("make sure the cygwin installer workspace exists, stdout(%s), stderr(%s)",
			strings.Split(strings.TrimSpace(stdout), "\n"), strings.Split(strings.TrimSpace(stderr), "\n")).
		Info()

	return nil
}

func (act *actionInstallNodeByWindowsSSH) prepareNativeInstallCommand(
	std *nodeUtils.NodeActionStandarder, client *sshx.Client, installerPath string,
) (string, error) {

	installParams, err := act.buildNodeInstallParams(std, installerPath)
	if err != nil {
		return "", err
	}

	installBat, err := act.buildBat(installParams)
	if err != nil {
		return "", fmt.Errorf("failed to build install bat: %w", err)
	}
	std.InstanceData().Log().
		Zh("安装节点命令: %s", installBat).
		En("install node cmd: %s", installBat).
		Info()

	installBatPath := winpath.Clean(winpath.Join(std.DeployInfo().InstallerRuntime.WorkDir, installBatName))
	if err := client.TransferFile(io.NopCloser(strings.NewReader(installBat)), windowsSSHTransferPath(installBatPath)); err != nil {
		return "", fmt.Errorf("failed to transfer bat file for windows ssh execution: %w", err)
	}

	return buildWindowsSSHNativeInstallCommand(std.DeployInfo().InstallerRuntime.WorkDir, installBatPath), nil
}

func (act *actionInstallNodeByWindowsSSH) prepareCygwinInstallCommand(
	std *nodeUtils.NodeActionStandarder, client *sshx.Client, installerPath string,
) (string, error) {

	installParams, err := act.buildNodeInstallParams(std, installerPath)
	if err != nil {
		return "", err
	}

	installShell, err := act.buildShell(installParams)
	if err != nil {
		return "", fmt.Errorf("failed to build install shell: %w", err)
	}
	std.InstanceData().Log().
		Zh("Cygwin 安装节点命令: %s", installShell).
		En("cygwin install node cmd: %s", installShell).
		Info()

	installShellPath := winpath.Clean(winpath.Join(std.DeployInfo().InstallerRuntime.WorkDir, installShellName))
	if err := client.TransferFile(io.NopCloser(strings.NewReader(installShell)), windowsSSHTransferPath(installShellPath)); err != nil {
		return "", fmt.Errorf("failed to transfer shell file for cygwin windows ssh execution: %w", err)
	}

	return buildWindowsSSHCygwinInstallCommand(std.DeployInfo().InstallerRuntime.WorkDir), nil
}

func (act *actionInstallNodeByWindowsSSH) buildNodeInstallParams(
	std *nodeUtils.NodeActionStandarder, installerPath string,
) (*installer.NodeInstallParams, error) {

	endpointSource := nodeUtils.SelectInstallEndpointSource(std)
	callbackEndpoints, downloadEndpoints, err := nodeUtils.GenerateNodeInstallerServerEndpoints(std, act.provider, endpointSource)
	if err != nil {
		return nil, fmt.Errorf("failed to generate node installer server endpoints: %w", err)
	}

	installParams := &installer.NodeInstallParams{
		NodeCommonParams: installer.NodeCommonParams{
			DeployEnv:     system.GetEnv(),
			Generation:    int(std.DeployInfo().Host.Dynamic.NodeGeneration),
			NodeRole:      string(std.DeployInfo().Host.Dynamic.NodeRole),
			BaseWorkDir:   std.DeployInfo().InstallerRuntime.BaseWorkDir,
			BaseDeployDir: std.DeployInfo().BaseRuntime.BaseDeployDir,
		},
		InstallerPath:   installerPath,
		DownloadSvrAddr: nodeUtils.BuildServerURLs(downloadEndpoints...),
		CallbackSvrAddr: nodeUtils.BuildServerURLs(callbackEndpoints...),
		DeployToken:     std.Token(),
		NodeVersion:     std.DeployInfo().Host.Dynamic.NodeVersion,
		OperInstID:      std.InstanceData().OperationInstanceID,
		RenewGSEProc:    std.DeployInfo().InstallOptions.RenewGSEProc,
		RenewGSETask:    std.DeployInfo().InstallOptions.RenewGSETask,
	}

	if !std.DeployInfo().InstallOptions.ReRegister && std.DeployInfo().Host.Dynamic.AgentID != "" {
		installParams.AdditionArgs = append(installParams.AdditionArgs,
			fmt.Sprintf("--agent_id %s", std.DeployInfo().Host.Dynamic.AgentID))
	}

	return installParams, nil
}

func (act *actionInstallNodeByWindowsSSH) buildBat(param *installer.NodeInstallParams) (string, error) {
	_, installCmd, err := param.ToWindowsScript()
	if err != nil {
		return "", fmt.Errorf("failed to render node install script: %w", err)
	}

	return installCmd, nil
}

func (act *actionInstallNodeByWindowsSSH) buildShell(param *installer.NodeInstallParams) (string, error) {
	_, installCmd, err := param.ToWindowsShellScript()
	if err != nil {
		return "", fmt.Errorf("failed to render node install shell script: %w", err)
	}

	return installCmd, nil
}

func windowsSSHTransferPath(winPath string) string {
	// Keep Windows paths as the canonical representation inside the workflow.
	// Adapt to slash form only at the SFTP boundary because sshx.TransferFile
	// uses POSIX path helpers to derive the destination directory.
	return winpath.ToSlash(winpath.Clean(winPath))
}

func buildWindowsSSHNativeInstallCommand(workDir string, installBatPath string) string {
	argumentList := fmt.Sprintf("/c \"%s\"", strings.ReplaceAll(installBatPath, "\"", "\"\""))

	return buildPowerShellCommand(fmt.Sprintf(
		"Start-Process -FilePath %s -ArgumentList %s -WorkingDirectory %s",
		powerShellSingleQuote("cmd.exe"),
		powerShellSingleQuote(argumentList),
		powerShellSingleQuote(workDir),
	))
}

func buildWindowsSSHCygwinInstallCommand(workDir string) string {
	return fmt.Sprintf(
		"cd %s && sh %s",
		shellDoubleQuote(winpath.ToSlash(winpath.Clean(workDir))),
		shellDoubleQuote(installShellName),
	)
}

func buildPowerShellCommand(script string) string {
	return "powershell.exe -NoProfile -NonInteractive -ExecutionPolicy Bypass -EncodedCommand " +
		encodePowerShellCommand(script)
}

func encodePowerShellCommand(script string) string {
	codeUnits := utf16.Encode([]rune(script))
	encoded := make([]byte, len(codeUnits)*utf16BytesPerCodeUnit)
	for i, codeUnit := range codeUnits {
		encoded[i*utf16BytesPerCodeUnit] = byte(codeUnit)
		encoded[i*utf16BytesPerCodeUnit+1] = byte(codeUnit >> utf16BitsPerByte)
	}

	return base64.StdEncoding.EncodeToString(encoded)
}

func powerShellSingleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func shellDoubleQuote(value string) string {
	return "\"" + strings.ReplaceAll(value, "\"", "\\\"") + "\""
}
