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
	"path"
	"strings"
	"time"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
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
	// ActionNameUninstallPagent defines the action name.
	ActionNameUninstallPagent = "uninstall_pagent"

	uninstallPagentScriptTimeout = 10 * time.Minute
)

// NewActionUninstallPagent get a new action.
func NewActionUninstallPagent(capability *Capability) action.Definition {
	return &actionUninstallPagent{
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		gseHandler:            capability.GSEHandler,
	}
}

// ActionParamUninstallPagent defines the action param.
type ActionParamUninstallPagent struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionUninstallPagent struct {
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	gseHandler            gse.IHandler
}

// Name returns the name of the action.
func (act *actionUninstallPagent) Name() string {
	return ActionNameUninstallPagent
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionUninstallPagent) DisplayNameZh() string {
	return "卸载 P-Agent"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionUninstallPagent) DisplayNameEn() string {
	return "Uninstall P-Agent"
}

// Version returns the version of the action.
func (act *actionUninstallPagent) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionUninstallPagent) Description() string {
	return "uninstall pagent through relay callback"
}

// Timeout returns the timeout of the action.
func (act *actionUninstallPagent) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionUninstallPagent) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionUninstallPagent) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionUninstallPagent) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: funlen,nonamedreturns
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionUninstallPagent) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamUninstallPagent)
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

	// select matching tools.
	toolName, err := tool.FormatInstallerName(std.DeployInfo().Host.Dynamic.NodeOsType, std.DeployInfo().Host.Dynamic.NodeCPUArch)
	if err != nil {
		return err
	}

	// get callback address from the selected relay.
	relayInfo, err := std.GetSelectedRelay()
	if err != nil {
		return fmt.Errorf("failed to get selected relay: %w", err)
	}
	callbackSvrAddr, _ := std.BuildRelayServerURLs(relayInfo)
	std.InstanceData().Log().
		Zh("relay 回调服务地址(%s)", callbackSvrAddr).
		En("relay callback svr addr(%s)", callbackSvrAddr).
		Info()

	uninstallParams := &UninstallParams{
		AgentID:          std.DeployInfo().Host.Dynamic.AgentID,
		InstallerName:    toolName,
		InstallerWorkDir: std.DeployInfo().InstallerRuntime.WorkDir,
		Generation:       std.DeployInfo().Host.Dynamic.NodeGeneration,
		NodeRole:         std.DeployInfo().Host.Dynamic.NodeRole,
		CallbackSvrAddr:  callbackSvrAddr,
		DeployToken:      std.Token(),
		OperInstID:       std.InstanceData().OperationInstanceID,
		BaseWorkDir:      std.DeployInfo().InstallerRuntime.BaseWorkDir,
		BaseDeployDir:    std.DeployInfo().BaseRuntime.BaseDeployDir,
	}

	if err := std.UpdateInstanceDataContent(ActionWaitInstallerComplete{
		NodeActionStandardParam: param.NodeActionStandardParam,
		EnsureAgentID:           false,
	}); err != nil {
		return fmt.Errorf("failed to update instance data content: %w", err)
	}

	// exec uninstall command
	if std.DeployInfo().Host.Dynamic.NodeOsType == criteria.OSWindows {
		return act.doUninstallWindows(std, uninstallParams)
	}

	return act.doUninstallUnix(std, uninstallParams)
}

// nolint: perfsprint
func (act *actionUninstallPagent) doUninstallUnix(std *nodeUtils.NodeActionStandarder, param *UninstallParams) error {
	installerPath := path.Clean(path.Join(param.InstallerWorkDir, param.InstallerName))

	args := []string{
		fmt.Sprintf("--deploy_env %s", system.GetEnv()),
		fmt.Sprintf("--generation %d", param.Generation),
		fmt.Sprintf("--node_role %s", param.NodeRole),
		fmt.Sprintf("--base_work_dir %s", param.BaseWorkDir),
		fmt.Sprintf("--base_deploy_dir %s", param.BaseDeployDir),
		fmt.Sprintf("--cbsvr_addr %s", param.CallbackSvrAddr),
		fmt.Sprintf("--deploy_token %s", param.DeployToken),
		fmt.Sprintf("--oper_inst_id %s", param.OperInstID),
	}
	if len(param.AdditionArgs) > 0 {
		args = append(args, param.AdditionArgs...)
	}

	uninstallLogPath := path.Clean(fmt.Sprintf("%s.stdout", installerPath))
	uninstallCmd := fmt.Sprintf("chmod +x %s && %s %s %s >%s 2>&1 &",
		installerPath, installerPath, installer.NodeCmdFullUninstall, strings.Join(args, " "), uninstallLogPath)
	std.InstanceData().Log().
		Zh("卸载节点命令: %s", uninstallCmd).
		En("uninstall node cmd: %s", uninstallCmd).
		Info()

	taskID, err := act.gseHandler.ExecuteScript(std.Context(),
		types.ScriptTypeBash,
		fmt.Sprintf(
			`mkdir -p %s && cd %s && echo "%s" > uninstall.sh && sh uninstall.sh`,
			param.InstallerWorkDir,
			param.InstallerWorkDir,
			uninstallCmd),
		uninstallPagentScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: param.AgentID,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to execute uninstall script: %w", err)
	}
	std.InstanceData().Log().
		Zh("卸载节点 task-id: %s", taskID).
		En("uninstall node task-id: %s", taskID).
		Info()

	return nil
}

// nolint: perfsprint
func (act *actionUninstallPagent) doUninstallWindows(std *nodeUtils.NodeActionStandarder, param *UninstallParams) error {
	installerPath := winpath.Clean(winpath.Join(param.InstallerWorkDir, param.InstallerName))

	args := []string{
		fmt.Sprintf("--deploy_env %s", system.GetEnv()),
		fmt.Sprintf("--generation %d", param.Generation),
		fmt.Sprintf("--node_role %s", param.NodeRole),
		fmt.Sprintf("--base_work_dir %s", param.BaseWorkDir),
		fmt.Sprintf("--base_deploy_dir %s", param.BaseDeployDir),
		fmt.Sprintf("--cbsvr_addr %s", param.CallbackSvrAddr),
		fmt.Sprintf("--deploy_token %s", param.DeployToken),
		fmt.Sprintf("--oper_inst_id %s", param.OperInstID),
	}
	if len(param.AdditionArgs) > 0 {
		args = append(args, param.AdditionArgs...)
	}

	uninstallLogPath := winpath.Clean(fmt.Sprintf("%s.stdout", installerPath))
	uninstallCmd := fmt.Sprintf("%s %s %s >%s 2>&1",
		installerPath, installer.NodeCmdFullUninstall, strings.Join(args, " "), uninstallLogPath)
	std.InstanceData().Log().
		Zh("卸载节点命令: %s", uninstallCmd).
		En("uninstall node cmd: %s", uninstallCmd).
		Info()

	taskID, err := act.gseHandler.ExecuteScript(std.Context(),
		types.ScriptTypeBat,
		fmt.Sprintf(
			`cd %s && %s`,
			param.InstallerWorkDir,
			uninstallCmd),
		uninstallPagentScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: param.AgentID,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to execute uninstall script: %w", err)
	}
	std.InstanceData().Log().
		Zh("卸载节点 task-id: %s", taskID).
		En("uninstall node task-id: %s", taskID).
		Info()

	return nil
}
