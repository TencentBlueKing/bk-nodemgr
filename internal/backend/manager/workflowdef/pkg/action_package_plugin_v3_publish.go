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
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNamePackagePublishPluginV3Pkg is the package plugin v3 publish action name.
	ActionNamePackagePublishPluginV3Pkg = "package_publish_plugin_v3_pkg"
)

// NewActionPackagePublishPluginV3Pkg creates a publish action for package import.
func NewActionPackagePublishPluginV3Pkg(capability *Capability) action.Definition {
	return &actionPackagePublishPluginV3Pkg{
		fileHandler:    capability.FileHandler,
		storagePackage: capability.StoragePackage,
		storageRelease: capability.StorageRelease,
	}
}

type actionPackagePublishPluginV3Pkg struct {
	fileHandler    file.IHandler
	storagePackage pkgStg.IStorage
	storageRelease release.IStorage
}

// ActionParamImportPackagePublish defines the publish action's param.
type ActionParamImportPackagePublish struct {
	pkgUtils.PackageActionStandardParam
}

func (act *actionPackagePublishPluginV3Pkg) Name() string {
	return ActionNamePackagePublishPluginV3Pkg
}

func (act *actionPackagePublishPluginV3Pkg) DisplayNameZh() string {
	return "发布插件V3资源包"
}

func (act *actionPackagePublishPluginV3Pkg) DisplayNameEn() string {
	return "Publish Plugin V3 Package"
}

func (act *actionPackagePublishPluginV3Pkg) Version() string {
	return "v1.0.0" // nolint: goconst
}

func (act *actionPackagePublishPluginV3Pkg) Description() string {
	return "publish uploaded plugin package"
}

func (act *actionPackagePublishPluginV3Pkg) Timeout() time.Duration {
	return 10 * time.Minute // nolint: mnd
}

func (act *actionPackagePublishPluginV3Pkg) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

func (act *actionPackagePublishPluginV3Pkg) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

func (act *actionPackagePublishPluginV3Pkg) Tags() []action.Tag {
	return []action.Tag{}
}

func (act *actionPackagePublishPluginV3Pkg) Do(ctx *action.InstanceContext) (err error) {
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
		Zh("发布插件V3安装包, 上传ID: %s", uploadInfo.UploadID).
		En("publish plugin v3 package, upload-id: %s", uploadInfo.UploadID).
		Info()

	if err := act.fileHandler.PublishReleasePluginV3(nCtx, uploadInfo.UploadID); err != nil {
		logger.G.Sys().Ctx(nCtx).WithErr(err).With("upload-id", uploadInfo.UploadID).Error("failed to publish release plugin v3 package")

		return fmt.Errorf("failed to publish release plugin package: %w", err)
	}

	cond := &types.ReleaseCondition{
		ExactInclude: &types.ReleaseExactFields{
			Platform: uploadInfo.Platforms,
			Name:     []string{uploadInfo.Name},
			Version:  []string{uploadInfo.Version},
		},
	}
	pkgs, _, err := act.storageRelease.ListReleasePlugin(nCtx, types.UnlimitedPage(), cond)
	if err != nil {
		return fmt.Errorf("failed to list release plugin: %w", err)
	}

	if len(pkgs) == 0 {
		std.InstanceData().Log().
			Zh("找不到已发布的插件包，请检查插件包名称或版本是否正确, 插件包名称: %s, 版本: %s", uploadInfo.Name, uploadInfo.Version).
			En("no plugin package found, please check the plugin package name or version is correct, plugin package name: %s, version: %s",
				uploadInfo.Name, uploadInfo.Version).
			Error()

		return fmt.Errorf("no plugin package found, please check the plugin package name or version is correct, plugin package name: %s, version: %s",
			uploadInfo.Name, uploadInfo.Version)
	}

	if len(pkgs) != len(uploadInfo.Platforms) {
		std.InstanceData().Log().
			Zh("已发布的插件包内系统架构与上传的包内数量不匹配, 已发布的系统架构数: %d，原始包内系统架构数: %d",
				len(pkgs), len(uploadInfo.Platforms)).
			En("published package os/arch count not match, published os/arch count: %d, original package os/arch count: %d",
				len(pkgs), len(uploadInfo.Platforms)).
			Warn()
	}

	info.Release = conv.SliceToSlice(pkgs, func(pkg *types.ReleasePlugin) types.Release {
		return pkg.Release
	})

	for _, release := range info.Release {
		std.InstanceData().Log().
			Zh("已发布插件V3安装包, 平台: %s", release.Platform.String()).
			En("published plugin v3 package, platform: %s", release.Platform.String()).
			Info()
	}

	logger.G.Sys().Ctx(nCtx).With("upload-id", uploadInfo.UploadID, "plugin-name", uploadInfo.Name).Info("published release plugin v3 package")

	std.InstanceData().Log().
		Zh("发布安装包成功，上传ID: %s，插件包: %s，版本: %s", uploadInfo.UploadID, uploadInfo.Name, uploadInfo.Version).
		En("published package succeed, upload-id: %s, plugin-name: %s, version: %s", uploadInfo.UploadID, uploadInfo.Name, uploadInfo.Version).
		Info()

	return nil
}
