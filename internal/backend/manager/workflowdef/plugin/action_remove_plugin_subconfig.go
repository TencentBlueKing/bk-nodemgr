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
	"path/filepath"
	"strings"
	"time"

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/winpath"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameRemovePluginSubConfig the name of action remove plugin sub config.
	ActionNameRemovePluginSubConfig = "remove_plugin_subconfig"

	// removePluginSubConfigScriptTimeout the timeout of the remove sub config script execution.
	removePluginSubConfigScriptTimeout = 60 * time.Second

	// removePluginSubConfigResultPollInterval the interval between two result polls.
	removePluginSubConfigResultPollInterval = 1 * time.Second
)

// NewActionRemovePluginSubConfig new an action to remove plugin sub config.
func NewActionRemovePluginSubConfig(capability *Capability) action.Definition {
	return &actionRemovePluginSubConfig{
		daoPluginDeployment:   capability.StoragePlugin,
		daoProcessConfig:      capability.StoragePlugin,
		gseHandler:            capability.GSEHandler,
		storageActionInstance: capability.StorageWorkflow,
	}
}

// ActParamRemovePluginSubConfig defines the parameters for actionRemovePluginSubConfig.
type ActParamRemovePluginSubConfig struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

type actionRemovePluginSubConfig struct {
	daoPluginDeployment   pluginStg.IDaoPluginDeployment
	daoProcessConfig      pluginStg.IDaoProcessConfig
	gseHandler            gse.IHandler
	storageActionInstance workflow.IStorageActionInstance
}

// Name returns the name of the action.
func (act *actionRemovePluginSubConfig) Name() string {
	return ActionNameRemovePluginSubConfig
}

// Version returns the version of the action.
func (act *actionRemovePluginSubConfig) Version() string {
	return "1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionRemovePluginSubConfig) Description() string {
	return "remove plugin sub config"
}

// Timeout returns the timeout of the action.
func (act *actionRemovePluginSubConfig) Timeout() time.Duration {
	return 3 * time.Minute // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actionRemovePluginSubConfig) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionRemovePluginSubConfig) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionRemovePluginSubConfig) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionRemovePluginSubConfig) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamRemovePluginSubConfig)
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
	deployInfo := std.DeployInfo()

	pluginConf, err := act.daoPluginDeployment.GetPluginDeploymentPluginConf(nCtx, std.Token())
	if err != nil {
		std.InstanceData().Log().
			Zh("获取插件部署配置失败, %v", err).
			En("failed to get plugin deployment config, %v", err).
			Error()

		return fmt.Errorf("get plugin deployment plugin config failed, err: %w", err)
	}

	configs, _, err := act.daoProcessConfig.ListProcessConfigs(nCtx, types.UnlimitedPage(), &types.ProcessConfigCondition{
		ExactInclude: &types.ProcessConfigExactFields{
			HostID:      []int64{deployInfo.Process.HostID},
			ProcessName: []string{deployInfo.Process.PluginName},
			Name:        pluginConf.RemoveConfigFileName,
		},
	})
	if err != nil {
		std.InstanceData().Log().
			Zh("获取待删除子配置目标失败, %v", err).
			En("failed to get remove plugin sub config targets, %v", err).
			Error()

		return err
	}

	if len(configs) == 0 {
		std.InstanceData().Log().
			Zh("获取待删除子配置目标失败, 没有找到待删除的子配置文件").
			En("failed to get remove plugin sub config targets, no remove sub config file found").
			Error()

		return fmt.Errorf("no remove sub config file found")
	}

	if len(configs) != len(pluginConf.RemoveConfigFileName) {
		std.InstanceData().Log().
			Zh("获取待删除子配置目标失败, 找到的子配置文件与配置文件名不一致").
			En("failed to get remove plugin sub config targets, remove sub config file name not match").
			Error()

		return fmt.Errorf("remove sub config file name not match")
	}

	for _, config := range configs {
		if config.IsMainConfig {
			std.InstanceData().Log().
				Zh("获取待删除子配置目标失败, 主配置文件不能删除, 文件名(%s)", config.Name).
				En("failed to get remove plugin sub config targets, main config file cannot be removed, file-name(%s)", config.Name).
				Error()

			return fmt.Errorf("main config file cannot be removed")
		}
	}

	if deployInfo.BaseRuntime.SubConfigDir == "" {
		std.InstanceData().Log().
			Zh("删除插件子配置失败, 子配置目标目录为空").
			En("failed to remove plugin sub config, sub config target dir is empty").
			Error()

		return errors.New("sub config target dir is empty")
	}

	osType := deployInfo.Process.Platform.OS
	fullPaths, err := act.buildRemoveSubConfigFullPaths(deployInfo.BaseRuntime.SubConfigDir, osType, configs)
	if err != nil {
		std.InstanceData().Log().
			Zh("删除插件子配置失败, %v", err).
			En("failed to remove plugin sub config, %v", err).
			Error()

		return err
	}

	if err = act.executeRemoveSubConfigScript(std, fullPaths); err != nil {
		return err
	}

	std.InstanceData().Log().
		Zh("删除插件子配置成功, 主机ID(%d), 插件名(%s), 文件数(%d)，文件名(%v)",
			deployInfo.Process.HostID, deployInfo.Process.PluginName, len(fullPaths), fullPaths).
		En("remove plugin sub config succeed, host-id(%d), plugin-name(%s), file-count(%d), file-name(%v)",
			deployInfo.Process.HostID, deployInfo.Process.PluginName, len(fullPaths), fullPaths).
		Info()

	return nil
}

