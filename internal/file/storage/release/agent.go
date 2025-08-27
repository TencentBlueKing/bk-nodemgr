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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/release"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IAgent defines the agent interface.
type IAgent interface {
	// GetReleaseAgent gets release agent by generation, type, platform and version.
	GetReleaseAgent(ctx context.Context, gen types.Generation, plat platform.Platform, version string) (*types.Release, error)

	// UpsertManyReleaseAgent upserts many release agent.
	UpsertManyReleaseAgent(ctx context.Context, gen types.Generation, releases []*types.Release) error

	// ExistReleaseAgent checks if release agent exists.
	ExistReleaseAgent(ctx context.Context, gen types.Generation, version string, plats ...platform.Platform) (bool, error)
}

// UpsertManyReleaseAgent upserts many release.
func (s *Storage) UpsertManyReleaseAgent(ctx context.Context, gen types.Generation, releases []*types.Release) error {
	for _, rls := range releases {
		if rls != nil {
			rls.UpdatedAt = time.Now()
		}
	}

	return s.daoRelease.UpsertMany(ctx, types.ReleaseTypeAgent, releases...)
}

// ExistReleaseAgent checks if release agent exists.
func (s *Storage) ExistReleaseAgent(ctx context.Context, gen types.Generation, version string, plats ...platform.Platform) (bool, error) {
	count, err := s.daoRelease.Count(ctx, types.ReleaseTypeAgent, release.WithGeneration(gen), release.WithType(types.ReleaseTypeAgent), release.WithVersion(version), release.WithPlatform(plats...))
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetReleaseAgent gets release by generation, type, platform and version.
func (s *Storage) GetReleaseAgent(ctx context.Context, gen types.Generation, plat platform.Platform, version string) (*types.Release, error) {
	return s.daoRelease.Get(ctx, types.ReleaseTypeAgent, gen, plat, version)
}
