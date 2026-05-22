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
	"time"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameReconfigPagent defines the action name.
	ActionNameReconfigPagent = "reconfig_pagent"

	reconfigPagentScriptTimeout = 10 * time.Minute
)

// NewActionReconfigPagent get a new action.
func NewActionReconfigPagent(capability *Capability) action.Definition {
	return &actionReconfigPagent{
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		gseHandler:            capability.GSEHandler,
	}
}

// ActionParamReconfigPagent defines the action param.
type ActionParamReconfigPagent struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionReconfigPagent struct {
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	gseHandler            gse.IHandler
}

// Name returns the name of the action.
func (act *actionReconfigPagent) Name() string {
	return ActionNameReconfigPagent
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionReconfigPagent) DisplayNameZh() string {
	return "重新配置 P-Agent"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionReconfigPagent) DisplayNameEn() string {
	return "Reconfigure P-Agent"
}

// Version returns the version of the action.
func (act *actionReconfigPagent) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionReconfigPagent) Description() string {
	return "reconfig pagent through relay callback"
}

// Timeout returns the timeout of the action.
func (act *actionReconfigPagent) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionReconfigPagent) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionReconfigPagent) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionReconfigPagent) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: funlen,nonamedreturns
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionReconfigPagent) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamReconfigPagent)
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

	reconfigParams := &installer.NodeReconfigParams{
		NodeCommonParams: installer.NodeCommonParams{
			DeployEnv:     system.GetEnv(),
			Generation:    int(std.DeployInfo().Host.Dynamic.NodeGeneration),
			NodeRole:      string(std.DeployInfo().Host.Dynamic.NodeRole),
			BaseWorkDir:   std.DeployInfo().InstallerRuntime.BaseWorkDir,
			BaseDeployDir: std.DeployInfo().BaseRuntime.BaseDeployDir,
		},
		InstallWorkDir:    std.DeployInfo().InstallerRuntime.WorkDir,
		InstallerFileName: toolName,
		CallbackSvrAddr:   callbackSvrAddr,
		DeployToken:       std.Token(),
		OperInstID:        std.InstanceData().OperationInstanceID,
	}

	if err := std.UpdateInstanceDataContent(ActionWaitInstallerComplete{
		NodeActionStandardParam: param.NodeActionStandardParam,
		EnsureAgentID:           true,
	}); err != nil {
		return fmt.Errorf("failed to update instance data content: %w", err)
	}

	// exec reconfig command
	if std.DeployInfo().Host.Dynamic.NodeOsType == criteria.OSWindows {
		return act.doReconfigWindows(std, reconfigParams)
	}

	return act.doReconfigUnix(std, reconfigParams)
}

// nolint: perfsprint
func (act *actionReconfigPagent) doReconfigUnix(std *nodeUtils.NodeActionStandarder, param *installer.NodeReconfigParams) error {
	_, reconfigCmd, err := param.ToUnixScript()
	if err != nil {
		return fmt.Errorf("failed to render node reconfig script: %w", err)
	}
	std.InstanceData().Log().
		Zh("重新配置节点命令: %s", reconfigCmd).
		En("reconfig node cmd: %s", reconfigCmd).
		Info()

	taskID, err := act.gseHandler.ExecuteScript(std.Context(),
		types.ScriptTypeBash,
		fmt.Sprintf(
			`mkdir -p %s && cd %s && echo "%s" > reconfig.sh && sh reconfig.sh`,
			param.InstallWorkDir,
			param.InstallWorkDir,
			reconfigCmd),
		reconfigPagentScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: std.DeployInfo().Host.Dynamic.AgentID,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to execute reconfig script: %w", err)
	}
	std.InstanceData().Log().
		Zh("重新配置节点 task-id: %s", taskID).
		En("reconfig node task-id: %s", taskID).
		Info()

	return nil
}

// nolint: perfsprint
func (act *actionReconfigPagent) doReconfigWindows(std *nodeUtils.NodeActionStandarder, param *installer.NodeReconfigParams) error {
	_, reconfigCmd, err := param.ToWindowsScript()
	if err != nil {
		return fmt.Errorf("failed to render node reconfig script: %w", err)
	}
	std.InstanceData().Log().
		Zh("重新配置节点命令: %s", reconfigCmd).
		En("reconfig node cmd: %s", reconfigCmd).
		Info()

	taskID, err := act.gseHandler.ExecuteScript(std.Context(),
		types.ScriptTypeBat,
		fmt.Sprintf(
			`cd %s && %s`,
			param.InstallWorkDir,
			reconfigCmd),
		reconfigPagentScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: std.DeployInfo().Host.Dynamic.AgentID,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to execute reconfig script: %w", err)
	}
	std.InstanceData().Log().
		Zh("重新配置节点 task-id: %s", taskID).
		En("reconfig node task-id: %s", taskID).
		Info()

	return nil
}