func (act *actionRemovePluginSubConfig) executeRemoveSubConfigScript(
	std *pluginUtils.PluginActionStandarder, fullPaths []string) error {

	nCtx := std.Context()
	deployInfo := std.DeployInfo()
	osType := deployInfo.Process.Platform.OS

	var scriptType types.ScriptType
	var scriptContent string
	if osType == criteria.OSWindows {
		scriptType, scriptContent = act.buildWindowsRemoveSubConfigScript(deployInfo.BaseRuntime.SubConfigDir, fullPaths)
	} else {
		scriptType, scriptContent = act.buildUnixRemoveSubConfigScript(deployInfo.BaseRuntime.SubConfigDir, osType, fullPaths)
	}

	std.InstanceData().Log().
		Zh("删除插件子配置脚本内容，脚本类型(%s)，脚本内容(%s)",
			scriptType, scriptContent).
		En("remove plugin sub config script content, script-type(%s), script-content(%s)",
			scriptType, scriptContent).
		Info()

	endpoint := types.Endpoint{AgentID: deployInfo.Process.Info.AgentID}
	taskID, err := act.gseHandler.ExecuteScript(nCtx, scriptType, scriptContent, removePluginSubConfigScriptTimeout,
		&types.EndpointWithAuth{Endpoint: endpoint})
	if err != nil {
		return fmt.Errorf("failed to execute remove sub config script: %w", err)
	}

	if err = act.queryScriptResult(nCtx, taskID, endpoint); err != nil {
		return fmt.Errorf("failed to remove plugin sub config files, task-id(%s): %w", taskID, err)
	}

	return nil
}

