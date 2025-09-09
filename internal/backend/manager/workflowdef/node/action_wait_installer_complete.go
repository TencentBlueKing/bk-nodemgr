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
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameWaitInstallerComplete defines the action name.
	ActionNameWaitInstallerComplete = "wait_installer_complete"
)

// NewActionWaitInstallerComplete get a new action.
func NewActionWaitInstallerComplete(
	storageActionInstance workflow.IStorageActionInstance, logger logger.ILogger) action.Definition {

	return &actionWaitInstallerComplete{
		storageActionInstance: storageActionInstance,
		logger:                logger,
	}
}

type actionWaitInstallerComplete struct {
	storageActionInstance workflow.IStorageActionInstance
	logger                logger.ILogger
}

// Name returns the name of the action.
func (act *actionWaitInstallerComplete) Name() string {
	return ActionNameWaitInstallerComplete
}

// Version returns the version of the action.
func (act *actionWaitInstallerComplete) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionWaitInstallerComplete) Description() string {
	return "Wait for the installer command to complete"
}

// Timeout returns the timeout of the action.
func (act *actionWaitInstallerComplete) Timeout() time.Duration {
	return 30 * time.Minute // nolint:mnd
}

// Tags returns the tags of the action.
func (act *actionWaitInstallerComplete) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount this func define how many times this action will retry.
func (act *actionWaitInstallerComplete) MaxRetryCount() uint {
	return 0
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionWaitInstallerComplete) DelayFn() func() {
	return func() {}
}

// Do this func define what the action will do.
func (act *actionWaitInstallerComplete) Do(ctx *action.InstanceContext) error {
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
			act.logger.Debugf("action is running, sleep 1 second, oper_inst_id(%s), action_name(%s)",
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
