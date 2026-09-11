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

// Package release provides the release storage interface.
// nolint: nonamedreturns
package release

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/release"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// getReleaseProxy gets release by generation, type, platform and version.
func (s *Storage) getReleaseProxy(nCtx contextx.IContext, gen types.Generation, plat platfmt.Platform, version string) (*types.ReleaseProxy, error) {
	rls, err := s.daoRelease.Get(nCtx, types.ReleaseTypeProxy,
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get release proxy: %w", err)
	}

	additionInfo := new(types.ReleaseAdditionInfoProxy)
	if err = conv.MapToStruct(rls.AdditionInfo, additionInfo); err != nil {
		return nil, fmt.Errorf("failed to get release proxy, failed to convert addition info to struct: %w", err)
	}

	return &types.ReleaseProxy{
		Release:                  *rls,
		ReleaseAdditionInfoProxy: *additionInfo,
	}, nil
}

// listReleaseProxy lists proxy releases by page and conditions.
func (s *Storage) listReleaseProxy(nCtx contextx.IContext, page types.Page,
	conditions ...*types.ReleaseCondition) ([]*types.ReleaseProxy, int64, error) {

	rls, total, err := s.listRelease(nCtx, types.ReleaseTypeProxy, page, conditions...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list release proxy: %w", err)
	}
	results := make([]*types.ReleaseProxy, len(rls))
	for i, item := range rls {
		additionInfo := new(types.ReleaseAdditionInfoProxy)
		if err = conv.MapToStruct(item.AdditionInfo, additionInfo); err != nil {
			return nil, 0, fmt.Errorf("failed to convert addition info to struct: %w", err)
		}
		results[i] = &types.ReleaseProxy{
			Release:                  *item,
			ReleaseAdditionInfoProxy: *additionInfo,
		}
	}

	return results, total, nil
}

// upsertManyReleaseProxy upsert many release.
func (s *Storage) upsertManyReleaseProxy(nCtx contextx.IContext, releaseProxys []*types.ReleaseProxy) error {
	releases := make([]*types.Release, 0, len(releaseProxys))
	for _, rls := range releaseProxys {
		if rls == nil {
			continue
		}

		additionInfo, err := conv.StructToMap(rls.ReleaseAdditionInfoProxy)
		if err != nil {
			return fmt.Errorf("failed to upsert many release proxy, failed to convert addition info to map: %w", err)
		}

		rls.Name = types.ReleaseNameProxy
		rls.UpdatedAt = time.Now()
		rls.Operator = nCtx.BKUsername()
		rls.AdditionInfo = additionInfo

		releases = append(releases, &rls.Release)
	}

	if err := s.daoRelease.UpsertMany(nCtx, types.ReleaseTypeProxy, releases...); err != nil {
		return fmt.Errorf("failed to upsert many release proxy: %w", err)
	}

	return nil
}

// existReleaseProxy checks if release proxy exists.
func (s *Storage) existReleaseProxy(nCtx contextx.IContext, gen types.Generation, version string, plats ...platfmt.Platform) (bool, error) {
	exist, err := s.daoRelease.Exist(nCtx, types.ReleaseTypeProxy,
		release.WithGeneration(gen),
		release.WithType(types.ReleaseTypeProxy),
		release.WithVersion(version),
		release.WithPlatform(plats...),
	)
	if err != nil {
		return false, fmt.Errorf("failed to check exist release proxy: %w", err)
	}

	return exist, nil
}
