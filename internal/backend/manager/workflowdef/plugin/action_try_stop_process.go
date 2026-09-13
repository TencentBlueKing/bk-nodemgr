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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/retrier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameTryStopProcess the name of action try stop process.
	ActionNameTryStopProcess = "try_stop_process"
)

// NewActionTryStopProcess new an action to try stop process.
func NewActionTryStopProcess(capability *Capability) action.Definition {
	return &actTryStopProcess{
		daoPluginDeployment: capability.StoragePlugin,
		daoPlugin:           capability.StoragePlugin,
		fileHandler:         capability.FileHandler,
		daoHost:             capability.StorageTopo,
		daoDomainGse:        capability.StorageTopo,
		gseHandlerProc:      capability.GSEHandler.NewHandlerProc(),
	}
}

// ActionParamTryStopProcess defines the parameters for actionTryStopProcess.
type ActionParamTryStopProcess struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

// actTryStopProcess ...
type actTryStopProcess struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoPlugin           pluginStg.IDaoPlugin
	fileHandler         file.IReleasePluginHandler
	daoHost             topoStg.IStorageHost
	daoDomainGse        topoStg.IStorageDomainGse
	gseHandlerProc      gse.IHandlerProc
}

// Name returns the name of the action.
func (act *actTryStopProcess) Name() string {
	return ActionNameTryStopProcess
}

// Version returns the version of the action.
func (act *actTryStopProcess) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actTryStopProcess) Description() string {
	return "try stop process by gse."
}

// Timeout returns the timeout of the action.
func (act *actTryStopProcess) Timeout() time.Duration {
	return 1 * time.Minute // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actTryStopProcess) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actTryStopProcess) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actTryStopProcess) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: funlen
// NOCC: golint/fnsize(plugin stop workflow action is one execution step).
func (act *actTryStopProcess) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActionParamTryStopProcess)
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

	nCtx := std.Context()
	host, err := act.daoHost.GetHostByID(nCtx, std.DeployInfo().Process.HostID)
	if err != nil {
		return fmt.Errorf("failed to get host by ID: %w", err)
	}

	plugin, err := act.daoPlugin.GetPlugin(nCtx, std.DeployInfo().Process.PluginName)
	if err != nil {
		return fmt.Errorf("failed to get plugin: %w", err)
	}

	programName := pluginProgramName(plugin.PkgName, host.Dynamic.NodeOsType)

	hostProcessInfo, err := act.gseHandlerProc.QueryProcessInfo(nCtx, plugin.Name, programName, host.Dynamic.AgentID)
	if err != nil {
		std.InstanceData().Log().
			Zh("查询主机上进程信息失败, agent-id(%s), plugin-name(%s), program-name(%s): %s",
				host.Dynamic.AgentID, plugin.Name, programName, err.Error()).
			En("failed to query process info from gse, agent-id(%s), plugin-name(%s), program-name(%s): %s",
				host.Dynamic.AgentID, plugin.Name, programName, err.Error()).
			Error()

		return fmt.Errorf("failed to query process info from gse: %w", err)
	}

	if hostProcessInfo.Status != types.ProcessStatusRunning {
		std.InstanceData().Log().
			Zh("主机(%d)上进程状态(%s)不为运行中, 无需停止进程",
				host.HostID, hostProcessInfo.Status).
			En("host(%d) process status(%s) is not running, no need to stop the process",
				host.HostID, hostProcessInfo.Status).
			Info()

		return nil
	}

	processSpec, err := act.buildStopProcessSpec(std, host, plugin, hostProcessInfo.Version)
	if err != nil {
		return fmt.Errorf("failed to build stop process spec: %w", err)
	}

	std.InstanceData().Log().
		Zh("主机上进程状态为运行中, 尝试执行停止插件进程, plugin-name(%s), host-id(%d), cmd(%s)",
			plugin.Name, host.HostID, processSpec.Controller.StopCmd).
		En("process status on host is running, try to execute stop plugin process, "+
			"plugin-name(%s), host-id(%d), cmd(%s)",
			plugin.Name, host.HostID, processSpec.Controller.StopCmd).
		Info()

	std.InstanceData().Log().
		Zh("主机上进程状态, pid(%d), version(%s), agent-id(%s), autostart(%t), status(%s)",
			hostProcessInfo.Pid, hostProcessInfo.Version, hostProcessInfo.AgentID,
			hostProcessInfo.AutoStart, hostProcessInfo.Status).
		En("process status on host, pid(%d), version(%s), agent-id(%s), autostart(%t), status(%s)",
			hostProcessInfo.Pid, hostProcessInfo.Version, hostProcessInfo.AgentID,
			hostProcessInfo.AutoStart, hostProcessInfo.Status).
		Info()

	result, err := act.gseHandlerProc.UnTrusteeshipAndStopProcess(nCtx, processSpec)
	if err != nil {
		std.InstanceData().Log().
			Zh("执行停止插件进程操作失败, result(%s), err(%s)", result, err.Error()).
			En("failed to execute stop plugin process operation, result(%s), err(%s)", result, err.Error()).
			Warn()
	} else {
		std.InstanceData().Log().
			Zh("成功执行停止插件进程操作, result(%s)", result).
			En("successfully execute stop plugin process operation, result(%s)", result).
			Info()
	}

	std.InstanceData().Log().
		Zh("检查进程状态").
		En("check process status").
		Info()
	polling := retrier.NewPolling(retrier.PollingOpts{
		Timeout:  act.Timeout(),
		Interval: time.Second,
	})

	var processInfo *types.ProcessInfo
	err = polling.Do(nCtx, func(_ int) error {
		processInfo, err = act.gseHandlerProc.QueryProcessInfo(nCtx, processSpec.PluginName, processSpec.Identity.Name, processSpec.AgentID)
		if err != nil {
			return fmt.Errorf("failed to query process info: %w", err)
		}

		if processInfo.Status == types.ProcessStatusRunning {
			std.InstanceData().Log().
				Zh("进程状态仍在运行").
				En("process status is still running").
				Info()

			return errors.New("process status is still running")
		}

		if processInfo.AutoStart {
			std.InstanceData().Log().
				Zh("进程自动启动为 true, 进程将被 GSE 再次重启").
				En("process autostart is true, process will be restart by gse again").
				Info()

			return errors.New("process autostart is true, process will be restart by gse again")
		}

		// update process info
		std.DeployInfo().Process.Info = *processInfo

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to wait process no running: %w", err)
	}

	std.InstanceData().Log().
		Zh("进程已停止, pid(%d), version(%s), agent-id(%s), autostart(%t), status(%s)",
			processInfo.Pid, processInfo.Version, processInfo.AgentID, processInfo.AutoStart, processInfo.Status).
		En("process stopped, pid(%d), version(%s), agent-id(%s), autostart(%t), status(%s)",
			processInfo.Pid, processInfo.Version, processInfo.AgentID, processInfo.AutoStart, processInfo.Status).
		Info()

	return nil
}

