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

package plugin

import (
	"errors"
	"fmt"
	"time"

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameCheckV2PluginNoAlive identifies the action that checks whether the V2 plugin is running.
	ActionNameCheckV2PluginNoAlive = "check_v2_plugin_no_alive"
	procNameSpaceV2                = "nodeman"
)

// NewActionCheckV2PluginNoAlive creates an action that checks whether the V2 plugin is running.
func NewActionCheckV2PluginNoAlive(capability *Capability) action.Definition {
	return &actionCheckV2PluginNoAlive{
		daoPluginDeployment: capability.StoragePlugin,
		daoHost:             capability.StorageTopo,
		daoPlugin:           capability.StoragePlugin,
		fileHandler:         capability.FileHandler,
		gseHandlerProc:      capability.GSEHandler.NewHandlerProc(gse.WithProcNameSpace(procNameSpaceV2)),
	}
}

// ActParamCheckV2PluginNoAlive defines the parameters for checking the V2 plugin's running status.
type ActParamCheckV2PluginNoAlive struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

type actionCheckV2PluginNoAlive struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoHost             topoStg.IStorageHost
	daoPlugin           pluginStg.IDaoPlugin
	fileHandler         file.IReleasePluginHandler
	gseHandlerProc      gse.IHandlerProc
}

// Name returns the action name.
func (act *actionCheckV2PluginNoAlive) Name() string {
	return ActionNameCheckV2PluginNoAlive
}

// Version returns the action version.
func (act *actionCheckV2PluginNoAlive) Version() string {
	return "1.0.0" // nolint: goconst
}

// Description returns the action description.
func (act *actionCheckV2PluginNoAlive) Description() string {
	return "check whether the V2 plugin is running"
}

// Timeout returns the action timeout.
func (act *actionCheckV2PluginNoAlive) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the action tags.
func (act *actionCheckV2PluginNoAlive) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the maximum retry count.
func (act *actionCheckV2PluginNoAlive) MaxRetryCount() uint {
	return 1 // nolint: mnd
}

// DelayFn returns the delay between failed attempts.
func (act *actionCheckV2PluginNoAlive) DelayFn(attempt int) func() {
	return action.DefaultBackoffDelayFn(act, attempt)
}

// Do checks whether the V2 plugin is running when the selected release conflicts with it.
func (act *actionCheckV2PluginNoAlive) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamCheckV2PluginNoAlive)
	if err = conv.MapToStruct(ctx.Data.Content, param); err != nil {
		return fmt.Errorf("failed to decode V2 plugin precheck parameters: %w", err)
	}

	std := pluginUtils.NewPluginActionStandarder(act.daoPluginDeployment)
	if err = std.Initialize(ctx, param.PluginActionStandardParam); err != nil {
		return fmt.Errorf("failed to initialize V2 plugin precheck: %w", err)
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	deployInfo := std.DeployInfo()
	hostID := deployInfo.Process.HostID
	pluginName := deployInfo.Process.PluginName
	version := deployInfo.InstallOptions.Version
	host, err := act.daoHost.GetHostByID(std.Context(), hostID)
	if err != nil {
		return fmt.Errorf("failed to get host for V2 plugin precheck, host-id(%d), plugin-name(%s): %w", hostID, pluginName, err)
	}
	if host.Dynamic == nil {
		return fmt.Errorf("host dynamic info is empty for V2 plugin precheck, host-id(%d), plugin-name(%s)", hostID, pluginName)
	}
	plugin, err := act.daoPlugin.GetPlugin(std.Context(), pluginName)
	if err != nil {
		return fmt.Errorf("failed to get plugin for V2 plugin precheck, host-id(%d), plugin-name(%s): %w", hostID, pluginName, err)
	}
	pluginPlatform := platform.NewPlatform(host.Dynamic.NodeOsType, host.Dynamic.NodeCPUArch)
	if version == "" {
		// Preserve default-version installs while keeping the precheck and installation on the same release.
		version, err = act.fileHandler.GetReleasePluginDefaultVersion(std.Context(), plugin.PkgName, host.Dynamic.NodeGeneration, pluginPlatform)
		if err != nil {
			return fmt.Errorf("failed to get default plugin version for V2 plugin precheck, plugin-pkg-name(%s): %w", plugin.PkgName, err)
		}
	}
	pluginPkg, err := act.fileHandler.GetReleasePlugin(std.Context(), types.ReleasePluginKey{
		Generation: host.Dynamic.NodeGeneration,
		Platform:   pluginPlatform,
		Version:    version,
		Name:       plugin.PkgName,
	})
	if err != nil {
		return fmt.Errorf("failed to get plugin release for V2 plugin precheck, plugin-pkg-name(%s), version(%s): %w", plugin.PkgName, version, err)
	}
	deployInfo.InstallOptions.Version = version
	if !pluginPkg.LegacyConflict {
		std.InstanceData().Log().
			Zh("此插件无需检查 V2 插件运行状态，跳过检查, namespace(%s), host-id(%d), plugin-name(%s)",
				procNameSpaceV2, deployInfo.Process.HostID, deployInfo.Process.PluginName).
			En("V2 plugin running status check is not required for this plugin, skip checking, namespace(%s), host-id(%d), plugin-name(%s)",
				procNameSpaceV2, deployInfo.Process.HostID, deployInfo.Process.PluginName).
			Info()

		return nil
	}

	if host.Dynamic.AgentID == "" {
		return fmt.Errorf("host agent id is empty for V2 plugin precheck, host-id(%d), plugin-name(%s)", hostID, pluginName)
	}

	processInfo, err := act.gseHandlerProc.QueryProcessInfo(std.Context(), pluginName, pluginName, host.Dynamic.AgentID)
	if err != nil {
		return fmt.Errorf("failed to query plugin process, namespace(%s), host-id(%d), plugin-name(%s): %w",
			procNameSpaceV2, hostID, pluginName, err)
	}
	if processInfo.Status == types.ProcessStatusRunning {
		std.InstanceData().Log().
			Zh("停止插件安装, V2插件正在运行, namespace(%s), 主机id(%d), 插件名(%s), 进程状态(%s)",
				procNameSpaceV2, hostID, pluginName, processInfo.Status).
			En("plugin installation blocked, V2 plugin is running, namespace(%s), host-id(%d), plugin-name(%s), process-status(%s)",
				procNameSpaceV2, hostID, pluginName, processInfo.Status).
			Error()

		return fmt.Errorf("v2 plugin is running, installation blocked, namespace(%s), host-id(%d), plugin-name(%s), status(%s)",
			procNameSpaceV2, hostID, pluginName, processInfo.Status)
	}

	std.InstanceData().Log().
		Zh("V2插件未运行, namespace(%s), 主机id(%d), 插件名(%s), 进程状态(%s)",
			procNameSpaceV2, hostID, pluginName, processInfo.Status).
		En("V2 plugin is not running, namespace(%s), host-id(%d), plugin-name(%s), process-status(%s)",
			procNameSpaceV2, hostID, pluginName, processInfo.Status).
		Info()

	return nil
}

// DisplayNameZh returns the Chinese display name.
func (act *actionCheckV2PluginNoAlive) DisplayNameZh() string {
	return "检测 V2 插件是否运行"
}

// DisplayNameEn returns the English display name.
func (act *actionCheckV2PluginNoAlive) DisplayNameEn() string {
	return "Check Whether the V2 Plugin Is Running"
}
