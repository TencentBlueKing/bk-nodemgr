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

package release

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// getReleaseProxy gets release by generation, release type, platform and version.
func (s *Storage) getReleaseProxy(
	nCtx contextx.IContext, gen types.Generation, plat platfmt.Platform, version string) (*types.ReleaseProxy, error) {

	rls, err := s.getRelease(nCtx, types.ReleaseTypeProxy, gen, plat, version, types.ReleaseNameProxy)
	if err != nil {
		return nil, err
	}

	additionInfo := new(types.ReleaseAdditionInfoProxy)
	err = conv.MapToStruct(rls.AdditionInfo, additionInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to convert addition info to struct: %w", err)
	}

	return &types.ReleaseProxy{
		Release:                  *rls,
		ReleaseAdditionInfoProxy: *additionInfo,
	}, nil
}

// listReleaseProxy lists release by page and conditions.
func (s *Storage) listReleaseProxy(
	nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) (
	[]*types.ReleaseProxy, int64, error) {

	rls, total, err := s.listRelease(nCtx, types.ReleaseTypeProxy, page, conditions...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list release proxy: %w", err)
	}

	releaseProxys := make([]*types.ReleaseProxy, len(rls))
	for idx, r := range rls {
		additionInfo := new(types.ReleaseAdditionInfoProxy)
		err = conv.MapToStruct(r.AdditionInfo, additionInfo)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert addition info to struct: %w", err)
		}

		releaseProxys[idx] = &types.ReleaseProxy{
			Release:                  *r,
			ReleaseAdditionInfoProxy: *additionInfo,
		}
	}

	return releaseProxys, total, nil
}
