/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pkg

import (
	"fmt"
	"time"

	pkgUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/pkg/utils"
	pkgStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/pkg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNamePackageExportPlugin is the plugin package export action name.
	ActionNamePackageExportPlugin = "package_export_plugin"
)

// NewActionPackageExportPlugin creates an action for exporting a plugin package.
func NewActionPackageExportPlugin(capability *Capability) action.Definition {
	return &actionPackageExportPlugin{
		fileHandler:    capability.FileHandler,
		storagePackage: capability.StoragePackage,
	}
}

type actionPackageExportPlugin struct {
	fileHandler    file.IHandler
	storagePackage pkgStg.IStorage
}

// ActionPackageExportPlugin defines the action's parameters.
type ActionPackageExportPlugin struct {
	pkgUtils.PackageActionStandardParam
}

func (act *actionPackageExportPlugin) Name() string {
	return ActionNamePackageExportPlugin
}

func (act *actionPackageExportPlugin) DisplayNameZh() string {
	return "导出插件资源包"
}

func (act *actionPackageExportPlugin) DisplayNameEn() string {
	return "Export Plugin Package"
}

func (act *actionPackageExportPlugin) Version() string {
	return "v1.0.0"
}

func (act *actionPackageExportPlugin) Description() string {
	return "export origin plugin package"
}

func (act *actionPackageExportPlugin) Timeout() time.Duration {
	return 30 * time.Minute // nolint: mnd
}

func (act *actionPackageExportPlugin) MaxRetryCount() uint {
	return 0
}

func (act *actionPackageExportPlugin) DelayFn(_ int) func() {
	return func() {}
}

func (act *actionPackageExportPlugin) Tags() []action.Tag {
	return []action.Tag{}
}

func (act *actionPackageExportPlugin) Do(ctx *action.InstanceContext) error {
	param := new(ActionPackageExportPlugin)
	if err := conv.MapToStruct(ctx.Data.Content, param); err != nil {
		return fmt.Errorf("failed to decode export parameters: %w", err)
	}
	std := pkgUtils.NewPackageActionStandarder(act.storagePackage)
	if err := std.Initialize(ctx, param.PackageActionStandardParam); err != nil {
		return err
	}
	nCtx := std.Context()
	info := std.DeployInfo()
	workflow, err := act.storagePackage.GetPackageExportWorkflowByTriggerID(nCtx, ctx.Data.TriggerID)
	if err != nil {
		return fmt.Errorf("failed to get package export workflow: %w", err)
	}

	name := info.ExportPluginPkgOptions.PluginPkgName
	version := info.ExportPluginPkgOptions.PluginPkgVersion

	std.InstanceData().Log().
		Zh("准备导出插件资源包, 插件名: %s, 版本: %s", name, version).
		En("prepare export plugin package, plugin-name: %s, version: %s", name, version).
		Info()

	exportID, err := act.fileHandler.ExportPrepareOriginPluginPackage(nCtx, name, version)
	if err != nil {
		return fmt.Errorf("failed to prepare plugin package export: %w", err)
	}

	fields := types.PackageExportFields{WorkflowID: true}
	exportData := &types.PackageExport{ExportID: exportID, WorkflowID: workflow.WorkflowID}
	if err := act.storagePackage.UpdatePackageExport(nCtx, fields, exportData); err != nil {
		return fmt.Errorf("failed to update package export workflow: %w", err)
	}

	std.InstanceData().Log().
		Zh("准备导出插件资源包成功, 导出ID: %s", exportID).
		En("prepare export plugin package succeed, export-id: %s", exportID).
		Info()

	return nil
}
