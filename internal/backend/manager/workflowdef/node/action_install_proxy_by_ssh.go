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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/nodeconfig"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/configfile"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/nodepkg"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/sshx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameInstallProxyBySSH defines the proxy SSH install action name.
	ActionNameInstallProxyBySSH = "install_proxy_by_ssh"
)

// NewActionInstallProxyBySSH creates a new proxy SSH install action.
func NewActionInstallProxyBySSH(capability *Capability) action.Definition {
	return &actionInstallProxyBySSH{
		nodeInstaller: &actionInstallNodeBySSH{
			fileHandler:           capability.FileHandler,
			fileCache:             capability.FileCache,
			storageHostCredit:     capability.StorageHostCredit,
			storageNodeDeployment: capability.StorageNode,
			storageHost:           capability.StorageTopo,
			storageNetworkUnit:    capability.StorageTopo,
			provider:              capability.DiscoverProvider,
			passwordVault:         capability.HostPasswordVault,
			storageActionInstance: capability.StorageWorkflow,
		},
		pagentInstaller: &actionInstallPagentBySSH{
			storageHostCredit:     capability.StorageHostCredit,
			storageNodeDeployment: capability.StorageNode,
			storageHost:           capability.StorageTopo,
			storageActionInstance: capability.StorageWorkflow,
			storageNetworkUnit:    capability.StorageTopo,
			provider:              capability.DiscoverProvider,
			passwordVault:         capability.HostPasswordVault,
			proxyMessager:         capability.ProxyMessager,
		},
	}
}

// ActParamInstallProxyBySSH defines parameters for actionInstallProxyBySSH.
type ActParamInstallProxyBySSH struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionInstallProxyBySSH struct {
	nodeInstaller   *actionInstallNodeBySSH
	pagentInstaller *actionInstallPagentBySSH
}

// Name returns the name of the action.
func (act *actionInstallProxyBySSH) Name() string {
	return ActionNameInstallProxyBySSH
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionInstallProxyBySSH) DisplayNameZh() string {
	return "通过 SSH 安装 Proxy"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionInstallProxyBySSH) DisplayNameEn() string {
	return "Install Proxy via SSH"
}

// Version returns the version of the action.
func (act *actionInstallProxyBySSH) Version() string {
	return "v1.0.0"
}

// Description returns the description of the action.
func (act *actionInstallProxyBySSH) Description() string {
	return "Install proxy with SSH-only or relay-assisted SSH according to the proxy install origin unit"
}

