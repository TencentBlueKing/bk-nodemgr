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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/retrier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameStopMismatchVersionPluginProcessV2 the name of action stop the plugin process whose version
	// mismatches the target version.
	ActionNameStopMismatchVersionPluginProcessV2 = "stop_mismatch_version_plugin_process_v2"
)

// NewActionStopMismatchVersionPluginProcessV2 new an action to stop the plugin process whose version mismatches
// the target version.
func NewActionStopMismatchVersionPluginProcessV2(capability *Capability) action.Definition {
	return &actionStopMismatchVersionPluginProcessV2{
		daoPluginDeployment: capability.StoragePlugin,
		daoProcessV2:        capability.StoragePlugin,
		fileHandler:         capability.FileHandler,
		gseHandlerProc:      capability.GSEHandler.NewHandlerProc(gse.WithProcNameSpace(procNameSpaceNodeMan)),
	}
}

// ActParamStopMismatchVersionPluginProcessV2 defines the parameters for actionStopMismatchVersionPluginProcessV2.
type ActParamStopMismatchVersionPluginProcessV2 struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

// actionStopMismatchVersionPluginProcessV2 stops the running plugin process whose version mismatches the target
// version, so the old process releases the plugin package files before the package is transferred.
type actionStopMismatchVersionPluginProcessV2 struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoProcessV2        pluginStg.IDaoProcessV2
	fileHandler         file.IReleasePluginHandler
	gseHandlerProc      gse.IHandlerProc
}

// Name returns the name of the action.
func (act *actionStopMismatchVersionPluginProcessV2) Name() string {
	return ActionNameStopMismatchVersionPluginProcessV2
}

// Version returns the version of the action.
func (act *actionStopMismatchVersionPluginProcessV2) Version() string {
	return "1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionStopMismatchVersionPluginProcessV2) Description() string {
	return "stop the plugin process whose version mismatches the target version"
}

// Timeout returns the timeout of the action.
func (act *actionStopMismatchVersionPluginProcessV2) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionStopMismatchVersionPluginProcessV2) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionStopMismatchVersionPluginProcessV2) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionStopMismatchVersionPluginProcessV2) DelayFn(attempt int) func() {
	return action.DefaultBackoffDelayFn(act, attempt)
}

