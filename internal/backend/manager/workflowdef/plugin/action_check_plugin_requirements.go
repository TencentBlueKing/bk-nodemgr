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
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/pageexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameCheckPluginRequirements selects the dependency files that need distribution.
	ActionNameCheckPluginRequirements = "check_plugin_requirements"

	// requirementListPageSize bounds each dependency file list request.
	requirementListPageSize = 500
	// requirementListTimeout bounds the whole paginated dependency file listing.
	requirementListTimeout = 1 * time.Minute
)

// NewActionCheckPluginRequirements creates an action to select dependency files for distribution.
func NewActionCheckPluginRequirements(capability *Capability) action.Definition {
	return &actionCheckPluginRequirements{
		daoPluginDeployment: capability.StoragePlugin,
		fileHandler:         capability.FileHandler,
	}
}

type actionCheckPluginRequirements struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	fileHandler         file.IPluginRequirementHandler
}

// Name returns the action name.
func (act *actionCheckPluginRequirements) Name() string {
	return ActionNameCheckPluginRequirements
}

// Version returns the action version.
func (act *actionCheckPluginRequirements) Version() string {
	return "1.0.0"
}

// Description returns the action description.
func (act *actionCheckPluginRequirements) Description() string {
	return "select plugin dependency files to distribute"
}

// Timeout returns the dependency file check timeout.
func (act *actionCheckPluginRequirements) Timeout() time.Duration {
	return 2 * time.Minute // nolint: mnd
}

// Tags returns the action tags.
func (act *actionCheckPluginRequirements) Tags() []action.Tag {
	return []action.Tag{}
}

// Queue returns the action queue.
func (act *actionCheckPluginRequirements) Queue() action.Queue {
	return action.QueueDefault
}

// MaxRetryCount returns the action retry count.
func (act *actionCheckPluginRequirements) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn returns the retry delay.
func (act *actionCheckPluginRequirements) DelayFn(attempt int) func() {
	return action.DefaultBackoffDelayFn(act, attempt)
}

// Do selects applicable dependency files that are missing or changed on the process and
// records their IDs in the deployment requirement options for apply_plugin_requirements.
// When ApplyAll is set, every applicable file is selected without comparing process records.
func (act *actionCheckPluginRequirements) Do(ctx *action.InstanceContext) (err error) {
	param := new(pluginUtils.PluginActionStandardParam)
	if err = conv.MapToStruct(ctx.Data.Content, param); err != nil {
		return fmt.Errorf("failed to parse plugin requirements check action parameters: %w", err)
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
	requirements, err := act.applicableRequirements(std)
	if err != nil {
		return err
	}

	selected := make([]string, 0, len(requirements))
	for _, record := range requirements {
		if deploy.RequirementOptions.ApplyAll ||
			requirementNeedsApply(record, deploy.Process.Requirements, deploy.BaseRuntime.ConfigDir) {

			selected = append(selected, record.RequirementFileID)
		}
	}
	deploy.RequirementOptions.RequirementIDs = selected

	std.InstanceData().Log().
		Zh("依赖文件检查完成, plugin-name(%s), host-id(%d), 适用(%d), 待下发(%d)",
			deploy.Process.PluginName, deploy.Process.HostID, len(requirements), len(selected)).
		En("dependency files checked, plugin-name(%s), host-id(%d), applicable(%d), to-apply(%d)",
			deploy.Process.PluginName, deploy.Process.HostID, len(requirements), len(selected)).
		Info()

	return nil
}

func (act *actionCheckPluginRequirements) applicableRequirements(std *pluginUtils.PluginActionStandarder) (
	[]*types.PluginRequirementFileRecord, error) {

	process := &std.DeployInfo().Process
	nCtx := std.Context()

	condition := &types.PluginRequirementFileCondition{
		ExactInclude: &types.PluginRequirementFileExactFields{
			PluginName: []string{process.PluginName},
		},
	}
	executor := pageexecutor.NewPageExecutor[*types.PluginRequirementFileRecord](requirementListPageSize, requirementListTimeout)
	fn := func(nCtx contextx.IContext, page types.Page) ([]*types.PluginRequirementFileRecord, error) {
		records, _, err := act.fileHandler.ListPluginRequirements(nCtx, page, condition)

		return records, err
	}
	pageResult, err := executor.Execute(nCtx, types.UnlimitedPage(), fn)
	if err != nil {
		return nil, fmt.Errorf("failed to list dependency files, plugin-name(%s): %w", process.PluginName, err)
	}
	records := pageResult.Items

	requirements := make([]*types.PluginRequirementFileRecord, 0, len(records))
	for _, record := range records {
		if record == nil {
			return nil, fmt.Errorf("dependency file list contains nil record, plugin-name(%s)", process.PluginName)
		}

		if record.PluginName != process.PluginName {
			return nil, fmt.Errorf("dependency file plugin name mismatch, requirement-file-id(%s), file-plugin-name(%s), plugin-name(%s)",
				record.RequirementFileID, record.PluginName, process.PluginName)
		}

		if !record.AppliesTo(process.Platform, process.Info.Version) {
			continue
		}

		requirements = append(requirements, record)
	}

	slices.SortFunc(requirements, func(left, right *types.PluginRequirementFileRecord) int {
		return cmp.Or(cmp.Compare(left.TargetFileName, right.TargetFileName), cmp.Compare(left.RequirementFileID, right.RequirementFileID))
	})
	destinations := make(map[string]struct{}, len(requirements))
	for _, record := range requirements {
		if err := types.ValidatePluginRequirementFileName(record.TargetFileName); err != nil {
			return nil, fmt.Errorf("failed to validate dependency file name, requirement-file-id(%s): %w",
				record.RequirementFileID, err)
		}
		destination := strings.ToLower(record.TargetFileName)
		if _, exists := destinations[destination]; exists {
			return nil, fmt.Errorf("duplicate dependency file destination, plugin-name(%s), target-file-name(%s)",
				process.PluginName, destination)
		}
		destinations[destination] = struct{}{}
	}

	std.InstanceData().Log().
		Zh("筛选适用依赖文件完成, plugin-name(%s), host-id(%d), 文件数(%d)",
			process.PluginName, process.HostID, len(requirements)).
		En("applicable dependency files selected, plugin-name(%s), host-id(%d), file-count(%d)",
			process.PluginName, process.HostID, len(requirements)).
		Info()

	return requirements, nil
}

// requirementNeedsApply reports whether the record must be distributed: no applied entry
// exists for it, or the applied content or destination differs from the current record.
func requirementNeedsApply(record *types.PluginRequirementFileRecord, applied []types.ProcessRequirement, targetDir string) bool {
	for _, item := range applied {
		if item.RequirementID != record.RequirementFileID {
			continue
		}

		return !strings.EqualFold(item.MD5, record.MD5) ||
			item.TargetFileName != record.TargetFileName ||
			item.TargetPath != targetDir
	}

	return true
}

// DisplayNameZh returns the Chinese display name.
func (act *actionCheckPluginRequirements) DisplayNameZh() string {
	return "检查插件依赖文件"
}

// DisplayNameEn returns the English display name.
func (act *actionCheckPluginRequirements) DisplayNameEn() string {
	return "Check Plugin Requirements"
}
