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
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/retrier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameWaitGseReady defines the action name.
	ActionNameWaitGseReady = "wait_gse_ready"
)

// NewActionWaitGseReady get a new action.
func NewActionWaitGseReady(capability *Capability) action.Definition {
	return &actionWaitGseReady{
		gseClient:             capability.GSEHandler,
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
	}
}

// ActParamWaitGseReady defines the action param for waiting GSE ready.
// Restart timing fields are loaded from the wait_gse_ready action private data.
type ActParamWaitGseReady struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionWaitGseReady struct {
	gseClient             gse.IHandler
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
}

// Name returns the name of the action.
func (act *actionWaitGseReady) Name() string {
	return ActionNameWaitGseReady
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionWaitGseReady) DisplayNameZh() string {
	return "等待 GSE 就绪"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionWaitGseReady) DisplayNameEn() string {
	return "Wait for GSE Ready"
}

// Version returns the version of the action.
func (act *actionWaitGseReady) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionWaitGseReady) Description() string {
	return "wait gse ready"
}

// Timeout returns the timeout of the action.
func (act *actionWaitGseReady) Timeout() time.Duration {
	// notice: in this action, the timeout should be equal or grander than the timeout of the workflow,
	// so we set it to 24 hours.
	return 24 * time.Hour // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actionWaitGseReady) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionWaitGseReady) MaxRetryCount() uint {
	// notice: this action is an polling action should not auto retry.
	return 0
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionWaitGseReady) DelayFn() func() {
	return func() {
		logger.G.Sys().With("action", act.Name()).Error("this action should not auto retry")
	}
}

// Do this func define what the action will do.
func (act *actionWaitGseReady) Do(ctx *action.InstanceContext) error {
	param := new(ActParamWaitGseReady)
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

	preRestartNodeStartTimeSec := conv.ToInt64Default(
		std.InstanceData().PrivateData[types.PDKeyPreRestartNodeStartTimeSec], 0)
	preRestartNodeStartTime := time.Unix(preRestartNodeStartTimeSec, 0)

	restartCommandIssueTimeSec := conv.ToInt64Default(
		std.InstanceData().PrivateData[types.PDKeyRestartCommandIssueTimeSec], 0)
	restartCommandIssueTime := time.Unix(restartCommandIssueTimeSec, 0)

	// Log restart-related fields for debugging.
	gracefulRestartTimeout := std.DeployInfo().RestartOptions.GracefulRestartTimeout
	std.InstanceData().Log().
		Zh("发起重启时间(%v), 无损等待超时(%v), 节点上一次启动时间(%v)",
			restartCommandIssueTime.Local(), gracefulRestartTimeout, preRestartNodeStartTime.Local()).
		En("restart issued time(%v), graceful timeout(%v), node last start time(%v)",
			restartCommandIssueTime.Local(), gracefulRestartTimeout, preRestartNodeStartTime.Local()).
		Info()

	// Create a context with GracefulRestartTimeout if applicable.
	// This ensures polling stops automatically when the restart timeout is exceeded.
	pollingCtx := std.Context()
	if canFastFailWithRestartTimeout(restartCommandIssueTimeSec, gracefulRestartTimeout) {
		var cancel context.CancelFunc
		pollingCtx, cancel = contextx.WithDeadline(std.Context(), restartCommandIssueTime.Add(gracefulRestartTimeout))
		defer cancel()
	}

	polling := retrier.NewPolling(retrier.PollingOpts{
		Timeout:  act.Timeout(),
		Interval: time.Second,
	})
	err = polling.Do(pollingCtx, func(_ int) error {
		states, err := act.gseClient.ListAgentState(pollingCtx, std.DeployInfo().Host.Dynamic.AgentID)
		if err != nil {
			return err
		}

		if len(states) != 1 {
			return fmt.Errorf("query agent state result no 1, states(%v)", states)
		}

		state := states[0]
		if err := ensureStateReady(std, state); err != nil {
			return err
		}

		return act.verifyRestartStartTime(pollingCtx, std, state, preRestartNodeStartTimeSec, preRestartNodeStartTime)
	})
	if err != nil {
		// Provide clearer error message when restart timeout was exceeded.
		if errors.Is(err, context.DeadlineExceeded) {
			elapsed := time.Since(restartCommandIssueTime)
			err = fmt.Errorf(
				"restart verification timeout exceeded: pre-start-time(%v), graceful-timeout(%s), elapsed(%s)",
				preRestartNodeStartTime.Local(),
				gracefulRestartTimeout,
				elapsed,
			)

			std.InstanceData().Log().
				Zh("等待 Agent 就绪超时: %s", err.Error()).
				En("wait agent ready timeout: %s", err.Error()).
				Warn()

			return err
		}

		std.InstanceData().Log().
			Zh("查询 Agent 状态失败: %s", err.Error()).
			En("failed to query agent state: %s", err.Error()).
			Warn()

		return err
	}

	return nil
}

func ensureStateReady(std *nodeUtils.NodeActionStandarder, state *types.AgentState) error {
	std.DeployInfo().Host.Dynamic.NodeStatus = state.NodeStatus
	if state.NodeStatus != types.NodeStatusRunning {
		return fmt.Errorf("agent state is not running, status(%s)", state.NodeStatus)
	}

	if state.Version != std.DeployInfo().Host.Dynamic.NodeVersion {
		return fmt.Errorf("agent version is not match, version(%s)", state.Version)
	}

	return nil
}

func (act *actionWaitGseReady) verifyRestartStartTime(nCtx contextx.IContext, std *nodeUtils.NodeActionStandarder,
	state *types.AgentState, preRestartNodeStartTimeSec int64, preRestartNodeStartTime time.Time) error {

	if preRestartNodeStartTimeSec == 0 {
		std.InstanceData().Log().
			Zh("未记录重启前启动时间，沿用原有状态校验通过路径").
			En("pre-restart start time is absent, using legacy state-only verification path").
			Info()

		return nil
	}

	agentInfos, err := act.gseClient.ListAgentInfo(nCtx, std.DeployInfo().Host.Dynamic.AgentID)
	if err != nil {
		return err
	}

	if len(agentInfos) != 1 {
		return fmt.Errorf("query agent info result no 1, agent_infos(%v)", agentInfos)
	}

	agentInfo := agentInfos[0]
	agentStartTime := time.Unix(int64(agentInfo.StartTime), 0)
	if !agentStartTime.After(preRestartNodeStartTime) {
		return fmt.Errorf(
			"agent start time has not advanced yet, pre-start-time(%v), current-start-time(%d)",
			preRestartNodeStartTime.Local(),
			agentInfo.StartTime,
		)
	}

	startTime := time.Unix(int64(agentInfo.StartTime), 0)
	std.InstanceData().Log().
		Zh("查询到 Agent 状态为运行中, 版本(%s), 启动时间(%v -> %v)",
			state.Version, preRestartNodeStartTime.Local(), startTime.Local()).
		En("found agent state is running, version(%s), start-time(%v -> %v)",
			state.Version, preRestartNodeStartTime.Local(), startTime.Local()).
		Info()

	return nil
}

func canFastFailWithRestartTimeout(restartCommandIssueTimeSec int64, gracefulRestartTimeout time.Duration) bool {
	return restartCommandIssueTimeSec != 0 && gracefulRestartTimeout > 0
}
