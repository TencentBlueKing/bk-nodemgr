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
	"errors"
	"fmt"
	"net"
	"path"
	"strconv"
	"strings"
	"time"

	nodedeployment "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-deployment"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/winpath"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameUpgradeNode defines the action name.
	ActionNameUpgradeNode = "upgrade_node"

	upgradeScriptTimeout = 10 * time.Minute
)

// NewActionUpgradeNode get a new action.
func NewActionUpgradeNode(
	storageNodeDeployment nodedeployment.IStorageNodeDeployment,
	gseHandler gse.IHandler,
	logger logger.ILogger,
	provider discover.Provider) action.Definition {

	return &actionUpgradeNode{
		storageNodeDeployment: storageNodeDeployment,
		gseHandler:            gseHandler,
		logger:                logger,
		provider:              provider,
	}
}

// ActionParamUpgradeNode defines the action param.
type ActionParamUpgradeNode struct {
	Token string `json:"token"`
}

// UpgradeParams this struct defines the parameters for upgrading agent.
type UpgradeParams struct {
	AgentID          string
	InstallerName    string
	InstallerWorkDir string
	Generation       types.Generation
	NodeRole         types.NodeRole
	CallbackSvrAddr  string
	FileSvrAddr      string
	NodeVersion      string
	DeployToken      string
	OperInstID       string
	BaseWorkDir      string
	BaseDeployDir    string
	AdditionArgs     []string
}

type actionUpgradeNode struct {
	storageNodeDeployment nodedeployment.IStorageNodeDeployment
	gseHandler            gse.IHandler
	logger                logger.ILogger
	provider              discover.Provider
}

// Name returns the name of the action.
func (act *actionUpgradeNode) Name() string {
	return ActionNameUpgradeNode
}

// Version returns the version of the action.
func (act *actionUpgradeNode) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionUpgradeNode) Description() string {
	return "upgrade node"
}

// Timeout returns the timeout of the action.
func (act *actionUpgradeNode) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionUpgradeNode) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionUpgradeNode) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionUpgradeNode) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: funlen,nonamedreturns
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionUpgradeNode) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActionParamUpgradeNode)
	err = conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		err = fmt.Errorf("failed to convert param: %w", err)

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

	// select matching tools.
	toolName, err := tool.FormatInstallerName(info.Host.Dynamic.NodeOsType, info.Host.Dynamic.NodeCPUArch)
	if err != nil {
		err = fmt.Errorf("failed to format tools name: %w", err)

		return err
	}

	// get service addresses. depends on whether it is a direct or indirect connection.
	fileSvcAddr, callbackSvcAddr, err := act.getServiceAddresses(info)
	if err != nil {
		return fmt.Errorf("failed to get service addresses: %w", err)
	}

	deployConstant, err := deployconstant.GetNodeDeployConf(info.Host.Dynamic.NodeGeneration, info.Host.Dynamic.NodeOsType)
	if err != nil {
		return fmt.Errorf("failed to get deploy constant %w", err)
	}

	upgradeParams := &UpgradeParams{
		AgentID:          info.Host.Dynamic.AgentID,
		InstallerName:    toolName,
		InstallerWorkDir: info.InstallerWorkDir,
		NodeVersion:      info.Host.Dynamic.NodeVersion,
		Generation:       info.Host.Dynamic.NodeGeneration,
		NodeRole:         info.Host.Dynamic.NodeRole,
		CallbackSvrAddr:  callbackSvcAddr,
		FileSvrAddr:      fileSvcAddr,
		DeployToken:      param.Token,
		OperInstID:       ctx.Data.OperationInstanceID,
		BaseWorkDir:      deployConstant.BaseWorkDir,
		BaseDeployDir:    deployConstant.BaseDeployDir,
	}

	// exec upgrade command
	if info.Host.Dynamic.NodeOsType == criteria.OSWindows {
		return act.doUpgradeWindows(ctx, upgradeParams)
	}

	return act.doUpgradeUnix(ctx, upgradeParams)
}

