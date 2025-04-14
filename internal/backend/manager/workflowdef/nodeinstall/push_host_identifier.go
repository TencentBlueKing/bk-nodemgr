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

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/nodedeployment"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/retrier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
)

// NewActionPushHostIdentifier ...
func NewActionPushHostIdentifier(
	cmdbClient cmdb.IHandler,
	storage nodedeployment.IStorage,
	logger logger.Logger,
) *PushHostIdentifier {

	return &PushHostIdentifier{
		cmdbClient: cmdbClient,
		storage:    storage,
		logger:     logger,
	}
}

// PushHostIdentifierParam ...
type PushHostIdentifierParam struct {
	Token string `json:"token"`
}

// PushHostIdentifier ...
type PushHostIdentifier struct {
	cmdbClient cmdb.IHandler
	storage    nodedeployment.IStorage
	logger     logger.Logger
}

// Name returns the name of the action.
func (action *PushHostIdentifier) Name() string {
	return ActionNamePushHostIdentifier
}

// Version returns the version of the action.
func (action *PushHostIdentifier) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (action *PushHostIdentifier) Description() string {
	return "Query agent state"
}

// Timeout returns the timeout of the action.
func (action *PushHostIdentifier) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (action *PushHostIdentifier) Tags() []operengine.ActionTag {
	return []operengine.ActionTag{}
}

// MaxRetryCount returns the max retry count of the action.
func (action *PushHostIdentifier) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (action *PushHostIdentifier) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (action *PushHostIdentifier) Do(ctx *operengine.ActionInstContext) error {
	param := new(PushHostIdentifierParam)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	info, err := action.storage.GetInfo(ctx.Ctx, param.Token)
	if err != nil {
		return err
	}

	tCtx, err := tenant.SetID(ctx.Ctx, info.TenantID)
	if err != nil {
		return err
	}

	polling := retrier.NewPolling(retrier.PollingOpts{
		Timeout:  action.Timeout(),
		Interval: time.Second,
		Logger:   action.logger,
	})

	taskID, err := action.cmdbClient.PushHostIdentifier(tCtx, info.HostID)
	if err != nil {
		return err
	}
	ctx.Data.Log(fmt.Sprintf("pushed host identifier, task-id(%s)", taskID))

	var success bool
	err = polling.Do(tCtx, func(_ int) error {
		successList, failedList, pendingList, err := action.cmdbClient.FindHostIdentifierPushResult(tCtx, taskID)
		if err != nil {
			action.logger.Errorf("failed to find host identifier push result, err: %s", err.Error())

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
		ctx.Data.Log("failed to push host identifier, err: " + err.Error())

		return err
	}

	if !success {
		ctx.Data.Log("failed to push host identifier, no success result")

		return errors.New("failed to push host identifier")
	}

	ctx.Data.Log("pushed host identifier")

	return nil
}
