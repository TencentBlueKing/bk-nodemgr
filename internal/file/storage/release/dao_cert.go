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

// listReleaseCert lists cert releases by page and conditions.
func (s *Storage) listReleaseCert(nCtx contextx.IContext, page types.Page,
	conditions ...*types.ReleaseCondition) ([]*types.ReleaseCert, int64, error) {

	rls, total, err := s.listRelease(nCtx, types.ReleaseTypeCert, page, conditions...)
	if err != nil {
		return nil, 0, err
	}
	results := make([]*types.ReleaseCert, len(rls))
	for i, item := range rls {
		results[i] = &types.ReleaseCert{Release: *item}
	}

	return results, total, nil
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
