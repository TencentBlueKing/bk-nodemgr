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
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/release"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// getReleaseProxy gets release by generation, type, platform and version.
func (s *Storage) getReleaseProxy(
	ctx contextx.IContext, gen types.Generation, plat platform.Platform, version string) (data *types.ReleaseProxy, err error) {

	var rls *types.Release
	rls, err = s.daoRelease.Get(ctx, types.ReleaseTypeProxy,
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get release proxy: %w", err)
	}

	additionInfo := new(types.ReleaseAdditionInfoProxy)
	if err = conv.MapToStruct(rls.AdditionInfo, additionInfo); err != nil {
		return nil, fmt.Errorf("failed to get release proxy: %v", err)
	}

	return &types.ReleaseProxy{
		Release:                  *rls,
		ReleaseAdditionInfoProxy: *additionInfo,
	}, nil
}

// upsertManyReleaseProxy upsert many release.
func (s *Storage) upsertManyReleaseProxy(ctx contextx.IContext, releaseProxys []*types.ReleaseProxy) (err error) {
	releases := make([]*types.Release, 0, len(releaseProxys))
	for _, rls := range releaseProxys {
		if rls == nil {
			continue
		}

		rls.Name = releaseNameProxy
		rls.UpdatedAt = time.Now()
		rls.Operator = ctx.BKUsername()
		rls.AdditionInfo, err = conv.StructToMap(rls.ReleaseAdditionInfoProxy)
		if err != nil {
			return fmt.Errorf("failed to upsert many release proxy: %v", err)
		}

		releases = append(releases, &rls.Release)
	}

	return s.daoRelease.UpsertMany(ctx, types.ReleaseTypeProxy, releases...)
}
