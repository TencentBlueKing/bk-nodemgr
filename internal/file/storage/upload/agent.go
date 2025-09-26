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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IAgent defines the interface of upload storage.
type IAgent interface {
	// GetAgentUpload gets a upload by upload-id.
	GetAgentUpload(nCtx contextx.IContext, uploadID string) (*types.Upload, error)

	// CreateAgentUpload creates a upload.
	CreateAgentUpload(nCtx contextx.IContext, up *types.Upload) (string, error)

	// DeleteAgentUpload deletes a upload by upload-id.
	DeleteAgentUpload(nCtx contextx.IContext, uploadID string) error
}

// GetAgentUpload gets a upload by upload-id.
func (s *Storage) GetAgentUpload(nCtx contextx.IContext, uploadID string) (data *types.Upload, err error) {
	// record metric.
	metric := s.metric().Start("get_agent")
	defer metric.End(err)

	return s.daoUpload.Get(nCtx, types.UploadCategoryOriginAgent, uploadID)
}

// CreateAgentUpload creates a upload.
func (s *Storage) CreateAgentUpload(nCtx contextx.IContext, up *types.Upload) (uploadID string, err error) {
	// record metric.
	metric := s.metric().Start("create_agent")
	defer metric.End(err)

	up.UploadID = identifier.GenUploadID()
	if err = s.daoUpload.Create(nCtx, types.UploadCategoryOriginAgent, up); err != nil {
		return "", err
	}

	return up.UploadID, nil
}

// DeleteAgentUpload deletes a upload by upload-id.
func (s *Storage) DeleteAgentUpload(nCtx contextx.IContext, uploadID string) (err error) {
	// record metric.
	metric := s.metric().Start("delete_agent")
	defer metric.End(err)

	return s.daoUpload.DeleteMany(nCtx, types.UploadCategoryOriginAgent, uploadID)
}
