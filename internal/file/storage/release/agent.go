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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/release"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// UpsertManyReleaseAgentGen2 upsert many release.
func (s *Storage) UpsertManyReleaseAgentGen2(ctx context.Context, releaseAgents []*types.ReleaseAgent) (err error) {
	// record metric.
	metric := s.metric().Start("upsert_many_agent")
	defer metric.End(err)

	releases := make([]*types.Release, 0, len(releaseAgents))
	for _, rls := range releaseAgents {
		if rls == nil {
			continue
		}

		rls.UpdatedAt = time.Now()
		rls.AdditionInfo, err = conv.StructToMap(rls.ReleaseAdditionInfoAgent)
		if err != nil {
			return fmt.Errorf("failed to upsert many release agent: %v", err)
		}

		releases = append(releases, &rls.Release)
	}

	return s.daoRelease.UpsertMany(ctx, types.ReleaseTypeAgent, types.Generation2, releases...)
}

// ExistReleaseAgent checks if release agent gen2 exists.
func (s *Storage) ExistReleaseAgent(ctx context.Context, gen types.Generation, version string, plats ...platform.Platform) (result bool, err error) {
	// record metric.
	metric := s.metric().Start("exist_agent")
	defer metric.End(err)

	count, err := s.daoRelease.Count(ctx, types.ReleaseTypeAgent, gen,
		release.WithGeneration(gen),
		release.WithType(types.ReleaseTypeAgent),
		release.WithVersion(version),
		release.WithPlatform(plats...))
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetReleaseAgent gets release by generation, type, platform and version.
func (s *Storage) GetReleaseAgent(
	ctx context.Context, gen types.Generation, plat platform.Platform, version string) (data *types.ReleaseAgent, err error) {

	// record metric.
	metric := s.metric().Start("get_agent")
	defer metric.End(err)

	var rls *types.Release
	if rls, err = s.daoRelease.Get(ctx, types.ReleaseTypeAgent, gen, plat, version); err != nil {
		return nil, fmt.Errorf("failed to get release agent: %w", err)
	}

	additionInfo := new(types.ReleaseAdditionInfoAgent)
	if err = conv.MapToStruct(rls.AdditionInfo, additionInfo); err != nil {
		return nil, fmt.Errorf("failed to get release agent: %v", err)
	}

	return &types.ReleaseAgent{
		Release:                  *rls,
		ReleaseAdditionInfoAgent: *additionInfo,
	}, nil
}
