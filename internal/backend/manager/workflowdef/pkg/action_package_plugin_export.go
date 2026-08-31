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
	"errors"
	"fmt"
	"strings"
	"time"

	pkgUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/pkg/utils"
	pkgStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/pkg"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

// ActionNamePackageExportPlugin is the plugin package export action name.
const ActionNamePackageExportPlugin = "package_export_plugin"

// NewActionPackageExportPlugin creates an action for exporting a plugin package.
func NewActionPackageExportPlugin(capability *Capability) action.Definition {
	return &actionPackageExportPlugin{
		fileHandler:    capability.FileHandler,
		storageRelease: capability.StorageRelease,
		storagePackage: capability.StoragePackage,
	}
}

type actionPackageExportPlugin struct {
	fileHandler    file.IHandler
	storageRelease release.IStorage
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

func (act *actionPackageExportPlugin) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActionPackageExportPlugin)
	if err := conv.MapToStruct(ctx.Data.Content, param); err != nil {
		return fmt.Errorf("failed to decode export parameters: %w", err)
	}
	std := pkgUtils.NewPackageActionStandarder(act.storagePackage)
	if err := std.Initialize(ctx, param.PackageActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if saveErr := std.Save(); saveErr != nil {
			err = errors.Join(err, saveErr)
		}
	}()

	nCtx := std.Context()
	info := std.DeployInfo()

	cond := &types.ReleaseCondition{
		ExactInclude: &types.ReleaseExactFields{
			Generation: []types.Generation{types.Generation2},
			Name:       []string{info.ExportPluginPkgOptions.PluginPkgName},
			Version:    []string{info.ExportPluginPkgOptions.PluginPkgVersion},
		},
	}
	releases, _, err := act.storageRelease.ListReleasePlugin(nCtx, types.UnlimitedPage(), cond)
	if err != nil {
		return fmt.Errorf("failed to list release plugin: %w", err)
	}

	originCategoryList := conv.SliceUnique(conv.SliceToSlice(releases, func(release *types.ReleasePlugin) types.UploadCategory {
		return release.OriginUploadCategory
	}))
	if len(originCategoryList) != 1 {
		return fmt.Errorf("target plugin release origin upload category mismatch, origin upload category: %v", originCategoryList)
	}
	originUploadCategory := originCategoryList[0]

	uploadIDs := conv.SliceUnique(conv.SliceToSlice(releases, func(release *types.ReleasePlugin) string {
		return release.OriginUploadID
	}))

	std.InstanceData().Log().
		Zh("准备导出插件资源包, 上传ID: %s", uploadIDs).
		En("prepare export plugin package, upload-id: %s", uploadIDs).
		Info()

	fileInfo, address, err := act.fileHandler.ExportPrepareOriginPluginPackage(std.Context(),
		types.ExportPrepareParam{
			PluginPkgName:       info.ExportPluginPkgOptions.PluginPkgName,
			PluginPkgVersion:    info.ExportPluginPkgOptions.PluginPkgVersion,
			UploadOriginPkgType: originUploadCategory,
			UploadIDs:           uploadIDs,
			// use operation-inst-id as file name suffix to avoid conflict with other operations.
			// in action can not easily get the workflow-id.
			// replace all ":" with "-" to avoid filename illegal characters.
			FileNameSuffix: strings.ReplaceAll(strings.TrimSpace(std.InstanceData().OperationID), ":", "-"),
		})
	if err != nil {
		return fmt.Errorf("failed to prepare plugin package export: %w", err)
	}

	info.Export = types.PackageDeploymentExportInfo{
		Filename: fileInfo.Name,
		Address:  address,
	}

	std.InstanceData().Log().
		Zh("准备导出插件资源包成功, 文件名: %s", fileInfo.Name).
		En("prepare export plugin package succeed, file name: %s", fileInfo.Name).
		Info()

	return nil
}
