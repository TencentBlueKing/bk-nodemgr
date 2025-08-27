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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IProxy defines the agent interface.
type IProxy interface {
	// GetReleaseProxy gets release proxy by generation, type, platform and version.
	GetReleaseProxy(ctx context.Context, gen types.Generation, plat platform.Platform, version string) (*types.Release, error)

	// UpsertManyReleaseProxy upserts many agent release.
	UpsertManyReleaseProxy(ctx context.Context, gen types.Generation, releases []*types.Release) error
}

// GetReleaseProxy gets release by generation, type, platform and version.
func (s *Storage) GetReleaseProxy(ctx context.Context, gen types.Generation, plat platform.Platform, version string) (*types.Release, error) {
	return s.daoRelease.Get(ctx, types.ReleaseTypeProxy, gen, plat, version)
}

// UpsertManyReleaseProxy upserts many release.
func (s *Storage) UpsertManyReleaseProxy(ctx context.Context, gen types.Generation, releases []*types.Release) error {
	for _, rls := range releases {
		if rls != nil {
			rls.UpdatedAt = time.Now()
		}
	}

	return s.daoRelease.UpsertMany(ctx, types.ReleaseTypeProxy, releases...)
}