// Timeout returns the timeout of the action.
func (act *actionInstallProxyBySSH) Timeout() time.Duration {
	return 5 * time.Minute // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actionInstallProxyBySSH) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionInstallProxyBySSH) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn defines how long to wait before retrying after failure.
func (act *actionInstallProxyBySSH) DelayFn() func() {
	return func() {
		time.Sleep(5 * time.Second) // nolint: mnd
	}
}

// Do installs proxy by SSH-only mode for cross-unit installs or relay-assisted SSH for same-unit indirect installs.
func (act *actionInstallProxyBySSH) Do(ctx *action.InstanceContext) error {
	param := new(ActParamInstallProxyBySSH)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	std := nodeUtils.NewNodeActionStandarder(act.nodeInstaller.storageNodeDeployment, act.nodeInstaller.storageHost)
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

	crossUnit := act.isCrossUnitInstall(std)
	if !crossUnit && std.DeployInfo().InstallOptions.DirectInstall {
		return fmt.Errorf("same-unit proxy install from direct origin unit is unsupported, networkunit-id(%d)",
			std.DeployInfo().Host.Dynamic.NetworkUnitID)
	}

	if crossUnit {
		return act.doCrossUnitInstall(std)
	}

	return act.doRelayInstall(std)
}

func (act *actionInstallProxyBySSH) isCrossUnitInstall(std *nodeUtils.NodeActionStandarder) bool {
	return std.DeployInfo().Host.Dynamic.ProxyInstallOriginUnitID != std.DeployInfo().Host.Dynamic.NetworkUnitID
}

func (act *actionInstallProxyBySSH) doRelayInstall(std *nodeUtils.NodeActionStandarder) error {
	credit := nodeUtils.NewCreditHandler(act.pagentInstaller.storageHostCredit, act.pagentInstaller.passwordVault)
	_, cKey, err := credit.GetSSHCredit(std)
	if err != nil {
		return fmt.Errorf("failed to get ssh credit: %w", err)
	}

	toolName, installerPath, err := act.pagentInstaller.setupInstallationTools(std)
	if err != nil {
		return err
	}

	relayInfo, err := std.GetSelectedRelay()
	if err != nil {
		return fmt.Errorf("failed to get selected relay: %w", err)
	}

	installCmd, err := act.pagentInstaller.buildInstallCmd(std, installerPath)
	if err != nil {
		return fmt.Errorf("failed to build install cmd: %w", err)
	}

	if err := act.pagentInstaller.notifyRelayToInstall(std, cKey, toolName, installCmd, relayInfo); err != nil {
		return err
	}

	if err := act.pagentInstaller.waitForRelayReportInstall(std); err != nil {
		return err
	}

	return saveWaitInstallerPrivateData(
		std.Context(),
		act.pagentInstaller.storageActionInstance,
		std.InstanceData().OperationInstanceID,
		false,
		true,
	)
}

func (act *actionInstallProxyBySSH) doCrossUnitInstall(std *nodeUtils.NodeActionStandarder) error {
	credit := nodeUtils.NewCreditHandler(act.nodeInstaller.storageHostCredit, act.nodeInstaller.passwordVault)
	cMethod, cKey, err := credit.GetSSHCredit(std)
	if err != nil {
		return fmt.Errorf("failed to get ssh credit: %w", err)
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
	}, sshx.DefaultTimeout)
	if err != nil {
		return fmt.Errorf("failed to generate new ssh client: %w", err)
	}

	if err = act.nodeInstaller.ensureWorkspace(std, client); err != nil {
		return fmt.Errorf("failed to ensure workspace through ssh: %w", err)
	}

	installerPath, err := act.nodeInstaller.ensureInstallerTool(std, client)
	if err != nil {
		return fmt.Errorf("failed to ensure installer tool through ssh: %w", err)
	}

	return act.doSSHOnlyInstall(std, client, installerPath)
}

func (act *actionInstallProxyBySSH) doSSHOnlyInstall(
	std *nodeUtils.NodeActionStandarder, client *sshx.Client, installerPath string) error {

	std.InstanceData().Log().
		Zh("检测到跨管控单元 Proxy 安装, 使用 SSH-Only 模式").
		En("cross-unit proxy install detected, using SSH-Only mode").
		Info()

	if err := act.ensureProxyArtifacts(std, client); err != nil {
		return fmt.Errorf("failed to ensure proxy artifacts: %w", err)
	}

	if err := act.executeSSHOnlyProxyInstallCMD(std, client, installerPath); err != nil {
		return fmt.Errorf("failed to execute ssh-only proxy install cmd: %w", err)
	}

	return saveWaitInstallerPrivateData(std.Context(), act.nodeInstaller.storageActionInstance,
		std.InstanceData().OperationInstanceID, true, true)
}

