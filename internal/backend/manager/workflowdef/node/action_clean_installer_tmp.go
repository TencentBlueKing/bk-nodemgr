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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
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

// CleanParams this struct defines the parameters for clean installer temp files.
type CleanParams struct {
	AgentID          string
	InstallerName    string
	InstallerWorkDir string
	Generation       types.Generation
	NodeRole         types.NodeRole
	BaseWorkDir      string
	BaseDeployDir    string
	AdditionArgs     []string
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
func (act *actionCleanInstaller) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: funlen,nonamedreturns
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionCleanInstaller) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamCleanInstaller)
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

	// select matching tools.
	toolName, err := tool.FormatInstallerName(std.DeployInfo().Host.Dynamic.NodeOsType, std.DeployInfo().Host.Dynamic.NodeCPUArch)
	if err != nil {
		return err
	}

	deployConstant, err := deployconstant.GetNodeDeployConf(std.DeployInfo().Host.Dynamic.NodeGeneration, std.DeployInfo().Host.Dynamic.NodeOsType)
	if err != nil {
		return fmt.Errorf("failed to get deploy constant: %w", err)
	}

	cleanParams := &CleanParams{
		AgentID:          std.DeployInfo().Host.Dynamic.AgentID,
		InstallerName:    toolName,
		InstallerWorkDir: std.DeployInfo().InstallerWorkDir,
		Generation:       std.DeployInfo().Host.Dynamic.NodeGeneration,
		NodeRole:         std.DeployInfo().Host.Dynamic.NodeRole,
		BaseWorkDir:      deployConstant.BaseWorkDir,
		BaseDeployDir:    deployConstant.BaseDeployDir,
	}

	// exec upgrade command
	if std.DeployInfo().Host.Dynamic.NodeOsType == criteria.OSWindows {
		return act.doCleanWindows(std, cleanParams)
	}

	return act.doCleanUnix(std, cleanParams)
}

// nolint: perfsprint
func (act *actionCleanInstaller) doCleanUnix(std *nodeUtils.NodeActionStandarder, param *CleanParams) error {
	installerPath := path.Clean(path.Join(param.InstallerWorkDir, param.InstallerName))

	args := []string{
		fmt.Sprintf("--deploy_env %s", system.GetEnv()),
		fmt.Sprintf("--generation %d", param.Generation),
		fmt.Sprintf("--node_role %s", param.NodeRole),
		fmt.Sprintf("--base_work_dir %s", param.BaseWorkDir),
		fmt.Sprintf("--base_deploy_dir %s", param.BaseDeployDir),
	}
	if len(param.AdditionArgs) > 0 {
		args = append(args, param.AdditionArgs...)
	}

	cleanLogPath := path.Clean(fmt.Sprintf("%s.stdout", installerPath))
	cleanCmd := fmt.Sprintf("chmod +x %s && %s %s %s >%s 2>&1 &",
		installerPath, installerPath, installer.NodeCmdStepCleanTmp, strings.Join(args, " "), cleanLogPath)
	std.InstanceData().LogI("clean installer cmd: " + cleanCmd)

	taskID, err := act.gseHandler.ExecuteScript(std.Context(),
		types.ScriptTypeBash,
		fmt.Sprintf(
			`mkdir -p %s && cd %s && echo "%s" > clean.sh && sh clean.sh`,
			param.InstallerWorkDir,
			param.InstallerWorkDir,
			cleanCmd),
		cleanScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: param.AgentID,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to execute clean installer script: %w", err)
	}
	std.InstanceData().LogI("clean installer task-id: " + taskID)

	return nil
}

// nolint: perfsprint
func (act *actionCleanInstaller) doCleanWindows(std *nodeUtils.NodeActionStandarder, param *CleanParams) error {
	installerPath := winpath.Clean(winpath.Join(param.InstallerWorkDir, param.InstallerName))

	args := []string{
		fmt.Sprintf("--deploy_env %s", system.GetEnv()),
		fmt.Sprintf("--generation %d", param.Generation),
		fmt.Sprintf("--node_role %s", param.NodeRole),
		fmt.Sprintf("--base_work_dir %s", param.BaseWorkDir),
		fmt.Sprintf("--base_deploy_dir %s", param.BaseDeployDir),
	}
	if len(param.AdditionArgs) > 0 {
		args = append(args, param.AdditionArgs...)
	}

	cleanLogPath := path.Clean(fmt.Sprintf("%s.stdout", installerPath))
	cleanCmd := fmt.Sprintf("%s %s %s >%s 2>&1",
		installerPath, installer.NodeCmdStepCleanTmp, strings.Join(args, " "), cleanLogPath)
	std.InstanceData().LogI("clean installer cmd: " + cleanCmd)

	taskID, err := act.gseHandler.ExecuteScript(std.Context(),
		types.ScriptTypeBat,
		fmt.Sprintf(
			`cd %s && %s`,
			param.InstallerWorkDir,
			cleanCmd),
		cleanScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: param.AgentID,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to execute clean installer script: %w", err)
	}
	std.InstanceData().LogI("clean installer task-id: " + taskID)

	return nil
}
