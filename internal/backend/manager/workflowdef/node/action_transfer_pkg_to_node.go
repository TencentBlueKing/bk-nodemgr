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
	"path/filepath"
	"time"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/winpath"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameTransferPkgToNode defines the action name.
	ActionNameTransferPkgToNode = "transfer_pkg_to_node"
)

// NewActionTransferPkgToNode get a new action.
func NewActionTransferPkgToNode(
	storageNodeDeployment nodeStg.IDaoNodeDeployment,
	fileHandler file.IHandler,
) action.Definition {

	return &actionTransferPkgToNode{
		storageNodeDeployment: storageNodeDeployment,
		fileHandler:           fileHandler,
	}
}

// ActionParamTransferPkgToNode defines the action param.
type ActionParamTransferPkgToNode struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionTransferPkgToNode struct {
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	fileHandler           file.IHandler
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
// nolint: funlen,nonamedreturns
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionTransferPkgToNode) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActionParamTransferPkgToNode)
	err = conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		err = fmt.Errorf("failed to convert param: %w", err)

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

	deployConstant, err := deployconstant.GetNodeDeployConf(std.DeployInfo().Host.Dynamic.NodeGeneration, std.DeployInfo().Host.Dynamic.NodeOsType)
	if err != nil {
		return fmt.Errorf("failed to get deploy constant: %w", err)
	}

	// installer workdir priority: user specified in info > deploy constant default.
	if std.DeployInfo().InstallerWorkDir == "" {
		std.DeployInfo().InstallerWorkDir = deployConstant.WorkDir
	}

	gp := gopool.NewPool()
	if !std.DeployInfo().TransferOptions.SelectDownloads || std.DeployInfo().TransferOptions.EnableReleasePackage {
		gp.Go(func() error {
			return act.transferRelease(ctx.Ctx, std.DeployInfo())
		})
	}
	if !std.DeployInfo().TransferOptions.SelectDownloads || std.DeployInfo().TransferOptions.EnableInstaller {
		gp.Go(func() error {
			return act.transferInstaller(ctx.Ctx, std.DeployInfo())
		})
	}
	if err := gp.Wait(); err != nil {
		logger.G.Sys().WithErr(err).With("host-id", std.DeployInfo().Host.HostID).Error("failed to transfer pkg to node")

		return err
	}

	logger.G.Sys().With("host-id", std.DeployInfo().Host.HostID).Info("transfer pkg to node all done")
	ctx.Data.LogI("transfer pkg to node all done")

	return nil
}

func (act *actionTransferPkgToNode) transferRelease(ctx contextx.IContext, info *types.DeploymentInfo) error {
	var rt types.ReleaseType
	switch info.Host.Dynamic.NodeRole {
	case types.NodeRoleProxy:
		rt = types.ReleaseTypeProxy

	case types.NodeRoleAgent:
		rt = types.ReleaseTypeAgent

	default:
		return fmt.Errorf("unsupported node role: %s", info.Host.Dynamic.NodeRole)
	}

	var dataDir string
	if info.Host.Dynamic.NodeOsType == criteria.OSWindows {
		dataDir = winpath.Join(info.InstallerWorkDir, "data")
	} else {
		dataDir = filepath.Join(info.InstallerWorkDir, "data")
	}

	transferHandler, err := act.fileHandler.LaunchTransferNode(ctx,
		info.Host.Dynamic.NodeGeneration,
		rt,
		platform.Platform{
			OS:   info.Host.Dynamic.NodeOsType,
			Arch: info.Host.Dynamic.NodeCPUArch,
		},
		info.Host.Dynamic.NodeVersion,
		dataDir,
		&info.Host)
	if err != nil {
		return fmt.Errorf("failed to launch transfer release. host-id(%d): %w", info.Host.HostID, err)
	}

	logger.G.Sys().With("task-id", transferHandler.GetTaskID(), "host-id", info.Host.HostID).Info("launched transfer release")

	result, err := transferHandler.WaitUntilDone(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait until transfer release done. task-id(%s), host-id(%d): %w",
			transferHandler.GetTaskID(), info.Host.HostID, err)
	}

	if !result.Terminated {
		return fmt.Errorf("transfer release not terminated. task-id(%s), host-id(%d)",
			transferHandler.GetTaskID(), info.Host.HostID)
	}

	if result.ErrorCode != 0 {
		return fmt.Errorf("transfer release failed. task-id(%s), host-id(%d), err-code(%d), err-msg(%s)",
			transferHandler.GetTaskID(), info.Host.HostID, result.ErrorCode, result.ErrorMessage)
	}

	logger.G.Sys().With("task-id", transferHandler.GetTaskID(), "host-id", info.Host.HostID).Info("transfer release done")

	return nil
}

func (act *actionTransferPkgToNode) transferInstaller(ctx contextx.IContext, info *types.DeploymentInfo) error {
	transferHandler, err := act.fileHandler.LaunchTransferInstaller(ctx,
		types.Generation2,
		platform.Platform{
			OS:   info.Host.Dynamic.NodeOsType,
			Arch: info.Host.Dynamic.NodeCPUArch,
		},
		info.InstallerWorkDir,
		&info.Host)
	if err != nil {
		return fmt.Errorf("failed to launch transfer installer. host-id(%d): %w", info.Host.HostID, err)
	}

	logger.G.Sys().With("task-id", transferHandler.GetTaskID(), "host-id", info.Host.HostID).Info("launched transfer installer")

	result, err := transferHandler.WaitUntilDone(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait until transfer installer done. task-id(%s), host-id(%d): %w",
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

	logger.G.Sys().With("task-id", transferHandler.GetTaskID(), "host-id", info.Host.HostID).Info("transfer installer done")

	return nil
}
