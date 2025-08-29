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
	"errors"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler upload Handler interface.
type IHandler interface {
	// Create creates upload.
	Create(ctx context.Context, category types.UploadCategory, upload *types.Upload) error

	// Get gets upload by upload id.
	Get(ctx context.Context, category types.UploadCategory, uploadID string) (*types.Upload, error)

	// DeleteMany deletes upload by upload-ids.
	DeleteMany(ctx context.Context, category types.UploadCategory, uploadIDs ...string) error
}

// Handler implements IHandler.
type Handler struct {
	client *mongo.Database
	logger logger.ILogger
	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

func (h *Handler) categoryDao(category types.UploadCategory) *dao {
	if d, ok := h.daoMap.Load(category); ok {
		return d.(*dao) // nolint: forcetypeassert
	}

	newDaoClient := newDao(string(category), h.client, h.logger)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		h.logger.Warnf("failed to ensure upload indexes, err: %v", errors.Join(base.ErrEnsureIndexesFailed(), err))
	}

	d, _ := h.daoMap.LoadOrStore(category, newDaoClient)

	// note: we can be sure that only the categoryDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao) // nolint: forcetypeassert
}

// New new a Handler.
func New(client *mongo.Database, logger logger.ILogger) *Handler {
	return &Handler{
		client: client,
		logger: logger,
		daoMap: sync.Map{},
	}
}

// Create creates upload.
func (h *Handler) Create(ctx context.Context, category types.UploadCategory, upload *types.Upload) error {
	return h.categoryDao(category).Create(ctx, &Upload{
		UploadID:  upload.UploadID,
		Category:  string(upload.Category),
		SavedName: upload.SavedName,
		Operator:  upload.Operator,
		CreatedAt: time.Now(),
	})
}

// Get gets upload by upload id.
func (h *Handler) Get(ctx context.Context, category types.UploadCategory, uploadID string) (*types.Upload, error) {
	filter := base.AliveFilter()
	filter = base.WithValues(FieldKeyUploadID, uploadID)(filter)

	upload, err := h.categoryDao(category).Get(ctx, filter)
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

// DeleteMany deletes upload by upload id.
func (h *Handler) DeleteMany(ctx context.Context, category types.UploadCategory, uploadIDs ...string) error {
	return h.categoryDao(category).deleteMany(ctx, uploadIDs...)
}
