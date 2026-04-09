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
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/winpath"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameTransferPluginPkgToNode defines the action name.
	ActionNameTransferPluginPkgToNode = "transfer_plugin_pkg_to_node"
)

// NewActionTransferPluginPkgToNode get a new action.
func NewActionTransferPluginPkgToNode(capability *Capability) action.Definition {
	return &actionTransferPluginPkgToNode{
		daoPluginDeployment: capability.StoragePlugin,
		fileHandler:         capability.FileHandler,
		daoHost:             capability.StorageTopo,
		gseHandler:          capability.GSEHandler,
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
	gseHandler          gse.IHandler
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
	return 10 * time.Minute // nolint: mnd
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

	std.InstanceData().Log().
		Zh("开始传输插件包到节点").
		En("transfer plugin pkg to node start.").
		Info()

	nCtx := std.Context()
	targetHost, err := act.daoHost.GetHostByID(nCtx, std.DeployInfo().Process.HostID)
	if err != nil {
		return fmt.Errorf("failed to get host by id, host-id(%d): %w", std.DeployInfo().Process.HostID, err)
	}

	gp := gopool.NewPool()
	if !std.DeployInfo().TransferOptions.SelectDownloads || std.DeployInfo().TransferOptions.EnableReleasePackage {
		gp.Go(func() error {
			std.InstanceData().Log().
				Zh("开始传输发布包").
				En("transfer release start.").
				Info()
			defer std.InstanceData().Log().
				Zh("传输发布包完成").
				En("transfer release done.").
				Info()

			return act.transferRelease(nCtx, std.DeployInfo(), targetHost)
		})
	}
	if !std.DeployInfo().TransferOptions.SelectDownloads || std.DeployInfo().TransferOptions.EnableInstaller {
		gp.Go(func() error {
			std.InstanceData().Log().
				Zh("开始传输安装器").
				En("transfer installer start.").
				Info()
			defer std.InstanceData().Log().
				Zh("传输安装器完成").
				En("transfer installer done.").
				Info()

			return act.transferInstaller(nCtx, std.DeployInfo(), targetHost)
		})
	}
	// offline mode: push config files since installer cannot callback to fetch them.
	if std.DeployInfo().InstallOptions.IsOffline {
		gp.Go(func() error {
			return act.pushOfflinePluginConfig(nCtx, std, targetHost)
		})
	}

	if err := gp.Wait(); err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("host-id", targetHost.HostID).Error("failed to transfer plugin pkg to node.")
		return err
	}

	logger.G.Biz(nCtx).With("host-id", targetHost.HostID).Info("transfer plugin pkg to node all done.")
	std.InstanceData().Log().
		Zh("传输插件包到节点全部完成").
		En("transfer plugin pkg to node all done.").
		Info()

	return nil
}

func (act *actionTransferPluginPkgToNode) transferRelease(nCtx contextx.IContext, info *types.PluginDeploymentInfo, targetHost *types.Host) error {
	var dataDir string
	if targetHost.Dynamic.NodeOsType == criteria.OSWindows {
		dataDir = winpath.Join(info.InstallerRuntime.WorkDir, "data", "plugin", info.Process.PluginPkgName)
	} else {
		dataDir = filepath.Join(info.InstallerRuntime.WorkDir, "data", "plugin", info.Process.PluginPkgName)
	}

	transferHandler, err := act.fileHandler.LaunchTransferPlugin(
		nCtx,
		info.Process.PluginPkgName,
		info.Process.Generation,
		info.Process.Platform,
		info.Process.Info.Version,
		dataDir,
		targetHost)
	if err != nil {
		return fmt.Errorf("failed to launch transfer release, host-id(%d): %w", targetHost.HostID, err)
	}

	logger.G.Biz(nCtx).With("task-id", transferHandler.GetTaskID(), "host-id", targetHost.HostID).Info("launched transfer release.")

	result, err := transferHandler.WaitUntilDone(nCtx)
	if err != nil {
		return fmt.Errorf("failed to wait until transfer release done, task-id(%s), host-id(%d): %w",
			transferHandler.GetTaskID(), targetHost.HostID, err)
	}

	if !result.Terminated {
		return fmt.Errorf("transfer release not terminated, task-id(%s), host-id(%d)",
			transferHandler.GetTaskID(), targetHost.HostID)
	}

	if result.ErrorCode != 0 {
		return fmt.Errorf("transfer release failed, task-id(%s), host-id(%d), err-code(%d), err-msg(%s)",
			transferHandler.GetTaskID(), targetHost.HostID, result.ErrorCode, result.ErrorMessage)
	}

	logger.G.Biz(nCtx).With("task-id", transferHandler.GetTaskID(), "host-id", targetHost.HostID).Info("transfer release done.")

	return nil
}

