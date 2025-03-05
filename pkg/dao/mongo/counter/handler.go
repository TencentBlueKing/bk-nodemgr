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
	"context"
	"errors"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

// Handler counter handler interface.
type Handler interface {
	// Generate generate a global new sequence in namespace key.
	Generate(ctx context.Context, key string) (int64, error)
}

type handler struct {
	client *mongo.Database
	logger logger.Logger
	dao    *dao
	once   sync.Once
}

const (
	defaultEnsureIndexesTimeout = 5 * time.Second
)

// New create a new counter handler.
func New(client *mongo.Database, logger logger.Logger) Handler {
	return &handler{
		client: client,
		logger: logger,
		dao:    newDao(client, logger),
	}
}

// Generate generate a global new sequence in namespace key.
func (h *handler) Generate(ctx context.Context, key string) (int64, error) {
	return h.getDao(ctx).generate(ctx, key)
}

func (h *handler) getDao(ctx context.Context) *dao {
	h.once.Do(func() {
		if err := h.dao.ensureIndexes(ctx); err != nil {
			h.logger.Warnf("failed to ensure counter indexes, err: %v", errors.Join(base.ErrEnsureIndexesFailed(), err))
		}
	})

	return h.dao
}
