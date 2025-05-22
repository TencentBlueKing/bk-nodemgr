/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operinstdata ...
package operinstdata

import (
	"errors"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler this define the handler interface.
type IHandler interface {
	IOperationInstData
	IActionInstData
}

// handler implements the IHandler interface.
type handler struct {
	client *mongo.Database
	logger logger.Logger
	// daoMap stores dao's containing operation information.
	// Do not edit the daoMap except with the operationDao func.
	daoMap sync.Map
}

// tenantDao get the dao by tenantID.
func (h *handler) tenantDao(tenantID string) *dao {
	if d, ok := h.daoMap.Load(tenantID); ok {
		return d.(*dao) // nolint:forcetypeassert
	}

	newDaoClient := newDao(tenantID, h.client, h.logger)
	if err := newDaoClient.baseOrm.EnsureIndexes(); err != nil {
		h.logger.Warnf("failed to ensure trigger indexes, err: %v", errors.Join(base.ErrEnsureIndexesFailed(), err))
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao) // nolint:forcetypeassert
}

// New create a new handler.
func New(client *mongo.Database, logger logger.Logger) IHandler {
	return &handler{
		client: client,
		logger: logger,
		daoMap: sync.Map{},
	}
}
