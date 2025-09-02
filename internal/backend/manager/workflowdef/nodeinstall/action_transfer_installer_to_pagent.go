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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameTransferInstallerToPagent defines the action name.
	ActionNameTransferInstallerToPagent = "transfer_installer_to_pagent"
)

// NewActionTransferInstallerToRelay get a new action.
func NewActionTransferInstallerToRelay(
	storageNodeDeployment nodedeployment.IStorageNodeDeployment,
	fileHandler file.IHandler,
	logger logger.ILogger) action.Definition {

	return &actionTransferInstallerToPagent{
		storageNodeDeployment: storageNodeDeployment,
		fileHandler:           fileHandler,
		logger:                logger,
	}
}

// ActionParamTransferInstallerToRelay defines the action param.
type ActionParamTransferInstallerToRelay struct {
	Token string `json:"token"`
}

type actionTransferInstallerToPagent struct {
	storageNodeDeployment nodedeployment.IStorageNodeDeployment
	fileHandler           file.IHandler
	logger                logger.ILogger
}

// Name returns the name of the action.
func (act *actionTransferInstallerToPagent) Name() string {
	return ActionNameTransferInstallerToPagent
}

// Version returns the version of the action.
func (act *actionTransferInstallerToPagent) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionTransferInstallerToPagent) Description() string {
	return "transfer installer to relay"
}

// Timeout returns the timeout of the action.
func (act *actionTransferInstallerToPagent) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionTransferInstallerToPagent) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionTransferInstallerToPagent) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionTransferInstallerToPagent) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: funlen,nonamedreturns
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionTransferInstallerToPagent) Do(ctx *action.InstanceContext) (err error) {
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

	deployConstant, err := deployconstant.GetDeployConf(info.Host.Dynamic.NodeGeneration, info.Host.Dynamic.NodeOsType)
	if err != nil {
		return fmt.Errorf("failed to get deploy constant, err: %w", err)
	}

	// installer workdir priority: user specified in info > deploy constant default.
	if info.InstallerWorkDir == "" {
		info.InstallerWorkDir = deployConstant.WorkDir
	}

	if !info.TransferOptions.SelectDownloads || info.TransferOptions.EnableInstaller {
		ctx.Data.LogI("start to transfer installer package")
		if err = act.transferInstaller(ctx.Ctx, info); err != nil {
			return fmt.Errorf("failed to transfer installer, err: %w", err)
		}
	}

	//TODO: close installer after transfer.

	act.logger.InfoCtxf(ctx.Ctx, "transfer pkg to node all done. host-id(%d)", info.Host.HostID)
	ctx.Data.LogI("transfer pkg to node all done")

	return nil
}

func (act *actionTransferInstallerToPagent) transferInstaller(ctx contextx.IContext, info *types.DeploymentInfo) error {
	transferHandler, err := act.fileHandler.LaunchTransferInstaller(ctx,
		types.Generation2,
		platform.Platform{
			OS:   info.Host.Dynamic.NodeOsType,
			Arch: info.Host.Dynamic.NodeCPUArch,
		},
		info.InstallerWorkDir,
		&info.Host)
	if err != nil {
		return fmt.Errorf("failed to launch transfer installer. host-id(%d), err: %w", info.Host.HostID, err)
	}

	act.logger.InfoCtxf(ctx, "launched transfer installer. task-id(%s), host-id(%d)",
		transferHandler.GetTaskID(), info.Host.HostID)

	result, err := transferHandler.WaitUntilDone(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait until transfer installer done. task-id(%s), host-id(%d), err: %w",
			transferHandler.GetTaskID(), info.Host.HostID, err)
	}

	if !result.Terminated {
		return fmt.Errorf("transfer installer not terminated. task-id(%s), host-id(%d)",
			transferHandler.GetTaskID(), info.Host.HostID)
	}

	if result.ErrorCode != 0 {
		return fmt.Errorf("transfer installer failed. task-id(%s), host-id(%d), err-code(%d), err-msg(%s)",
			transferHandler.GetTaskID(), info.Host.HostID, result.ErrorCode, result.ErrorMessage)
	}

	act.logger.InfoCtxf(ctx, "transfer installer done. task-id(%s), host-id(%d)",
		transferHandler.GetTaskID(), info.Host.HostID)

	return nil
}
