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
	// ActionNameRestartNode defines the action name.
	ActionNameRestartNode = "restart_node"
)

// NewActionRestartNode get a new action.
func NewActionRestartNode(capability *Capability) action.Definition {
	return &actionRestartNode{
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		gseHandler:            capability.GSEHandler,
	}
}

// ActionParamRestartNode defines the action param.
type ActionParamRestartNode struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
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
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	gseHandler            gse.IHandler
}

// Name returns the name of the action.
func (act *actionRestartNode) Name() string {
	return ActionNameRestartNode
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionRestartNode) DisplayNameZh() string {
	return "重启节点"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionRestartNode) DisplayNameEn() string {
	return "Restart Node"
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
// nolint: funlen,nonamedreturns
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionRestartNode) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamRestartNode)
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

	// check if this node version is >= lowest version which supports the soft restart through cluster.
	// proxy node do not support soft restart.
	if std.DeployInfo().CurrentVersionSupports.OperateAgentRestart && std.DeployInfo().Host.Dynamic.NodeRole == types.NodeRoleAgent {
		return act.restartThroughCluster(std, std.DeployInfo())
	}

	if !std.DeployInfo().RestartOptions.ForceRestart {
		return errors.New("current node version do not support soft restart")
	}

	return act.restartThroughCommand(std, std.DeployInfo())
}

func (act *actionRestartNode) restartThroughCluster(std *nodeUtils.NodeActionStandarder, info *types.DeploymentInfo) error {
	result, err := act.gseHandler.OperateAgent(std.Context(), types.OperateAgent{
		Type:                   types.OperateAgentTypeRestart,
		CurrentAgentVersion:    "",
		TargetAgentVersionSign: "",
		Timeout:                info.RestartOptions.GracefulRestartTimeout,
		Force:                  info.RestartOptions.ForceRestart,
		Remark:                 "restart by nodemgr: " + std.InstanceData().OperationInstanceID,
	}, info.Host.Dynamic.AgentID)
	if err != nil {
		return fmt.Errorf("failed to operate agent for restarting: %w", err)
	}

	if len(result.MissingAgentIDs) > 0 {
		return fmt.Errorf("failed to operate agent for restarting. not-available-agent-ids(%v)", result.MissingAgentIDs)
	}
	std.InstanceData().Log().
		Zh("通过集群操作代理重启节点，agent-id(%s)，强制(%t)，超时(%.2fs)",
			info.Host.Dynamic.AgentID, info.RestartOptions.ForceRestart, info.RestartOptions.GracefulRestartTimeout.Seconds()).
		En("restart node through operating agent with cluster. agent-id(%s), force(%t), timeout(%.2fs)",
			info.Host.Dynamic.AgentID, info.RestartOptions.ForceRestart, info.RestartOptions.GracefulRestartTimeout.Seconds()).
		Info()

	return nil
}

func (act *actionRestartNode) restartThroughCommand(std *nodeUtils.NodeActionStandarder, info *types.DeploymentInfo) error {
	// select matching tools.
	toolName, err := tool.FormatInstallerName(info.Host.Dynamic.NodeOsType, info.Host.Dynamic.NodeCPUArch)
	if err != nil {
		return err
	}

	deployConstant, err := deployconstant.GetNodeDeployConf(info.Host.Dynamic.NodeGeneration, info.Host.Dynamic.NodeOsType)
	if err != nil {
		return fmt.Errorf("failed to get deploy constant: %w", err)
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
		return act.restartThroughCommandWindows(std, restartParams)
	}

	return act.restartThroughCommandUnix(std, restartParams)
}

// nolint: perfsprint
func (act *actionRestartNode) restartThroughCommandUnix(std *nodeUtils.NodeActionStandarder, param *RestartParams) error {
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
	restartCmd := fmt.Sprintf("chmod +x %s && %s %s %s >%s 2>&1 &",
		installerPath, installerPath, installer.NodeCmdStepRestart, strings.Join(args, " "), restartLogPath)
	std.InstanceData().Log().
		Zh("重启命令: %s", restartCmd).
		En("restart cmd: %s", restartCmd).
		Info()

	taskID, err := act.gseHandler.ExecuteScript(std.Context(),
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
	std.InstanceData().Log().
		Zh("重启节点任务ID: %s", taskID).
		En("restart node task-id: %s", taskID).
		Info()

	return nil
}

// nolint: perfsprint
func (act *actionRestartNode) restartThroughCommandWindows(std *nodeUtils.NodeActionStandarder, param *RestartParams) error {
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
	restartCmd := fmt.Sprintf("%s %s %s >%s 2>&1",
		installerPath, installer.NodeCmdStepRestart, strings.Join(args, " "), restartLogPath)
	std.InstanceData().Log().
		Zh("重启命令: %s", restartCmd).
		En("restart cmd: %s", restartCmd).
		Info()

	taskID, err := act.gseHandler.ExecuteScript(std.Context(),
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
	std.InstanceData().Log().
		Zh("重启节点任务ID: %s", taskID).
		En("restart node task-id: %s", taskID).
		Info()

	return nil
}
