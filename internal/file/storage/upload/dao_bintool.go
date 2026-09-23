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

// getBinToolUpload gets a upload by upload-id.
func (s *Storage) getBinToolUpload(nCtx contextx.IContext, uploadID string) (*types.Upload, error) {
	data, err := s.daoUpload.Get(nCtx, types.UploadCategoryOriginBinTool, uploadID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bintool upload, upload-id(%s): %w", uploadID, err)
	}

	return data, nil
}

// createBinToolUpload creates a upload.
func (s *Storage) createBinToolUpload(nCtx contextx.IContext, up *types.Upload) (string, error) {
	if up == nil {
		return "", fmt.Errorf("failed to create bintool upload, upload is nil")
	}

	up.UploadID = identifier.GenUploadID()
	if err := s.daoUpload.Create(nCtx, types.UploadCategoryOriginBinTool, up); err != nil {
		return "", fmt.Errorf("failed to create bintool upload: %w", err)
	}

	return up.UploadID, nil
}

// deleteBinToolUpload deletes a upload by upload-id.
func (s *Storage) deleteBinToolUpload(nCtx contextx.IContext, uploadID string) error {
	if err := s.daoUpload.DeleteMany(nCtx, types.UploadCategoryOriginBinTool, uploadID); err != nil {
		return fmt.Errorf("failed to delete bintool upload, upload-id(%s): %w", uploadID, err)
	}

	return nil
}

// distinctBinToolSavedName distincts saved names of bintool uploads.
func (s *Storage) distinctBinToolSavedName(nCtx contextx.IContext, names []string) ([]string, error) {
	bound, err := s.daoUpload.DistinctSavedName(nCtx, types.UploadCategoryOriginBinTool, names)
	if err != nil {
		return nil, fmt.Errorf("failed to distinct bintool upload saved names: %w", err)
	}

	return bound, nil
}
