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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameWaitInstallComplete defines the action name.
	ActionNameWaitInstallComplete = "wait_install_complete"
)

// NewActionWaitInstallComplete get a new action.
func NewActionWaitInstallComplete(
	storageActionInstance workflow.IStorageActionInstance, logger logger.Logger) action.Definition {

	return &actionWaitInstallComplete{
		storageActionInstance: storageActionInstance,
		logger:                logger,
	}
}

type actionWaitInstallComplete struct {
	storageActionInstance workflow.IStorageActionInstance
	logger                logger.Logger
}

// Name returns the name of the action.
func (act *actionWaitInstallComplete) Name() string {
	return ActionNameWaitInstallComplete
}

// Version returns the version of the action.
func (act *actionWaitInstallComplete) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionWaitInstallComplete) Description() string {
	return "Wait for the action to complete"
}

// Timeout returns the timeout of the action.
func (act *actionWaitInstallComplete) Timeout() time.Duration {
	return 30 * time.Minute // nolint:mnd
}

// Tags returns the tags of the action.
func (act *actionWaitInstallComplete) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount this func define how many times this action will retry.
func (act *actionWaitInstallComplete) MaxRetryCount() uint {
	return 0
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionWaitInstallComplete) DelayFn() func() {
	return func() {}
}

// Do this func define what the action will do.
func (act *actionWaitInstallComplete) Do(ctx *action.InstanceContext) error {
	for {
		select {
		case <-ctx.Ctx.Done():
			return nil
		default:
		}

		lifecycle, err := act.storageActionInstance.GetActionInstanceLifecycle(
			ctx.Ctx,
			ctx.Data.OperationInstanceID,
			ctx.Data.Name)
		if err != nil {
			act.logger.Errorf("get action_inst_data lifecycle failed, err: %v", err)

			return err
		}

		// check action state is running or not
		switch lifecycle.State {
		case action.StateRunning:
			act.logger.Infof("action is running, sleep 1 second, oper_inst_id(%s), action_name(%s)",
				ctx.Data.OperationInstanceID, ctx.Data.Name)
			time.Sleep(1 * time.Second)

		case action.StateFailed:
			act.logger.Errorf("wait install complete failed. oper_inst_id(%s), action_name(%s)",
				ctx.Data.OperationInstanceID, ctx.Data.Name)

			return fmt.Errorf("wait install complete failed. oper_inst_id(%s), action_name(%s)",
				ctx.Data.OperationInstanceID, ctx.Data.Name)

		default:
			return nil
		}
	}
}
