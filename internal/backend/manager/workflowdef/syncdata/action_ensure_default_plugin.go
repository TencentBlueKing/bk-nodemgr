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

package syncdata

import (
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/syncdata/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/release"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/pageexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameEnsureDefaultPlugin defines the default plugin reconciliation action name.
	ActionNameEnsureDefaultPlugin = "ensure_default_plugin"

	defaultPluginReleasePageSize = 500
)

// NewActionEnsureDefaultPlugin creates the default plugin reconciliation action.
func NewActionEnsureDefaultPlugin(capability *Capability) action.Definition {
	return &actionEnsureDefaultPlugin{
		domainPlugin: capability.StoragePlugin,
		fileHandler:  capability.FileHandler,
	}
}

type actionEnsureDefaultPlugin struct {
	domainPlugin pluginStg.IDomainPlugin
	fileHandler  file.IReleasePluginHandler
}

// ActionParamEnsureDefaultPlugin defines the default plugin reconciliation action parameters.
type ActionParamEnsureDefaultPlugin struct {
	utils.SyncDataActionStandardParam
}

// Name returns the name.
func (act *actionEnsureDefaultPlugin) Name() string {
	return ActionNameEnsureDefaultPlugin
}

// DisplayNameZh returns the Chinese display name.
func (act *actionEnsureDefaultPlugin) DisplayNameZh() string {
	return "补齐默认插件"
}

// DisplayNameEn returns the English display name.
func (act *actionEnsureDefaultPlugin) DisplayNameEn() string {
	return "Ensure Default Plugins"
}

// Version returns the version.
func (act *actionEnsureDefaultPlugin) Version() string {
	return "v1.0.0" // nolint:goconst
}

// Description returns the description.
func (act *actionEnsureDefaultPlugin) Description() string {
	return "ensure every enabled plugin package has a default plugin"
}

// Timeout returns the timeout.
func (act *actionEnsureDefaultPlugin) Timeout() time.Duration {
	return 10 * time.Minute // nolint:mnd
}

// MaxRetryCount returns the max retry count.
func (act *actionEnsureDefaultPlugin) MaxRetryCount() uint {
	// The scheduled workflow retries failed packages on the next one-minute run.
	return 0
}

// DelayFn returns the retry delay function.
func (act *actionEnsureDefaultPlugin) DelayFn(attempt int) func() {
	return action.DefaultBackoffDelayFn(act, attempt)
}

// Tags returns the tags.
func (act *actionEnsureDefaultPlugin) Tags() []action.Tag {
	return []action.Tag{}
}

// Do reconciles default plugins for all enabled plugin releases.
func (act *actionEnsureDefaultPlugin) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamEnsureDefaultPlugin)
	if err := conv.MapToStruct(ctx.Data.Content, param); err != nil {
		return err
	}

	std := utils.NewSyncDataActionStandarder()
	if err := std.Initialize(ctx, param.SyncDataActionStandardParam); err != nil {
		return err
	}

	releases, err := act.listEnabledPluginReleases(std.Context())
	if err != nil {
		return fmt.Errorf("failed to list enabled plugin releases: %w", err)
	}

	var failures []error
	processedPackages := make(map[string]struct{})
	for _, release := range releases {
		if _, ok := processedPackages[release.Name]; ok {
			continue
		}
		processedPackages[release.Name] = struct{}{}

		if err := act.domainPlugin.EnsureDefaultPlugin(std.Context(), release); err != nil {
			failures = append(failures, fmt.Errorf(
				"plugin package %q, generation %d, platform %s, version %q: %w",
				release.Name, release.Generation, release.Platform.String(), release.Version, err,
			))
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf("failed to ensure default plugins for %d plugin packages: %w", len(failures), errors.Join(failures...))
	}

	return nil
}

func (act *actionEnsureDefaultPlugin) listEnabledPluginReleases(nCtx contextx.IContext) ([]*types.ReleasePlugin, error) {
	cond := &types.ReleaseCondition{
		ExactInclude: &types.ReleaseExactFields{Enabled: []bool{true}},
	}
	page := types.UnlimitedPage()
	page.Sort = types.WithSortFields(
		release.FieldKeyName,
		release.FieldKeyGeneration,
		release.FieldKeyCPUArch,
		release.FieldKeyOSType,
		release.FieldKeyVersion,
	)
	executor := pageexecutor.NewPageExecutor[*types.ReleasePlugin](defaultPluginReleasePageSize, act.Timeout())
	result, err := executor.Execute(nCtx, page, func(ctx contextx.IContext, page types.Page) ([]*types.ReleasePlugin, error) {
		releases, _, err := act.fileHandler.ListReleasePlugin(ctx, page, cond)
		return releases, err
	})
	if err != nil {
		return nil, err
	}

	return result.Items, nil
}
