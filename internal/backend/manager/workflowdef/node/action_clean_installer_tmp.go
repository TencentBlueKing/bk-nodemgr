/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
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
	// ActionNameCleanInstaller defines the action name.
	ActionNameCleanInstaller = "clean_installer"

	cleanScriptTimeout = 1 * time.Minute
)

// NewActionCleanInstaller get a new action.
func NewActionCleanInstaller(capability *Capability) action.Definition {
	return &actionCleanInstaller{
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		gseHandler:            capability.GSEHandler,
	}
}

// ActionParamCleanInstaller defines the action param.
type ActionParamCleanInstaller struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionCleanInstaller struct {
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	gseHandler            gse.IHandler
}

// Name returns the name of the action.
func (act *actionCleanInstaller) Name() string {
	return ActionNameCleanInstaller
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionCleanInstaller) DisplayNameZh() string {
	return "清理安装临时文件"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionCleanInstaller) DisplayNameEn() string {
	return "Clean Installer Temp Files"
}

// Version returns the version of the action.
func (act *actionCleanInstaller) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionCleanInstaller) Description() string {
	return "clean installer"
}

// Timeout returns the timeout of the action.
func (act *actionCleanInstaller) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionCleanInstaller) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionCleanInstaller) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionCleanInstaller) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: funlen,nonamedreturns
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionCleanInstaller) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActionParamCleanInstaller)
	err = conv.MapToStruct(ctx.Data.Content, param)
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

	// select matching tools.
	toolName, err := tool.FormatInstallerName(std.DeployInfo().Host.Dynamic.NodeOsType, std.DeployInfo().Host.Dynamic.NodeCPUArch)
	if err != nil {
		return err
	}

	cleanParams := &installer.NodeStepCleanTmpParams{
		NodeCommonParams: installer.NodeCommonParams{
			DeployEnv:     system.GetEnv(),
			Generation:    int(std.DeployInfo().Host.Dynamic.NodeGeneration),
			NodeRole:      string(std.DeployInfo().Host.Dynamic.NodeRole),
			BaseWorkDir:   std.DeployInfo().InstallerRuntime.BaseWorkDir,
			BaseDeployDir: std.DeployInfo().BaseRuntime.BaseDeployDir,
		},
		InstallWorkDir:    std.DeployInfo().InstallerRuntime.WorkDir,
		InstallerFileName: toolName,
	}

	// exec upgrade command
	if std.DeployInfo().Host.Dynamic.NodeOsType == criteria.OSWindows {
		return act.doCleanWindows(std, cleanParams)
	}

	return act.doCleanUnix(std, cleanParams)
}

// nolint: perfsprint
func (act *actionCleanInstaller) doCleanUnix(std *nodeUtils.NodeActionStandarder, param *installer.NodeStepCleanTmpParams) error {
	_, cleanCmd, err := param.ToUnixScript()
	if err != nil {
		return fmt.Errorf("failed to render node clean tmp script: %w", err)
	}
	std.InstanceData().Log().
		Zh("清理安装器命令: %s", cleanCmd).
		En("clean installer cmd: %s", cleanCmd).
		Info()

	taskID, err := act.gseHandler.ExecuteScript(std.Context(),
		types.ScriptTypeBash,
		fmt.Sprintf(
			`mkdir -p %s && cd %s && echo "%s" > clean.sh && sh clean.sh`,
			param.InstallWorkDir,
			param.InstallWorkDir,
			cleanCmd),
		cleanScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: std.DeployInfo().Host.Dynamic.AgentID,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to execute clean installer script: %w", err)
	}
	std.InstanceData().Log().
		Zh("清理安装器任务ID: %s", taskID).
		En("clean installer task-id: %s", taskID).
		Info()

	return nil
}

// nolint: perfsprint
func (act *actionCleanInstaller) doCleanWindows(std *nodeUtils.NodeActionStandarder, param *installer.NodeStepCleanTmpParams) error {
	_, cleanCmd, err := param.ToWindowsScript()
	if err != nil {
		return fmt.Errorf("failed to render node clean tmp script: %w", err)
	}
	std.InstanceData().Log().
		Zh("清理安装器命令: %s", cleanCmd).
		En("clean installer cmd: %s", cleanCmd).
		Info()

	taskID, err := act.gseHandler.ExecuteScript(std.Context(),
		types.ScriptTypeBat,
		fmt.Sprintf(
			`cd %s && %s`,
			param.InstallWorkDir,
			cleanCmd),
		cleanScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: std.DeployInfo().Host.Dynamic.AgentID,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to execute clean installer script: %w", err)
	}
	std.InstanceData().Log().
		Zh("清理安装器任务ID: %s", taskID).
		En("clean installer task-id: %s", taskID).
		Info()

	return nil
}
