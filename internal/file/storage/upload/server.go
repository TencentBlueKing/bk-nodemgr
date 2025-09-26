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

// IServer defines the interface of upload storage.
type IServer interface {
	// GetServerUpload gets a upload by upload-id.
	GetServerUpload(nCtx contextx.IContext, uploadID string) (*types.Upload, error)

	// CreateServerUpload creates a upload.
	CreateServerUpload(nCtx contextx.IContext, up *types.Upload) (string, error)

	// DeleteServerUpload deletes a upload by upload-id.
	DeleteServerUpload(nCtx contextx.IContext, uploadID string) error
}

// GetServerUpload gets a upload by upload-id.
func (s *Storage) GetServerUpload(nCtx contextx.IContext, uploadID string) (data *types.Upload, err error) {
	// record metric.
	metric := s.metric().Start("get_server")
	defer metric.End(err)

	return s.daoUpload.Get(nCtx, types.UploadCategoryOriginServer, uploadID)
}

// CreateServerUpload creates a upload.
func (s *Storage) CreateServerUpload(nCtx contextx.IContext, up *types.Upload) (uploadID string, err error) {
	// record metric.
	metric := s.metric().Start("create_server")
	defer metric.End(err)

	up.UploadID = identifier.GenUploadID()
	if err = s.daoUpload.Create(nCtx, types.UploadCategoryOriginServer, up); err != nil {
		return "", err
	}

	return up.UploadID, nil
}

// DeleteServerUpload deletes a upload by upload-id.
func (s *Storage) DeleteServerUpload(nCtx contextx.IContext, uploadID string) (err error) {
	// record metric.
	metric := s.metric().Start("delete_server")
	defer metric.End(err)

	return s.daoUpload.DeleteMany(nCtx, types.UploadCategoryOriginServer, uploadID)
}
