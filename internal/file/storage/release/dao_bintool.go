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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// upsertReleaseBinTool upserts release bintool.
func (s *Storage) upsertReleaseBinTool(nCtx contextx.IContext, bintool types.ReleaseBinTool) error {
	rls := &bintool.Release
	rls.Operator = nCtx.BKUsername()
	rls.Name = types.ReleaseNameBinTool
	rls.Version = types.ReleaseVersionBinTool
	rls.UpdatedAt = time.Now()

	if err := s.daoRelease.UpsertMany(nCtx, types.ReleaseTypeBinTool, rls); err != nil {
		return fmt.Errorf("failed to upsert release bintool: %w", err)
	}

	return nil
}

// deleteReleaseBinTool deletes release bintool.
func (s *Storage) deleteReleaseBinTool(nCtx contextx.IContext, gen types.Generation, fileName string) error {
	if err := s.daoRelease.Delete(nCtx, types.ReleaseTypeBinTool, release.WithGeneration(gen), release.WithFileName(fileName)); err != nil {
		return fmt.Errorf("failed to delete release bintool: %w", err)
	}

	return nil
}

// getReleaseBinTool gets release bintool.
func (s *Storage) getReleaseBinTool(nCtx contextx.IContext, gen types.Generation) (*types.ReleaseBinTool, error) {
	rls, err := s.daoRelease.Get(nCtx, types.ReleaseTypeBinTool, release.WithGeneration(gen))
	if err != nil {
		return nil, fmt.Errorf("failed to get release bintool: %w", err)
	}

	return &types.ReleaseBinTool{
		Release: *rls,
	}, nil
}

// existReleaseBinTool checks if release bintool exists.
func (s *Storage) existReleaseBinTool(nCtx contextx.IContext, gen types.Generation) (bool, error) {
	result, err := s.daoRelease.Exist(nCtx, types.ReleaseTypeBinTool, release.WithGeneration(gen))
	if err != nil {
		return false, fmt.Errorf("failed to check exist release bintool: %w", err)
	}

	return result, nil
}
