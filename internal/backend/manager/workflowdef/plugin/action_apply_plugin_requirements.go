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
 * either express or implied. See the License for the specific
 * language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package plugin

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameApplyPluginRequirements applies the dependency files selected by check_plugin_requirements.
	ActionNameApplyPluginRequirements = "apply_plugin_requirements"
)

// NewActionApplyPluginRequirements creates an action to distribute dependency files.
func NewActionApplyPluginRequirements(capability *Capability) action.Definition {
	return &actionApplyPluginRequirements{
		daoPluginDeployment: capability.StoragePlugin,
		daoHost:             capability.StorageTopo,
		fileHandler:         capability.FileHandler,
	}
}

type actionApplyPluginRequirements struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoHost             topoStg.IStorageHost
	fileHandler         file.IPluginRequirementHandler
}

// Name returns the action name.
func (act *actionApplyPluginRequirements) Name() string {
	return ActionNameApplyPluginRequirements
}

// Version returns the action version.
func (act *actionApplyPluginRequirements) Version() string {
	return "1.0.0"
}

// Description returns the action description.
func (act *actionApplyPluginRequirements) Description() string {
	return "distribute plugin dependency files"
}

// Timeout returns the file transfer timeout.
func (act *actionApplyPluginRequirements) Timeout() time.Duration {
	return 10 * time.Minute // nolint: mnd
}

// Tags returns the action tags.
func (act *actionApplyPluginRequirements) Tags() []action.Tag {
	return []action.Tag{}
}

// Queue returns the action queue.
func (act *actionApplyPluginRequirements) Queue() action.Queue {
	return action.QueueDefault
}

// MaxRetryCount returns the action retry count.
func (act *actionApplyPluginRequirements) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn returns the retry delay.
func (act *actionApplyPluginRequirements) DelayFn(attempt int) func() {
	return action.DefaultBackoffDelayFn(act, attempt)
}

// Do distributes the dependency files selected by check_plugin_requirements into the plugin
// configuration directory and stages the merged records for update_process to record.
func (act *actionApplyPluginRequirements) Do(ctx *action.InstanceContext) (err error) {
	param := new(pluginUtils.PluginActionStandardParam)
	if err = conv.MapToStruct(ctx.Data.Content, param); err != nil {
		return fmt.Errorf("failed to parse plugin requirements action parameters: %w", err)
	}

	// initialize standard data.
	std := pluginUtils.NewPluginActionStandarder(act.daoPluginDeployment)
	if err = std.Initialize(ctx, *param); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	deploy := std.DeployInfo()
	process := &deploy.Process
	if len(deploy.RequirementOptions.RequirementIDs) == 0 {
		std.InstanceData().Log().
			Zh("无待下发依赖文件, 跳过下发, plugin-name(%s), host-id(%d)", process.PluginName, process.HostID).
			En("no dependency files to apply, skip transfer, plugin-name(%s), host-id(%d)", process.PluginName, process.HostID).
			Info()

		return nil
	}

	requirements, err := act.selectedRequirements(std)
	if err != nil {
		return err
	}
	applied, err := act.applyRequirements(std, requirements)
	if err != nil {
		return err
	}

	// Persisted by the deferred std.Save; update_process records the requirements
	// to the process record after the process is restarted and verified alive.
	// Merging keeps the entries that check_plugin_requirements skipped as unchanged.
	process.Requirements = mergeProcessRequirements(process.Requirements, applied)

	std.InstanceData().Log().
		Zh("插件依赖文件下发成功, plugin-name(%s), host-id(%d), 文件数(%d)",
			process.PluginName, process.HostID, len(applied)).
		En("plugin dependency files applied, plugin-name(%s), host-id(%d), file-count(%d)",
			process.PluginName, process.HostID, len(applied)).
		Info()

	return nil
}

func (act *actionApplyPluginRequirements) selectedRequirements(std *pluginUtils.PluginActionStandarder) (
	[]*types.PluginRequirementFileRecord, error) {

	ids := std.DeployInfo().RequirementOptions.RequirementIDs
	records := make([]*types.PluginRequirementFileRecord, 0, len(ids))
	for _, id := range ids {
		record, err := act.fileHandler.GetPluginRequirement(std.Context(), id)
		if err != nil {
			return nil, fmt.Errorf("failed to get dependency file, requirement-file-id(%s): %w", id, err)
		}
		records = append(records, record)
	}

	return records, nil
}

// mergeProcessRequirements replaces applied entries by requirement ID and keeps the rest,
// so files skipped as unchanged retain their records.
func mergeProcessRequirements(existing, applied []types.ProcessRequirement) []types.ProcessRequirement {
	merged := make([]types.ProcessRequirement, 0, len(existing)+len(applied))
	for _, item := range existing {
		if slices.ContainsFunc(applied, func(record types.ProcessRequirement) bool {
			return record.RequirementID == item.RequirementID
		}) {

			continue
		}
		merged = append(merged, item)
	}

	return append(merged, applied...)
}

