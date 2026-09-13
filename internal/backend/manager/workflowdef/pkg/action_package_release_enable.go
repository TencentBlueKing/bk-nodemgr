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
	// ActionNamePackageReleasePluginEnable is the package release plugin enable action name.
	ActionNamePackageReleasePluginEnable = "package_release_plugin_enable"
)

// NewActionPackageReleasePluginEnable creates the enable action for release plugin package.
func NewActionPackageReleasePluginEnable(capability *Capability) action.Definition {
	return &actionPackageReleasePluginEnable{
		fileHandler:    capability.FileHandler,
		storagePackage: capability.StoragePackage,
	}
}

type actionPackageReleasePluginEnable struct {
	fileHandler    file.IHandler
	storagePackage pkgStg.IStorage
}

// ActionParamImportPackageReleasePluginEnable defines the enable release plugin package action's param.
type ActionParamImportPackageReleasePluginEnable struct {
	pkgUtils.PackageActionStandardParam
}

func (act *actionPackageReleasePluginEnable) Name() string {
	return ActionNamePackageReleasePluginEnable
}

func (act *actionPackageReleasePluginEnable) DisplayNameZh() string {
	return "启用插件资源包"
}

func (act *actionPackageReleasePluginEnable) DisplayNameEn() string {
	return "Enable Release Plugin Package"
}

func (act *actionPackageReleasePluginEnable) Version() string {
	return "v1.0.0" // nolint: goconst
}

func (act *actionPackageReleasePluginEnable) Description() string {
	return "enable release plugin package"
}

func (act *actionPackageReleasePluginEnable) Timeout() time.Duration {
	return 5 * time.Minute // nolint: mnd
}

func (act *actionPackageReleasePluginEnable) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

func (act *actionPackageReleasePluginEnable) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

func (act *actionPackageReleasePluginEnable) Tags() []action.Tag {
	return []action.Tag{}
}

func (act *actionPackageReleasePluginEnable) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActionParamImportPackageReleasePluginEnable)
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
	if len(info.Release) == 0 {
		return errors.New("plugin releases is empty")
	}

	for _, release := range info.Release {
		if release.Name == "" {
			return errors.New("plugin name is empty")
		}
		if release.Version == "" {
			return errors.New("plugin version is empty")
		}
		if !release.Platform.Validate() {
			return fmt.Errorf("invalid plugin platform: %s", release.Platform.String())
		}

		std.InstanceData().Log().
			Zh("启用插件资源包, 插件包: %s, 版本: %s, 平台: %s", release.Name, release.Version, release.Platform.String()).
			En("enable plugin package, plugin-name: %s, version: %s, platform: %s", release.Name, release.Version, release.Platform.String()).
			Info()

		key := types.ReleasePluginKey{
			Generation: release.Generation,
			Platform:   release.Platform,
			Version:    release.Version,
			Name:       release.Name,
		}
		if err := act.fileHandler.EnableReleasePlugin(nCtx, key); err != nil {
			logger.G.Sys().Ctx(nCtx).WithErr(err).With("plugin-name", release.Name).Error("failed to enable release plugin package")

			return fmt.Errorf("failed to enable release plugin package: %w", err)
		}
		std.InstanceData().Log().
			Zh("启用安装包成功，插件包: %s，版本: %s，平台: %s", release.Name, release.Version, release.Platform.String()).
			En("enable package succeed, plugin-name: %s, version: %s, platform: %s", release.Name, release.Version, release.Platform.String()).
			Info()
	}

	logger.G.Sys().Ctx(nCtx).With("count", len(info.Release)).Info("enabled release plugin package")

	return nil
}
