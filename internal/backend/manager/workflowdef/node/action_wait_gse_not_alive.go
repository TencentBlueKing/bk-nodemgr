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
	// ActionNameWaitGseNotAlive defines the action name.
	ActionNameWaitGseNotAlive = "wait_gse_not_alive"
)

// NewActionWaitGseNotAlive get a new action.
func NewActionWaitGseNotAlive(capability *Capability) action.Definition {
	return &actionWaitGseNotAlive{
		gseClient:             capability.GSEHandler,
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
	}
}

// ActParamWaitGseNotAlive defines the action param for waiting GSE not alive.
type ActParamWaitGseNotAlive struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionWaitGseNotAlive struct {
	gseClient             gse.IHandler
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
}

// Name returns the name of the action.
func (act *actionWaitGseNotAlive) Name() string {
	return ActionNameWaitGseNotAlive
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionWaitGseNotAlive) DisplayNameZh() string {
	return "等待 GSE 不存活"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionWaitGseNotAlive) DisplayNameEn() string {
	return "Wait for GSE Not Alive"
}

// Version returns the version of the action.
func (act *actionWaitGseNotAlive) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionWaitGseNotAlive) Description() string {
	return "wait gse not alive"
}

// Timeout returns the timeout of the action.
func (act *actionWaitGseNotAlive) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionWaitGseNotAlive) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionWaitGseNotAlive) MaxRetryCount() uint {
	return 0
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionWaitGseNotAlive) DelayFn(_ int) func() {
	return func() {
		logger.G.Sys().With("action", act.Name()).Error("this action should not auto retry")
	}
}

// Do this func define what the action will do.
func (act *actionWaitGseNotAlive) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamWaitGseNotAlive)
	err = conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	std := nodeUtils.NewNodeActionStandarder(act.storageNodeDeployment, act.storageHost)
	if err = std.Initialize(ctx, param.NodeActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	agentID := std.DeployInfo().Host.Dynamic.AgentID
	var lastStatus types.NodeStatus
	polling := retrier.NewPolling(retrier.PollingOpts{
		Timeout:  act.Timeout(),
		Interval: time.Second,
	})
	err = polling.Do(std.Context(), func(_ int) error {
		states, err := act.gseClient.ListAgentState(std.Context(), agentID)
		if err != nil {
			return err
		}

		if len(states) != 1 {
			return fmt.Errorf("query agent state result no 1, states(%v)", states)
		}

		state := states[0]
		lastStatus = state.NodeStatus
		std.DeployInfo().Host.Dynamic.NodeStatus = state.NodeStatus
		if state.NodeStatus == types.NodeStatusRunning {
			return fmt.Errorf("agent is still running, agent-id(%s)", agentID)
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			err = fmt.Errorf("wait agent not alive timeout: agent-id(%s), last-status(%s)", agentID, lastStatus)
		}

		std.InstanceData().Log().
			Zh("等待 Agent 不存活失败: %s", err.Error()).
			En("failed to wait agent not alive: %s", err.Error()).
			Warn()

		return err
	}

	std.InstanceData().Log().
		Zh("确认 Agent 不存活成功. agent-id(%s), status(%s)", agentID, lastStatus).
		En("confirmed agent is not alive. agent-id(%s), status(%s)", agentID, lastStatus).
		Info()

	return nil
}
