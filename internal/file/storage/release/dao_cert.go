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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// getReleaseCert gets release cert.
func (s *Storage) getReleaseCert(ctx context.Context) (*types.ReleaseCert, error) {
	var (
		rls *types.Release
		err error
	)
	if rls, err = s.daoRelease.Get(ctx, types.ReleaseTypeCert, types.Generation2, platform.EmptyPlatform(), ""); err != nil {
		return nil, fmt.Errorf("failed to get release cert: %w", err)
	}

	return &types.ReleaseCert{
		Release: *rls,
	}, nil
}

// existReleaseCert checks if release cert exists.
func (s *Storage) existReleaseCert(ctx context.Context) (bool, error) {
	result, err := s.daoRelease.Exist(ctx, types.ReleaseTypeCert, types.Generation2, platform.EmptyPlatform(), "")
	if err != nil {
		return false, fmt.Errorf("failed to check if release cert exists: %w", err)
	}

	return result, nil
}

// upsertReleaseCert upserts release cert.
func (s *Storage) upsertReleaseCert(ctx context.Context, cert types.ReleaseCert) error {
	err := s.daoRelease.UpsertMany(ctx, types.ReleaseTypeCert, types.Generation2, &types.Release{
		Generation: types.Generation2,
		Type:       types.ReleaseTypeCert,
		Platform:   platform.EmptyPlatform(),
		Version:    "",
		FileName:   cert.FileName,
		MD5:        cert.MD5,
		UpdatedAt:  time.Now(),
	})

	if err != nil {
		return fmt.Errorf("failed to upsert release cert: %w", err)
	}

	return nil
}

// deleteReleaseCert deletes release cert.
func (s *Storage) deleteReleaseCert(ctx context.Context, fileName string) error {
	err := s.daoRelease.Delete(ctx, types.ReleaseTypeCert, types.Generation2, platform.EmptyPlatform(), fileName)
	if err != nil {
		return fmt.Errorf("failed to delete release cert: %w", err)
	}

	return nil
}
