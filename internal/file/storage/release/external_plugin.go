/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package release provides the release storage interface.
// nolint: nonamedreturns
package release

import (
	"context"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/release"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/epluginpkg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ExistReleaseExternalPluginGen2 checks if release plugin exists.
func (s *Storage) ExistReleaseExternalPluginGen2(
	ctx contextx.IContext, pluginName string, version string, plats ...platform.Platform) (result bool, err error) {

	// record metric.
	metric := s.metric().Start("exist_external_plugin")
	defer metric.End(err)

	fileNames := make([]string, 0, len(plats))
	for _, plat := range plats {
		pluginFileName, err := epluginpkg.FormatPkgName(pluginName, types.ReleaseTypeExternalPlugin, types.Generation2, plat, version)
		if err != nil {
			return false, err
		}

		fileNames = append(fileNames, pluginFileName)
	}

	var num int64
	if num, err = s.daoRelease.Count(ctx, types.ReleaseTypeExternalPlugin, types.Generation2, release.WithFileName(fileNames...)); err != nil {
		return false, err
	}

	return num > 0, nil
}

// UpsertManyReleaseExternalPluginGen2 upsert many release.
func (s *Storage) UpsertManyReleaseExternalPluginGen2(ctx context.Context, releaseExternalPlugins []*types.ReleaseExternalPlugin) (err error) {
	// record metric.
	metric := s.metric().Start("upsert_many_external_plugin")
	defer metric.End(err)

	releases := make([]*types.Release, 0, len(releaseExternalPlugins))
	for _, rls := range releaseExternalPlugins {
		if rls == nil {
			continue
		}

		rls.UpdatedAt = time.Now()
		rls.AdditionInfo, err = conv.StructToMap(rls.ReleaseAdditionInfoExternalPlugin)
		if err != nil {
			return fmt.Errorf("failed to upsert many release agent: %v", err)
		}

		releases = append(releases, &rls.Release)
	}

	return s.daoRelease.UpsertMany(ctx, types.ReleaseTypeExternalPlugin, types.Generation2, releases...)
}
