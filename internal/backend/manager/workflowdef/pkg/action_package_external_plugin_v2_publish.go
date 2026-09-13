/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pkg

import (
	"errors"
	"fmt"
	"time"

	pkgUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/pkg/utils"
	pkgStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/pkg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNamePackagePublishExternalPluginV2Pkg is the external v2 publish action name.
	ActionNamePackagePublishExternalPluginV2Pkg = "package_publish_external_plugin_v2_pkg"
)

// NewActionPackagePublishExternalPluginV2Pkg creates the external v2 publish action.
func NewActionPackagePublishExternalPluginV2Pkg(capability *Capability) action.Definition {
	return &actionPackagePublishExternalPluginV2Pkg{
		fileHandler:    capability.FileHandler,
		storagePackage: capability.StoragePackage,
	}
}

type actionPackagePublishExternalPluginV2Pkg struct {
	fileHandler    file.IHandler
	storagePackage pkgStg.IStorage
}

func (act *actionPackagePublishExternalPluginV2Pkg) Name() string {
	return ActionNamePackagePublishExternalPluginV2Pkg
}

func (act *actionPackagePublishExternalPluginV2Pkg) DisplayNameZh() string {
	return "发布外部插件V2资源包"
}

func (act *actionPackagePublishExternalPluginV2Pkg) DisplayNameEn() string {
	return "Publish External Plugin V2 Package"
}

func (act *actionPackagePublishExternalPluginV2Pkg) Version() string {
	return "v1.0.0" // nolint: goconst
}

func (act *actionPackagePublishExternalPluginV2Pkg) Description() string {
	return "publish uploaded external plugin v2 package"
}

func (act *actionPackagePublishExternalPluginV2Pkg) Timeout() time.Duration {
	return 10 * time.Minute // nolint: mnd
}

func (act *actionPackagePublishExternalPluginV2Pkg) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

func (act *actionPackagePublishExternalPluginV2Pkg) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

func (act *actionPackagePublishExternalPluginV2Pkg) Tags() []action.Tag {
	return []action.Tag{}
}

func (act *actionPackagePublishExternalPluginV2Pkg) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActionParamImportPackagePublish)
	err = conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	std := pkgUtils.NewPackageActionStandarder(act.storagePackage)
	if err := std.Initialize(ctx, param.PackageActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	nCtx := std.Context()
	info := std.DeployInfo()
	uploadInfo := info.Upload
	if err := uploadInfo.Validate(); err != nil {
		return fmt.Errorf("validate upload info failed: %w", err)
	}

	std.InstanceData().Log().
		Zh("发布外部插件V2安装包, 上传ID: %s", uploadInfo.UploadID).
		En("publish external plugin v2 package, upload-id: %s", uploadInfo.UploadID).
		Info()

	if err := act.fileHandler.PublishReleaseExternalPluginV2(nCtx, uploadInfo.UploadID); err != nil {
		logger.G.Sys().Ctx(nCtx).WithErr(err).With("upload-id", uploadInfo.UploadID).
			Error("failed to publish release external plugin v2 package")

		return fmt.Errorf("failed to publish release external plugin v2 package: %w", err)
	}

	cond := &types.ReleaseCondition{
		ExactInclude: &types.ReleaseExactFields{
			Generation: []types.Generation{types.Generation2},
			Name:       []string{uploadInfo.Name},
			Version:    []string{uploadInfo.Version},
		},
	}
	pkgReleases, _, err := act.fileHandler.ListReleasePlugin(nCtx, types.UnlimitedPage(), cond)
	if err != nil {
		return fmt.Errorf("failed to list release plugin: %w", err)
	}

	if len(pkgReleases) == 0 {
		std.InstanceData().Log().
			Zh("找不到已发布的插件包，请检查插件包名称或版本是否正确, 插件包名称: %s, 版本: %s",
				uploadInfo.Name, uploadInfo.Version).
			En("no plugin package found, please check the plugin package name or version is correct, "+
				"plugin package name: %s, version: %s", uploadInfo.Name, uploadInfo.Version).
			Error()

		return fmt.Errorf("no plugin package found, please check the plugin package name or version is correct, "+
			"plugin package name: %s, version: %s", uploadInfo.Name, uploadInfo.Version)
	}

	if err := pkgUtils.ComparePlatforms(uploadInfo.Platforms, pkgReleases); err != nil {
		std.InstanceData().Log().
			Zh("已发布的V2插件包平台集合与上传包的预期平台集合不匹配: %s", err).
			En("published v2 package platform set does not match the uploaded package: %s", err).
			Error()

		return err
	}

	info.Release = conv.SliceToSlice(pkgReleases, func(pkg *types.ReleasePlugin) types.Release {
		return pkg.Release
	})

	for _, pluginRelease := range info.Release {
		std.InstanceData().Log().
			Zh("已发布外部插件V2安装包, 平台: %s", pluginRelease.Platform.String()).
			En("published external plugin v2 package, platform: %s", pluginRelease.Platform.String()).
			Info()
	}

	logger.G.Sys().Ctx(nCtx).With("upload-id", uploadInfo.UploadID, "plugin-name", uploadInfo.Name).
		Info("published release external plugin v2 package")

	std.InstanceData().Log().
		Zh("发布安装包成功，上传ID: %s，插件包: %s，版本: %s", uploadInfo.UploadID, uploadInfo.Name, uploadInfo.Version).
		En("published package succeed, upload-id: %s, plugin-name: %s, version: %s",
			uploadInfo.UploadID, uploadInfo.Name, uploadInfo.Version).
		Info()

	return nil
}
