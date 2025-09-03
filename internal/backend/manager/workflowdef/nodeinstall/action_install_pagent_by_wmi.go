/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodeinstall

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	nodedeployment "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-deployment"
	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/relayconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/winpath"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
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
	Token    string `json:"token"`
	Operator string `json:"operator"`
}

// pagentInstallParamsWMI this struct defines the parameters for installing agent.
type pagentInstallParamsWin struct {
	InstallerPath string
	Generation    types.Generation
	NodeRole      types.NodeRole
	NodeVersion   string
	DeployToken   string
	OperInstID    string
	BaseWorkDir   string
	BaseDeployDir string
	AdditionArgs  []string
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
// nolint: perfsprint,funlen,fnsize
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionInstallPagentByWMI) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamInstallAgentByWMI)
	err = conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		err = fmt.Errorf("failed to convert param, err: %w", err)

		return err
	}

	info, err := act.storageNodeDeployment.GetInfo(ctx.Ctx, param.Token)
	if err != nil {
		return err
	}
	// let the callback server known which action to mark and log.
	info.BlockingActionName = ActionNameWaitInstallerComplete

	defer func() {
		if storeErr := act.storageNodeDeployment.UpdateInfo(ctx.Ctx, param.Token, info); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	// select matching tools, and use sftp to transfer it.
	toolName, err := tool.FormatInstallerName(info.Host.Dynamic.NodeOsType, info.Host.Dynamic.NodeCPUArch)
	if err != nil {
		err = fmt.Errorf("failed to format tools name, err: %w", err)

		return err
	}

	installerPath := winpath.Clean(winpath.Join(info.InstallerWorkDir, toolName))

	deployConstant, err := deployconstant.GetDeployConf(info.Host.Dynamic.NodeGeneration, info.Host.Dynamic.NodeOsType)
	if err != nil {
		return fmt.Errorf("failed to get deploy constant, err: %w", err)
	}

	installParams := &pagentInstallParamsWin{
		NodeVersion:   info.Host.Dynamic.NodeVersion,
		Generation:    info.Host.Dynamic.NodeGeneration,
		InstallerPath: installerPath,
		NodeRole:      info.Host.Dynamic.NodeRole,
		DeployToken:   param.Token,
		OperInstID:    ctx.Data.OperationInstanceID,
		BaseWorkDir:   deployConstant.BaseWorkDir,
		BaseDeployDir: deployConstant.BaseDeployDir,
	}

	if !info.InstallOptions.ReRegister && info.Host.Dynamic.AgentID != "" {
		installParams.AdditionArgs = append(installParams.AdditionArgs,
			fmt.Sprintf("--agent_id %s", info.Host.Dynamic.AgentID))
	}

	// exec install command
	installBat := act.buildBat(installParams)
	ctx.Data.LogI(fmt.Sprintf("install node cmd: %s", installBat))

	password, err := act.queryPassword(contextx.NewTenantUserContext(ctx.Ctx, info.Host.TenantID, param.Operator),
		param.Operator, act.storageHostCredit, act.passwordVault, info)
	if err != nil {
		return err
	}

	targetWorkDir := winpath.Join(installParams.BaseWorkDir, system.GetEnv())
	if err := act.notifyRelayToInstall(ctx, info, password, toolName, targetWorkDir, installBat, &info.RelayInfo); err != nil {
		return err
	}
	ctx.Data.LogI("notify relay to install pagent by wmi successfully")

	var outStr string
	outStr, err = act.waitForRelayReportInstall(ctx)
	if err != nil {
		return err
	}
	ctx.Data.LogI("relay run install pagent by wmi successfully. result out str: " + outStr)

	return nil
}

func (act *actionInstallPagentByWMI) notifyRelayToInstall(ctx *action.InstanceContext,
	info *types.DeploymentInfo, password, toolsName, targetWorkDir string, args []string, relayInfo *types.RelayInfo) error {

	event := protoRelay.InstallPagentByWMIReq{
		ActionName:       ctx.Data.Name,
		OperInstID:       ctx.Data.OperationInstanceID,
		IP:               info.Host.Dynamic.LoginIP,
		Port:             info.Host.Dynamic.LoginPort,
		User:             info.Host.Dynamic.LoginUser,
		LoginMode:        string(info.Host.Dynamic.LoginMode),
		Password:         password,
		InstallerWorkDir: info.InstallerWorkDir,
		ToolsName:        toolsName,
		InstallerCmd:     args,
		TargetWorkDir:    targetWorkDir,
		InstallerBatName: installBatName,
	}
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event failed: %w", err)
	}

	errCh := act.proxyMessager.PushToClient(ctx.Ctx,
		protoRelay.ServerPushEventTypeInstallByWMI, data, relayInfo.AgentID)

	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("notify relay to install failed: %w", err)
		}
	case <-ctx.Ctx.Done():
		return ctx.Ctx.Err()
	}

	act.logger.Infof("notify relay to install pagent.")

	return nil
}

