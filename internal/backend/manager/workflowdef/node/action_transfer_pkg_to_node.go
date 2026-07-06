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
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
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
func NewActionTransferPkgToNode(capability *Capability) action.Definition {
	return &actionTransferPkgToNode{
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		fileHandler:           capability.FileHandler,
	}
}

// ActionParamTransferPkgToNode defines the action param.
type ActionParamTransferPkgToNode struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionTransferPkgToNode struct {
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	fileHandler           file.IHandler
}

// Name returns the name of the action.
func (act *actionTransferPkgToNode) Name() string {
	return ActionNameTransferPkgToNode
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionTransferPkgToNode) DisplayNameZh() string {
	return "传输安装包到节点"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionTransferPkgToNode) DisplayNameEn() string {
	return "Transfer Package to Node"
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
func (act *actionTransferPkgToNode) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: funlen,nonamedreturns
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionTransferPkgToNode) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamTransferPkgToNode)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := nodeUtils.NewNodeActionStandarder(act.storageNodeDeployment, act.storageHost)
	if err = std.Initialize(ctx, param.NodeActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	if err = std.DeployInfo().Host.Dynamic.NodeOsType.Validate(); err != nil {
		return fmt.Errorf("host dynamic has invalid node_os_type. host-id(%d): %w", std.DeployInfo().Host.HostID, err)
	}

	if err = std.DeployInfo().Host.Dynamic.NodeCPUArch.Validate(); err != nil {
		return fmt.Errorf("host dynamic has invalid node_cpu_arch. host-id(%d): %w", std.DeployInfo().Host.HostID, err)
	}

	gp := gopool.NewPool()
	if !std.DeployInfo().TransferOptions.DisableReleasePackage {
		gp.Go(func() error {
			return act.transferRelease(std)
		})
	}
	if !std.DeployInfo().TransferOptions.DisableInstaller {
		gp.Go(func() error {
			return act.transferInstaller(std)
		})
	}
	if err := gp.Wait(); err != nil {
		logger.G.Sys().Ctx(std.Context()).WithErr(err).With("host-id", std.DeployInfo().Host.HostID).Error("failed to transfer pkg to node")

		return err
	}

	logger.G.Sys().Ctx(std.Context()).With("host-id", std.DeployInfo().Host.HostID).Info("transfer pkg to node all done")
	std.InstanceData().Log().
		Zh("传输安装包到节点全部完成").
		En("transfer pkg to node all done").
		Info()

	return nil
}

func (act *actionTransferPkgToNode) transferRelease(std *nodeUtils.NodeActionStandarder) error {
	info := std.DeployInfo()
	nCtx := std.Context()

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
		dataDir = winpath.Join(info.InstallerRuntime.WorkDir, "data")
	} else {
		dataDir = filepath.Join(info.InstallerRuntime.WorkDir, "data")
	}

	std.InstanceData().Log().
		Zh("准备传输 release 包, host-id(%d)", std.DeployInfo().Host.HostID).
		En("preparing to transfer release package. host-id(%d)", std.DeployInfo().Host.HostID).
		Info()

	transferHandler, err := act.fileHandler.LaunchTransferNode(nCtx,
		info.Host.Dynamic.NodeGeneration,
		rt,
		platfmt.Platform{
			OS:   info.Host.Dynamic.NodeOsType,
			Arch: info.Host.Dynamic.NodeCPUArch,
		},
		info.Host.Dynamic.NodeVersion,
		dataDir,
		&info.Host)
	if err != nil {
		return fmt.Errorf("failed to launch transfer release. host-id(%d): %w", info.Host.HostID, err)
	}

	std.InstanceData().Log().
		Zh("开始传输 release 包. release-type(%s), task-id(%s)", rt, transferHandler.GetTaskID()).
		En("start transfer release. release-type(%s), task-id(%s)", rt, transferHandler.GetTaskID()).
		Info()

	logger.G.Sys().Ctx(nCtx).With("task-id", transferHandler.GetTaskID(), "host-id", info.Host.HostID).Info("launched transfer release")

	result, err := transferHandler.WaitUntilDone(nCtx)
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

	logger.G.Sys().Ctx(nCtx).With("task-id", transferHandler.GetTaskID(), "host-id", info.Host.HostID).Info("transfer release done")

	std.InstanceData().Log().
		Zh("传输 release 包完成").
		En("transfer release done").
		Info()

	return nil
}

func (act *actionTransferPkgToNode) transferInstaller(std *nodeUtils.NodeActionStandarder) error {
	info := std.DeployInfo()
	nCtx := std.Context()

	std.InstanceData().Log().
		Zh("准备传输 installer 包, host-id(%d)", std.DeployInfo().Host.HostID).
		En("preparing to transfer installer package. host-id(%d)", std.DeployInfo().Host.HostID).
		Info()

	transferHandler, err := act.fileHandler.LaunchTransferInstaller(nCtx,
		types.Generation2,
		platfmt.Platform{
			OS:   info.Host.Dynamic.NodeOsType,
			Arch: info.Host.Dynamic.NodeCPUArch,
		},
		info.InstallerRuntime.WorkDir,
		&info.Host)
	if err != nil {
		return fmt.Errorf("failed to launch transfer installer. host-id(%d): %w", info.Host.HostID, err)
	}

	std.InstanceData().Log().
		Zh("开始传输 installer 安装包. task-id(%s)", transferHandler.GetTaskID()).
		En("start transfer installer. task-id(%s)", transferHandler.GetTaskID()).
		Info()

	logger.G.Sys().Ctx(nCtx).With("task-id", transferHandler.GetTaskID(), "host-id", info.Host.HostID).Info("launched transfer installer")

	result, err := transferHandler.WaitUntilDone(nCtx)
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

	logger.G.Sys().Ctx(nCtx).With("task-id", transferHandler.GetTaskID(), "host-id", info.Host.HostID).Info("transfer installer done")

	std.InstanceData().Log().
		Zh("传输 installer 安装包完成").
		En("transfer installer done").
		Info()

	return nil
}
