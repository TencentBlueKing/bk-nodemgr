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

// Package upload provides the upload storage interface.
package upload

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// getPluginV2Upload gets a upload by upload-id.
func (s *Storage) getPluginV2Upload(nCtx contextx.IContext, uploadID string) (*types.Upload, error) {
	data, err := s.daoUpload.Get(nCtx, types.UploadCategoryOriginPluginV2, uploadID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin v2 upload, upload-id(%s): %w", uploadID, err)
	}

	return data, nil
}

// createPluginV2Upload creates a upload.
func (s *Storage) createPluginV2Upload(nCtx contextx.IContext, up *types.Upload) (string, error) {
	if up == nil {
		return "", fmt.Errorf("failed to create plugin v2 upload, upload is nil")
	}

	up.UploadID = identifier.GenUploadID()
	if err := s.daoUpload.Create(nCtx, types.UploadCategoryOriginPluginV2, up); err != nil {
		return "", fmt.Errorf("failed to create plugin v2 upload: %w", err)
	}

	return up.UploadID, nil
}

// deletePluginV2Upload deletes a upload by upload-id.
func (s *Storage) deletePluginV2Upload(nCtx contextx.IContext, uploadID string) error {
	if err := s.daoUpload.DeleteMany(nCtx, types.UploadCategoryOriginPluginV2, uploadID); err != nil {
		return fmt.Errorf("failed to delete plugin v2 upload, upload-id(%s): %w", uploadID, err)
	}

	return nil
}
