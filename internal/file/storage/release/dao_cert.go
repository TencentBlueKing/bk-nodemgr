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

// getReleaseCert gets release cert.
func (s *Storage) getReleaseCert(nCtx contextx.IContext) (*types.ReleaseCert, error) {
	rls, err := s.daoRelease.Get(nCtx, types.ReleaseTypeCert,
		// cert was designed in generation 2.
		release.WithGeneration(types.Generation2),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get release cert: %w", err)
	}

	return &types.ReleaseCert{
		Release: *rls,
	}, nil
}

// existReleaseCert checks if release cert exists.
func (s *Storage) existReleaseCert(nCtx contextx.IContext) (bool, error) {
	result, err := s.daoRelease.Exist(nCtx, types.ReleaseTypeCert,
		// cert was designed in generation 2.
		release.WithGeneration(types.Generation2),
	)
	if err != nil {
		return false, fmt.Errorf("failed to check exist release cert: %w", err)
	}

	return result, nil
}

// upsertReleaseCert upserts release cert.
func (s *Storage) upsertReleaseCert(nCtx contextx.IContext, cert types.ReleaseCert) error {
	rls := &cert.Release
	rls.Operator = nCtx.BKUsername()
	rls.Name = types.ReleaseNameCert
	rls.Version = types.ReleaseVersionCert
	rls.UpdatedAt = time.Now()

	if err := s.daoRelease.UpsertMany(nCtx, types.ReleaseTypeCert, rls); err != nil {
		return fmt.Errorf("failed to upsert release cert: %w", err)
	}

	return nil
}

// deleteReleaseCert deletes release cert.
func (s *Storage) deleteReleaseCert(nCtx contextx.IContext, fileName string) error {
	if err := s.daoRelease.Delete(nCtx, types.ReleaseTypeCert, release.WithFileName(fileName)); err != nil {
		return fmt.Errorf("failed to delete release cert: %w", err)
	}

	return nil
}
