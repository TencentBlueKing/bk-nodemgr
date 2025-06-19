/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package upload

import (
	"context"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler upload handler interface.
type IHandler interface {
	// Create creates upload.
	Create(ctx context.Context, upload *types.Upload) error

	// Get gets upload by upload id.
	Get(ctx context.Context, uploadID string) (*types.Upload, error)

	// DeleteMany deletes upload by upload-ids.
	DeleteMany(ctx context.Context, uploadIDs ...string) error
}

type handler struct {
	logger logger.Logger
	dao    *dao
}

// New new a handler.
func New(client *mongo.Database, logger logger.Logger) IHandler {
	return &handler{
		logger: logger,
		dao:    newDao(client, logger),
	}
}

// Create creates upload.
func (h *handler) Create(ctx context.Context, upload *types.Upload) error {
	return h.dao.Create(ctx, &Upload{
		UploadID:  upload.UploadID,
		Category:  string(upload.Category),
		SavedName: upload.SavedName,
		Operator:  upload.Operator,
		CreatedAt: time.Now(),
	})
}

// Get gets upload by upload id.
func (h *handler) Get(ctx context.Context, uploadID string) (*types.Upload, error) {
	filter := base.AliveFilter()
	filter = base.WithValues(FieldKeyUploadID, uploadID)(filter)

	upload, err := h.dao.Get(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &types.Upload{
		UploadID:  upload.UploadID,
		Category:  types.UploadCategory(upload.Category),
		SavedName: upload.SavedName,
		Operator:  upload.Operator,
		CreatedAt: upload.CreatedAt,
	}, nil
}

// Delete deletes upload by upload id.
func (h *handler) DeleteMany(ctx context.Context, uploadIDs ...string) error {
	return h.dao.deleteMany(ctx, uploadIDs...)
}
