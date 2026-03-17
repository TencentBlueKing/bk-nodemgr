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
	"strings"
	"time"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/relayconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/winpath"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameInstallPagentByWMI defines the action name.
	ActionNameInstallPagentByWMI = "install_pagent_by_wmi"
)

// NewActionInstallPagentByWMI get a new action.
func NewActionInstallPagentByWMI(capability *Capability) action.Definition {
	return &actionInstallPagentByWMI{
		storageHostCredit:     capability.StorageHostCredit,
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		storageActionInstance: capability.StorageWorkflow,

		passwordVault: capability.HostPasswordVault,

		proxyMessager: capability.ProxyMessager,
	}
}

// ActParamInstallPagentBywmi ...
type ActParamInstallPagentBywmi struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionInstallPagentByWMI struct {
	storageHostCredit     credit.IStorageHostCredit
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	storageActionInstance workflow.IStorageActionInstance

	passwordVault creditvault.IHostPasswordVault

	proxyMessager relayhandler.IServerMessager
}

// Name returns the name of the action.
func (act *actionInstallPagentByWMI) Name() string {
	return ActionNameInstallPagentByWMI
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionInstallPagentByWMI) DisplayNameZh() string {
	return "通过 WMI 安装 P-Agent"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionInstallPagentByWMI) DisplayNameEn() string {
	return "Install P-Agent via WMI"
}

// Version returns the version of the action.
func (act *actionInstallPagentByWMI) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionInstallPagentByWMI) Description() string {
	return "let relay to connect to the target machine, transfer files through sftp, and execute the installation command"
}

// Timeout returns the timeout of the action.
func (act *actionInstallPagentByWMI) Timeout() time.Duration {
	return 3 * time.Minute // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actionInstallPagentByWMI) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionInstallPagentByWMI) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionInstallPagentByWMI) DelayFn() func() {
	return func() {
		time.Sleep(5 * time.Second) // nolint: mnd
	}
}

// Do this func define what the action will do.
// To ensure readability, this action uses fmt.Sprintf to concatenate characters.
// nolint: perfsprint,funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionInstallPagentByWMI) Do(ctx *action.InstanceContext) error {
	param := new(ActParamInstallPagentBywmi)
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

	// get wmi credit.
	credit := nodeUtils.NewCreditHandler(act.storageHostCredit, act.passwordVault)
	_, cKey, err := credit.GetWMICredit(std)
	if err != nil {
		return fmt.Errorf("failed to get wmi credit: %w", err)
	}

	// select matching tools, and use sftp to transfer it.
	toolName, installerPath, deployConstant, err := act.setupInstallationTools(std)
	if err != nil {
		return err
	}

	relayInfo, err := std.GetSelectedRelay()
	if err != nil {
		return fmt.Errorf("failed to get selected relay: %w", err)
	}

	downloadURLs, callbackURLs := std.BuildRelayServerURLs(relayInfo)

	// build install command.
	installCmd := act.buildInstallCmd(std, installerPath, deployConstant, downloadURLs, callbackURLs)

	// notify relay to install.
	if err := act.notifyRelayToInstall(std, cKey, toolName, installCmd, relayInfo); err != nil {
		return err
	}

	// wait for relay report install result.
	if err := act.waitForRelayReportInstall(std); err != nil {
		return err
	}

	if err := std.UpdateInstanceDataContent(ActionWaitInstallerComplete{
		NodeActionStandardParam: param.NodeActionStandardParam,
		EnsureAgentID:           true,
	}); err != nil {
		return fmt.Errorf("failed to update instance data content: %w", err)
	}

	return nil
}

func (act *actionInstallPagentByWMI) setupInstallationTools(std *nodeUtils.NodeActionStandarder) (
	string, string, deployconstant.NodeDeployConf, error) {

	toolName, err := tool.FormatInstallerName(std.DeployInfo().Host.Dynamic.NodeOsType,
		std.DeployInfo().Host.Dynamic.NodeCPUArch)
	if err != nil {
		return "", "", deployconstant.NodeDeployConf{}, fmt.Errorf("failed to format tools name: %w", err)
	}

	deployConstant, err := deployconstant.GetNodeDeployConf(std.DeployInfo().Host.Dynamic.NodeGeneration,
		std.DeployInfo().Host.Dynamic.NodeOsType)
	if err != nil {
		return "", "", deployconstant.NodeDeployConf{}, fmt.Errorf("failed to get deploy conf: %w", err)
	}

	installerPath := winpath.Clean(winpath.Join(std.DeployInfo().InstallerWorkDir, toolName))

	std.InstanceData().Log().
		Zh("设置安装工具，工具名(%s)，安装器路径(%s)", toolName, installerPath).
		En("setup installation tools, tool name(%s), installerPath(%s)", toolName, installerPath).
		Info()

	return toolName, installerPath, deployConstant, nil
}

