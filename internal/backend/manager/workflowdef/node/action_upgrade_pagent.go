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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
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
		storageHost:           capability.StorageTopo,
		storageNetworkUnit:    capability.StorageTopo,
		provider:              capability.DiscoverProvider,
		gseHandler:            capability.GSEHandler,
		storageActionInstance: capability.StorageWorkflow,
	}
}

// ActionParamUpgradePagent defines the action param.
type ActionParamUpgradePagent struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionUpgradePagent struct {
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	storageNetworkUnit    topoStg.IStorageNetworkUnit
	provider              discover.Provider
	gseHandler            gse.IHandler
	storageActionInstance workflow.IStorageActionInstance
}

// Name returns the name of the action.
func (act *actionUpgradePagent) Name() string {
	return ActionNameUpgradePagent
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionUpgradePagent) DisplayNameZh() string {
	return "升级 P-Agent"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionUpgradePagent) DisplayNameEn() string {
	return "Upgrade P-Agent"
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

	// get upgrade params.
	toolName, err := tool.FormatInstallerName(std.DeployInfo().Host.Dynamic.NodeOsType, std.DeployInfo().Host.Dynamic.NodeCPUArch)
	if err != nil {
		return err
	}

	callbackEndpoints, downloadEndpoints, err := nodeUtils.GenerateNodeInstallerServerEndpoints(
		std, act.provider, nodeUtils.NodeInstallerEndpointSourceRelay)
	if err != nil {
		return fmt.Errorf("failed to generate node installer server endpoints: %w", err)
	}

	upgradeParams := &installer.NodeUpgradeParams{
		NodeCommonParams: installer.NodeCommonParams{
			DeployEnv:     system.GetEnv(),
			Generation:    int(std.DeployInfo().Host.Dynamic.NodeGeneration),
			NodeRole:      string(std.DeployInfo().Host.Dynamic.NodeRole),
			BaseWorkDir:   std.DeployInfo().InstallerRuntime.BaseWorkDir,
			BaseDeployDir: std.DeployInfo().BaseRuntime.BaseDeployDir,
		},
		InstallWorkDir:    std.DeployInfo().InstallerRuntime.WorkDir,
		InstallerFileName: toolName,
		DownloadSvrAddr:   nodeUtils.BuildServerURLs(downloadEndpoints...),
		CallbackSvrAddr:   nodeUtils.BuildServerURLs(callbackEndpoints...),
		DeployToken:       std.Token(),
		NodeVersion:       std.DeployInfo().Host.Dynamic.NodeVersion,
		OperInstID:        std.InstanceData().OperationInstanceID,
	}

	std.InstanceData().Log().
		Zh("构建升级参数成功. params(%v)", upgradeParams).
		En("build upgrade params success. params(%v)", upgradeParams).
		Info()

	err = saveWaitInstallerPrivateData(
		std.Context(), act.storageActionInstance, std.InstanceData().OperationInstanceID,
		false, true)
	if err != nil {
		return err
	}

	// exec upgrade command
	if std.DeployInfo().Host.Dynamic.NodeOsType == criteria.OSWindows {
		return act.doUpgradeWindows(std, upgradeParams)
	}

	return act.doUpgradeUnix(std, upgradeParams)
}

func (act *actionUpgradePagent) doUpgradeUnix(std *nodeUtils.NodeActionStandarder, param *installer.NodeUpgradeParams) error {
	_, upgradeCmd, err := param.ToUnixScript()
	if err != nil {
		return fmt.Errorf("failed to render node upgrade script: %w", err)
	}

	std.InstanceData().Log().
		Zh("升级节点命令: %s", upgradeCmd).
		En("upgrade node command: %s", upgradeCmd).
		Info()

	taskID, err := act.gseHandler.ExecuteScript(std.Context(),
		types.ScriptTypeBash,
		fmt.Sprintf(
			`mkdir -p %s && cd %s && echo "%s" > upgrade.sh && sh upgrade.sh`,
			param.InstallWorkDir,
			param.InstallWorkDir,
			upgradeCmd),
		upgradeScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: std.DeployInfo().Host.Dynamic.AgentID,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to execute upgrade script: %w", err)
	}

	std.InstanceData().Log().
		Zh("升级节点 task id: %s", taskID).
		En("upgrade node task id: %s", taskID).
		Info()

	return nil
}

func (act *actionUpgradePagent) doUpgradeWindows(std *nodeUtils.NodeActionStandarder, param *installer.NodeUpgradeParams) error {
	_, upgradeCmd, err := param.ToWindowsScript()
	if err != nil {
		return fmt.Errorf("failed to render node upgrade script: %w", err)
	}

	std.InstanceData().Log().
		Zh("升级节点命令: %s", upgradeCmd).
		En("upgrade node command: %s", upgradeCmd).
		Info()

	taskID, err := act.gseHandler.ExecuteScript(std.Context(),
		types.ScriptTypeBat,
		fmt.Sprintf(
			`cd %s && %s`,
			param.InstallWorkDir,
			upgradeCmd),
		upgradeScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: std.DeployInfo().Host.Dynamic.AgentID,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to execute upgrade script: %w", err)
	}

	std.InstanceData().Log().
		Zh("升级节点 task id: %s", taskID).
		En("upgrade node task id: %s", taskID).
		Info()

	return nil
}