// Do this func define what the action will do.
func (act *actionStopMismatchVersionPluginProcessV2) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamStopMismatchVersionPluginProcessV2)
	err = conv.MapToStruct(ctx.Data.Content, param)
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
	hostID := deployInfo.Process.HostID
	pluginName := deployInfo.Process.PluginName
	targetVersion := deployInfo.InstallOptions.Version

	runningProcess, err := act.daoProcessV2.GetProcessV2(nCtx, hostID, pluginName)
	if err != nil && !errors.Is(err, base.ErrRecordNoFound()) {
		return fmt.Errorf("failed to get the process v2, host-id(%d), plugin-name(%s): %w", hostID, pluginName, err)
	}

	if runningProcess == nil ||
		runningProcess.Info.Status != types.ProcessStatusRunning ||
		targetVersion == "" ||
		runningProcess.Info.Version == targetVersion {

		std.InstanceData().Log().
			Zh("无需停止版本不一致的 V2 插件进程, 主机id(%d), 插件名(%s), 目标版本(%s)",
				hostID, pluginName, targetVersion).
			En("no need to stop the mismatch version v2 plugin process, host-id(%d), plugin-name(%s), "+
				"target-version(%s)", hostID, pluginName, targetVersion).
			Info()

		return nil
	}

	runningVersion := runningProcess.Info.Version
	processSpec, err := act.buildStopProcessSpec(std, runningProcess)
	if err != nil {
		return err
	}

	result, err := act.gseHandlerProc.UnTrusteeshipAndStopProcess(nCtx, processSpec)
	if err != nil {
		return fmt.Errorf("failed to stop the mismatch version v2 plugin process, cmd(%s): %w",
			processSpec.Controller.StopCmd, err)
	}

	std.InstanceData().Log().
		Zh("成功执行停止 V2 插件进程操作, 主机id(%d), 插件名(%s), 运行版本(%s), 目标版本(%s), cmd(%s), result(%s)",
			hostID, pluginName, runningVersion, targetVersion, processSpec.Controller.StopCmd, result).
		En("successfully executed the stop v2 plugin process operation, host-id(%d), plugin-name(%s), "+
			"running-version(%s), target-version(%s), cmd(%s), result(%s)",
			hostID, pluginName, runningVersion, targetVersion, processSpec.Controller.StopCmd, result).
		Info()

	polling := retrier.NewPolling(retrier.PollingOpts{
		Timeout:  act.Timeout(),
		Interval: time.Second,
	})

	var processInfo *types.ProcessInfo
	err = polling.Do(nCtx, func(_ int) error {
		processInfo, err = act.gseHandlerProc.QueryProcessInfo(nCtx, pluginName, processSpec.Identity.Name, processSpec.AgentID)
		if err != nil {
			return fmt.Errorf("failed to query process info: %w", err)
		}

		if processInfo.Status != types.ProcessStatusStopped {
			std.InstanceData().Log().
				Zh("V2 进程状态未停止, status(%s)", processInfo.Status).
				En("V2 process status is not stopped, status(%s)", processInfo.Status).
				Info()

			return fmt.Errorf("V2 process status is not stopped, status(%s)", processInfo.Status)
		}

		if processInfo.AutoStart {
			std.InstanceData().Log().
				Zh("V2 进程仍被 GSE 托管").
				En("V2 process is still trusteeship by gse").
				Info()

			return fmt.Errorf("V2 process is still trusteeship by gse")
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to wait V2 plugin process stopped: %w", err)
	}

	// Persist the observed state without replacing the deployment's target version.
	if err = act.daoProcessV2.UpdateProcessInfoV2(nCtx, hostID, pluginName, processInfo); err != nil {
		return fmt.Errorf("failed to update process v2 info: %w", err)
	}

	return nil
}

func (act *actionStopMismatchVersionPluginProcessV2) buildStopProcessSpec(std *pluginV2Utils.PluginActionStandarder, runningProcess *types.Process) (
	types.ProcessSpec, error) {

	pluginPkg, err := act.getRunningProcessPluginPkg(std, runningProcess.Info.Version)
	if err != nil {
		return types.ProcessSpec{}, err
	}

	deployInfo := std.DeployInfo()
	processSpec := deployInfo.Process.ToProcessSpec()
	processSpec.AgentID = runningProcess.Info.AgentID
	processSpec.Controller = pluginPkg.PluginController
	processSpec.Identity = runningProcess.Identity
	if err := processSpec.Identity.Validate(); err != nil {
		// Legacy process records may not have an identity; rebuild the config path from the running release.
		processSpec.Identity = deployInfo.Process.Identity
		processSpec.Identity.ConfigPath = ""
		for _, configTemplate := range pluginPkg.ConfigTemplates {
			if !configTemplate.IsMainConfig {
				continue
			}
			processSpec.Identity.ConfigPath = tool.JoinPath(deployInfo.Process.Platform.OS,
				deployInfo.BaseRuntime.PluginHomeDir, configTemplate.FilePath, configTemplate.Name)

			break
		}
	}

	return processSpec, nil
}

// getRunningProcessPluginPkg gets the plugin package of the running process.
func (act *actionStopMismatchVersionPluginProcessV2) getRunningProcessPluginPkg(
	std *pluginV2Utils.PluginActionStandarder, version string) (*types.ReleasePlugin, error) {

	deployProcess := std.DeployInfo().Process
	releaseKey := types.ReleasePluginKey{
		Generation: deployProcess.Generation,
		Platform:   deployProcess.Platform,
		Version:    version,
		Name:       deployProcess.PluginPkgName,
	}

	pluginPkg, err := act.fileHandler.GetReleasePlugin(std.Context(), releaseKey)
	if err == nil {
		return pluginPkg, nil
	}

	std.InstanceData().Log().
		Zh("未获取到运行版本(%s)的插件包信息, 改用默认版本的插件包信息拼接停止进程参数, 主机id(%d), 插件名(%s), 错误(%s)",
			version, deployProcess.HostID, deployProcess.PluginName, err).
		En("failed to get the plugin package of the running version(%s), use the default version plugin "+
			"package to build the stop process spec instead, host-id(%d), plugin-name(%s), error(%s)",
			version, deployProcess.HostID, deployProcess.PluginName, err).
		Warn()

	if releaseKey.Version, err = act.fileHandler.GetReleasePluginDefaultVersion(
		std.Context(), releaseKey.Name, releaseKey.Generation, releaseKey.Platform); err != nil {
		return nil, fmt.Errorf("failed to get the default version of plugin-pkg(%s): %w", releaseKey.Name, err)
	}

	pluginPkg, err = act.fileHandler.GetReleasePlugin(std.Context(), releaseKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get the plugin package of the default version(%s): %w", releaseKey.Version, err)
	}

	return pluginPkg, nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionStopMismatchVersionPluginProcessV2) DisplayNameZh() string {
	return "停止版本不一致的 V2 插件进程"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionStopMismatchVersionPluginProcessV2) DisplayNameEn() string {
	return "Stop The Mismatch Version V2 Plugin Process"
}
