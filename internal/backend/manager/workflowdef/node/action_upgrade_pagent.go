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
	"net"
	"path"
	"strconv"
	"strings"
	"time"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/winpath"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameUpgradePagent defines the action name.
	ActionNameUpgradePagent = "upgrade_pagent"
)

// NewActionUpgradePagent get a new action.
func NewActionUpgradePagent(capability *Capability) action.Definition {
	return &actionUpgradePagent{
		storageNodeDeployment: capability.StorageNode,
		gseHandler:            capability.GSEHandler,
		provider:              capability.DiscoverProvider,
	}
}

// ActionParamUpgradePagent defines the action param.
type ActionParamUpgradePagent struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionUpgradePagent struct {
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	gseHandler            gse.IHandler
	provider              discover.Provider
}

// Name returns the name of the action.
func (act *actionUpgradePagent) Name() string {
	return ActionNameUpgradePagent
}

// Version returns the version of the action.
func (act *actionUpgradePagent) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionUpgradePagent) Description() string {
	return "upgrade pagent"
}

// Timeout returns the timeout of the action.
func (act *actionUpgradePagent) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionUpgradePagent) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionUpgradePagent) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionUpgradePagent) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: funlen,nonamedreturns
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionUpgradePagent) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamUpgradePagent)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := nodeUtils.NewNodeActionStandarder(act.storageNodeDeployment)
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

	// get upgrade params.
	upgradeParams, err := act.setupUpgradeParams(std)
	if err != nil {
		return err
	}

	std.UpdateInstanceDataContent(ActionWaitInstallerComplete{
		NodeActionStandardParam: param.NodeActionStandardParam,
		EnsureAgentID:           true,
	})

	// exec upgrade command
	if std.DeployInfo().Host.Dynamic.NodeOsType == criteria.OSWindows {
		return act.doUpgradeWindows(std, upgradeParams)
	}

	return act.doUpgradeUnix(std, upgradeParams)
}

func (act *actionUpgradePagent) setupUpgradeParams(
	std *nodeUtils.NodeActionStandarder) (*UpgradeParams, error) {

	toolName, err := tool.FormatInstallerName(std.DeployInfo().Host.Dynamic.NodeOsType, std.DeployInfo().Host.Dynamic.NodeCPUArch)
	if err != nil {
		return nil, fmt.Errorf("failed to format tools name: %w", err)
	}

	deployConstant, err := deployconstant.GetNodeDeployConf(std.DeployInfo().Host.Dynamic.NodeGeneration, std.DeployInfo().Host.Dynamic.NodeOsType)
	if err != nil {
		return nil, fmt.Errorf("failed to get deploy constant: %w", err)
	}

	// get service addresses form relay config file.
	downloadSvcAddr, callbackSvcAddr := act.getServiceAddresses(std)

	upgradeParams := &UpgradeParams{
		AgentID:          std.DeployInfo().Host.Dynamic.AgentID,
		InstallerName:    toolName,
		InstallerWorkDir: std.DeployInfo().InstallerWorkDir,
		NodeVersion:      std.DeployInfo().Host.Dynamic.NodeVersion,
		Generation:       std.DeployInfo().Host.Dynamic.NodeGeneration,
		NodeRole:         std.DeployInfo().Host.Dynamic.NodeRole,
		CallbackSvrAddr:  callbackSvcAddr,
		DownloadSvrAddr:  downloadSvcAddr,
		DeployToken:      std.Token(),
		OperInstID:       std.InstanceData().OperationInstanceID,
		BaseWorkDir:      deployConstant.BaseWorkDir,
		BaseDeployDir:    deployConstant.BaseDeployDir,
	}

	std.InstanceData().LogI(fmt.Sprintf("build upgrade params success. params(%v)", upgradeParams))

	return upgradeParams, nil
}

