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
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// upsertManyReleaseAgent upsert many release.
func (s *Storage) upsertManyReleaseAgent(nCtx contextx.IContext, releaseAgents []*types.ReleaseAgent) error {
	releases := make([]*types.Release, 0, len(releaseAgents))
	for _, rls := range releaseAgents {
		if rls == nil {
			continue
		}

		additionInfo, err := conv.StructToMap(rls.ReleaseAdditionInfoAgent)
		if err != nil {
			return fmt.Errorf("failed to upsert many release agent, failed to convert addition info to map: %w", err)
		}

		rls.Name = types.ReleaseNameAgent
		rls.Operator = nCtx.BKUsername()
		rls.UpdatedAt = time.Now()
		rls.AdditionInfo = additionInfo

		releases = append(releases, &rls.Release)
	}

	if err := s.daoRelease.UpsertMany(nCtx, types.ReleaseTypeAgent, releases...); err != nil {
		return fmt.Errorf("failed to upsert many release agent: %w", err)
	}

	return nil
}

// existReleaseAgent checks if release agent exists.
func (s *Storage) existReleaseAgent(nCtx contextx.IContext, gen types.Generation, version string, plats ...platfmt.Platform) (bool, error) {
	exist, err := s.daoRelease.Exist(nCtx, types.ReleaseTypeAgent,
		release.WithGeneration(gen),
		release.WithType(types.ReleaseTypeAgent),
		release.WithVersion(version),
		release.WithPlatform(plats...))
	if err != nil {
		return false, fmt.Errorf("failed to check exist release agent: %w", err)
	}

	return exist, nil
}

// getReleaseAgent gets release by generation, type, platform and version.
func (s *Storage) getReleaseAgent(nCtx contextx.IContext, gen types.Generation, plat platfmt.Platform, version string) (*types.ReleaseAgent, error) {
	rls, err := s.daoRelease.Get(nCtx, types.ReleaseTypeAgent,
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get release agent: %w", err)
	}

	additionInfo := new(types.ReleaseAdditionInfoAgent)
	if err = conv.MapToStruct(rls.AdditionInfo, additionInfo); err != nil {
		return nil, fmt.Errorf("failed to get release agent, failed to convert addition info to struct: %w", err)
	}

	return &types.ReleaseAgent{
		Release:                  *rls,
		ReleaseAdditionInfoAgent: *additionInfo,
	}, nil
}