func (act *actionRemovePluginSubConfig) buildWindowsRemoveSubConfigScript(
	subConfigDir string, fullPaths []string) (types.ScriptType, string) {

	basePrefix := ensureWindowsTrailingSeparator(cleanRemoveSubConfigDir(subConfigDir, criteria.OSWindows))
	commands := []string{"@echo off", "setlocal", fmt.Sprintf(`set "BASE=%s"`, escapeBatPath(basePrefix))}
	for _, fullPath := range fullPaths {
		label := fmt.Sprintf("skip%d", len(commands))
		commands = append(commands,
			fmt.Sprintf(`set "TARGET=%s"`, escapeBatPath(fullPath)),
			fmt.Sprintf(`if /I not "%%TARGET:~0,%d%%"=="%%BASE%%" exit /b 1`, len(basePrefix)),
			`if exist "%TARGET%\" exit /b 1`,
			fmt.Sprintf(`if not exist "%%TARGET%%" goto %s`, label),
			`del /F /Q "%TARGET%"`,
			`if errorlevel 1 exit /b 1`,
			fmt.Sprintf(":%s", label),
		)
	}
	commands = append(commands, "endlocal")

	return types.ScriptTypeBat, strings.Join(commands, "\r\n") + "\r\n"
}

func (act *actionRemovePluginSubConfig) buildUnixRemoveSubConfigScript(
	subConfigDir string, osType criteria.OSType, fullPaths []string) (types.ScriptType, string) {

	basePrefix := ensureUnixTrailingSeparator(cleanRemoveSubConfigDir(subConfigDir, osType))
	commands := []string{"#!/bin/bash", "set -e", fmt.Sprintf("base=%s", quoteShellArg(basePrefix))}
	for _, fullPath := range fullPaths {
		commands = append(commands,
			fmt.Sprintf("target=%s", quoteShellArg(fullPath)),
			`case "$target" in`,
			`  "$base"*) ;;`,
			`  *) exit 1 ;;`,
			`esac`,
			`if [ -d "$target" ]; then exit 1; fi`,
			`if [ -e "$target" ]; then rm -f -- "$target"; fi`,
		)
	}

	return types.ScriptTypeBash, strings.Join(commands, "\n") + "\n"
}

func ensureWindowsTrailingSeparator(path string) string {
	if strings.HasSuffix(path, string(winpath.DirSeparator)) || strings.HasSuffix(path, "/") {
		return path
	}

	return path + string(winpath.DirSeparator)
}

func ensureUnixTrailingSeparator(path string) string {
	if strings.HasSuffix(path, "/") {
		return path
	}

	return path + "/"
}

func ensureTrailingSeparator(path string, osType criteria.OSType) string {
	if osType == criteria.OSWindows {
		return ensureWindowsTrailingSeparator(path)
	}

	return ensureUnixTrailingSeparator(path)
}

func cleanRemoveSubConfigDir(path string, osType criteria.OSType) string {
	if osType == criteria.OSWindows {
		return winpath.Clean(path)
	}

	return filepath.Clean(path)
}

func escapeBatPath(path string) string {
	return strings.ReplaceAll(path, "%", "%%")
}

func quoteShellArg(arg string) string {
	return "'" + strings.ReplaceAll(arg, "'", "'\\''") + "'"
}

// buildRemoveSubConfigFullPaths computes the deduplicated full paths to remove
// for the given process config records, each joined under the sub-config dir
// and validated to stay inside it.
func (act *actionRemovePluginSubConfig) buildRemoveSubConfigFullPaths(
	subConfigDir string, osType criteria.OSType, configs []*types.ProcessConfig) ([]string, error) {

	if err := pluginUtils.CheckDirPathSafe(subConfigDir, osType); err != nil {
		return nil, fmt.Errorf("check sub config dir safe failed, dir(%s): %w", subConfigDir, err)
	}

	pathSet := make(map[string]struct{}, len(configs))
	paths := make([]string, 0, len(configs))
	for _, config := range configs {
		fullPath, err := act.buildRemoveSubConfigFullPath(subConfigDir, osType, config.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to build remove sub config path, file-name(%s): %w", config.Name, err)
		}

		if _, ok := pathSet[fullPath]; ok {
			continue
		}
		pathSet[fullPath] = struct{}{}
		paths = append(paths, fullPath)
	}

	return paths, nil
}

