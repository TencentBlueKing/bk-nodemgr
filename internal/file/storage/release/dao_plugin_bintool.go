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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// upsertReleasePluginBinTool upserts release plugin bintool.
func (s *Storage) upsertReleasePluginBinTool(nCtx contextx.IContext, pluginBinTool types.ReleasePluginBinTool) error {
	rls := &pluginBinTool.Release
	rls.Operator = nCtx.BKUsername()
	rls.Version = types.ReleaseVersionPluginBinTool
	rls.UpdatedAt = time.Now()

	if err := s.daoRelease.UpsertMany(nCtx, types.ReleaseTypePluginBinTool, rls); err != nil {
		return fmt.Errorf("failed to upsert release plugin bintool, name(%s): %w", rls.Name, err)
	}

	return nil
}

// getReleasePluginBinTool gets release plugin bintool.
func (s *Storage) getReleasePluginBinTool(nCtx contextx.IContext, gen types.Generation, name string) (*types.ReleasePluginBinTool, error) {
	rls, err := s.daoRelease.Get(nCtx, types.ReleaseTypePluginBinTool,
		release.WithGeneration(gen),
		release.WithName(name),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get release plugin bintool, name(%s): %w", name, err)
	}

	return &types.ReleasePluginBinTool{
		Release: *rls,
	}, nil
}

// listReleasePluginBinTool lists plugin bintool releases by page and conditions.
func (s *Storage) listReleasePluginBinTool(nCtx contextx.IContext, page types.Page,
	conditions ...*types.ReleaseCondition) ([]*types.ReleasePluginBinTool, int64, error) {

	rls, total, err := s.listRelease(nCtx, types.ReleaseTypePluginBinTool, page, conditions...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list release plugin bintool: %w", err)
	}
	results := make([]*types.ReleasePluginBinTool, len(rls))
	for i, item := range rls {
		results[i] = &types.ReleasePluginBinTool{Release: *item}
	}

	return results, total, nil
}

// existReleasePluginBinTool checks if release plugin bintool exists.
func (s *Storage) existReleasePluginBinTool(nCtx contextx.IContext, gen types.Generation) (bool, error) {
	result, err := s.daoRelease.Exist(nCtx, types.ReleaseTypePluginBinTool, release.WithGeneration(gen))
	if err != nil {
		return false, fmt.Errorf("failed to check exist release plugin bintool: %w", err)
	}

	return result, nil
}