func (act *actTryStopProcess) buildStopProcessSpec(std *pluginUtils.PluginActionStandarder, host *types.Host, plugin *types.Plugin, version string) (
	types.ProcessSpec, error) {

	nCtx := std.Context()
	releaseKey := types.ReleasePluginKey{
		Generation: host.Dynamic.NodeGeneration,
		Platform:   platform.NewPlatform(host.Dynamic.NodeOsType, host.Dynamic.NodeCPUArch),
		Version:    version,
		Name:       plugin.PkgName,
	}
	pluginPkg, err := act.fileHandler.GetReleasePlugin(nCtx, releaseKey)
	if err != nil {
		std.InstanceData().Log().
			Zh("查询运行版本的插件包失败, 回退到用户指定版本, plugin-pkg-name(%s), version(%s), fallback-version(%s): %s",
				releaseKey.Name, version, std.DeployInfo().InstallOptions.Version, err.Error()).
			En("failed to get plugin release for running version, falling back to user-specified version, "+
				"plugin-pkg-name(%s), version(%s), fallback-version(%s): %s",
				releaseKey.Name, version, std.DeployInfo().InstallOptions.Version, err.Error()).
			Warn()

		// When taking over a legacy plugin whose reported release cannot be retrieved,
		// fall back to the user-specified release to build stop parameters, even though
		// its version may not match the running plugin.
		releaseKey.Version = std.DeployInfo().InstallOptions.Version
		pluginPkg, err = act.fileHandler.GetReleasePlugin(nCtx, releaseKey)
	}
	if err != nil {
		return types.ProcessSpec{}, fmt.Errorf("failed to get plugin release, name(%s), version(%s): %w",
			releaseKey.Name, releaseKey.Version, err)
	}

	pluginConstant, err := deployconstant.GetPluginDeployConf(host.Dynamic.NodeGeneration, host.Dynamic.NodeOsType)
	if err != nil {
		return types.ProcessSpec{}, fmt.Errorf("failed to get plugin deploy conf: %w", err)
	}
	customDeployConfig, err := act.daoDomainGse.GetNetworkUnitCustomDeployConfig(nCtx,
		host.Dynamic.NetworkUnitID, host.Dynamic.NodeOsType)
	if err != nil {
		return types.ProcessSpec{}, fmt.Errorf("failed to get network unit custom deploy config: %w", err)
	}
	pluginConstant.BaseDeployDir = conv.NonEmptyOr(customDeployConfig.PluginRuntime.BaseDeployDir, pluginConstant.BaseDeployDir)

	if err := pluginUtils.EnsureHostLoginUser(std, host); err != nil {
		return types.ProcessSpec{}, fmt.Errorf("failed to ensure host login user: %w", err)
	}

	process := types.Process{
		PluginName:    plugin.Name,
		PluginPkgName: plugin.PkgName,
		Platform:      releaseKey.Platform,
	}
	runtime := types.PluginDeploymentBaseRuntime{
		PluginHomeDir: pluginConstant.GeneratePluginHomeDir(plugin.Group, plugin.Name),
		RunDir:        pluginConstant.GeneratePluginRunDir(plugin.Group, plugin.Name),
		LogDir:        conv.NonEmptyOr(customDeployConfig.PluginRuntime.LogDir, pluginConstant.LogDir),
	}

	return buildPluginProcessSpec(process, host, pluginPkg, runtime), nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actTryStopProcess) DisplayNameZh() string {
	return "尝试停止进程"
}

// DisplayNameEn returns the English display name of the action.
func (act *actTryStopProcess) DisplayNameEn() string {
	return "Try Stop Process"
}
