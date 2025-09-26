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
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler upload Handler interface.
type IHandler interface {
	// Create creates upload.
	Create(nCtx contextx.IContext, category types.UploadCategory, upload *types.Upload) error

	// Get gets upload by upload id.
	Get(nCtx contextx.IContext, category types.UploadCategory, uploadID string) (*types.Upload, error)

	// DeleteMany deletes upload by upload-ids.
	DeleteMany(nCtx contextx.IContext, category types.UploadCategory, uploadIDs ...string) error
}

// Handler implements IHandler.
type Handler struct {
	client *mongo.Database
	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

func (h *Handler) categoryDao(category types.UploadCategory) *dao {
	if d, ok := h.daoMap.Load(category); ok {
		return d.(*dao) // nolint: forcetypeassert
	}

	newDaoClient := newDao(string(category), h.client)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).Warn("failed to ensure upload indexes")
	}

	d, _ := h.daoMap.LoadOrStore(category, newDaoClient)

	// note: we can be sure that only the categoryDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao) // nolint: forcetypeassert
}

// New new a Handler.
func New(client *mongo.Database) *Handler {
	return &Handler{
		client: client,
		daoMap: sync.Map{},
	}
}

// Create creates upload.
func (h *Handler) Create(nCtx contextx.IContext, category types.UploadCategory, upload *types.Upload) error {
	return h.categoryDao(category).Create(nCtx, &Upload{
		UploadID:  upload.UploadID,
		Category:  string(upload.Category),
		SavedName: upload.SavedName,
		Operator:  upload.Operator,
		CreatedAt: time.Now(),
	})
}

// Get gets upload by upload id.
func (h *Handler) Get(nCtx contextx.IContext, category types.UploadCategory, uploadID string) (*types.Upload, error) {
	filter := base.AliveFilter()
	filter = base.WithValues(FieldKeyUploadID, uploadID)(filter)

	upload, err := h.categoryDao(category).Get(nCtx, filter)
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
func (h *Handler) DeleteMany(nCtx contextx.IContext, category types.UploadCategory, uploadIDs ...string) error {
	return h.categoryDao(category).deleteMany(nCtx, uploadIDs...)
}
