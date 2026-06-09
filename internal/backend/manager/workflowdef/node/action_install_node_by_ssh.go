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
	"errors"
	"fmt"
	"io"
	"path"
	"time"

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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/sshx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameInstallNodeBySSH defines the action name.
	ActionNameInstallNodeBySSH = "install_node_by_ssh"
)

// NewActionInstallNodeBySSH get a new action.
func NewActionInstallNodeBySSH(capability *Capability) action.Definition {
	return &actionInstallNodeBySSH{
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

// ActParamInstallAgentBySSH ...
type ActParamInstallAgentBySSH struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionInstallNodeBySSH struct {
	fileHandler file.IHandler
	fileCache   filecache.IFileCache

	storageHostCredit     credit.IStorageHostCredit
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	storageActionInstance workflow.IStorageActionInstance
	provider              discover.Provider
	passwordVault         creditvault.IHostPasswordVault
}

// Name returns the name of the action.
func (act *actionInstallNodeBySSH) Name() string {
	return ActionNameInstallNodeBySSH
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionInstallNodeBySSH) DisplayNameZh() string {
	return "通过 SSH 安装节点"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionInstallNodeBySSH) DisplayNameEn() string {
	return "Install Node via SSH"
}

// Version returns the version of the action.
func (act *actionInstallNodeBySSH) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionInstallNodeBySSH) Description() string {
	return "Use ssh to connect to the target machine, transfer files through sftp, and execute the installation command"
}

// Timeout returns the timeout of the action.
func (act *actionInstallNodeBySSH) Timeout() time.Duration {
	return 5 * time.Minute // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actionInstallNodeBySSH) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionInstallNodeBySSH) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionInstallNodeBySSH) DelayFn() func() {
	return func() {
		time.Sleep(5 * time.Second) // nolint: mnd
	}
}

// Do this func define what the action will do.
// To ensure readability, this action uses fmt.Sprintf to concatenate characters.
// nolint: perfsprint,funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionInstallNodeBySSH) Do(ctx *action.InstanceContext) error {
	param := new(ActParamInstallAgentBySSH)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := nodeUtils.NewNodeActionStandarder(act.storageNodeDeployment, act.storageHost)
	if err = std.Initialize(ctx, param.NodeActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	// let the callback server known which action to mark and log.
	if err := std.SaveBlockingActionName(ActionNameWaitInstallerComplete); err != nil {
		return fmt.Errorf("failed to save blocking action name: %w", err)
	}

	// get ssh credit.
	credit := nodeUtils.NewCreditHandler(act.storageHostCredit, act.passwordVault)
	cMethod, cKey, err := credit.GetSSHCredit(std)
	if err != nil {
		return fmt.Errorf("failed to get ssh credit: %w", err)
	}

	// generate the ssh client.
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
	}, sshx.DefaultTimeout)
	if err != nil {
		return fmt.Errorf("failed to generate new ssh client: %w", err)
	}

	// ensure the workspace dir.
	if err = act.ensureWorkspace(std, client); err != nil {
		return fmt.Errorf("failed to ensure workspace through ssh: %w", err)
	}

	// ensure the installer tool.
	installerPath, err := act.ensureInstallerTool(std, client)
	if err != nil {
		return fmt.Errorf("failed to ensure installer tool through ssh: %w", err)
	}

	// normal install.
	if err := act.executeInstallCMD(std, client, installerPath); err != nil {
		return fmt.Errorf("failed to execute install cmd: %w", err)
	}

	if err = saveWaitInstallerPrivateData(std.Context(), act.storageActionInstance,
		std.InstanceData().OperationInstanceID, false, true); err != nil {
		return err
	}

	return nil
}

func (act *actionInstallNodeBySSH) ensureWorkspace(std *nodeUtils.NodeActionStandarder, client *sshx.Client) error {
	if _, stderr, err := client.RunCommand("mkdir -p " + std.DeployInfo().InstallerRuntime.WorkDir); err != nil {
		err = fmt.Errorf("failed to run command. command(mkdir -p %s), stderr(%s): %w", std.DeployInfo().InstallerRuntime.WorkDir, stderr, err)

		return err
	}

	return nil
}

