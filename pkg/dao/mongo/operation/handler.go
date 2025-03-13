/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operation ...
package operation

import (
	"context"
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
	"go.mongodb.org/mongo-driver/mongo"
)

// Handler operation handler interface.
type Handler interface {
	// Upsert insert or update an operation.
	Upsert(ctx context.Context, operation *operengine.Operation) error

	// FindOne if not specified, the first undeleted record is returned.
	FindOne(ctx context.Context, opts ...OptFn) (*operengine.Operation, error)
}

type handler struct {
	dao *dao
}

// New ...
func New(client *mongo.Database, logger logger.Logger) Handler {
	return &handler{
		dao: newDao(client, logger),
	}
}

// FindOne ...
func (h *handler) FindOne(ctx context.Context, opts ...OptFn) (*operengine.Operation, error) {
	if ctx == nil {
		return nil, errors.New("ctx is nil")
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	operations, err := h.dao.find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(operations) == 0 {
		return nil, nil
	}

	operation := operations[0]

	data := &operengine.Operation{
		TriggerID:   operation.TriggerID,
		OperationID: operation.OperationID,
		DefSnapshot: operengine.OperDefSnapshot{
			OperDefName: operation.DefSnapshot.OperDefName,
			ActionNames: operation.DefSnapshot.ActionNames,
		},
		OperInstIDs: operation.OperInstIDs,
	}

	return data, nil

}

// Upsert ...
func (h *handler) Upsert(ctx context.Context, operation *operengine.Operation) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if operation == nil {
		return nil
	}

	data := &Operation{
		OperationID: operation.OperationID,
		TriggerID:   operation.TriggerID,
		OperInstIDs: operation.OperInstIDs,
		DefSnapshot: DefSnapshot{
			OperDefName: operation.DefSnapshot.OperDefName,
			ActionNames: operation.DefSnapshot.ActionNames,
		},
	}

	if err := h.dao.upsert(ctx, data); err != nil {
		return err
	}

	return nil
}
