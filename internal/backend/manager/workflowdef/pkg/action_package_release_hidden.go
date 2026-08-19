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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNamePackageReleasePluginHidden is the package release plugin hidden action name.
	ActionNamePackageReleasePluginHidden = "package_release_plugin_hidden"
)

// NewActionPackageReleasePluginHidden creates the hidden action for release plugin package.
func NewActionPackageReleasePluginHidden(capability *Capability) action.Definition {
	return &actionPackageReleasePluginHidden{
		storageRelease: capability.StorageRelease,
		storagePackage: capability.StoragePackage,
	}
}

type actionPackageReleasePluginHidden struct {
	storageRelease release.IPlugin
	storagePackage pkgStg.IStorage
}

// ActionParamImportPackageReleasePluginHidden defines the hidden release plugin package action's param.
type ActionParamImportPackageReleasePluginHidden struct {
	pkgUtils.PackageActionStandardParam
}

func (act *actionPackageReleasePluginHidden) Name() string {
	return ActionNamePackageReleasePluginHidden
}

func (act *actionPackageReleasePluginHidden) DisplayNameZh() string {
	return "隐藏插件资源包"
}

func (act *actionPackageReleasePluginHidden) DisplayNameEn() string {
	return "Hide Release Plugin Package"
}

func (act *actionPackageReleasePluginHidden) Version() string {
	return "v1.0.0" // nolint: goconst
}

func (act *actionPackageReleasePluginHidden) Description() string {
	return "hide release plugin package"
}

func (act *actionPackageReleasePluginHidden) Timeout() time.Duration {
	return 5 * time.Minute // nolint: mnd
}

func (act *actionPackageReleasePluginHidden) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

func (act *actionPackageReleasePluginHidden) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

func (act *actionPackageReleasePluginHidden) Tags() []action.Tag {
	return []action.Tag{}
}

func (act *actionPackageReleasePluginHidden) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamImportPackageReleasePluginHidden)
	if err := conv.MapToStruct(ctx.Data.Content, param); err != nil {
		return err
	}

	std := pkgUtils.NewPackageActionStandarder(act.storagePackage)
	if err := std.Initialize(ctx, param.PackageActionStandardParam); err != nil {
		return err
	}

	nCtx := std.Context()
	info := std.DeployInfo()
	if info.Name == "" {
		return errors.New("plugin name is empty")
	}
	if info.Version == "" {
		return errors.New("plugin version is empty")
	}
	if len(info.Platforms) == 0 {
		return errors.New("plugin platforms is empty")
	}

	for _, platform := range info.Platforms {
		if !platform.Validate() {
			return fmt.Errorf("invalid plugin platform: %s", platform.String())
		}

		std.InstanceData().Log().
			Zh("隐藏插件资源包, 插件包: %s, 版本: %s, 平台: %s", info.Name, info.Version, platform.String()).
			En("hide plugin package, plugin-name: %s, version: %s, platform: %s", info.Name, info.Version, platform.String()).
			Info()

		key := types.ReleasePluginKey{
			Generation: info.Generation,
			Platform:   platform,
			Version:    info.Version,
			Name:       info.Name,
		}
		if err := act.storageRelease.SetHiddenReleasePlugin(nCtx, key, true); err != nil {
			logger.G.Sys().Ctx(nCtx).WithErr(err).With("plugin-name", info.Name).Error("failed to hide release plugin package")

			return fmt.Errorf("failed to hide release plugin package: %w", err)
		}

		std.InstanceData().Log().
			Zh("隐藏安装包成功，插件包: %s，版本: %s，平台: %s", info.Name, info.Version, platform.String()).
			En("hide package succeed, plugin-name: %s, version: %s, platform: %s", info.Name, info.Version, platform.String()).
			Info()
	}

	logger.G.Sys().Ctx(nCtx).With("plugin-name", info.Name, "version", info.Version).Info("hide release plugin package")

	return nil
}
