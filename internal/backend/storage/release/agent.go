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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/release"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IAgent define the agent interface.
type IAgent interface {
	// GetReleaseAgent gets release by generation, release type, platform and version.
	GetReleaseAgent(ctx contextx.IContext, gen types.Generation, plat platform.Platform, version string) (*types.ReleaseAgent, error)
}

// GetReleaseAgent gets release by generation, release type, platform and version.
func (s *Storage) GetReleaseAgent(ctx contextx.IContext, gen types.Generation, plat platform.Platform, version string) (*types.ReleaseAgent, error) {
	rls, err := s.daoRelease.Get(ctx, types.ReleaseTypeAgent,
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get release agent: %v", err)
	}

	additionInfo := new(types.ReleaseAdditionInfoAgent)
	err = conv.MapToStruct(rls.AdditionInfo, additionInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to get release agent: %v", err)
	}

	releaseAgent := &types.ReleaseAgent{
		Release:                  *rls,
		ReleaseAdditionInfoAgent: *additionInfo,
	}

	return releaseAgent, nil
}
