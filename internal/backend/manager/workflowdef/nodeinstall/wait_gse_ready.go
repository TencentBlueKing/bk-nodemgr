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
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/nodedeployment"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/retrier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
)

// NewActionWaitGseRunning ...
func NewActionWaitGseRunning(
	gseClient gse.IHandler,
	iDaoNodeDeployment nodedeployment.IDaoNodeDeployment,
	logger logger.Logger,
) operengine.ActionDef {

	return &WaitGseReady{
		gseClient:          gseClient,
		iDaoNodeDeployment: iDaoNodeDeployment,
		logger:             logger,
	}
}

// WaitGseRunningParam ...
type WaitGseRunningParam struct {
	Token string `json:"token"`
}

// WaitGseReady ...
type WaitGseReady struct {
	gseClient          gse.IHandler
	iDaoNodeDeployment nodedeployment.IDaoNodeDeployment
	logger             logger.Logger
}

// Name returns the name of the action.
func (action *WaitGseReady) Name() string {
	return ActionNameWaitGseReady
}

// Version returns the version of the action.
func (action *WaitGseReady) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (action *WaitGseReady) Description() string {
	return "wait gse ready"
}

// Timeout returns the timeout of the action.
func (action *WaitGseReady) Timeout() time.Duration {
	// notice: in this action, the timeout should be equal or grander than the timeout of the workflow,
	// so we set it to 24 hours.
	return 24 * time.Hour // nolint: mnd
}

// Tags returns the tags of the action.
func (action *WaitGseReady) Tags() []operengine.ActionTag {
	return []operengine.ActionTag{}
}

// MaxRetryCount returns the max retry count of the action.
func (action *WaitGseReady) MaxRetryCount() uint {
	// notice: this action is an polling action should not auto retry.
	return 0
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (action *WaitGseReady) DelayFn() func() {
	return func() {
		action.logger.Errorf("this action should not auto retry, action-name(%s)", action.Name())
	}
}

// Do this func define what the action will do.
func (action *WaitGseReady) Do(ctx *operengine.ActionInstContext) error {
	param := new(WaitGseRunningParam)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	info, err := action.iDaoNodeDeployment.GetInfo(ctx.Ctx, param.Token)
	if err != nil {
		return err
	}

	polling := retrier.NewPolling(retrier.PollingOpts{
		Timeout:  action.Timeout(),
		Interval: time.Second,
		Logger:   action.logger,
	})
	err = polling.Do(ctx.Ctx, func(_ int) error {
		states, err := action.gseClient.ListAgentState(ctx.Ctx, info.Dynamic.AgentID)
		if err != nil {
			return err
		}

		if len(states) != 1 {
			return fmt.Errorf("query agent state result no 1, states(%v)", states)
		}

		state := states[0]
		info.Dynamic.NodeStatus = state.StatusCode.ToNodeStatus()
		if state.StatusCode != types.AgentStatusCodeRunning {
			return fmt.Errorf("agent state is not running, status(%s)", state.StatusCode.String())
		}

		return nil
	})
	if err != nil {
		ctx.Data.Log("failed to query agent state, err: " + err.Error())

		return err
	}

	if err := action.iDaoNodeDeployment.UpdateInfo(ctx.Ctx, param.Token, info); err != nil {
		return fmt.Errorf("update node deployment info failed, err: %w", err)
	}

	return nil
}
