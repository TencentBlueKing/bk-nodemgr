/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package counter

import (
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

// Handler counter handler interface.
type Handler interface {
	// Generate generate a global new sequence in namespace key.
	Generate(nCtx contextx.IContext, key string) (int64, error)
}

type handler struct {
	client *mongo.Database

	dao  *dao
	once sync.Once
}

// New create a new counter handler.
func New(client *mongo.Database) Handler {
	return &handler{
		client: client,
		dao:    newDao(client),
	}
}

// Generate generate a global new sequence in namespace key.
func (h *handler) Generate(nCtx contextx.IContext, key string) (int64, error) {
	if key == "" {
		return -1, base.ErrEmptyParamData()
	}

	return h.getDao().generate(nCtx, key)
}

func (h *handler) getDao() *dao {
	h.once.Do(func() {
		if err := h.dao.ensureIndexes(); err != nil {
			logger.G.Sys().WithErr(err).Warn("failed to ensure counter indexes")
		}
	})

	return h.dao
}
