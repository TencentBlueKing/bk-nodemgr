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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	nodedeployment "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-deployment"
	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/relayconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/winpath"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/wmix"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameInstallPagentByWMI defines the action name.
	ActionNameInstallPagentByWMI = "install_pagent_by_wmi"
)

// NewActionInstallPagentByWMI get a new action.
func NewActionInstallPagentByWMI(
	proxyMessager relayhandler.IServerMessager,
	storageNodeDeployment nodedeployment.IStorageNodeDeployment,
	storageHostCredit credit.IStorageHostCredit,
	storageActionInstance workflow.IStorageActionInstance,
	passwordVault creditvault.IHostPasswordVault,
	logger logger.ILogger,
) action.Definition {

	return &actionInstallPagentByWMI{
		storageHostCredit:     storageHostCredit,
		storageNodeDeployment: storageNodeDeployment,
		storageActionInstance: storageActionInstance,

		passwordVault: passwordVault,

		proxyMessager: proxyMessager,

		logger: logger,
	}
}

// ActParamInstallPagentBywmi ...
type ActParamInstallPagentBywmi struct {
	utils.NodeActionStandardParam `json:",inline"`
}

type actionInstallPagentByWMI struct {
	logger logger.ILogger

	storageHostCredit     credit.IStorageHostCredit
	storageNodeDeployment nodedeployment.IStorageNodeDeployment
	storageActionInstance workflow.IStorageActionInstance

	passwordVault creditvault.IHostPasswordVault

	proxyMessager relayhandler.IServerMessager
}

// Name returns the name of the action.
func (act *actionInstallPagentByWMI) Name() string {
	return ActionNameInstallPagentByWMI
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
func (act *actionInstallPagentByWMI) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamInstallPagentBywmi)
	err = conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		err = fmt.Errorf("failed to convert param: %w", err)

		return err
	}

	// initialize standard data.
	std := utils.NewNodeActionStandarder(act.storageNodeDeployment)
	if err = std.Initialize(ctx, param.NodeActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	// let the callback server known which action to mark and log.
	std.DeployInfo().BlockingActionName = ActionNameWaitInstallerComplete

	// get ssh credit.
	credit := utils.NewCreditHandler(act.storageHostCredit, act.passwordVault)
	cMethod, cKey, err := credit.GetWMICredit(std)
	if err != nil {
		return fmt.Errorf("failed to get ssh credit: %w", err)
	}

	// select matching tools, and use sftp to transfer it.
	toolName, installerPath, deployConstant, err := act.setupInstallationTools(std)
	if err != nil {
		return err
	}

	// build install command.
	installCmd := act.buildInstallParams(std, installerPath, deployConstant)

	// notify relay to install pagent by ssh.
	targetWorkDir := winpath.Join(deployConstant.BaseWorkDir, system.GetEnv())
	if err := act.notifyRelayToInstall(std, cMethod, cKey, toolName, targetWorkDir, installCmd); err != nil {
		return err
	}

	// wait for relay report install result.
	if err := act.waitForRelayReportInstall(std); err != nil {
		return err
	}

	std.InstanceData().LogI("install pagent by wmi successfully")

	return nil
}

func (act *actionInstallPagentByWMI) setupInstallationTools(std *utils.NodeActionStandarder) (
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

	std.InstanceData().LogI(fmt.Sprintf("setup installation tools,tool name(%s), installerPath(%s)", toolName, installerPath))

	return toolName, installerPath, deployConstant, nil
}

func (act *actionInstallPagentByWMI) notifyRelayToInstall(
	std *utils.NodeActionStandarder,
	cMethod wmix.AuthMethod, cKey,
	toolsName, targetWorkDir string,
	args []string) error {

	event := protoRelay.InstallPagentByWMIReq{
		ActionName:       std.InstanceData().Name,
		OperInstID:       std.InstanceData().OperationInstanceID,
		IP:               std.DeployInfo().Host.Dynamic.LoginIP,
		Port:             std.DeployInfo().Host.Dynamic.LoginPort,
		User:             std.DeployInfo().Host.Dynamic.LoginUser,
		LoginMode:        string(cMethod),
		Password:         cKey,
		InstallerWorkDir: std.DeployInfo().InstallerWorkDir,
		ToolsName:        toolsName,
		InstallerCmd:     args,
		TargetWorkDir:    targetWorkDir,
		InstallerBatName: installBatName,
	}
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event failed: %w", err)
	}

	errCh := act.proxyMessager.PushToClient(std.Context(),
		protoRelay.ServerPushEventTypeInstallByWMI, data, std.DeployInfo().RelayInfo.AgentID)

	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("notify relay to install failed: %w", err)
		}
	case <-std.Context().Done():
		return std.Context().Err()
	}

	act.logger.Infof("notify relay to install pagent.")

	return nil
}

func (act *actionInstallPagentByWMI) waitForRelayReportInstall(
	std *utils.NodeActionStandarder) error {

	timeoutCtx, cancel := context.WithTimeout(std.Context(), waitForRelayReportTimeout)
	defer cancel()

	ticker := time.NewTicker(waitForRelayReportInterval)
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
				act.logger.Warnf("get private data failed, retrying. oper_inst_id(%s), action_name(%s): %v",
					std.InstanceData().OperationInstanceID, std.InstanceData().Name, err)

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

			std.InstanceData().LogI("wait for relay report install result successfully. result stdout: " + outStr)

			return nil
		}
	}
}

// TODO: add relay file and callback address.
func (act *actionInstallPagentByWMI) buildInstallParams(
	std *utils.NodeActionStandarder,
	installerPath string, deployConstant deployconstant.NodeDeployConf) []string {

	installParams := &InstallParamsWin{
		NodeVersion:   std.DeployInfo().Host.Dynamic.NodeVersion,
		Generation:    std.DeployInfo().Host.Dynamic.NodeGeneration,
		InstallerPath: installerPath,
		NodeRole:      std.DeployInfo().Host.Dynamic.NodeRole,
		DeployToken:   std.Token(),
		OperInstID:    std.InstanceData().OperationInstanceID,
		BaseWorkDir:   deployConstant.BaseWorkDir,
		BaseDeployDir: deployConstant.BaseDeployDir,
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
	}
	if len(installParams.AdditionArgs) > 0 {
		args = append(args, installParams.AdditionArgs...)
	}

	std.InstanceData().LogI(fmt.Sprintf("build install params: %v", args))

	return args
}
