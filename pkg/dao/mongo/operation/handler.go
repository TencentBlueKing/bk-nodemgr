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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"

	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler operation handler interface.
type IHandler interface {
	// Upsert insert or update an operation.
	Upsert(ctx context.Context, operation *operation.Operation) error

	// List list operation by page and opts.
	List(ctx context.Context, page types.Page, opts ...OptFn) ([]*operation.Operation, int64, error)
}

type handler struct {
	logger logger.ILogger
	dao    *dao
}

// New new a handler.
func New(client *mongo.Database, logger logger.ILogger) IHandler {
	return &handler{
		logger: logger,
		dao:    newDao(client, logger),
	}
}

// List lists operation by page and opts.
func (h *handler) List(ctx context.Context, page types.Page, opts ...OptFn) ([]*operation.Operation, int64, error) {
	if err := page.Validate(); err != nil {
		return nil, 0, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.dao.Count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := base.ParsePage(page)

	datas, err := h.dao.List(ctx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}

	operations := make([]*operation.Operation, len(datas))
	for idx, data := range datas {
		operations[idx] = convertOperationFromDB(data)
	}

	return operations, num, nil
}

// Upsert insert or update an operation.
func (h *handler) Upsert(ctx context.Context, operation *operation.Operation) error {
	if operation == nil {
		return base.ErrEmptyParamData()
	}

	data := convertOperationToDB(operation)

	if err := h.dao.upsert(ctx, data); err != nil {
		return err
	}

	return nil
}

func convertOperationFromDB(dbOp *Operation) *operation.Operation {
	if dbOp == nil {
		return nil
	}

	defSnapshot := convertDefFromDB(dbOp.DefSnapshot)
	param := convertParamFromDB(dbOp.Parameters)

	param.ExtraContent = dbOp.Extras

	return &operation.Operation{
		TriggerID:   dbOp.TriggerID,
		OperationID: dbOp.OperationID,
		Definition:  defSnapshot,
		InstanceIDs: dbOp.OperInstIDs,
		Param:       param,
	}
}

func convertOperationToDB(bizOp *operation.Operation) *Operation {
	if bizOp == nil {
		return nil
	}
	var defSnapshot DefSnapshot
	if bizOp.Definition != nil {
		defSnapshot = convertDefToDB(bizOp.Definition)
	}

	opera := &Operation{
		OperationID: bizOp.OperationID,
		TriggerID:   bizOp.TriggerID,
		OperInstIDs: bizOp.InstanceIDs,
		DefSnapshot: defSnapshot,
		Parameters:  convertParamToDB(bizOp.Param),
		Extras:      bizOp.Param.ExtraContent,
	}
	if bizOp.InstanceIDs == nil || len(bizOp.InstanceIDs) == 0 {
		opera.OperInstEmpty = true
	}

	return opera
}

func convertParamFromDB(param Parameters) operation.Param {
	return operation.Param{
		ParentOperationID: param.ParentOperationID,
		Timeout:           param.Timeout,
		InitContent:       param.InitContent,
		RetryStartPoint:   param.RetryStartPoint,
	}
}

func convertParamToDB(param operation.Param) Parameters {
	return Parameters{
		ParentOperationID: param.ParentOperationID,
		Timeout:           param.Timeout,
		InitContent:       param.InitContent,
		RetryStartPoint:   param.RetryStartPoint,
	}
}

func convertDefFromDB(defoper DefSnapshot) operation.Definition {
	return &operation.DefinitionSnapshot{
		SnapshotName:               defoper.OperDefName,
		SnapshotActionDefNames:     defoper.ActionNames,
		SnapshotDefaultParameters:  convertParamFromDB(defoper.DefaultParameters),
		RetryStartPoint:            defoper.DefaultParameters.RetryStartPoint,
		SnapshotExtraExecutionName: defoper.ExtraExecutionName,
	}
}

func convertDefToDB(defoper operation.Definition) DefSnapshot {
	return DefSnapshot{
		OperDefName:        defoper.Name(),
		ActionNames:        defoper.ActionDefNames(),
		DefaultParameters:  convertParamToDB(defoper.DefaultParameters()),
		ExtraExecutionName: defoper.ExtraExecutionName(),
	}
}
