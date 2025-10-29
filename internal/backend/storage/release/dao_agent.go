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
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// getReleaseAgent gets release by generation, release type, platform and version.
func (s *Storage) getReleaseAgent(
	nCtx contextx.IContext, gen types.Generation, plat platfmt.Platform, version string) (*types.ReleaseAgent, error) {

	rls, err := s.getRelease(nCtx, types.ReleaseTypeAgent, gen, types.ReleaseNameAgent, plat, version)
	if err != nil {
		return nil, err
	}

	additionInfo := new(types.ReleaseAdditionInfoAgent)
	err = conv.MapToStruct(rls.AdditionInfo, additionInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to convert addition info to struct: %w", err)
	}

	return &types.ReleaseAgent{
		Release:                  *rls,
		ReleaseAdditionInfoAgent: *additionInfo,
	}, nil
}

// listRelease lists release by page and conditions.
func (s *Storage) listReleaseAgent(
	nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) (
	[]*types.ReleaseAgent, int64, error) {

	rls, total, err := s.listRelease(nCtx, types.ReleaseTypeAgent, page, conditions...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list release agent: %w", err)
	}

	releaseAgents := make([]*types.ReleaseAgent, len(rls))
	for idx, r := range rls {
		additionInfo := new(types.ReleaseAdditionInfoAgent)
		err = conv.MapToStruct(r.AdditionInfo, additionInfo)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert addition info to struct: %w", err)
		}

		releaseAgents[idx] = &types.ReleaseAgent{
			Release:                  *r,
			ReleaseAdditionInfoAgent: *additionInfo,
		}
	}

	return releaseAgents, total, nil
}

// countReleaseAgent counts release by conditions.
func (s *Storage) countReleaseAgent(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error) {
	opts, err := convertReleaseConditionsToOptions(conditions...)
	if err != nil {
		return 0, fmt.Errorf("failed to convert release conditions to options: %w", err)
	}

	num, err := s.daoRelease.Count(nCtx, types.ReleaseTypeAgent, opts...)
	if err != nil {
		return 0, fmt.Errorf("failed to count release agent: %w", err)
	}

	return num, nil
}
