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
	"fmt"
	"path"
	"strings"
	"time"

	nodedeployment "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-deployment"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/winpath"
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
func NewActionCleanInstaller(storageNodeDeployment nodedeployment.IStorageNodeDeployment,
	gseHandler gse.IHandler,
	logger logger.ILogger) action.Definition {

	return &actionCleanInstaller{
		storageNodeDeployment: storageNodeDeployment,
		gseHandler:            gseHandler,
		logger:                logger,
	}
}

// ActionParamCleanInstaller defines the action param.
type ActionParamCleanInstaller struct {
	Token string `json:"token"`
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
	storageNodeDeployment nodedeployment.IStorageNodeDeployment
	gseHandler            gse.IHandler
	logger                logger.ILogger
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
func (act *actionCleanInstaller) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActionParamCleanInstaller)
	err = conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		err = fmt.Errorf("failed to convert param, err: %w", err)

		return err
	}

	info, err := act.storageNodeDeployment.GetInfo(ctx.Ctx, param.Token)
	if err != nil {
		return err
	}

	// select matching tools.
	toolName, err := tool.FormatInstallerName(info.Host.Dynamic.NodeOsType, info.Host.Dynamic.NodeCPUArch)
	if err != nil {
		err = fmt.Errorf("failed to format tools name, err: %w", err)

		return err
	}

	deployConstant, err := deployconstant.GetDeployConf(info.Host.Dynamic.NodeGeneration, info.Host.Dynamic.NodeOsType)
	if err != nil {
		return fmt.Errorf("failed to get deploy constant, err: %w", err)
	}

	cleanParams := &CleanParams{
		AgentID:          info.Host.Dynamic.AgentID,
		InstallerName:    toolName,
		InstallerWorkDir: info.InstallerWorkDir,
		Generation:       info.Host.Dynamic.NodeGeneration,
		NodeRole:         info.Host.Dynamic.NodeRole,
		BaseWorkDir:      deployConstant.BaseWorkDir,
		BaseDeployDir:    deployConstant.BaseDeployDir,
	}

	// exec upgrade command
	if info.Host.Dynamic.NodeOsType == criteria.OSWindows {
		return act.doCleanWindows(ctx, cleanParams)
	}

	return act.doCleanUnix(ctx, cleanParams)
}

// nolint: perfsprint
func (act *actionCleanInstaller) doCleanUnix(ctx *action.InstanceContext, param *CleanParams) error {
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
	cleanCmd := fmt.Sprintf("chmod +x %s && %s step clean-tmp %s >%s 2>&1 &",
		installerPath, installerPath, strings.Join(args, " "), cleanLogPath)
	ctx.Data.LogI("clean installer cmd: " + cleanCmd)

	taskID, err := act.gseHandler.ExecuteScript(ctx.Ctx,
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
	ctx.Data.LogI("clean installer task-id: " + taskID)

	return nil
}

// nolint: perfsprint
func (act *actionCleanInstaller) doCleanWindows(ctx *action.InstanceContext, param *CleanParams) error {
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
	cleanCmd := fmt.Sprintf("%s step clean-tmp %s >%s 2>&1",
		installerPath, strings.Join(args, " "), cleanLogPath)
	ctx.Data.LogI("clean installer cmd: " + cleanCmd)

	taskID, err := act.gseHandler.ExecuteScript(ctx.Ctx,
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
	ctx.Data.LogI("clean installer task-id: " + taskID)

	return nil
}
