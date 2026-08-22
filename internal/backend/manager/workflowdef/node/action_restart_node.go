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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
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
		storageActionInstance: capability.StorageWorkflow,
	}
}

// ActionParamRestartNode defines the action param.
type ActionParamRestartNode struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionRestartNode struct {
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	gseHandler            gse.IHandler
	storageActionInstance workflow.IStorageActionInstance
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
func (act *actionRestartNode) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: funlen,nonamedreturns
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionRestartNode) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActionParamRestartNode)
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

	// Capture pre-restart state for verification in wait_gse_ready action.
	var preRestartNodeStartTime uint64
	if captureErr := act.capturePreRestartNodeStartTime(std, &preRestartNodeStartTime); captureErr != nil {
		std.InstanceData().Log().
			Zh("重启前查询 Agent 启动时间失败，将回退为仅状态校验: %s", captureErr.Error()).
			En("failed to query pre-restart agent start time, fallback to state-only verification: %s", captureErr.Error()).
			Warn()
	}

	// check if this node version is >= lowest version which supports the soft restart through cluster.
	// proxy node do not support soft restart.
	if std.DeployInfo().CurrentVersionSupports.OperateAgentRestart && std.DeployInfo().Host.Dynamic.NodeRole == types.NodeRoleAgent {
		if err := act.restartThroughCluster(std, std.DeployInfo()); err != nil {
			return err
		}
	} else {
		if !std.DeployInfo().RestartOptions.ForceRestart {
			return errors.New("current node version do not support soft restart")
		}

		if err := act.restartThroughCommand(std, std.DeployInfo()); err != nil {
			return err
		}
	}

	if err = act.storageActionInstance.UpsertActionInstancePrivateData(std.Context(),
		std.InstanceData().OperationInstanceID,
		ActionNameWaitGseReady,
		map[string]any{
			types.PDKeyPreRestartNodeStartTimeSec: preRestartNodeStartTime,
			types.PDKeyRestartCommandIssueTimeSec: time.Now().Unix(),
		}); err != nil {
		return fmt.Errorf("failed to save wait gse ready private data: %w", err)
	}

	return nil
}

func (act *actionRestartNode) capturePreRestartNodeStartTime(
	std *nodeUtils.NodeActionStandarder,
	preRestartNodeStartTime *uint64,
) error {

	agentInfos, err := act.gseHandler.ListAgentInfo(std.Context(), std.DeployInfo().Host.Dynamic.AgentID)
	if err != nil {
		return fmt.Errorf("failed to list agent info before restart: %w", err)
	}

	if len(agentInfos) != 1 {
		return fmt.Errorf("query agent info result no 1, agent_infos(%v)", agentInfos)
	}

	*preRestartNodeStartTime = agentInfos[0].StartTime
	std.InstanceData().Log().
		Zh("记录重启前 Agent 启动时间: %d", *preRestartNodeStartTime).
		En("record pre-restart agent start time: %d", *preRestartNodeStartTime).
		Info()

	return nil
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
		Zh("通过集群操作代理重启节点, agent-id(%s), 强制(%t), 超时(%.2fs)",
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

	restartParams := &installer.NodeStepRestartParams{
		NodeCommonParams: installer.NodeCommonParams{
			DeployEnv:     system.GetEnv(),
			Generation:    int(info.Host.Dynamic.NodeGeneration),
			NodeRole:      string(info.Host.Dynamic.NodeRole),
			BaseWorkDir:   info.InstallerRuntime.BaseWorkDir,
			BaseDeployDir: info.BaseRuntime.BaseDeployDir,
		},
		InstallWorkDir:    info.InstallerRuntime.WorkDir,
		InstallerFileName: toolName,
		Force:             true,
	}

	// exec upgrade command
	if info.Host.Dynamic.NodeOsType == criteria.OSWindows {
		return act.restartThroughCommandWindows(std, restartParams)
	}

	return act.restartThroughCommandUnix(std, restartParams)
}

// nolint: perfsprint
func (act *actionRestartNode) restartThroughCommandUnix(std *nodeUtils.NodeActionStandarder, param *installer.NodeStepRestartParams) error {
	_, restartCmd, err := param.ToUnixScript()
	if err != nil {
		return fmt.Errorf("failed to render node restart script: %w", err)
	}
	std.InstanceData().Log().
		Zh("重启命令: %s", restartCmd).
		En("restart cmd: %s", restartCmd).
		Info()

	taskID, err := act.gseHandler.ExecuteScript(std.Context(),
		types.ScriptTypeBash,
		fmt.Sprintf(
			`mkdir -p %s && cd %s && echo "%s" > restart.sh && sh restart.sh`,
			param.InstallWorkDir,
			param.InstallWorkDir,
			restartCmd),
		cleanScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: std.DeployInfo().Host.Dynamic.AgentID,
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
func (act *actionRestartNode) restartThroughCommandWindows(std *nodeUtils.NodeActionStandarder, param *installer.NodeStepRestartParams) error {
	_, restartCmd, err := param.ToWindowsScript()
	if err != nil {
		return fmt.Errorf("failed to render node restart script: %w", err)
	}
	std.InstanceData().Log().
		Zh("重启命令: %s", restartCmd).
		En("restart cmd: %s", restartCmd).
		Info()

	taskID, err := act.gseHandler.ExecuteScript(std.Context(),
		types.ScriptTypeBat,
		fmt.Sprintf(
			`cd %s && %s`,
			param.InstallWorkDir,
			restartCmd),
		cleanScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: std.DeployInfo().Host.Dynamic.AgentID,
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
