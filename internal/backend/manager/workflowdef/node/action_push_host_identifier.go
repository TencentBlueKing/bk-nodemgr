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

	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/retrier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNamePushHostIdentifier defines the action name.
	ActionNamePushHostIdentifier = "push_host_identifier"
)

// NewActionPushHostIdentifier get a new action.
func NewActionPushHostIdentifier(
	cmdbClient cmdb.IHandler,
	storage nodeStg.IStorage,
	logger logger.ILogger,
) action.Definition {

	return &actionPushHostIdentifier{
		cmdbClient: cmdbClient,
		storage:    storage,
		logger:     logger,
	}
}

// ActParamPushHostIdentifier ...
type ActParamPushHostIdentifier struct {
	Token    string `json:"token"`
	Operator string `json:"operator"`
}

// PushHostIdentifier ...
type actionPushHostIdentifier struct {
	cmdbClient cmdb.IHandler
	storage    nodeStg.IStorage
	logger     logger.ILogger
}

// Name returns the name of the action.
func (act *actionPushHostIdentifier) Name() string {
	return ActionNamePushHostIdentifier
}

// Version returns the version of the action.
func (act *actionPushHostIdentifier) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionPushHostIdentifier) Description() string {
	return "push host identifier via cmdb"
}

// Timeout returns the timeout of the action.
func (act *actionPushHostIdentifier) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionPushHostIdentifier) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionPushHostIdentifier) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionPushHostIdentifier) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionPushHostIdentifier) Do(ctx *action.InstanceContext) error {
	param := new(ActParamPushHostIdentifier)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	info, err := act.storage.GetNodeDeploymentInfo(ctx.Ctx, param.Token)
	if err != nil {
		return err
	}

	tenantUserCtx := contextx.NewTenantUserContext(ctx.Ctx, info.Host.TenantID, param.Operator)
	polling := retrier.NewPolling(retrier.PollingOpts{
		Timeout:  act.Timeout(),
		Interval: time.Second,
		Logger:   act.logger,
	})

	taskID, err := act.cmdbClient.PushHostIdentifier(tenantUserCtx, info.Host.HostID)
	if err != nil {
		return err
	}
	ctx.Data.LogI(fmt.Sprintf("pushed host identifier, task-id(%s)", taskID))

	var success bool
	err = polling.Do(tenantUserCtx, func(_ int) error {
		successList, failedList, pendingList, err := act.cmdbClient.FindHostIdentifierPushResult(tenantUserCtx, taskID)
		if err != nil {
			act.logger.Errorf("failed to find host identifier push result, err: %s", err.Error())

			return err
		}

		if len(pendingList) > 0 {
			return errors.New("pending host identifier push result")
		}

		if len(successList)+len(failedList) == 0 {
			return errors.New("invalid host identifier push result, no success or failed")
		}

		if len(successList) > 0 {
			success = true
		}

		return nil
	})
	if err != nil {
		ctx.Data.LogE("failed to push host identifier, err: " + err.Error())

		return err
	}

	if !success {
		ctx.Data.LogE("failed to push host identifier, no success result")

		return errors.New("failed to push host identifier")
	}

	ctx.Data.LogI("pushed host identifier")

	return nil
}
