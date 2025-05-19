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
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"

	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler operation handler interface.
type IHandler interface {
	// Upsert insert or update an operation.
	Upsert(ctx context.Context, operation *operation.Operation) error

	// FindOne if not specified, the first undeleted record is returned.
	FindOne(ctx context.Context, opts ...OptFn) (*operation.Operation, error)

	// Listlist operation by page and opts.
	List(ctx context.Context, page types.Page, opts ...OptFn) ([]*operation.Operation, int64, error)
}

type handler struct {
	client *mongo.Database
	logger logger.Logger
	// daoMap stores dao's containing operation information.
	// Do not edit the daoMap except with the operationDao func.
	daoMap sync.Map
}

func (h *handler) tenantDao(tenantID string) *dao {
	if d, ok := h.daoMap.Load(tenantID); ok {
		return d.(*dao)
	}

	newDaoClient := newDao(tenantID, h.client, h.logger)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		h.logger.Warnf("failed to ensure networkarea indexes, err: %v", errors.Join(base.ErrEnsureIndexesFailed(), err))
	}

	daoclient, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	// return d.(*dao)
	val, ok := daoclient.(*dao)
	if !ok {
		h.logger.Errorf("invalid type stored in daoMap: %T", daoclient)
		return newDaoClient
	}

	return val
}

// New new a handler.
func New(client *mongo.Database, logger logger.Logger) IHandler {
	return &handler{
		client: client,
		logger: logger,
		daoMap: sync.Map{},
	}
}

// List lists operation by page and opts.
func (h *handler) List(ctx context.Context, page types.Page, opts ...OptFn) ([]*operation.Operation, int64, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, 0, err
	}

	if ctx == nil {
		return nil, 0, errors.New("ctx is nil")
	}
	if err := page.Validate(); err != nil {
		return nil, 0, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.tenantDao(tenantID).Count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	if num == 0 {
		return nil, 0, nil
	}

	findOpt := base.ParsePage(page)

	datas, err := h.tenantDao(tenantID).List(ctx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}

	workflows := make([]*operation.Operation, len(datas))
	for idx, data := range datas {
		workflows[idx] = convertOperationFromDB(data)
	}

	return workflows, num, nil
}

// FindOne finds one operation by opts.
func (h *handler) FindOne(ctx context.Context, opts ...OptFn) (*operation.Operation, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		return nil, errors.New("ctx is nil")
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	operations, err := h.tenantDao(tenantID).find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(operations) == 0 {
		return nil, nil
	}
	if operations[0] == nil {
		return nil, errors.New("nil operation in result set")
	}

	return convertOperationFromDB(operations[0]), nil
}

// Upsert insert or update an operation.
func (h *handler) Upsert(ctx context.Context, operation *operation.Operation) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	if operation == nil {
		return base.ErrEmptyParamData()
	}

	data := convertOperationToDB(operation)

	if err := h.tenantDao(tenantID).upsert(ctx, data); err != nil {
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

	return &Operation{
		OperationID: bizOp.OperationID,
		TriggerID:   bizOp.TriggerID,
		OperInstIDs: bizOp.InstanceIDs,
		DefSnapshot: defSnapshot,
		Parameters:  convertParamToDB(bizOp.Param),
	}
}

func convertParamFromDB(param Parameters) operation.OperationParam {
	return operation.OperationParam{
		ParentOperationID: param.ParentOperationID,
		Timeout:           param.Timeout,
		InitContent:       param.InitContent,
	}
}

func convertParamToDB(param operation.OperationParam) Parameters {
	return Parameters{
		ParentOperationID: param.ParentOperationID,
		Timeout:           param.Timeout,
		InitContent:       param.InitContent,
	}
}

func convertDefFromDB(defoper DefSnapshot) operation.Definition {
	return &operation.DefinitionSnapshot{
		SnapshotName:              defoper.OperDefName,
		SnapshotActionDefNames:    defoper.ActionNames,
		SnapshotDefaultParameters: convertParamFromDB(defoper.DefaultParameters),
	}
}

func convertDefToDB(defoper operation.Definition) DefSnapshot {
	return DefSnapshot{
		OperDefName:       defoper.Name(),
		ActionNames:       defoper.ActionDefNames(),
		DefaultParameters: convertParamToDB(defoper.DefaultParameters()),
	}
}