func (act *actionInstallPagentByWMI) notifyRelayToInstall(
	std *nodeUtils.NodeActionStandarder,
	cKey string,
	toolsName string,
	installCmd string,
	relayInfo *types.RelayInfo) error {

	event := protoRelay.InstallPagentByWMIReq{
		ActionName:       std.InstanceData().Name,
		OperInstID:       std.InstanceData().OperationInstanceID,
		IP:               std.DeployInfo().Host.Dynamic.LoginIP,
		Port:             std.DeployInfo().Host.Dynamic.LoginPort,
		User:             std.DeployInfo().Host.Dynamic.LoginUser,
		LoginMode:        string(std.DeployInfo().Host.Dynamic.LoginMode),
		Password:         cKey,
		InstallerWorkDir: std.DeployInfo().InstallerWorkDir,
		ToolsName:        toolsName,
		InstallerCmd:     installCmd,
		InstallerBatName: installBatName,
	}
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	errCh := act.proxyMessager.PushToClient(std.Context(),
		protoRelay.ServerPushEventTypeInstallByWMI, data, relayInfo.AgentID)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("failed to notify relay to install. agent-id(%s): %w", relayInfo.AgentID, err)
		}

		return nil
	case <-std.Context().Done():
		return fmt.Errorf("context cancelled. agent-id(%s): %w", relayInfo.AgentID, std.Context().Err())
	case <-time.After(queryClientTimeoutWMI):
		return fmt.Errorf("wait client timed out. agent-id(%s)", relayInfo.AgentID)
	}
}

// nolint: gocognit
func (act *actionInstallPagentByWMI) waitForRelayReportInstall(
	std *nodeUtils.NodeActionStandarder) error {

	timeoutCtx, cancel := contextx.WithTimeout(contextx.From(std.Context()), waitForRelayReportTimeoutWMI)
	defer cancel()

	ticker := time.NewTicker(waitForRelayReportIntervalWMI)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("wait for relay report install result timed out. oper_inst_id(%s), action_name(%s)",
				std.InstanceData().OperationInstanceID, std.InstanceData().Name)

		case <-ticker.C:
			privateData, err := act.storageActionInstance.GetActionInstancePrivateData(
				timeoutCtx, std.InstanceData().OperationInstanceID, std.InstanceData().Name)
			if err != nil {
				logger.G.Sys().
					WithErr(err).
					With("oper-inst-id", std.InstanceData().OperationInstanceID, "action_name", std.InstanceData().Name).
					Error("failed to get action instance private data")

				continue
			}

			relayInstallResultRaw, exists := privateData[relayconstant.InstallResultKey]
			if !exists {
				continue
			}

			relayInstallResult, ok := relayInstallResultRaw.(map[string]any)
			if !ok {
				return errors.New("unexpected type for relay install result")
			}

			errMsgRaw := relayInstallResult[relayconstant.InstallResultErrMsgKey]
			errMsg, ok := errMsgRaw.(string)
			if !ok {
				return errors.New("unexpected type for error message")
			}

			if errMsg != "" {
				return errors.New(errMsg)
			}

			outStrRaw := relayInstallResult[relayconstant.InstallResultOutStrKey]
			outStr, ok := outStrRaw.(string)
			if !ok {
				return errors.New("unexpected type for output string")
			}

			std.InstanceData().Log().
				Zh("等待 relay 报告安装结果成功，结果 stdout: %s", outStr).
				En("wait for relay report install result successfully. result stdout: %s", outStr).
				Info()

			return nil
		}
	}
}

func (act *actionInstallPagentByWMI) buildInstallCmd(
	std *nodeUtils.NodeActionStandarder,
	installerPath string,
	deployConstant deployconstant.NodeDeployConf,
	downloadURLs, callbackURLs string) string {

	installParams := &InstallParamsWin{
		NodeVersion:     std.DeployInfo().Host.Dynamic.NodeVersion,
		Generation:      std.DeployInfo().Host.Dynamic.NodeGeneration,
		InstallerPath:   installerPath,
		NodeRole:        std.DeployInfo().Host.Dynamic.NodeRole,
		DeployToken:     std.Token(),
		OperInstID:      std.InstanceData().OperationInstanceID,
		BaseWorkDir:     deployConstant.BaseWorkDir,
		BaseDeployDir:   deployConstant.BaseDeployDir,
		DownloadSvrAddr: downloadURLs,
		CallbackSvrAddr: callbackURLs,
	}

	if !std.DeployInfo().InstallOptions.ReRegister && std.DeployInfo().Host.Dynamic.AgentID != "" {
		installParams.AdditionArgs = append(installParams.AdditionArgs,
			fmt.Sprintf("--agent_id %s", std.DeployInfo().Host.Dynamic.AgentID))
	}

	args := []string{
		fmt.Sprintf("--deploy_env %s", system.GetEnv()),
		fmt.Sprintf("--generation %d", installParams.Generation),
		fmt.Sprintf("--node_role %s", installParams.NodeRole),
		fmt.Sprintf("--base_work_dir %s", installParams.BaseWorkDir),
		fmt.Sprintf("--base_deploy_dir %s", installParams.BaseDeployDir),
		fmt.Sprintf("--deploy_token %s", installParams.DeployToken),
		fmt.Sprintf("--node_version %s", installParams.NodeVersion),
		fmt.Sprintf("--oper_inst_id %s", installParams.OperInstID),
		fmt.Sprintf("--dlsvr_addr %s", installParams.DownloadSvrAddr),
		fmt.Sprintf("--cbsvr_addr %s", installParams.CallbackSvrAddr),
	}
	if len(installParams.AdditionArgs) > 0 {
		args = append(args, installParams.AdditionArgs...)
	}

	installLogPath := winpath.Clean(fmt.Sprintf("%s.stdout", installParams.InstallerPath))

	installCmd := fmt.Sprintf("cd %s && %s %s %s >%s 2>&1",
		winpath.Join(installParams.BaseWorkDir, system.GetEnv()),
		installParams.InstallerPath, installer.NodeCmdFullInstall, strings.Join(args, " "), installLogPath)

	std.InstanceData().Log().
		Zh("构建安装参数: %v", args).
		En("build install params: %v", args).
		Info()

	return installCmd
}
