/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package upload

import (
	"context"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IServer defines the interface of upload storage.
type IServer interface {
	// GetServerUpload gets a upload by upload-id.
	GetServerUpload(ctx context.Context, uploadID string) (*types.Upload, error)

	// CreateServerUpload creates a upload.
	CreateServerUpload(ctx context.Context, up *types.Upload) (string, error)

	// DeleteServerUpload deletes a upload by upload-id.
	DeleteServerUpload(ctx context.Context, uploadID string) error
}

// GetServerUpload gets a upload by upload-id.
func (s *Storage) GetServerUpload(ctx context.Context, uploadID string) (*types.Upload, error) {
	return s.daoUpload.Get(ctx, types.UploadCategoryOriginServer, uploadID)
}

// CreateServerUpload creates a upload.
func (s *Storage) CreateServerUpload(ctx context.Context, up *types.Upload) (string, error) {
	up.UploadID = identifier.GenUploadID()

	return up.UploadID, s.daoUpload.Create(ctx, types.UploadCategoryOriginServer, up)
}

// DeleteServerUpload deletes a upload by upload-id.
func (s *Storage) DeleteServerUpload(ctx context.Context, uploadID string) error {
	return s.daoUpload.DeleteMany(ctx, types.UploadCategoryOriginServer, uploadID)
}