func (act *actionTransferPluginPkgToNode) transferInstaller(nCtx contextx.IContext, info *types.PluginDeploymentInfo, targetHost *types.Host) error {
	transferHandler, err := act.fileHandler.LaunchTransferInstaller(nCtx,
		types.Generation2,
		platfmt.Platform{
			OS:   targetHost.Dynamic.NodeOsType,
			Arch: targetHost.Dynamic.NodeCPUArch,
		},
		info.InstallerRuntime.WorkDir,
		targetHost)
	if err != nil {
		return fmt.Errorf("failed to launch transfer installer, host-id(%d): %w", targetHost.HostID, err)
	}

	logger.G.Biz(nCtx).With("task-id", transferHandler.GetTaskID(), "host-id", targetHost.HostID).Info("launched transfer installer.")

	result, err := transferHandler.WaitUntilDone(nCtx)
	if err != nil {
		return fmt.Errorf("failed to wait until transfer installer done, task-id(%s), host-id(%d): %w",
			transferHandler.GetTaskID(), targetHost.HostID, err)
	}

	if !result.Terminated {
		return fmt.Errorf("transfer installer not terminated, task-id(%s), host-id(%d)",
			transferHandler.GetTaskID(), targetHost.HostID)
	}

	if result.ErrorCode != 0 {
		return fmt.Errorf("transfer installer failed, task-id(%s), host-id(%d), err-code(%d), err-msg(%s)",
			transferHandler.GetTaskID(), targetHost.HostID, result.ErrorCode, result.ErrorMessage)
	}

	logger.G.Biz(nCtx).With("task-id", transferHandler.GetTaskID(), "host-id", targetHost.HostID).Info("transfer installer done.")

	return nil
}

// pushOfflinePluginConfig pushes plugin config files to ConfigDir via GSE in offline mode.
// In offline mode, installer cannot callback to fetch config, so we push it beforehand.
func (act *actionTransferPluginPkgToNode) pushOfflinePluginConfig(
	nCtx contextx.IContext,
	std *pluginUtils.PluginActionStandarder,
	targetHost *types.Host,
) error {

	if targetHost.Dynamic.LoginUser == "" {
		return fmt.Errorf("host login user is empty, host-id(%d)", targetHost.HostID)
	}

	var configDir string
	if targetHost.Dynamic.NodeOsType == criteria.OSWindows {
		configDir = winpath.Join(std.DeployInfo().InstallerRuntime.WorkDir, "data", "plugin", std.DeployInfo().Process.PluginPkgName, "config")
	} else {
		configDir = filepath.Join(std.DeployInfo().InstallerRuntime.WorkDir, "data", "plugin", std.DeployInfo().Process.PluginPkgName, "config")
	}

	if err := pluginUtils.CheckDirPathSafe(configDir, std.DeployInfo().Process.Platform.OS); err != nil {
		return fmt.Errorf("check config store dir safe failed, dir(%s): %w", configDir, err)
	}

	pluginConf, err := act.daoPluginDeployment.GetPluginDeploymentPluginConfConfigFilesDetail(nCtx, std.Token())
	if err != nil {
		return fmt.Errorf("failed to get plugin config: %w", err)
	}

	endpoints := []*types.Endpoint{{AgentID: targetHost.Dynamic.AgentID}}
	tasks := make([]*types.PushFileDetail, 0, len(pluginConf))
	for _, conf := range pluginConf {
		if conf == nil || !conf.IsMainConfig {
			continue
		}

		tasks = append(tasks, &types.PushFileDetail{
			FileName:    conf.Name,
			FileContent: conf.Content,
			StoreDir:    configDir,
			Owner:       targetHost.Dynamic.LoginUser,
			Endpoints:   endpoints,
		})

		std.InstanceData().Log().
			Zh("准备推送插件主配置文件(%s)到主机(%d), 目录(%s)", conf.Name, targetHost.HostID, configDir).
			En("prepare to push plugin main config file(%s) to host(%d) in dir(%s)", conf.Name, targetHost.HostID, configDir).
			Info()
	}

	if len(tasks) == 0 {
		std.InstanceData().Log().
			Zh("无需推送插件主配置").
			En("no plugin main config need to push").
			Info()

		return nil
	}

	std.InstanceData().Log().
		Zh("开始推送 %d 个插件主配置文件到主机(%d)", len(tasks), targetHost.HostID).
		En("start to push %d plugin main config files to host(%d)", len(tasks), targetHost.HostID).
		Info()

	// Need GSE handler - will be added to capability
	if act.gseHandler == nil {
		return fmt.Errorf("gse handler is nil, cannot push config files")
	}

	taskID, err := act.gseHandler.PushFile(nCtx, tasks...)
	if err != nil {
		return fmt.Errorf("failed to push config files: %w", err)
	}

	pushResult, err := act.gseHandler.QueryPushFileFinalResult(nCtx, taskID)
	if err != nil {
		return fmt.Errorf("failed to query push config result: %w", err)
	}

	if !pushResult.Terminated {
		return fmt.Errorf("push config not terminated, task-id(%s)", taskID)
	}

	if pushResult.ErrorCode != 0 {
		return fmt.Errorf("push config failed, task-id(%s), err-code(%d), err-msg(%s)", taskID, pushResult.ErrorCode, pushResult.ErrorMessage)
	}

	std.InstanceData().Log().
		Zh("插件主配置推送成功, task-id(%s)", taskID).
		En("plugin main config pushed successfully, task-id(%s)", taskID).
		Info()

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionTransferPluginPkgToNode) DisplayNameZh() string {
	return "传输插件包到节点"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionTransferPluginPkgToNode) DisplayNameEn() string {
	return "Transfer Plugin Package to Node"
}
