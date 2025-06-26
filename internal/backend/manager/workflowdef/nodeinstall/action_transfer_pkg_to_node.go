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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameTransferPkgToNode defines the action name.
	ActionNameTransferPkgToNode = "upgrade_node"

	actionTransferPkgTargetDirUnix    = "/tmp"
	actionTransferPkgTargetDirWindows = "c:\\tmp"
)

// NewActionTransferPkgToNode get a new action.
func NewActionTransferPkgToNode(
	storageNodeDeployment nodedeployment.IStorageNodeDeployment,
	fileHandler file.IHandler,
	logger logger.Logger) action.Definition {

	return &actionTransferPkgToNode{
		storageNodeDeployment: storageNodeDeployment,
		fileHandler:           fileHandler,
		logger:                logger,
	}
}

// ActionParamTransferPkgToNode defines the action param.
type ActionParamTransferPkgToNode struct {
	Token string `json:"token"`
}

type actionTransferPkgToNode struct {
	storageNodeDeployment nodedeployment.IStorageNodeDeployment
	fileHandler           file.IHandler
	logger                logger.Logger
}

// Name returns the name of the action.
func (act *actionTransferPkgToNode) Name() string {
	return ActionNameTransferPkgToNode
}

// Version returns the version of the action.
func (act *actionTransferPkgToNode) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionTransferPkgToNode) Description() string {
	return "transfer pkg to node"
}

// Timeout returns the timeout of the action.
func (act *actionTransferPkgToNode) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionTransferPkgToNode) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionTransferPkgToNode) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionTransferPkgToNode) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: funlen,fnsize,nonamedreturns
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionTransferPkgToNode) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActionParamTransferPkgToNode)
	err = conv.MapToStruct(ctx.Data.Content, param)
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

	var rt types.ReleaseType
	switch info.Dynamic.NodeRole {
	case types.NodeRoleProxy:
		rt = types.ReleaseTypeProxy

	case types.NodeRoleAgent:
		rt = types.ReleaseTypeAgent

	default:
		return fmt.Errorf("unsupported node role: %s", info.Dynamic.NodeRole)
	}

	if info.TmpDir == "" {
		info.TmpDir = actionTransferPkgTargetDirUnix
		if info.Dynamic.NodeOsType == criteria.OSWindows {
			info.TmpDir = actionTransferPkgTargetDirWindows
		}
	}

	transferHandler, err := act.fileHandler.LaunchTransferRelease(ctx.Ctx,
		info.Dynamic.NodeGeneration,
		rt,
		platform.Platform{
			OS:   info.Dynamic.NodeOsType,
			Arch: info.Dynamic.NodeCPUArch,
		},
		info.Dynamic.NodeVersion,
		info.TmpDir,
		&info.Host)
	if err != nil {
		return fmt.Errorf("failed to launch transfer release. host-id(%d), err: %w", info.HostID, err)
	}

	act.logger.InfoCtxf(ctx.Ctx, "launched transfer release. task-id(%s), host-id(%d)",
		transferHandler.GetTaskID(), info.HostID)

	result, err := transferHandler.WaitUntilDone(ctx.Ctx)
	if err != nil {
		return fmt.Errorf("failed to wait until transfer done. task-id(%s), host-id(%d), err: %w",
			transferHandler.GetTaskID(), info.HostID, err)
	}

	if !result.Terminated {
		return fmt.Errorf("transfer not terminated. task-id(%s), host-id(%d)", transferHandler.GetTaskID(), info.HostID)
	}

	if result.ErrorCode != 0 {
		return fmt.Errorf("transfer failed. task-id(%s), host-id(%d), err-code(%d), err-msg(%s)",
			transferHandler.GetTaskID(), info.HostID, result.ErrorCode, result.ErrorMessage)
	}

	act.logger.InfoCtxf(ctx.Ctx, "transfer done. task-id(%s), host-id(%d)", transferHandler.GetTaskID(), info.HostID)

	return nil
}
