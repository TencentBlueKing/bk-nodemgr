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

// getPluginV3Upload gets a upload by upload-id.
func (s *Storage) getPluginV3Upload(nCtx contextx.IContext, uploadID string) (*types.Upload, error) {
	data, err := s.daoUpload.Get(nCtx, types.UploadCategoryOriginPluginV3, uploadID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin v3 upload, upload-id(%s): %w", uploadID, err)
	}

	return data, nil
}

// createPluginV3Upload creates a upload.
func (s *Storage) createPluginV3Upload(nCtx contextx.IContext, up *types.Upload) (string, error) {
	if up == nil {
		return "", fmt.Errorf("failed to create plugin v3 upload, upload is nil")
	}

	up.UploadID = identifier.GenUploadID()
	if err := s.daoUpload.Create(nCtx, types.UploadCategoryOriginPluginV3, up); err != nil {
		return "", fmt.Errorf("failed to create plugin v3 upload: %w", err)
	}

	return up.UploadID, nil
}

// deletePluginV3Upload deletes a upload by upload-id.
func (s *Storage) deletePluginV3Upload(nCtx contextx.IContext, uploadID string) error {
	if err := s.daoUpload.DeleteMany(nCtx, types.UploadCategoryOriginPluginV3, uploadID); err != nil {
		return fmt.Errorf("failed to delete plugin v3 upload, upload-id(%s): %w", uploadID, err)
	}

	return nil
}

// distinctPluginV3SavedName distincts saved names of plugin v3 uploads.
func (s *Storage) distinctPluginV3SavedName(nCtx contextx.IContext, names []string) ([]string, error) {
	bound, err := s.daoUpload.DistinctSavedName(nCtx, types.UploadCategoryOriginPluginV3, names)
	if err != nil {
		return nil, fmt.Errorf("failed to distinct plugin v3 upload saved names: %w", err)
	}

	return bound, nil
}
