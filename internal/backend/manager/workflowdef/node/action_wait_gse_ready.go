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

// ActParamWaitGseReady ...
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

	polling := retrier.NewPolling(retrier.PollingOpts{
		Timeout:  act.Timeout(),
		Interval: time.Second,
	})
	err = polling.Do(std.Context(), func(_ int) error {
		states, err := act.gseClient.ListAgentState(std.Context(), std.DeployInfo().Host.Dynamic.AgentID)
		if err != nil {
			return err
		}

		if len(states) != 1 {
			return fmt.Errorf("query agent state result no 1, states(%v)", states)
		}

		state := states[0]
		std.DeployInfo().Host.Dynamic.NodeStatus = state.NodeStatus
		if state.NodeStatus != types.NodeStatusRunning {
			return fmt.Errorf("agent state is not running, status(%s)", state.NodeStatus)
		}

		if state.Version != std.DeployInfo().Host.Dynamic.NodeVersion {
			return fmt.Errorf("agent version is not match, version(%s)", state.Version)
		}

		return nil
	})
	if err != nil {
		std.InstanceData().LogE("failed to query agent state: " + err.Error())

		return err
	}

	return nil
}