func (act *actionRemovePluginSubConfig) buildRemoveSubConfigFullPath(
	subConfigDir string, osType criteria.OSType, fileName string) (string, error) {

	if err := act.checkRemoveSubConfigFileName(osType, fileName); err != nil {
		return "", err
	}
	basePrefix := ensureTrailingSeparator(cleanRemoveSubConfigDir(subConfigDir, osType), osType)
	if osType == criteria.OSWindows {
		fullPath := winpath.Join(subConfigDir, fileName)
		if !strings.HasPrefix(strings.ToLower(fullPath), strings.ToLower(basePrefix)) {
			return "", fmt.Errorf("computed path escapes sub config dir, path(%s)", fullPath)
		}

		return fullPath, nil
	}

	fullPath := filepath.Join(subConfigDir, fileName)
	if !strings.HasPrefix(fullPath, basePrefix) {
		return "", fmt.Errorf("computed path escapes sub config dir, path(%s)", fullPath)
	}

	return fullPath, nil
}

func (act *actionRemovePluginSubConfig) checkRemoveSubConfigFileName(osType criteria.OSType, fileName string) error {
	if fileName == "" {
		return errors.New("file name is empty")
	}
	if fileName == "." || fileName == ".." {
		return fmt.Errorf("invalid file name: %q", fileName)
	}
	if strings.ContainsRune(fileName, 0) {
		return fmt.Errorf("file name contains NUL: %q", fileName)
	}
	if strings.ContainsAny(fileName, "\r\n") {
		return fmt.Errorf("file name contains line break: %q", fileName)
	}
	if osType == criteria.OSWindows {
		if strings.ContainsAny(fileName, `\/`) {
			return fmt.Errorf("file name contains separator: %q", fileName)
		}
		if strings.ContainsAny(fileName, `<>:"|*?`) {
			return fmt.Errorf("file name contains invalid windows character: %q", fileName)
		}

		return nil
	}
	if strings.ContainsRune(fileName, '/') {
		return fmt.Errorf("file name contains '/': %q", fileName)
	}

	return nil
}

// queryScriptResult polls the script execution result until it terminates, treating any failure status or non-zero error code as a retryable error.
func (act *actionRemovePluginSubConfig) queryScriptResult(nCtx contextx.IContext, taskID string, endpoint types.Endpoint) error {
	ticker := time.NewTicker(removePluginSubConfigResultPollInterval)
	defer ticker.Stop()

	for {
		results, err := act.gseHandler.QueryScriptExecutionResult(nCtx, taskID, &types.EndpointWithRestrict{Endpoint: endpoint})
		if err != nil {
			return fmt.Errorf("failed to query script execution result, task-id(%s): %w", taskID, err)
		}
		if len(results) == 0 {
			return fmt.Errorf("script execution result is empty, task-id(%s)", taskID)
		}

		result := results[0]
		switch result.Status {
		case types.ScriptStatusFinished:
			if result.ErrorCode != 0 {
				return fmt.Errorf("script execution failed, task-id(%s), status(%s), error-code(%d), error-message(%s), exit-code(%d)",
					taskID, result.Status, result.ErrorCode, result.ErrorMessage, result.ExitCode)
			}

			return nil
		case types.ScriptStatusFailed, types.ScriptStatusTimeout, types.ScriptStatusAgentRestarted, types.ScriptStatusStopped:
			return fmt.Errorf("script execution failed, task-id(%s), status(%s), error-code(%d), error-message(%s), exit-code(%d)",
				taskID, result.Status, result.ErrorCode, result.ErrorMessage, result.ExitCode)
		default:
			// still running (or status not reported yet): keep polling until cancelled.
			select {
			case <-nCtx.Done():
				return fmt.Errorf("script execution interrupted: %w", nCtx.Err())
			case <-ticker.C:
			}
		}
	}
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionRemovePluginSubConfig) DisplayNameZh() string {
	return "删除插件子配置"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionRemovePluginSubConfig) DisplayNameEn() string {
	return "Remove Plugin Sub Config"
}