// nolint: perfsprint
func (act *actionUpgradeNode) doUpgradeUnix(ctx *action.InstanceContext, param *UpgradeParams) error {
	installerPath := path.Clean(path.Join(param.InstallerWorkDir, param.InstallerName))

	args := []string{
		fmt.Sprintf("--deploy_env %s", system.GetEnv()),
		fmt.Sprintf("--generation %d", param.Generation),
		fmt.Sprintf("--node_role %s", param.NodeRole),
		fmt.Sprintf("--base_work_dir %s", param.BaseWorkDir),
		fmt.Sprintf("--base_deploy_dir %s", param.BaseDeployDir),
		fmt.Sprintf("--filesvr_addr %s", param.FileSvrAddr),
		fmt.Sprintf("--cbsvr_addr %s", param.CallbackSvrAddr),
		fmt.Sprintf("--deploy_token %s", param.DeployToken),
		fmt.Sprintf("--node_version %s", param.NodeVersion),
		fmt.Sprintf("--oper_inst_id %s", param.OperInstID),
		"--skip_download",
	}
	if len(param.AdditionArgs) > 0 {
		args = append(args, param.AdditionArgs...)
	}

	upgradeLogPath := path.Clean(fmt.Sprintf("%s.stdout", installerPath))
	upgradeCmd := fmt.Sprintf("chmod +x %s && %s full-upgrade %s >%s 2>&1 &",
		installerPath, installerPath, strings.Join(args, " "), upgradeLogPath)
	ctx.Data.LogI("upgrade node cmd: " + upgradeCmd)

	taskID, err := act.gseHandler.ExecuteScript(ctx.Ctx,
		types.ScriptTypeBash,
		fmt.Sprintf(
			`mkdir -p %s && cd %s && echo "%s" > upgrade.sh && sh upgrade.sh`,
			param.InstallerWorkDir,
			param.InstallerWorkDir,
			upgradeCmd),
		upgradeScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: param.AgentID,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to execute upgrade script: %w", err)
	}
	ctx.Data.LogI("upgrade node task-id: " + taskID)

	return nil
}

// nolint: perfsprint
func (act *actionUpgradeNode) doUpgradeWindows(ctx *action.InstanceContext, param *UpgradeParams) error {
	installerPath := winpath.Clean(winpath.Join(param.InstallerWorkDir, param.InstallerName))

	args := []string{
		fmt.Sprintf("--deploy_env %s", system.GetEnv()),
		fmt.Sprintf("--generation %d", param.Generation),
		fmt.Sprintf("--node_role %s", param.NodeRole),
		fmt.Sprintf("--base_work_dir %s", param.BaseWorkDir),
		fmt.Sprintf("--base_deploy_dir %s", param.BaseDeployDir),
		fmt.Sprintf("--filesvr_addr %s", param.FileSvrAddr),
		fmt.Sprintf("--cbsvr_addr %s", param.CallbackSvrAddr),
		fmt.Sprintf("--deploy_token %s", param.DeployToken),
		fmt.Sprintf("--node_version %s", param.NodeVersion),
		fmt.Sprintf("--oper_inst_id %s", param.OperInstID),
		"--skip_download",
	}
	if len(param.AdditionArgs) > 0 {
		args = append(args, param.AdditionArgs...)
	}

	upgradeLogPath := winpath.Clean(fmt.Sprintf("%s.stdout", installerPath))
	upgradeCmd := fmt.Sprintf("%s full-upgrade %s >%s 2>&1",
		installerPath, strings.Join(args, " "), upgradeLogPath)
	ctx.Data.LogI("upgrade node cmd: " + upgradeCmd)

	taskID, err := act.gseHandler.ExecuteScript(ctx.Ctx,
		types.ScriptTypeBat,
		fmt.Sprintf(
			`cd %s && %s`,
			param.InstallerWorkDir,
			upgradeCmd),
		upgradeScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: param.AgentID,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to execute upgrade script: %w", err)
	}
	ctx.Data.LogI("upgrade node task-id: " + taskID)

	return nil
}

func (act *actionUpgradeNode) getServiceAddresses(info *types.DeploymentInfo) (
	string, string, error) {

	if !info.UpgradeOptions.DirectLink {
		fileSvrAddr := getHTTPAddress(info.RelayInfo.InnerIP, info.RelayInfo.FileSvcPort)
		callbackSvrAddr := getHTTPAddress(info.RelayInfo.InnerIP, info.RelayInfo.CallbackSvcPort)
		act.logger.Infof("file server address(%s), callback server address(%s)", fileSvrAddr, callbackSvrAddr)

		return fileSvrAddr, callbackSvrAddr, nil
	}

	randSelector := discover.NewRandomSelector()

	fileSvrEndpoint, err := act.provider.GetEndpoint(
		discover.ServiceNameFile,
		discover.EndpointNameFileBasic,
		randSelector)
	if err != nil {
		return "", "", fmt.Errorf("failed to get file endpoint: %w", err)
	}

	callbackSvrEndpoint, err := act.provider.GetEndpoint(
		discover.ServiceNameBackend,
		discover.EndpointNameBackendCallback,
		randSelector)
	if err != nil {
		return "", "", fmt.Errorf("failed to get backend callback endpoint: %w", err)
	}

	fileSvrAddr := "http://" + fileSvrEndpoint.GetIPV4Address()
	callbackSvrAddr := "http://" + callbackSvrEndpoint.GetIPV4Address()
	act.logger.Infof("file server address(%s), callback server address(%s)", fileSvrAddr, callbackSvrAddr)

	return fileSvrAddr, callbackSvrAddr, nil
}

func getHTTPAddress(ip string, port int64) string {
	return "http://" + net.JoinHostPort(ip, strconv.Itoa(int(port)))
}
