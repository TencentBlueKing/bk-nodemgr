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

// IExternalPlugin defines the interface of upload storage.
type IExternalPlugin interface {
	// GetExternalPluginV2Upload gets a upload by upload-id.
	GetExternalPluginV2Upload(nCtx contextx.IContext, uploadID string) (*types.Upload, error)

	// CreateExternalPluginV2Upload creates a upload.
	CreateExternalPluginV2Upload(nCtx contextx.IContext, up *types.Upload) (string, error)

	// DeleteExternalPluginV2Upload deletes a upload by upload-id.
	DeleteExternalPluginV2Upload(nCtx contextx.IContext, uploadID string) error
}

// GetExternalPluginV2Upload gets a upload by upload-id.
func (s *Storage) GetExternalPluginV2Upload(nCtx contextx.IContext, uploadID string) (data *types.Upload, err error) {
	// record metric.
	metric := s.metric().Start("get_external_plugin_v2")
	defer metric.End(err)

	return s.daoUpload.Get(nCtx, types.UploadCategoryOriginExternalPluginV2, uploadID)
}

// CreateExternalPluginV2Upload creates a upload.
func (s *Storage) CreateExternalPluginV2Upload(nCtx contextx.IContext, up *types.Upload) (uploadID string, err error) {
	// record metric.
	metric := s.metric().Start("create_external_plugin_v2")
	defer metric.End(err)

	up.UploadID = identifier.GenUploadID()
	if err = s.daoUpload.Create(nCtx, types.UploadCategoryOriginExternalPluginV2, up); err != nil {
		return "", err
	}

	return up.UploadID, nil
}

// DeleteExternalPluginV2Upload deletes a upload by upload-id.
func (s *Storage) DeleteExternalPluginV2Upload(nCtx contextx.IContext, uploadID string) (err error) {
	// record metric.
	metric := s.metric().Start("delete_external_plugin_v2")
	defer metric.End(err)

	return s.daoUpload.DeleteMany(nCtx, types.UploadCategoryOriginExternalPluginV2, uploadID)
}