func (act *actionInstallProxyBySSH) ensureProxyArtifacts(std *nodeUtils.NodeActionStandarder, client *sshx.Client) error {
	nodeConf, err := act.nodeInstaller.storageNodeDeployment.GetNodeDeploymentNodeConf(std.Context(), std.Token())
	if err != nil {
		return fmt.Errorf("failed to get node conf: %w", err)
	}

	dataDir := path.Join(std.DeployInfo().InstallerRuntime.WorkDir, "data")
	configDir := path.Join(dataDir, "config")
	if _, stderr, err := client.RunCommand("mkdir -p " + configDir); err != nil {
		return fmt.Errorf("failed to mkdir config dir, stderr(%s): %w", stderr, err)
	}

	configKeys := map[string]string{
		types.ConfigKeyAgent: "gse_agent.conf",
		types.ConfigKeyFile:  "gse_file_proxy.conf",
		types.ConfigKeyData:  "gse_data_proxy.conf",
	}

	for key, filename := range configKeys {
		rendered, err := nodeconfig.RenderNodeConfig(key, nodeConf)
		if err != nil {
			return fmt.Errorf("failed to render config %s: %w", key, err)
		}

		configBytes, err := configfile.FormatConfigFileJson(rendered)
		if err != nil {
			return fmt.Errorf("failed to marshal rendered config %s: %w", key, err)
		}

		remotePath := path.Join(configDir, filename)
		if err := client.TransferFile(io.NopCloser(strings.NewReader(string(configBytes))), remotePath); err != nil {
			return fmt.Errorf("failed to transfer config %s: %w", filename, err)
		}

		std.InstanceData().Log().
			Zh("已推送配置文件: %s", remotePath).
			En("pushed config file: %s", remotePath).
			Info()
	}

	checkList, err := nodeconfig.BuildCheckList(std.DeployInfo(), nodeConf)
	if err != nil {
		return fmt.Errorf("failed to build checklist: %w", err)
	}

	checkListBytes, err := json.Marshal(checkList)
	if err != nil {
		return fmt.Errorf("failed to marshal checklist: %w", err)
	}

	checkListPath := path.Join(dataDir, "precheck.json")
	if err := client.TransferFile(io.NopCloser(strings.NewReader(string(checkListBytes))), checkListPath); err != nil {
		return fmt.Errorf("failed to transfer checklist: %w", err)
	}

	std.InstanceData().Log().
		Zh("已推送预检清单: %s", checkListPath).
		En("pushed checklist: %s", checkListPath).
		Info()

	if err := act.pushProxyReleasePackage(std, client); err != nil {
		return fmt.Errorf("failed to push release package: %w", err)
	}

	return nil
}

func (act *actionInstallProxyBySSH) pushProxyReleasePackage(std *nodeUtils.NodeActionStandarder, client *sshx.Client) error {
	plat := platfmt.NewPlatform(
		std.DeployInfo().Host.Dynamic.NodeOsType,
		std.DeployInfo().Host.Dynamic.NodeCPUArch,
	)

	installerPkgName, err := nodepkg.FormatPkgFileName(
		std.DeployInfo().Host.Dynamic.NodeGeneration,
		types.ReleaseTypeProxy,
		plat,
		std.DeployInfo().Host.Dynamic.NodeVersion,
	)
	if err != nil {
		return fmt.Errorf("failed to format installer pkg name: %w", err)
	}

	reader, _, err := act.openReleaseReader(std)
	if err != nil {
		return fmt.Errorf("failed to get release package: %w", err)
	}

	dataDir := path.Join(std.DeployInfo().InstallerRuntime.WorkDir, "data")
	remotePath := path.Join(dataDir, installerPkgName)
	if err := client.TransferFile(reader, remotePath); err != nil {
		return fmt.Errorf("failed to transfer release package: %w", err)
	}

	std.InstanceData().Log().
		Zh("已推送 release 包: %s", remotePath).
		En("pushed release package: %s", remotePath).
		Info()

	return nil
}

