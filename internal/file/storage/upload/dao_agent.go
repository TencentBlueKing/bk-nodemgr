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
package upload

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// getAgentUpload gets a upload by upload-id.
func (s *Storage) getAgentUpload(nCtx contextx.IContext, uploadID string) (*types.Upload, error) {
	data, err := s.daoUpload.Get(nCtx, types.UploadCategoryOriginAgent, uploadID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent upload, upload-id(%s): %w", uploadID, err)
	}

	return data, nil
}

// createAgentUpload creates a upload.
func (s *Storage) createAgentUpload(nCtx contextx.IContext, up *types.Upload) (string, error) {
	if up == nil {
		return "", fmt.Errorf("failed to create agent upload, upload is nil")
	}

	up.UploadID = identifier.GenUploadID()
	if err := s.daoUpload.Create(nCtx, types.UploadCategoryOriginAgent, up); err != nil {
		return "", fmt.Errorf("failed to create agent upload: %w", err)
	}

	return up.UploadID, nil
}

// deleteAgentUpload deletes a upload by upload-id.
func (s *Storage) deleteAgentUpload(nCtx contextx.IContext, uploadID string) error {
	if err := s.daoUpload.DeleteMany(nCtx, types.UploadCategoryOriginAgent, uploadID); err != nil {
		return fmt.Errorf("failed to delete agent upload, upload-id(%s): %w", uploadID, err)
	}

	return nil
}
