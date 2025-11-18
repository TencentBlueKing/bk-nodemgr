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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"

	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler operation handler interface.
type IHandler interface {
	// Upsert insert or update an operation.
	Upsert(nCtx contextx.IContext, operation *operation.Operation) error

	// List list operation by page and opts.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*operation.Operation, int64, error)

	// Delete deletes operations.
	Delete(nCtx contextx.IContext, operationID ...string) error

	// DeleteByTriggerID deletes operations by trigger id.
	DeleteByTriggerID(nCtx contextx.IContext, triggerID ...string) error

	// PullOperInstIDs pull operation instance ids by operation id.
	PullOperInstIDs(nCtx contextx.IContext, operationID string, operInstIDs ...string) error
}

type handler struct {
	dao *dao
}

// New new a handler.
func New(client *mongo.Database) IHandler {
	h := &handler{
		dao: newDao(client),
	}

	if err := h.dao.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).Warn("failed to ensure operation indexes")
	}

	return h
}

// List lists operation by page and opts.
func (h *handler) List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*operation.Operation, int64, error) {
	if err := page.Validate(); err != nil {
		return nil, 0, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.dao.Count(nCtx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := base.ParsePage(page)

	datas, err := h.dao.List(nCtx, filter, findOpt)
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
func (h *handler) Upsert(nCtx contextx.IContext, operation *operation.Operation) error {
	if operation == nil {
		return base.ErrEmptyParamData()
	}

	data := convertOperationToDB(operation)

	if err := h.dao.upsert(nCtx, data); err != nil {
		return err
	}

	return nil
}

// Delete deletes operations.
func (h *handler) Delete(nCtx contextx.IContext, operationID ...string) error {
	if len(operationID) == 0 {
		return base.ErrEmptyParamData()
	}

	return h.dao.delete(nCtx, operationID...)
}

// DeleteByTriggerID deletes operations by trigger id.
func (h *handler) DeleteByTriggerID(nCtx contextx.IContext, triggerID ...string) error {
	if len(triggerID) == 0 {
		return base.ErrEmptyParamData()
	}

	return h.dao.deleteByTriggerID(nCtx, triggerID...)
}

// PullOperInstIDs pull operation instance ids by operation id.
func (h *handler) PullOperInstIDs(nCtx contextx.IContext, operationID string, operInstIDs ...string) error {
	if operationID == "" || len(operInstIDs) == 0 {
		return base.ErrEmptyParamData()
	}

	return h.dao.pullField(nCtx, operationID, "oper_inst_ids", operInstIDs)
}

func convertOperationFromDB(dbOp *Operation) *operation.Operation {
	if dbOp == nil {
		return nil
	}

	defSnapshot := convertDefFromDB(dbOp.DefSnapshot)
	param := convertParamFromDB(dbOp.Parameters)
	retryFlags := convertRetryFlagsFromDB(dbOp.RetryFlags)

	return &operation.Operation{
		TriggerID:   dbOp.TriggerID,
		OperationID: dbOp.OperationID,
		Definition:  defSnapshot,
		InstanceIDs: dbOp.OperInstIDs,
		Param:       param,
		RetryFlags:  retryFlags,
	}
}

func convertOperationToDB(oper *operation.Operation) *Operation {
	if oper == nil {
		return nil
	}
	var defSnapshot DefSnapshot
	if oper.Definition != nil {
		defSnapshot = convertDefToDB(oper.Definition)
	}

	operation := &Operation{
		OperationID: oper.OperationID,
		TriggerID:   oper.TriggerID,
		OperInstIDs: oper.InstanceIDs,
		DefSnapshot: defSnapshot,
		Parameters:  convertParamToDB(oper.Param),
		RetryFlags:  convertRetryFlagsToDB(oper.RetryFlags),
	}

	if len(oper.InstanceIDs) == 0 || (len(oper.RetryFlags) > 0 && oper.RetryFlags[len(oper.RetryFlags)-1].RetryInstanceID == "") {
		operation.Instantiated = true
	}

	return operation
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

func convertRetryFlagsFromDB(retryFlags []RetryFlag) []operation.RetryFlag {
	flags := make([]operation.RetryFlag, 0, len(retryFlags))
	for _, flag := range retryFlags {
		flags = append(flags, operation.RetryFlag{
			Mode:             operation.RetryMode(flag.Mode),
			SourceInstanceID: flag.SourceInstanceID,
			RetryInstanceID:  flag.RetryInstanceID,
		})
	}

	return flags
}

func convertRetryFlagsToDB(retryFlags []operation.RetryFlag) []RetryFlag {
	flags := make([]RetryFlag, 0, len(retryFlags))
	for _, flag := range retryFlags {
		flags = append(flags, RetryFlag{
			Mode:             string(flag.Mode),
			SourceInstanceID: flag.SourceInstanceID,
			RetryInstanceID:  flag.RetryInstanceID,
		})
	}

	return flags
}
