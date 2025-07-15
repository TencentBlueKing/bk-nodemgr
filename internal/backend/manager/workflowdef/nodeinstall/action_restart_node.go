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
	// ActionNameRestartNode defines the action name.
	ActionNameRestartNode = "restart_node"
)

// NewActionStartNode get a new action.
func NewActionStartNode(storageNodeDeployment nodedeployment.IStorageNodeDeployment,
	gseHandler gse.IHandler,
	logger logger.Logger) action.Definition {

	return &actionRestartNode{
		storageNodeDeployment: storageNodeDeployment,
		gseHandler:            gseHandler,
		logger:                logger,
	}
}

// ActionParamRestartNode defines the action param.
type ActionParamRestartNode struct {
	Token string `json:"token"`
}

// RestartParams this struct defines the parameters for restarting node through command.
type RestartParams struct {
	AgentID          string
	InstallerName    string
	InstallerWorkDir string
	Generation       types.Generation
	NodeRole         types.NodeRole
	BaseWorkDir      string
	BaseDeployDir    string
	AdditionArgs     []string
}

type actionRestartNode struct {
	storageNodeDeployment nodedeployment.IStorageNodeDeployment
	gseHandler            gse.IHandler
	logger                logger.Logger
}

// Name returns the name of the action.
func (act *actionRestartNode) Name() string {
	return ActionNameRestartNode
}

// Version returns the version of the action.
func (act *actionRestartNode) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionRestartNode) Description() string {
	return "restart node"
}

// Timeout returns the timeout of the action.
func (act *actionRestartNode) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionRestartNode) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionRestartNode) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionRestartNode) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: funlen,fnsize,nonamedreturns
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionRestartNode) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActionParamRestartNode)
	err = conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		err = fmt.Errorf("failed to convert param, err: %w", err)

		return err
	}

	info, err := act.storageNodeDeployment.GetInfo(ctx.Ctx, param.Token)
	if err != nil {
		return err
	}

	// check if this node version is >= lowest version which supports the soft restart through cluster.
	if info.CurrentVersionSupports.OperateAgentRestart {
		return act.restartThroughCluster(ctx, info)
	}

	if !info.UpgradeOptions.ForceRestart {
		return errors.New("current node version do not support soft restart")
	}

	return act.restartThroughCommand(ctx, info)
}

func (act *actionRestartNode) restartThroughCluster(ctx *action.InstanceContext, info *types.DeploymentInfo) error {
	result, err := act.gseHandler.OperateAgent(ctx.Ctx, types.OperateAgent{
		Type:                   types.OperateAgentTypeRestart,
		CurrentAgentVersion:    "",
		TargetAgentVersionSign: "",
		Timeout:                info.UpgradeOptions.GracefulRestartTimeout,
		Force:                  info.UpgradeOptions.ForceRestart,
		Remark:                 "restart by nodemgr: " + ctx.Data.OperationInstanceID,
	}, info.Host.Dynamic.AgentID)
	if err != nil {
		return fmt.Errorf("failed to operate agent for restarting: %w", err)
	}

	if len(result.MissingAgentIDs) > 0 {
		return fmt.Errorf("failed to operate agent for restarting. not-available-agent-ids(%v)", result.MissingAgentIDs)
	}
	ctx.Data.LogI(fmt.Sprintf("restart node through operating agent with cluster. agent-id(%s), force(%t), timeout(%.2fs)",
		info.Host.Dynamic.AgentID, info.UpgradeOptions.ForceRestart, info.UpgradeOptions.GracefulRestartTimeout.Seconds()))

	return nil
}

func (act *actionRestartNode) restartThroughCommand(ctx *action.InstanceContext, info *types.DeploymentInfo) error {
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

	restartParams := &RestartParams{
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
		return act.restartThroughCommandWindows(ctx, restartParams)
	}

	return act.restartThroughCommandUnix(ctx, restartParams)
}

func (act *actionRestartNode) restartThroughCommandUnix(ctx *action.InstanceContext, param *RestartParams) error {
	installerPath := path.Clean(path.Join(param.InstallerWorkDir, param.InstallerName))
	args := []string{
		fmt.Sprintf("--deploy_env %s", system.GetEnv()),
		fmt.Sprintf("--generation %d", param.Generation),
		fmt.Sprintf("--node_role %s", param.NodeRole),
		fmt.Sprintf("--base_work_dir %s", param.BaseWorkDir),
		fmt.Sprintf("--base_deploy_dir %s", param.BaseDeployDir),
		"--force",
	}
	if len(param.AdditionArgs) > 0 {
		args = append(args, param.AdditionArgs...)
	}

	restartLogPath := path.Clean(fmt.Sprintf("%s.stdout", installerPath))
	restartCmd := fmt.Sprintf("chmod +x %s && %s step restart %s >%s 2>&1 &",
		installerPath, installerPath, strings.Join(args, " "), restartLogPath)
	ctx.Data.LogI("restart cmd: " + restartCmd)

	taskID, err := act.gseHandler.ExecuteScript(ctx.Ctx,
		types.ScriptTypeBash,
		fmt.Sprintf(
			`mkdir -p %s && cd %s && echo "%s" > restart.sh && sh restart.sh`,
			param.InstallerWorkDir,
			param.InstallerWorkDir,
			restartCmd),
		cleanScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: param.AgentID,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to execute restart node script: %w", err)
	}
	ctx.Data.LogI("restart node task-id: " + taskID)

	return nil
}

func (act *actionRestartNode) restartThroughCommandWindows(ctx *action.InstanceContext, param *RestartParams) error {
	installerPath := winpath.Clean(winpath.Join(param.InstallerWorkDir, param.InstallerName))
	args := []string{
		fmt.Sprintf("--deploy_env %s", system.GetEnv()),
		fmt.Sprintf("--generation %d", param.Generation),
		fmt.Sprintf("--node_role %s", param.NodeRole),
		fmt.Sprintf("--base_work_dir %s", param.BaseWorkDir),
		fmt.Sprintf("--base_deploy_dir %s", param.BaseDeployDir),
		"--force",
	}
	if len(param.AdditionArgs) > 0 {
		args = append(args, param.AdditionArgs...)
	}

	restartLogPath := winpath.Clean(fmt.Sprintf("%s.stdout", installerPath))
	restartCmd := fmt.Sprintf("%s step restart %s >%s 2>&1",
		installerPath, strings.Join(args, " "), restartLogPath)
	ctx.Data.LogI("restart cmd: " + restartCmd)

	taskID, err := act.gseHandler.ExecuteScript(ctx.Ctx,
		types.ScriptTypeBat,
		fmt.Sprintf(
			`cd %s && %s`,
			param.InstallerWorkDir,
			restartCmd),
		cleanScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: param.AgentID,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to execute restart node script: %w", err)
	}
	ctx.Data.LogI("restart node task-id: " + taskID)

	return nil
}