func (act *actionApplyPluginRequirements) applyRequirements(
	std *pluginUtils.PluginActionStandarder, requirements []*types.PluginRequirementFileRecord) (
	[]types.ProcessRequirement, error) {

	deploy := std.DeployInfo()
	nCtx := std.Context()

	applied := make([]types.ProcessRequirement, 0, len(requirements))
	targetDir := deploy.BaseRuntime.ConfigDir
	if targetDir == "" {
		return nil, fmt.Errorf("plugin configuration directory is empty, host-id(%d), plugin-name(%s)",
			deploy.Process.HostID, deploy.Process.PluginName)
	}

	targetHost, err := act.daoHost.GetHostByID(nCtx, deploy.Process.HostID)
	if err != nil {
		return nil, fmt.Errorf("failed to get host by id, host-id(%d): %w", deploy.Process.HostID, err)
	}

	for _, requirement := range requirements {
		std.InstanceData().Log().
			Zh("准备传输依赖文件, requirement-file-id(%s), target-file-name(%s), host-id(%d)",
				requirement.RequirementFileID, requirement.TargetFileName, targetHost.HostID).
			En("preparing to transfer dependency file. requirement-file-id(%s), target-file-name(%s), host-id(%d)",
				requirement.RequirementFileID, requirement.TargetFileName, targetHost.HostID).
			Info()

		transfer, err := act.fileHandler.LaunchTransferPluginRequirement(nCtx, requirement.RequirementFileID, targetDir, targetHost)
		if err != nil {
			return nil, fmt.Errorf("failed to launch transfer dependency file, requirement-file-id(%s), host-id(%d): %w",
				requirement.RequirementFileID, targetHost.HostID, err)
		}

		std.InstanceData().Log().
			Zh("开始传输依赖文件. task-id(%s), requirement-file-id(%s)", transfer.GetTaskID(), requirement.RequirementFileID).
			En("start transfer dependency file. task-id(%s), requirement-file-id(%s)", transfer.GetTaskID(), requirement.RequirementFileID).
			Info()

		logger.G.Sys().Ctx(nCtx).
			With("task-id", transfer.GetTaskID(), "requirement-file-id", requirement.RequirementFileID, "host-id", targetHost.HostID).
			Info("launched transfer dependency file")

		result, err := transfer.WaitUntilDone(nCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to wait until transfer dependency file done, task-id(%s), requirement-file-id(%s): %w",
				transfer.GetTaskID(), requirement.RequirementFileID, err)
		}
		if result == nil {
			return nil, fmt.Errorf("transfer dependency file returned no result, task-id(%s), requirement-file-id(%s)",
				transfer.GetTaskID(), requirement.RequirementFileID)
		}
		if !result.Terminated {
			return nil, fmt.Errorf("transfer dependency file not terminated, task-id(%s), requirement-file-id(%s)",
				transfer.GetTaskID(), requirement.RequirementFileID)
		}
		if result.ErrorCode != 0 {
			return nil, fmt.Errorf("transfer dependency file failed, task-id(%s), requirement-file-id(%s), err-code(%d), err-msg(%s)",
				transfer.GetTaskID(), requirement.RequirementFileID, result.ErrorCode, result.ErrorMessage)
		}

		fileInfo := transfer.GetFileInfo()
		if fileInfo.Name != requirement.TargetFileName || !strings.EqualFold(fileInfo.MD5, requirement.MD5) {
			return nil, fmt.Errorf("transfer dependency file metadata mismatch, task-id(%s), requirement-file-id(%s)",
				transfer.GetTaskID(), requirement.RequirementFileID)
		}

		logger.G.Sys().Ctx(nCtx).
			With("task-id", transfer.GetTaskID(), "requirement-file-id", requirement.RequirementFileID, "host-id", targetHost.HostID).
			Info("transfer dependency file done")

		std.InstanceData().Log().
			Zh("传输依赖文件完成, requirement-file-id(%s)", requirement.RequirementFileID).
			En("transfer dependency file done, requirement-file-id(%s)", requirement.RequirementFileID).
			Info()

		applied = append(applied, types.ProcessRequirement{
			RequirementID:  requirement.RequirementFileID,
			TargetPath:     targetDir,
			TargetFileName: requirement.TargetFileName,
			MD5:            fileInfo.MD5,
		})
	}

	return applied, nil
}

// DisplayNameZh returns the Chinese display name.
func (act *actionApplyPluginRequirements) DisplayNameZh() string {
	return "下发插件依赖文件"
}

// DisplayNameEn returns the English display name.
func (act *actionApplyPluginRequirements) DisplayNameEn() string {
	return "Apply Plugin Requirements"
}