func (act *actionInstallProxyBySSH) openReleaseReader(std *nodeUtils.NodeActionStandarder) (io.ReadCloser, string, error) {
	gen := std.DeployInfo().Host.Dynamic.NodeGeneration
	osType := std.DeployInfo().Host.Dynamic.NodeOsType
	cpuArch := std.DeployInfo().Host.Dynamic.NodeCPUArch
	version := std.DeployInfo().Host.Dynamic.NodeVersion
	plat := platfmt.NewPlatform(osType, cpuArch)

	pkgFileName, err := nodepkg.FormatPkgFileName(gen, types.ReleaseTypeOriginProxy, plat, version)
	if err != nil {
		return nil, "", fmt.Errorf("failed to format release pkg name: %w", err)
	}

	if act.nodeInstaller.fileCache == nil {
		resp, err := act.nodeInstaller.fileHandler.DownloadReleaseProxy(std.Context(), gen, plat, version)
		if err != nil {
			return nil, "", fmt.Errorf("failed to download release proxy: %w", err)
		}

		return resp.Data, pkgFileName, nil
	}

	fileInfo, err := act.nodeInstaller.fileHandler.InfoReleaseProxy(std.Context(), gen, plat, version)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get release proxy info: %w", err)
	}

	cachedFile, _, err := act.nodeInstaller.fileCache.GetOrFetch(std.Context(), fileInfo.Name, fileInfo.MD5,
		func(nCtx contextx.IContext) (io.ReadCloser, error) {
			resp, dlErr := act.nodeInstaller.fileHandler.DownloadReleaseProxy(nCtx, gen, plat, version)
			if dlErr != nil {
				return nil, dlErr
			}

			return resp.Data, nil
		})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get release proxy from file cache: %w", err)
	}

	reader, err := cachedFile.Content(std.Context())
	if err != nil {
		return nil, "", fmt.Errorf("failed to open cached release content: %w", err)
	}

	return reader, pkgFileName, nil
}

func (act *actionInstallProxyBySSH) executeSSHOnlyProxyInstallCMD(
	std *nodeUtils.NodeActionStandarder, client *sshx.Client, installerPath string) error {

	installParams := &installer.NodeInstallParams{
		NodeCommonParams: installer.NodeCommonParams{
			DeployEnv:     system.GetEnv(),
			Generation:    int(std.DeployInfo().Host.Dynamic.NodeGeneration),
			NodeRole:      string(std.DeployInfo().Host.Dynamic.NodeRole),
			BaseWorkDir:   std.DeployInfo().InstallerRuntime.BaseWorkDir,
			BaseDeployDir: std.DeployInfo().BaseRuntime.BaseDeployDir,
		},
		InstallerPath: installerPath,
		DeployToken:   std.Token(),
		NodeVersion:   std.DeployInfo().Host.Dynamic.NodeVersion,
		OperInstID:    std.InstanceData().OperationInstanceID,
		SkipCallback:  true,
		SkipDownload:  true,
		RenewGSEProc:  std.DeployInfo().InstallOptions.RenewGSEProc,
		RenewGSETask:  std.DeployInfo().InstallOptions.RenewGSETask,
	}

	if !std.DeployInfo().InstallOptions.ReRegister && std.DeployInfo().Host.Dynamic.AgentID != "" {
		installParams.AdditionArgs = append(installParams.AdditionArgs,
			fmt.Sprintf("--agent_id %s", std.DeployInfo().Host.Dynamic.AgentID))
	}

	installCmd, err := act.nodeInstaller.buildCMD(installParams)
	if err != nil {
		return fmt.Errorf("failed to build ssh-only proxy install cmd: %w", err)
	}
	std.InstanceData().Log().
		Zh("安装 Proxy 命令(跨管控单元 SSH-only): %s", installCmd).
		En("install proxy cmd (cross-unit SSH-only): %s", installCmd).
		Info()

	outStr, _, err := client.RunCommand(fmt.Sprintf(
		`mkdir -p %s && cd %s && echo "%s" > install.sh && sh install.sh`,
		std.DeployInfo().InstallerRuntime.WorkDir,
		std.DeployInfo().InstallerRuntime.WorkDir,
		installCmd),
	)
	if err != nil {
		return fmt.Errorf("failed to run install proxy: %w", err)
	}

	std.InstanceData().Log().
		Zh("安装 Proxy 结果: %s", outStr).
		En("install proxy result: %s", outStr).
		Info()

	return nil
}
