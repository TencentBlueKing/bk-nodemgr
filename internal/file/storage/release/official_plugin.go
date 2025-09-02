/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package release

import (
	"context"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/release"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/opluginpkg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IOfficialPlugin defines the interface of official plugin.
type IOfficialPlugin interface {
	// ExistReleaseOfficialPluginGen2 checks if release official plugin exists.
	ExistReleaseOfficialPluginGen2(ctx contextx.IContext, pluginName string, version string, plat ...platform.Platform) (bool, error)

	// UpsertManyReleaseOfficialPluginGen2 upserts many release official plugin gen2.
	UpsertManyReleaseOfficialPluginGen2(ctx context.Context, releaseOfficialPlugins []*types.ReleaseOfficialPlugin) error
}

// ExistReleaseOfficialPluginGen2 checks if release plugin exists.
func (s *Storage) ExistReleaseOfficialPluginGen2(ctx contextx.IContext, pluginName string, version string, plats ...platform.Platform) (bool, error) {
	fileNames := make([]string, 0, len(plats))
	for _, plat := range plats {
		pluginFileName, err := opluginpkg.FormatPkgName(pluginName, types.ReleaseTypeOfficialPlugin, types.Generation2, plat, version)
		if err != nil {
			return false, err
		}

		fileNames = append(fileNames, pluginFileName)
	}

	count, err := s.daoRelease.Count(ctx, types.ReleaseTypeOfficialPlugin, types.Generation2,
		release.WithFileName(fileNames...))
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// UpsertManyReleaseOfficialPluginGen2 upsert many release.
func (s *Storage) UpsertManyReleaseOfficialPluginGen2(ctx context.Context, releaseOfficialPlugins []*types.ReleaseOfficialPlugin) error {
	var err error
	releases := make([]*types.Release, 0, len(releaseOfficialPlugins))

	for _, rls := range releaseOfficialPlugins {
		if rls == nil {
			continue
		}

		rls.UpdatedAt = time.Now()
		rls.AdditionInfo, err = conv.StructToMap(rls.ReleaseAdditionInfoOfficialPlugin)
		if err != nil {
			return fmt.Errorf("failed to upsert many release agent: %v", err)
		}

		releases = append(releases, &rls.Release)
	}

	return s.daoRelease.UpsertMany(ctx, types.ReleaseTypeOfficialPlugin, types.Generation2, releases...)
}
