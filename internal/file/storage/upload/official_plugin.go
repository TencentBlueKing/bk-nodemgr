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

// IOfficialPlugin defines the interface of upload storage.
type IOfficialPlugin interface {
	// GetOfficialPluginUpload gets a upload by upload-id.
	GetOfficialPluginUpload(nCtx contextx.IContext, uploadID string) (*types.Upload, error)

	// CreateOfficialPluginUpload creates a upload.
	CreateOfficialPluginUpload(nCtx contextx.IContext, up *types.Upload) (string, error)

	// DeleteOfficialPluginUpload deletes a upload by upload-id.
	DeleteOfficialPluginUpload(nCtx contextx.IContext, uploadID string) error
}

// GetOfficialPluginUpload gets a upload by upload-id.
func (s *Storage) GetOfficialPluginUpload(nCtx contextx.IContext, uploadID string) (data *types.Upload, err error) {
	// record metric.
	metric := s.metric().Start("get_official_plugin")
	defer metric.End(err)

	return s.daoUpload.Get(nCtx, types.UploadCategoryOriginOfficialPlugin, uploadID)
}

// CreateOfficialPluginUpload creates a upload.
func (s *Storage) CreateOfficialPluginUpload(nCtx contextx.IContext, up *types.Upload) (uploadID string, err error) {
	// record metric.
	metric := s.metric().Start("create_official_plugin")
	defer metric.End(err)

	up.UploadID = identifier.GenUploadID()
	if err = s.daoUpload.Create(nCtx, types.UploadCategoryOriginOfficialPlugin, up); err != nil {
		return "", err
	}

	return up.UploadID, nil
}

// DeleteOfficialPluginUpload deletes a upload by upload-id.
func (s *Storage) DeleteOfficialPluginUpload(nCtx contextx.IContext, uploadID string) (err error) {
	// record metric.
	metric := s.metric().Start("delete_official_plugin")
	defer metric.End(err)

	return s.daoUpload.DeleteMany(nCtx, types.UploadCategoryOriginOfficialPlugin, uploadID)
}
