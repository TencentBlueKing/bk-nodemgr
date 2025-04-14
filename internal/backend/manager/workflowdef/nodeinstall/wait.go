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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
)

// NewActionWaitComplete new an action to wait action finished by other code.
func NewActionWaitComplete(storage operengine.OperInstStorage, logger logger.Logger) operengine.ActionDef {
	return &WaitComplete{
		storage: storage,
		logger:  logger,
	}
}

// WaitComplete is an action to wait action finished by other code.
type WaitComplete struct {
	storage operengine.OperInstStorage
	logger  logger.Logger
}

// Name returns the name of the action.
func (action *WaitComplete) Name() string {
	return ActionNameWaitComplete
}

// Version returns the version of the action.
func (action *WaitComplete) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (action *WaitComplete) Description() string {
	return "Wait for the action to complete"
}

// Timeout returns the timeout of the action.
func (action *WaitComplete) Timeout() time.Duration {
	return 30 * time.Minute // nolint:mnd
}

// Tags returns the tags of the action.
func (action *WaitComplete) Tags() []operengine.ActionTag {
	return []operengine.ActionTag{}
}

// MaxRetryCount this func define how many times this action will retry.
func (action *WaitComplete) MaxRetryCount() uint {
	return 0
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (action *WaitComplete) DelayFn() func() {
	return func() {}
}

// Do this func define what the action will do.
func (action *WaitComplete) Do(ctx *operengine.ActionInstContext) error {
	for {
		select {
		case <-ctx.Ctx.Done():
			return nil
		default:
		}

		lifecycle, err := action.storage.GetActInstLifecycle(ctx.Ctx, ctx.Data.OperInstID, ctx.Data.Name)
		if err != nil {
			action.logger.Errorf("get action_inst_data lifecycle failed, err: %v", err)

			return err
		}

		// check action state is running or not
		switch lifecycle.State {
		case operengine.ActionInstStateRunning:
			action.logger.Infof("action is running, sleep 1 second, oper_inst_id(%s), action_name(%s)",
				ctx.Data.OperInstID, ctx.Data.Name)
			time.Sleep(1 * time.Second)
		default:
			return nil
		}
	}
}
