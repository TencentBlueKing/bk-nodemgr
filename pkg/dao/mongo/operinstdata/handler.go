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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler this define the Handler interface.
type IHandler interface {
	IOperationInstData
	IActionInstData
}

// Handler implements the IHandler interface.
type Handler struct {
	logger logger.ILogger
	dao    *dao
}

// New create a new Handler.
func New(client *mongo.Database, logger logger.ILogger) *Handler {
	h := &Handler{
		dao:    newDao(client, logger),
		logger: logger,
	}

	if err := h.dao.EnsureIndexes(); err != nil {
		h.logger.Warnf("failed to ensure accesspoint indexes, err: %v",
			errors.Join(base.ErrEnsureIndexesFailed(), err))
	}

	return h
}
