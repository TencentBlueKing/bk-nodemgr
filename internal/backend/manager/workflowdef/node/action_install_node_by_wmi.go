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

package node

import (
	"errors"
	"fmt"
	"io"
	"strings"
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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tmp"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/winpath"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/wmix"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameInstallNodeByWMI defines the action name.
	ActionNameInstallNodeByWMI = "install_node_by_wmi"
)

// NewActionInstallNodeByWMI get a new action.
func NewActionInstallNodeByWMI(capability *Capability) action.Definition {
	return &actionInstallNodeByWMI{
		fileHandler:           capability.FileHandler,
		fileCache:             capability.FileCache,
		storageHostCredit:     capability.StorageHostCredit,
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		storageNetworkUnit:    capability.StorageTopo,
		provider:              capability.DiscoverProvider,
		passwordVault:         capability.HostPasswordVault,
		storageActionInstance: capability.StorageWorkflow,
	}
}

// ActParamInstallAgentByWMI ...
type ActParamInstallAgentByWMI struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionInstallNodeByWMI struct {
	fileHandler           file.IHandler
	fileCache             filecache.IFileCache
	storageHostCredit     credit.IStorageHostCredit
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	storageNetworkUnit    topoStg.IStorageNetworkUnit
	provider              discover.IProvider
	passwordVault         creditvault.IHostPasswordVault
	storageActionInstance workflow.IStorageActionInstance
}

// Name returns the name of the action.
func (act *actionInstallNodeByWMI) Name() string {
	return ActionNameInstallNodeByWMI
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionInstallNodeByWMI) DisplayNameZh() string {
	return "通过 WMI 安装节点"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionInstallNodeByWMI) DisplayNameEn() string {
	return "Install Node via WMI"
}

// Version returns the version of the action.
func (act *actionInstallNodeByWMI) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionInstallNodeByWMI) Description() string {
	return "Use wmi to connect to the target machine, transfer files through sftp, and execute the installation command"
}

// Timeout returns the timeout of the action.
func (act *actionInstallNodeByWMI) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionInstallNodeByWMI) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionInstallNodeByWMI) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionInstallNodeByWMI) DelayFn(_ int) func() {
	return func() {
		time.Sleep(5 * time.Second) // nolint: mnd
	}
}

const installBatName = "install.bat"

// Do this func define what the action will do.
// To ensure readability, this action uses fmt.Sprintf to concatenate characters.
// nolint: perfsprint,funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionInstallNodeByWMI) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamInstallAgentByWMI)
	err = conv.MapToStruct(ctx.Data.Content, param)
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

	if err = act.installByWMI(std); err != nil {
		return err
	}

	err = saveWaitInstallerPrivateData(
		std.Context(),
		act.storageActionInstance,
		std.InstanceData().OperationInstanceID,
		false, true)
	if err != nil {
		return err
	}

	return nil
}

func (act *actionInstallNodeByWMI) installByWMI(std *nodeUtils.NodeActionStandarder) error {
	// get wmi credit.
	credit := nodeUtils.NewCreditHandler(act.storageHostCredit, act.passwordVault)
	cMethod, cKey, err := credit.GetWMICredit(std)
	if err != nil {
		return fmt.Errorf("failed to get wmi credit: %w", err)
	}

	// generate the wmi client.
	client, err := wmix.NewClient(&wmix.Config{
		IP:         std.DeployInfo().Host.Dynamic.LoginIP,
		User:       std.DeployInfo().Host.Dynamic.LoginUser,
		AuthMethod: cMethod,
		Password: func() string {
			if cMethod == wmix.AuthMethodPassword {
				return cKey
			}

			return ""
		}(),
		Timeout: wmix.DefaultTimeout,
	})
	if err != nil {
		return fmt.Errorf("failed to generate new wmi client: %w", err)
	}

	// ensure the workspace dir.
	if err = act.ensureWorkspace(std, client); err != nil {
		return fmt.Errorf("failed to ensure workspace through wmi: %w", err)
	}

	// ensure the installer tool.
	installerPath, err := act.ensureInstallerTool(std, client)
	if err != nil {
		return fmt.Errorf("failed to ensure installer tool through wmi: %w", err)
	}

	// execute install cmd.
	if err := act.executeInstallCMD(std, client, installerPath); err != nil {
		return fmt.Errorf("failed to execute install cmd: %w", err)
	}

	return nil
}

func (act *actionInstallNodeByWMI) ensureWorkspace(std *nodeUtils.NodeActionStandarder, client *wmix.Client) error {
	stdout, stderr, err := client.RunCommand(std.Context(), "mkdir "+std.DeployInfo().InstallerRuntime.WorkDir)
	if err != nil {
		return fmt.Errorf("failed to run command. command(mkdir %s), stdout(%s), stderr(%s): %w",
			std.DeployInfo().InstallerRuntime.WorkDir, stdout, stderr, err)
	}

	std.InstanceData().Log().
		Zh("确保安装器工作目录存在, stdout(%s), stderr(%s)",
			strings.Split(strings.TrimSpace(stdout), "\n"), strings.Split(strings.TrimSpace(stderr), "\n")).
		En("make sure the installer workspace exists, stdout(%s), stderr(%s)",
			strings.Split(strings.TrimSpace(stdout), "\n"), strings.Split(strings.TrimSpace(stderr), "\n")).
		Info()

	return nil
}

