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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/retrier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNamePushHostIdentifier defines the action name.
	ActionNamePushHostIdentifier = "push_host_identifier"
)

// NewActionPushHostIdentifier get a new action.
func NewActionPushHostIdentifier(capability *Capability) action.Definition {
	return &actionPushHostIdentifier{
		cmdbClient:            capability.CMDBHandler,
		storageNodeDeployment: capability.StorageNode,
	}
}

// ActParamPushHostIdentifier ...
type ActParamPushHostIdentifier struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

// PushHostIdentifier ...
type actionPushHostIdentifier struct {
	cmdbClient            cmdb.IHandler
	storageNodeDeployment nodeStg.IDaoNodeDeployment
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

	// initialize standard data.
	std := nodeUtils.NewNodeActionStandarder(act.storageNodeDeployment)
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

	taskID, err := act.cmdbClient.PushHostIdentifier(std.Context(), std.DeployInfo().Host.HostID)
	if err != nil {
		return err
	}
	std.InstanceData().LogI(fmt.Sprintf("pushed host identifier, task-id(%s)", taskID))

	var success bool
	err = polling.Do(std.Context(), func(_ int) error {
		successList, _, pendingList, err := act.cmdbClient.FindHostIdentifierPushResult(std.Context(), taskID)
		if err != nil {
			logger.G.Sys().WithErr(err).Error("failed to find host identifier push result")

			return err
		}

		for _, hostID := range pendingList {
			if hostID == std.DeployInfo().Host.HostID {
				return errors.New("pending host identifier push result")
			}
		}

		for _, hostID := range successList {
			if hostID == std.DeployInfo().Host.HostID {
				success = true

				return nil
			}
		}

		return nil
	})
	if err != nil {
		std.InstanceData().LogE("failed to push host identifier: " + err.Error())

		return err
	}

	if !success {
		std.InstanceData().LogE("failed to push host identifier, no success result")

		return errors.New("failed to push host identifier")
	}

	std.InstanceData().LogI("pushed host identifier")

	return nil
}