func (act *actionInstallNodeBySSH) ensureInstallerTool(std *nodeUtils.NodeActionStandarder, client *sshx.Client) (string, error) {
	// select matching tools, and use sftp to transfer it.
	toolName, err := tool.FormatInstallerName(std.DeployInfo().Host.Dynamic.NodeOsType, std.DeployInfo().Host.Dynamic.NodeCPUArch)
	if err != nil {
		return "", fmt.Errorf("failed to format installer tool name: %w", err)
	}

	toolReader, err := act.openInstallerReader(std)
	if err != nil {
		return "", err
	}

	// TransferFile accepts io.ReadCloser and closes it via its own defer,
	// so we must not close toolReader here to avoid a double-close.
	installerPath := path.Clean(path.Join(std.DeployInfo().InstallerRuntime.WorkDir, toolName))
	if err = client.TransferFile(toolReader, installerPath); err != nil {
		return "", fmt.Errorf("failed to transfer installer tool to host: %w", err)
	}

	std.InstanceData().Log().
		Zh("已传输安装器到主机, 路径(%s)", installerPath).
		En("transferred installer to host, path(%s)", installerPath).
		Info()

	// make sure tool is executable
	if _, stderr, err := client.RunCommand("chmod +x " + installerPath); err != nil {
		return "", fmt.Errorf("failed to run command. command(chmod +x %s), stderr(%s): %w", installerPath, stderr, err)
	}

	return installerPath, nil
}

// openInstallerReader returns a reader for the installer binary, using the local
// file cache when available to avoid redundant downloads across workflow retries.
func (act *actionInstallNodeBySSH) openInstallerReader(std *nodeUtils.NodeActionStandarder) (io.ReadCloser, error) {
	osType := std.DeployInfo().Host.Dynamic.NodeOsType
	cpuArch := std.DeployInfo().Host.Dynamic.NodeCPUArch

	if act.fileCache == nil {
		toolFile, err := act.fileHandler.DownloadInstaller(std.Context(), osType, cpuArch)
		if err != nil {
			return nil, fmt.Errorf("failed to download installer from file service: %w", err)
		}

		return toolFile.Data, nil
	}

	// Query file metadata first to enable cache lookup by MD5.
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

func (act *actionInstallNodeBySSH) executeInstallCMD(std *nodeUtils.NodeActionStandarder, client *sshx.Client, installerPath string) error {
	selector := discover.NewRoundRobinSelector()
	downloadEndpoints, err := act.provider.SelectEndpoints(
		discover.ServiceNameFile,
		discover.EndpointNameFileDownload,
		nodeUtils.DefaultEndpointSelectionCount,
		selector)
	if err != nil {
		return fmt.Errorf("failed to select file endpoints: %w", err)
	}

	callbackEndpoints, err := act.provider.SelectEndpoints(
		discover.ServiceNameBackend,
		discover.EndpointNameBackendCallback,
		nodeUtils.DefaultEndpointSelectionCount,
		selector)
	if err != nil {
		return fmt.Errorf("failed to select backend callback endpoints: %w", err)
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

	installCmd, err := act.buildCMD(installParams)
	if err != nil {
		return fmt.Errorf("failed to build install cmd: %w", err)
	}
	std.InstanceData().Log().
		Zh("安装节点命令: %s", installCmd).
		En("install node cmd: %s", installCmd).
		Info()

	// exec install command.
	outStr, stderr, err := client.RunCommand(fmt.Sprintf(
		`mkdir -p %s && cd %s && echo "%s" > install.sh && sh install.sh`,
		std.DeployInfo().InstallerRuntime.WorkDir,
		std.DeployInfo().InstallerRuntime.WorkDir,
		installCmd),
	)
	if err != nil {
		err = fmt.Errorf("failed to run install node: %w", err)

		return err
	}

	std.InstanceData().Log().
		Zh("安装指令结果: stdout(%s), stderr(%s)", outStr, stderr).
		En("install command result: stdout(%s), stderr(%s)", outStr, stderr).
		Info()

	return nil
}

func (act *actionInstallNodeBySSH) buildCMD(param *installer.NodeInstallParams) (string, error) {
	_, installCmd, err := param.ToUnixScript()
	if err != nil {
		return "", fmt.Errorf("failed to render node install script: %w", err)
	}

	return installCmd, nil
}