func (act *actionInstallPagentByWMI) waitForRelayReportInstall(
	ctx *action.InstanceContext) (string, error) {

	timeoutCtx, cancel := context.WithTimeout(ctx.Ctx, waitForRelayReportTimeout)
	defer cancel()

	ticker := time.NewTicker(waitForRelayReportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return "", fmt.Errorf("wait for relay report install result timed out. oper_inst_id(%s), action_name(%s)",
				ctx.Data.OperationInstanceID, ctx.Data.Name)

		case <-ticker.C:
			privateData, err := act.storageActionInstance.GetActionInstancePrivateData(
				timeoutCtx, ctx.Data.OperationInstanceID, ctx.Data.Name)
			if err != nil {
				act.logger.Warnf("get private data failed, retrying. oper_inst_id(%s), action_name(%s): %v",
					ctx.Data.OperationInstanceID, ctx.Data.Name, err)

				continue
			}

			relayInstallResultRaw, exists := privateData[relayconstant.InstallResultKey]
			if !exists {
				continue
			}

			relayInstallResult, ok := relayInstallResultRaw.(map[string]any)
			if !ok {
				return "", errors.New("unexpected type for relay install result")
			}

			errMsgRaw := relayInstallResult[relayconstant.InstallResultErrMsgKey]
			errMsg, ok := errMsgRaw.(string)
			if !ok {
				return "", errors.New("unexpected type for error message")
			}

			if errMsg != "" {
				return "", errors.New(errMsg)
			}

			outStrRaw := relayInstallResult[relayconstant.InstallResultOutStrKey]
			outStr, ok := outStrRaw.(string)
			if !ok {
				return "", errors.New("unexpected type for output string")
			}

			return outStr, nil
		}
	}
}

// To ensure readability, this action uses fmt.Sprintf to concatenate characters.
// nolint: perfsprint
func (act *actionInstallPagentByWMI) buildBat(param *pagentInstallParamsWin) []string {
	args := []string{
		fmt.Sprintf("--deploy_env %s", system.GetEnv()),
		fmt.Sprintf("--generation %d", param.Generation),
		fmt.Sprintf("--node_role %s", param.NodeRole),
		fmt.Sprintf("--base_work_dir %s", param.BaseWorkDir),
		fmt.Sprintf("--base_deploy_dir %s", param.BaseDeployDir),
		fmt.Sprintf("--deploy_token %s", param.DeployToken),
		fmt.Sprintf("--node_version %s", param.NodeVersion),
		fmt.Sprintf("--oper_inst_id %s", param.OperInstID),
	}
	if len(param.AdditionArgs) > 0 {
		args = append(args, param.AdditionArgs...)
	}

	return args
}

func (act *actionInstallPagentByWMI) queryPassword(
	ctx contextx.ITenantContext,
	operator string,
	storageHostCredit credit.IStorageHostCredit,
	passwordVault creditvault.IHostPasswordVault,
	info *types.DeploymentInfo) (string, error) {

	switch info.Host.Dynamic.LoginMode {
	case types.LoginModePassword:
		passwd, err := storageHostCredit.LoadHostCredit(
			ctx,
			info.Host.Dynamic.LoginCreditID)
		if err != nil {
			return "", fmt.Errorf("failed to load password from storageHostCredit storage: %w", err)
		}

		return string(passwd), nil

	case types.LoginModeKeyFile:
		return "", errors.New("implete me")
	case types.LoginModePasswordVault:
		passwd, err := passwordVault.LoadPassword(
			ctx,
			operator,
			info.Host.Static.NetworkAreaID,
			info.Host.Dynamic.LoginIP,
			info.Host.Dynamic.LoginUser)
		if err != nil {
			return "", fmt.Errorf("failed to load password from password vault: %w", err)
		}

		return string(passwd), nil
	default:
		return "", fmt.Errorf("unsupported login mode, mode(%s)", info.Host.Dynamic.LoginMode)
	}
}