// nolint: perfsprint
func (act *actionUpgradePagent) doUpgradeUnix(std *nodeUtils.NodeActionStandarder, param *UpgradeParams) error {
	installerPath := path.Clean(path.Join(param.InstallerWorkDir, param.InstallerName))

	args := []string{
		fmt.Sprintf("--deploy_env %s", system.GetEnv()),
		fmt.Sprintf("--generation %d", param.Generation),
		fmt.Sprintf("--node_role %s", param.NodeRole),
		fmt.Sprintf("--base_work_dir %s", param.BaseWorkDir),
		fmt.Sprintf("--base_deploy_dir %s", param.BaseDeployDir),
		fmt.Sprintf("--dlsvr_addr %s", param.DownloadSvrAddr),
		fmt.Sprintf("--cbsvr_addr %s", param.CallbackSvrAddr),
		fmt.Sprintf("--deploy_token %s", param.DeployToken),
		fmt.Sprintf("--node_version %s", param.NodeVersion),
		fmt.Sprintf("--oper_inst_id %s", param.OperInstID),
	}
	if len(param.AdditionArgs) > 0 {
		args = append(args, param.AdditionArgs...)
	}

	upgradeLogPath := path.Clean(fmt.Sprintf("%s.stdout", installerPath))
	upgradeCmd := fmt.Sprintf("chmod +x %s && %s %s %s >%s 2>&1 &",
		installerPath, installerPath, installer.NodeCmdFullUpgrade, strings.Join(args, " "), upgradeLogPath)
	std.InstanceData().LogI("upgrade node command: " + upgradeCmd)

	taskID, err := act.gseHandler.ExecuteScript(std.Context(),
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

	std.InstanceData().LogI("upgrade node task id: " + taskID)

	return nil
}

// nolint: perfsprint
func (act *actionUpgradePagent) doUpgradeWindows(std *nodeUtils.NodeActionStandarder, param *UpgradeParams) error {
	installerPath := winpath.Clean(winpath.Join(param.InstallerWorkDir, param.InstallerName))

	args := []string{
		fmt.Sprintf("--deploy_env %s", system.GetEnv()),
		fmt.Sprintf("--generation %d", param.Generation),
		fmt.Sprintf("--node_role %s", param.NodeRole),
		fmt.Sprintf("--base_work_dir %s", param.BaseWorkDir),
		fmt.Sprintf("--base_deploy_dir %s", param.BaseDeployDir),
		fmt.Sprintf("--dlsvr_addr %s", param.DownloadSvrAddr),
		fmt.Sprintf("--cbsvr_addr %s", param.CallbackSvrAddr),
		fmt.Sprintf("--deploy_token %s", param.DeployToken),
		fmt.Sprintf("--node_version %s", param.NodeVersion),
		fmt.Sprintf("--oper_inst_id %s", param.OperInstID),
	}
	if len(param.AdditionArgs) > 0 {
		args = append(args, param.AdditionArgs...)
	}

	upgradeLogPath := winpath.Clean(fmt.Sprintf("%s.stdout", installerPath))
	upgradeCmd := fmt.Sprintf("%s %s %s >%s 2>&1",
		installerPath, installer.NodeCmdFullUpgrade, strings.Join(args, " "), upgradeLogPath)
	std.InstanceData().LogI("upgrade node command: " + upgradeCmd)

	taskID, err := act.gseHandler.ExecuteScript(std.Context(),
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

	std.InstanceData().LogI("upgrade node task id: " + taskID)

	return nil
}

func (act *actionUpgradePagent) getServiceAddresses(std *nodeUtils.NodeActionStandarder) (
	string, string) {

	downloadSvrAddr := getHTTPAddress(std.DeployInfo().RelayInfo.InnerIP, std.DeployInfo().RelayInfo.DownloadSvcPort)
	callbackSvrAddr := getHTTPAddress(std.DeployInfo().RelayInfo.InnerIP, std.DeployInfo().RelayInfo.CallbackSvcPort)

	std.InstanceData().LogI(fmt.Sprintf("relay download svr addr(%s), callback svr addr(%s)", downloadSvrAddr, callbackSvrAddr))

	return downloadSvrAddr, callbackSvrAddr
}

func getHTTPAddress(ip string, port int64) string {
	return "http://" + net.JoinHostPort(ip, strconv.Itoa(int(port)))
}
