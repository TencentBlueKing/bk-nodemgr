/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package upload provides the upload storage interface.
// nolint: nonamedreturns
package upload

import (
	"context"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ICert defines the interface of upload storage.
type ICert interface {
	// GetCertUpload gets a upload by upload-id.
	GetCertUpload(ctx context.Context, uploadID string) (*types.Upload, error)

	// CreateCertUpload creates a upload.
	CreateCertUpload(ctx context.Context, up *types.Upload) (string, error)

	// DeleteCertUpload deletes a upload by upload-id.
	DeleteCertUpload(ctx context.Context, uploadID string) error
}

// GetCertUpload gets a upload by upload-id.
func (s *Storage) GetCertUpload(ctx context.Context, uploadID string) (data *types.Upload, err error) {
	// record metric.
	metric := s.metric().Start("get_cert")
	defer metric.End(err)

	return s.daoUpload.Get(ctx, types.UploadCategoryOriginCert, uploadID)
}

// CreateCertUpload creates a upload.
func (s *Storage) CreateCertUpload(ctx context.Context, up *types.Upload) (uploadID string, err error) {
	// record metric.
	metric := s.metric().Start("create_cert")
	defer metric.End(err)

	up.UploadID = identifier.GenUploadID()
	if err = s.daoUpload.Create(ctx, types.UploadCategoryOriginCert, up); err != nil {
		return "", err
	}

	return up.UploadID, nil
}

// DeleteCertUpload deletes a upload by upload-id.
func (s *Storage) DeleteCertUpload(ctx context.Context, uploadID string) (err error) {
	// record metric.
	metric := s.metric().Start("delete_cert")
	defer metric.End(err)

	return s.daoUpload.DeleteMany(ctx, types.UploadCategoryOriginCert, uploadID)
}
