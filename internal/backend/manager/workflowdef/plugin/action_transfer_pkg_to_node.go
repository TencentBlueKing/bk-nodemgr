/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package plugin

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
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
	// ActionNameTransferPluginPkgToNode defines the action name.
	ActionNameTransferPluginPkgToNode = "transfer_plugin_pkg_to_node"
)

// NewActionTransferPluginPkgToNode get a new action.
func NewActionTransferPluginPkgToNode(daoPluginDeployment pluginStg.IDaoPluginDeployment, daoHost topoStg.IStorageHost, fileHandler file.IHandler) action.Definition {

	return &actionTransferPluginPkgToNode{
		daoPluginDeployment: daoPluginDeployment,
		fileHandler:         fileHandler,
		daoHost:             daoHost,
	}
}

// ActionParamTransferPluginPkgToNode defines the action param.
type ActionParamTransferPluginPkgToNode struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

type actionTransferPluginPkgToNode struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoHost             topoStg.IStorageHost
	fileHandler         file.IHandler
}

// Name returns the name of the action.
func (act *actionTransferPluginPkgToNode) Name() string {
	return ActionNameTransferPluginPkgToNode
}

// Version returns the version of the action.
func (act *actionTransferPluginPkgToNode) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionTransferPluginPkgToNode) Description() string {
	return "transfer plugin pkg to node"
}

// Timeout returns the timeout of the action.
func (act *actionTransferPluginPkgToNode) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionTransferPluginPkgToNode) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionTransferPluginPkgToNode) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionTransferPluginPkgToNode) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: funlen,nonamedreturns
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionTransferPluginPkgToNode) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActionParamTransferPluginPkgToNode)
	err = conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		err = fmt.Errorf("failed to convert param, err: %w", err)

		return err
	}

	// initialize standard data.
	std := pluginUtils.NewPluginActionStandarder(act.daoPluginDeployment)
	if err = std.Initialize(ctx, param.PluginActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	ctx.Data.LogI("transfer plugin pkg to node start.")

	nCtx := std.Context()
	targetHost, err := act.daoHost.GetHostByID(nCtx, std.DeployInfo().Plugin.HostID)
	if err != nil {
		return fmt.Errorf("failed to get host by id. host-id(%d): %w", std.DeployInfo().Plugin.HostID, err)
	}

	deployConstant, err := deployconstant.GetPluginDeployConf(std.DeployInfo().Plugin.Generation, std.DeployInfo().Plugin.Platform.OS)
	if err != nil {
		return fmt.Errorf("failed to get deploy constant, err: %w", err)
	}

	// installer workdir priority: user specified in info > deploy constant default.
	if std.DeployInfo().InstallerWorkDir == "" {
		std.DeployInfo().InstallerWorkDir = deployConstant.WorkDir
	}

	gp := gopool.NewPool()
	if !std.DeployInfo().TransferOptions.SelectDownloads || std.DeployInfo().TransferOptions.EnableReleasePackage {
		gp.Go(func() error {
			ctx.Data.LogI("transfer release start.")
			defer ctx.Data.LogI("transfer release done.")

			if err := act.transferRelease(nCtx, std.DeployInfo(), targetHost); err != nil {
				return fmt.Errorf("failed to transfer release. host-id(%d), err: %w", targetHost.HostID, err)
			}

			return nil
		})
	}
	if !std.DeployInfo().TransferOptions.SelectDownloads || std.DeployInfo().TransferOptions.EnableInstaller {
		gp.Go(func() error {
			return act.transferInstaller(nCtx, std.DeployInfo(), targetHost)
		})
	}

	if err := gp.Wait(); err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("host-id", targetHost.HostID).Error("failed to transfer plugin pkg to node.")
		return err
	}

	logger.G.Biz(nCtx).With("host-id", targetHost.HostID).Info("transfer plugin pkg to node all done.")
	ctx.Data.LogI("transfer plugin pkg to node all done.")

	return nil
}

