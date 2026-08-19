/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package pluginv2

import (
	"errors"
	"fmt"
	"time"

	pluginV2Utils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/pluginv2/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	releaseStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameFetchPluginProcessV2 the name of action fetch plugin process.
	ActionNameFetchPluginProcessV2 = "fetch_plugin_process_v2"
)

// NewActionFetchPluginProcessV2 new an action to fetch plugin process.
func NewActionFetchPluginProcessV2(capability *Capability) action.Definition {
	return &actionFetchPluginProcessV2{
		daoPluginDeployment: capability.StoragePlugin,
		daoReleasePlugin:    capability.StorageRelease,
		daoHost:             capability.StorageTopo,
		gseHandlerProc:      capability.GSEHandler.NewHandlerProc(gse.WithProcNameSpace(procNameSpaceNodeMan)),
	}
}

// ActParamFetchPluginProcessV2 defines the parameters for actionFetchPluginProcessV2.
type ActParamFetchPluginProcessV2 struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

type actionFetchPluginProcessV2 struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoReleasePlugin    releaseStg.IPlugin
	daoHost             topoStg.IStorageHost
	gseHandlerProc      gse.IHandlerProc
}

// Name returns the name of the action.
func (act *actionFetchPluginProcessV2) Name() string {
	return ActionNameFetchPluginProcessV2
}

// Version returns the version of the action.
func (act *actionFetchPluginProcessV2) Version() string {
	return "1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionFetchPluginProcessV2) Description() string {
	return "fetch plugin process v2"
}

// Timeout returns the timeout of the action.
func (act *actionFetchPluginProcessV2) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionFetchPluginProcessV2) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionFetchPluginProcessV2) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionFetchPluginProcessV2) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionFetchPluginProcessV2) Do(ctx *action.InstanceContext) error {
	param := new(ActParamFetchPluginProcessV2)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := pluginV2Utils.NewPluginActionStandarder(act.daoPluginDeployment)
	if err = std.Initialize(ctx, param.PluginActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	nCtx := std.Context()
	deployInfo := std.DeployInfo()
	host, err := act.daoHost.GetHostByID(nCtx, deployInfo.Process.HostID)
	if err != nil {
		std.InstanceData().Log().
			Zh("获取主机信息失败, 主机id(%d), 错误(%s)", deployInfo.Process.HostID, err).
			En("fetch host info failed, host-id(%d), error(%s)", deployInfo.Process.HostID, err).
			Error()

		return err
	}

	if err := pluginV2Utils.EnsureHostLoginUser(std, host); err != nil {
		std.InstanceData().Log().
			Zh("获取主机登录用户失败, 主机id(%d), 错误(%s)", deployInfo.Process.HostID, err).
			En("fetch host login user failed, host-id(%d), error(%s)", deployInfo.Process.HostID, err).
			Error()

		return err
	}

	deployInfo.Process.HostID = host.HostID
	deployInfo.Process.BizID = host.Static.BizID
	deployInfo.Process.Generation = host.Dynamic.NodeGeneration
	deployInfo.Process.Platform = platform.NewPlatform(host.Dynamic.NodeOsType, host.Dynamic.NodeCPUArch)

	// In the v2 process, we can consider the process name and the plugin name to be consistent
	processInfo, err := act.gseHandlerProc.QueryProcessInfo(nCtx, deployInfo.Process.PluginName, deployInfo.Process.PluginName, host.Dynamic.AgentID)
	if err != nil {
		return fmt.Errorf("failed to query process info: %w", err)
	}

	std.InstanceData().Log().
		Zh("成功查询插件进程信息, 主机id(%d), 插件名(%s), 进程信息(%+v)", deployInfo.Process.HostID, deployInfo.Process.PluginName, processInfo).
		En("successfully queried plugin process info, host-id(%d), plugin-name(%s), process-info(%+v)",
			deployInfo.Process.HostID, deployInfo.Process.PluginName, processInfo).
		Info()

	deployInfo.Process.Info = *processInfo
	if conv.IsEmpty(deployInfo.Process.Info.Version) {
		version, err := act.daoReleasePlugin.GetReleasePluginDefaultVersion(
			nCtx, deployInfo.Process.PluginName, deployInfo.Process.Generation, deployInfo.Process.Platform)
		if err != nil {
			std.InstanceData().Log().
				Zh("获取插件默认版本失败, 主机id(%d), 插件名(%s), 错误(%s)", deployInfo.Process.HostID, deployInfo.Process.PluginName, err).
				En("fetch plugin default version failed, host-id(%d), plugin-name(%s), error(%s)",
					deployInfo.Process.HostID, deployInfo.Process.PluginName, err).
				Error()

			return err
		}

		deployInfo.Process.Info.Version = version
	}

	pkg, err := act.daoReleasePlugin.GetReleasePlugin(nCtx, types.ReleasePluginKey{
		Generation: deployInfo.Process.Generation,
		Platform:   deployInfo.Process.Platform,
		Version:    deployInfo.Process.Info.Version,
		Name:       deployInfo.Process.PluginName,
	})
	if err != nil {
		std.InstanceData().Log().
			Zh("获取插件包信息失败, 主机id(%d), 插件名(%s), 错误(%s)", deployInfo.Process.HostID, deployInfo.Process.PluginName, err).
			En("fetch plugin package info failed, host-id(%d), plugin-name(%s), error(%s)",
				deployInfo.Process.HostID, deployInfo.Process.PluginName, err).
			Error()

		return err
	}

	deployInfo.Process.Platform = pkg.Platform
	deployInfo.Process.Generation = pkg.Generation
	deployInfo.Process.PluginPkgName = pkg.Name
	deployInfo.Process.Controller = pkg.PluginController

	pidFileName := fmt.Sprintf("%s.pid", deployInfo.Process.PluginPkgName)
	pidFilePath := tool.JoinPath(deployInfo.Process.Platform.OS, deployInfo.BaseRuntime.RunDir, pidFileName)
	deployInfo.Process.Identity = types.ProcessIdentity{
		Name:       deployInfo.Process.PluginName,
		SetupPath:  deployInfo.BaseRuntime.PluginHomeDir,
		PidPath:    pidFilePath,
		ConfigPath: deployInfo.BaseRuntime.ConfigDir,
		LogPath:    deployInfo.BaseRuntime.LogDir,
		User:       host.Dynamic.LoginUser,
	}

	std.InstanceData().Log().
		Zh("成功获取插件进程信息, 主机id(%d), 插件名(%s), 进程信息(%+v)",
			deployInfo.Process.HostID, deployInfo.Process.PluginName, deployInfo.Process.Info).
		En("successfully fetched plugin process info, host-id(%d), plugin-name(%s), process-info(%+v)",
			deployInfo.Process.HostID, deployInfo.Process.PluginName, deployInfo.Process.Info).
		Info()

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionFetchPluginProcessV2) DisplayNameZh() string {
	return "获取 V2 插件进程"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionFetchPluginProcessV2) DisplayNameEn() string {
	return "Fetch V2 Plugin Process"
}
