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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// GetReleaseCertGen2 gets release cert gen2.
func (s *Storage) GetReleaseCertGen2(ctx context.Context) (data *types.ReleaseCert, err error) {
	// record metric.
	metric := s.metric().Start("get_cert")
	defer metric.End(err)

	var rls *types.Release
	if rls, err = s.daoRelease.Get(ctx, types.ReleaseTypeCert, types.Generation2, platform.EmptyPlatform(), ""); err != nil {
		return nil, err
	}

	return &types.ReleaseCert{
		Release: *rls,
	}, nil
}

// ExistReleaseCertGen2 checks if release cert gen2 exists.
func (s *Storage) ExistReleaseCertGen2(ctx context.Context) (result bool, err error) {
	// record metric.
	metric := s.metric().Start("exist_cert")
	defer metric.End(err)

	return s.daoRelease.Exist(ctx, types.ReleaseTypeCert, types.Generation2, platform.EmptyPlatform(), "")
}

// UpsertReleaseCertGen2 upserts release cert gen2.
func (s *Storage) UpsertReleaseCertGen2(ctx context.Context, cert types.ReleaseCert) (err error) {
	// record metric.
	metric := s.metric().Start("upsert_cert")
	defer metric.End(err)

	return s.daoRelease.UpsertMany(ctx, types.ReleaseTypeCert, types.Generation2, &types.Release{
		Generation: types.Generation2,
		Type:       types.ReleaseTypeCert,
		Platform:   platform.EmptyPlatform(),
		Version:    "",
		FileName:   cert.FileName,
		MD5:        cert.MD5,
		UpdatedAt:  time.Now(),
	})
}

// DeleteReleaseCertGen2 deletes release cert gen2.
func (s *Storage) DeleteReleaseCertGen2(ctx context.Context, fileName string) (err error) {
	// record metric.
	metric := s.metric().Start("delete_cert")
	defer metric.End(err)

	return s.daoRelease.Delete(ctx, types.ReleaseTypeCert, types.Generation2, platform.EmptyPlatform(), fileName)
}