func (act *actionTransferPluginPkgToNode) transferRelease(nCtx contextx.IContext, info *types.PluginDeploymentInfo, targetHost *types.Host) error {
	rt, err := types.ConvPluginTypeToReleaseType(info.Plugin.Type)
	if err != nil {
		return fmt.Errorf("failed to convert plugin type to release type: %w", err)
	}

	var dataDir string
	if targetHost.Dynamic.NodeOsType == criteria.OSWindows {
		dataDir = winpath.Join(info.InstallerWorkDir, "data", "plugin", info.Plugin.Name)
	} else {
		dataDir = filepath.Join(info.InstallerWorkDir, "data", "plugin", info.Plugin.Name)
	}

	transferHandler, err := act.fileHandler.LaunchTransferPlugin(nCtx,
		info.Plugin.Name,
		info.Plugin.Generation,
		rt,
		info.Plugin.Platform,
		info.Plugin.Version,
		dataDir,
		targetHost)
	if err != nil {
		return fmt.Errorf("failed to launch transfer release. host-id(%d), err: %w", targetHost.HostID, err)
	}

	logger.G.Biz(nCtx).Info("launched transfer release. task-id(%s), host-id(%d)",
		transferHandler.GetTaskID(), targetHost.HostID)

	result, err := transferHandler.WaitUntilDone(nCtx)
	if err != nil {
		return fmt.Errorf("failed to wait until transfer release done. task-id(%s), host-id(%d), err: %w",
			transferHandler.GetTaskID(), targetHost.HostID, err)
	}

	if !result.Terminated {
		return fmt.Errorf("transfer release not terminated. task-id(%s), host-id(%d)",
			transferHandler.GetTaskID(), targetHost.HostID)
	}

	if result.ErrorCode != 0 {
		return fmt.Errorf("transfer release failed. task-id(%s), host-id(%d), err-code(%d), err-msg(%s)",
			transferHandler.GetTaskID(), targetHost.HostID, result.ErrorCode, result.ErrorMessage)
	}

	logger.G.Biz(nCtx).Info("transfer release done. task-id(%s), host-id(%d)",
		transferHandler.GetTaskID(), targetHost.HostID)

	return nil
}

func (act *actionTransferPluginPkgToNode) transferInstaller(ctx contextx.IContext, info *types.PluginDeploymentInfo, targetHost *types.Host) error {
	transferHandler, err := act.fileHandler.LaunchTransferInstaller(ctx,
		types.Generation2,
		platform.Platform{
			OS:   targetHost.Dynamic.NodeOsType,
			Arch: targetHost.Dynamic.NodeCPUArch,
		},
		info.InstallerWorkDir,
		targetHost)
	if err != nil {
		return fmt.Errorf("failed to launch transfer installer. host-id(%d), err: %w", targetHost.HostID, err)
	}

	logger.G.Biz(ctx).Info("launched transfer installer. task-id(%s), host-id(%d)",
		transferHandler.GetTaskID(), targetHost.HostID)

	result, err := transferHandler.WaitUntilDone(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait until transfer installer done. task-id(%s), host-id(%d), err: %w",
			transferHandler.GetTaskID(), targetHost.HostID, err)
	}

	if !result.Terminated {
		return fmt.Errorf("transfer installer not terminated. task-id(%s), host-id(%d)",
			transferHandler.GetTaskID(), targetHost.HostID)
	}

	if result.ErrorCode != 0 {
		return fmt.Errorf("transfer installer failed. task-id(%s), host-id(%d), err-code(%d), err-msg(%s)",
			transferHandler.GetTaskID(), targetHost.HostID, result.ErrorCode, result.ErrorMessage)
	}

	logger.G.Biz(ctx).Info("transfer installer done. task-id(%s), host-id(%d)",
		transferHandler.GetTaskID(), targetHost.HostID)

	return nil
}
