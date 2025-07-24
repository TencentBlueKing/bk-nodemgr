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
	"time"

	nodedeployment "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-deployment"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameCheckPkgExists defines the action name.
	ActionNameCheckPkgExists = "check_pkg_exists"
)

// NewActionCheckPkgExists get a new action.
func NewActionCheckPkgExists(
	storageNodeDeployment nodedeployment.IStorageNodeDeployment,
	storageActionInstance workflow.IStorageActionInstance,
	logger logger.Logger,
	proxyMessanger relayhandler.ServerMessager,

) action.Definition {

	return &actionCheckPkgExists{
		storageNodeDeployment: storageNodeDeployment,
		storageActionInstance: storageActionInstance,
		logger:                logger,
		proxyMessanger:        proxyMessanger,
	}
}

// ActionParamCheckPkgExists defines the action param.
type ActionParamCheckPkgExists struct {
	Token string `json:"token"`
}

// actionCheckPkgExists ...
type actionCheckPkgExists struct {
	storageNodeDeployment nodedeployment.IStorageNodeDeployment
	storageActionInstance workflow.IStorageActionInstance
	proxyMessanger        relayhandler.ServerMessager
	logger                logger.Logger
}

// Name returns the name of the action.
func (act *actionCheckPkgExists) Name() string {
	return ActionNamePushHostIdentifier
}

// Version returns the version of the action.
func (act *actionCheckPkgExists) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionCheckPkgExists) Description() string {
	return "check relay pkg exists"
}

// Timeout returns the timeout of the action.
func (act *actionCheckPkgExists) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionCheckPkgExists) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionCheckPkgExists) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionCheckPkgExists) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionCheckPkgExists) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamCheckPkgExists)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		err = fmt.Errorf("failed to convert param, err: %w", err)

		return err
	}

	info, err := act.storageNodeDeployment.GetInfo(ctx.Ctx, param.Token)
	if err != nil {
		return err
	}

	defer func() {
		if storeErr := act.storageNodeDeployment.UpdateInfo(ctx.Ctx, param.Token, info); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	if act.sendCheckPkgExists() != nil {
		return fmt.Errorf("failed to check pkg exists, err: %w", err)
	}

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

	return nil
}

func (act *actionCheckPkgExists) sendCheckPkgExists() error {
	return nil
}