func (act *actionInstallNodeByWMI) ensureInstallerTool(std *nodeUtils.NodeActionStandarder, client *wmix.Client) (string, error) {
	toolName, err := tool.FormatInstallerName(std.DeployInfo().Host.Dynamic.NodeOsType, std.DeployInfo().Host.Dynamic.NodeCPUArch)
	if err != nil {
		return "", fmt.Errorf("failed to format installer tool name: %w", err)
	}

	content, err := act.openInstallerReader(std)
	if err != nil {
		return "", err
	}

	// NewTempFileWithSpecialName reads from content and closes it internally;
	// do not close content again to avoid a double-close.
	tmpInstallerFile, err := tmp.NewTempFileWithSpecialName(content, toolName)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file for installer: %w", err)
	}
	defer func() {
		if cleanErr := tmpInstallerFile.CleanUp(); cleanErr != nil {
			std.InstanceData().Log().
				Zh("清理临时文件失败: %v", cleanErr).
				En("failed to clean temp file: %v", cleanErr).
				Error()
		}
	}()

	stdout, stderr, err := client.UploadFile(std.Context(), tmpInstallerFile.Path(), std.DeployInfo().InstallerRuntime.WorkDir)
	if err != nil {
		return "", fmt.Errorf("failed to transfer installer tool to host, stdout(%s), stderr(%s): %w",
			stdout, stderr, err)
	}

	installerPath := winpath.Clean(winpath.Join(std.DeployInfo().InstallerRuntime.WorkDir, toolName))
	std.InstanceData().Log().
		Zh("已传输安装器到主机, 路径(%s)", installerPath).
		En("transferred installer to host, path(%s)", installerPath).
		Info()

	return installerPath, nil
}

// openInstallerReader returns a reader for the installer binary, using the local
// file cache when available to avoid redundant downloads across workflow retries.
func (act *actionInstallNodeByWMI) openInstallerReader(std *nodeUtils.NodeActionStandarder) (io.ReadCloser, error) {
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

func (act *actionInstallNodeByWMI) executeInstallCMD(std *nodeUtils.NodeActionStandarder, client *wmix.Client, installerPath string) error {
	endpointSource := nodeUtils.SelectInstallEndpointSource(std)
	callbackEndpoints, downloadEndpoints, err := nodeUtils.GenerateNodeInstallerServerEndpoints(std, act.provider, endpointSource)
	if err != nil {
		return fmt.Errorf("failed to generate node installer server endpoints: %w", err)
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

	installBat, err := act.buildBat(installParams)
	if err != nil {
		return fmt.Errorf("failed to build install bat: %w", err)
	}
	std.InstanceData().Log().
		Zh("安装节点命令: %s", installBat).
		En("install node cmd: %s", installBat).
		Info()

	// exec install command.
	tmpInstallBat, err := tmp.NewTempFileWithSpecialName(io.NopCloser(strings.NewReader(installBat)), installBatName)
	if err != nil {
		return fmt.Errorf("failed to create temp bat file for wmi execution: %w", err)
	}
	defer func() {
		if err := tmp.Clean(); err != nil {
			std.InstanceData().Log().
				Zh("清理临时文件失败: %v", err).
				En("failed to clean temp file: %v", err).
				Error()
		}
	}()

	_, _, err = client.UploadFile(std.Context(), tmpInstallBat.Path(), std.DeployInfo().InstallerRuntime.WorkDir)
	if err != nil {
		return fmt.Errorf("failed to transfer bat file for wmi execution: %w", err)
	}

	installCMD := winpath.Clean(winpath.Join(std.DeployInfo().InstallerRuntime.WorkDir, installBatName))
	stdout, stderr, err := client.RunSilentCommand(std.Context(), installCMD)
	if err != nil {
		return fmt.Errorf("failed to run install node: %w", err)
	}

	std.InstanceData().Log().
		Zh("安装节点 stdout: %s", strings.Split(strings.TrimSpace(stdout), "\n")).
		En("install node stdout: %s", strings.Split(strings.TrimSpace(stdout), "\n")).
		Info()
	std.InstanceData().Log().
		Zh("安装节点 stderr: %s", strings.Split(strings.TrimSpace(stderr), "\n")).
		En("install node stderr: %s", strings.Split(strings.TrimSpace(stderr), "\n")).
		Info()

	return nil
}

func (act *actionInstallNodeByWMI) buildBat(param *installer.NodeInstallParams) (string, error) {
	_, installCmd, err := param.ToWindowsScript()
	if err != nil {
		return "", fmt.Errorf("failed to render node install script: %w", err)
	}

	return installCmd, nil
}
