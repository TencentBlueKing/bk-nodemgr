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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IAgent defines the interface of upload storage.
type IAgent interface {
	// GetAgentUpload gets a upload by upload-id.
	GetAgentUpload(ctx context.Context, uploadID string) (*types.Upload, error)

	// CreateAgentUpload creates a upload.
	CreateAgentUpload(ctx context.Context, up *types.Upload) (string, error)

	// DeleteAgentUpload deletes a upload by upload-id.
	DeleteAgentUpload(ctx context.Context, uploadID string) error
}

// GetAgentUpload gets a upload by upload-id.
func (s *Storage) GetAgentUpload(ctx context.Context, uploadID string) (data *types.Upload, err error) {
	// record metric.
	metric := s.metric().Start("get_agent")
	defer metric.End(err)

	return s.daoUpload.Get(ctx, types.UploadCategoryOriginAgent, uploadID)
}

// CreateAgentUpload creates a upload.
func (s *Storage) CreateAgentUpload(ctx context.Context, up *types.Upload) (uploadID string, err error) {
	// record metric.
	metric := s.metric().Start("create_agent")
	defer metric.End(err)

	up.UploadID = identifier.GenUploadID()
	if err = s.daoUpload.Create(ctx, types.UploadCategoryOriginAgent, up); err != nil {
		return "", err
	}

	return up.UploadID, nil
}

// DeleteAgentUpload deletes a upload by upload-id.
func (s *Storage) DeleteAgentUpload(ctx context.Context, uploadID string) (err error) {
	// record metric.
	metric := s.metric().Start("delete_agent")
	defer metric.End(err)

	return s.daoUpload.DeleteMany(ctx, types.UploadCategoryOriginAgent, uploadID)
}
