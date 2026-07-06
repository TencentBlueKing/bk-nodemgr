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
	"path"
	"time"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/nodepkg"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
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
func (act *actionInstallProxyBySSH) DelayFn(_ int) func() {
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
		return act.doRelaySSHOnlyInstall(std)
	}

	return act.doRelayCallbackInstall(std)
}

func (act *actionInstallProxyBySSH) isCrossUnitInstall(std *nodeUtils.NodeActionStandarder) bool {
	return std.DeployInfo().Host.Dynamic.ProxyInstallOriginUnitID != std.DeployInfo().Host.Dynamic.NetworkUnitID
}

// doRelayCallbackInstall handles the relay-assisted SSH installation for same-unit installs.
func (act *actionInstallProxyBySSH) doRelayCallbackInstall(std *nodeUtils.NodeActionStandarder) error {
	std.InstanceData().Log().
		Zh("检测到同管控单元 Proxy 安装, 使用 Relay Callback 模式").
		En("same-unit proxy install detected, using Relay Callback mode").
		Info()

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

// doRelaySSHOnlyInstall handles the SSH-only installation for cross-unit installs.
func (act *actionInstallProxyBySSH) doRelaySSHOnlyInstall(std *nodeUtils.NodeActionStandarder) error {
	std.InstanceData().Log().
		Zh("检测到跨管控单元 Proxy 安装, 使用 Relay SSH-Only 模式").
		En("cross-unit proxy install detected, using Relay SSH-Only mode").
		Info()

	credit := nodeUtils.NewCreditHandler(act.nodeInstaller.storageHostCredit, act.nodeInstaller.passwordVault)
	_, cKey, err := credit.GetSSHCredit(std)
	if err != nil {
		return fmt.Errorf("failed to get ssh credit: %w", err)
	}

	relayInfo, err := std.GetSelectedRelay()
	if err != nil {
		return fmt.Errorf("failed to get selected relay: %w", err)
	}

	installerName, err := tool.FormatInstallerName(
		std.DeployInfo().Host.Dynamic.NodeOsType,
		std.DeployInfo().Host.Dynamic.NodeCPUArch,
	)
	if err != nil {
		return fmt.Errorf("failed to format installer name: %w", err)
	}

	installerPath := path.Clean(path.Join(std.DeployInfo().InstallerRuntime.WorkDir, installerName))
	installerCmd, err := act.buildSSHOnlyProxyInstallCmd(std, installerPath)
	if err != nil {
		return err
	}

	releaseName, err := nodepkg.FormatPkgFileName(
		std.DeployInfo().Host.Dynamic.NodeGeneration,
		types.ReleaseTypeProxy,
		platfmt.NewPlatform(std.DeployInfo().Host.Dynamic.NodeOsType, std.DeployInfo().Host.Dynamic.NodeCPUArch),
		std.DeployInfo().Host.Dynamic.NodeVersion,
	)
	if err != nil {
		return fmt.Errorf("failed to format release pkg name: %w", err)
	}

	if err := act.notifyRelayToInstallProxyBySSH(std, cKey, installerName, releaseName, installerCmd, relayInfo); err != nil {
		return fmt.Errorf("failed to notify relay to install proxy by ssh: %w", err)
	}

	if err := act.pagentInstaller.waitForRelayReportInstall(std); err != nil {
		return fmt.Errorf("failed to wait for relay report install: %w", err)
	}

	return saveWaitInstallerPrivateData(
		std.Context(),
		act.pagentInstaller.storageActionInstance,
		std.InstanceData().OperationInstanceID,
		false,
		true,
	)
}

func (act *actionInstallProxyBySSH) buildSSHOnlyProxyInstallCmd(std *nodeUtils.NodeActionStandarder, installerPath string) (string, error) {
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

	_, installCmd, err := installParams.ToUnixScript()
	if err != nil {
		return "", fmt.Errorf("failed to build ssh-only proxy install cmd: %w", err)
	}
	std.InstanceData().Log().
		Zh("安装 Proxy 命令(跨管控单元 SSH-only): %s", installCmd).
		En("install proxy cmd (cross-unit SSH-only): %s", installCmd).
		Info()

	return fmt.Sprintf(`echo "%s" > install.sh && sh install.sh`, installCmd), nil
}

func (act *actionInstallProxyBySSH) notifyRelayToInstallProxyBySSH(
	std *nodeUtils.NodeActionStandarder, cKey, installerName, releaseName, installerCmd string, relayInfo *types.RelayInfo,
) error {

	event := protoRelay.InstallProxyBySSHReq{
		ActionName:       std.InstanceData().Name,
		OperInstID:       std.InstanceData().OperationInstanceID,
		IP:               std.DeployInfo().Host.Dynamic.LoginIP,
		Port:             std.DeployInfo().Host.Dynamic.LoginPort,
		User:             std.DeployInfo().Host.Dynamic.LoginUser,
		LoginMode:        string(std.DeployInfo().Host.Dynamic.LoginMode),
		Password:         cKey,
		InstallerWorkDir: std.DeployInfo().InstallerRuntime.WorkDir,
		InstallerName:    installerName,
		ReleaseName:      releaseName,
		Token:            std.Token(),
		InstallerCmd:     installerCmd,
	}
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	errCh := act.pagentInstaller.proxyMessager.PushToClient(std.Context(), protoRelay.ServerPushEventTypeInstallProxyBySSH, data, relayInfo.AgentID)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("failed to push to relay. event-type(%s), agent-id(%s): %w",
				protoRelay.ServerPushEventTypeInstallProxyBySSH, relayInfo.AgentID, err)
		}

		return nil
	case <-std.Context().Done():
		return fmt.Errorf("context cancelled. event-type(%s), agent-id(%s): %w",
			protoRelay.ServerPushEventTypeInstallProxyBySSH, relayInfo.AgentID, std.Context().Err())
	case <-time.After(queryClientTimeout):
		return fmt.Errorf("wait client timed out. event-type(%s), agent-id(%s)", protoRelay.ServerPushEventTypeInstallProxyBySSH, relayInfo.AgentID)
	}
}
