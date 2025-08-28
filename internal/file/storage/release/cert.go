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

// ICert defines the cert interface.
type ICert interface {
	// GetReleaseCert gets release cert.
	GetReleaseCert(ctx context.Context) (*types.ReleaseCert, error)

	// ExistReleaseCert checks if release cert exists.
	ExistReleaseCert(ctx context.Context) (bool, error)

	// UpsertReleaseCert upserts release cert.
	UpsertReleaseCert(ctx context.Context, cert types.ReleaseCert) error

	// DeleteReleaseCert deletes release cert.
	DeleteReleaseCert(ctx context.Context, fileName string) error
}

// GetReleaseCert gets release cert.
func (s *Storage) GetReleaseCert(ctx context.Context) (*types.ReleaseCert, error) {
	r, err := s.daoRelease.Get(ctx, types.ReleaseTypeCert, types.GenerationAll, platform.EmptyPlatform(), "")
	if err != nil {
		return nil, err
	}

	return &types.ReleaseCert{
		Release: *r,
	}, nil
}

// ExistReleaseCert checks if release cert exists.
func (s *Storage) ExistReleaseCert(ctx context.Context) (bool, error) {
	return s.daoRelease.Exist(ctx, types.ReleaseTypeCert, types.GenerationAll, platform.EmptyPlatform(), "")
}

// UpsertReleaseCert upserts release cert.
func (s *Storage) UpsertReleaseCert(ctx context.Context, cert types.ReleaseCert) error {
	return s.daoRelease.UpsertMany(ctx, types.ReleaseTypeCert, &types.Release{
		Generation: types.GenerationAll,
		Type:       types.ReleaseTypeCert,
		Platform:   platform.EmptyPlatform(),
		Version:    "",
		FileName:   cert.FileName,
		MD5:        cert.MD5,
		UpdatedAt:  time.Now(),
	})
}

// DeleteReleaseCert deletes release cert.
func (s *Storage) DeleteReleaseCert(ctx context.Context, fileName string) error {
	return s.daoRelease.Delete(ctx, types.ReleaseTypeCert, types.GenerationAll, platform.EmptyPlatform(), fileName)
}
