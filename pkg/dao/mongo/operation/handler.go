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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"

	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler operation handler interface.
// nolint: interfacebloat
type IHandler interface {
	// Upsert insert or update an operation.
	Upsert(nCtx contextx.IContext, operation *operation.Operation) error

	// List list operation by page and opts.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*operation.Operation, int64, error)

	// ListWithoutCount list operation by page and opts, without count.
	ListWithoutCount(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*operation.Operation, error)

	// ListOperationIDByParentOperationID lists operation IDs by parent operation ID.
	ListOperationIDByParentOperationID(nCtx contextx.IContext, page types.Page, parentID ...string) (
		[]string, error)

	// ListOperationIDByParentOperInstID lists operation IDs by parent operation instance ID.
	ListOperationIDByParentOperInstID(nCtx contextx.IContext, page types.Page, parentID ...string) (
		[]string, error)

	// Count count process by conditions.
	Count(nCtx contextx.IContext, opts ...OptFn) (int64, error)

	// Exist check operation whether exists by opts.
	Exist(nCtx contextx.IContext, opts ...OptFn) (bool, error)

	// Delete deletes operations.
	Delete(nCtx contextx.IContext, operationID ...string) error

	// DeleteByTriggerID deletes operations by trigger id.
	DeleteByTriggerID(nCtx contextx.IContext, triggerID ...string) error

	// PullOperInstIDs pull operation instance ids by operation id.
	PullOperInstIDs(nCtx contextx.IContext, operationID string, operInstIDs ...string) error

	// UpdateLatestInstBriefData updates an operation's latest instance brief data.
	UpdateLatestInstBriefData(nCtx contextx.IContext, operationID string, briefData *operation.InstanceBriefData) error

	// GetLatestOperationInstanceStatusDistributionByTriggerID gets the latest operation instance status distribution by trigger id.
	GetLatestOperationInstanceStatusDistributionByTriggerID(nCtx contextx.IContext, triggerID ...string) (
		map[string]*operation.InstanceStatusDistribution, error)

	// DistinctLatestInstState distinct with field latest_inst_state.
	DistinctLatestInstState(nCtx contextx.IContext, opts ...OptFn) ([]operation.State, error)
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

// ListWithoutCount lists operation by page and opts, without count.
func (h *handler) ListWithoutCount(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*operation.Operation, error) {
	if err := page.Validate(); err != nil {
		return nil, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	findOpt := base.ParsePage(page)

	datas, err := h.dao.List(nCtx, filter, findOpt)
	if err != nil {
		return nil, err
	}

	operations := make([]*operation.Operation, len(datas))
	for idx, data := range datas {
		operations[idx] = convertOperationFromDB(data)
	}

	return operations, nil
}

// ListOperationIDByParentOperationID lists operation IDs by parent operation ID.
func (h *handler) ListOperationIDByParentOperationID(nCtx contextx.IContext, page types.Page, parentID ...string) (
	[]string, error) {

	if err := page.Validate(); err != nil {
		return nil, err
	}

	filter := base.AliveFilter()
	filter = WithParentOperationID(parentID...)(filter)

	datas, err := h.dao.List(nCtx, filter, base.ParsePage(page), FieldKeyOperationID)
	if err != nil {
		return nil, err
	}

	operationIDs := make([]string, len(datas))
	for idx, data := range datas {
		operationIDs[idx] = data.OperationID
	}

	return operationIDs, nil
}

// ListOperationIDByParentOperInstID lists operation IDs by parent operation instance ID.
func (h *handler) ListOperationIDByParentOperInstID(nCtx contextx.IContext, page types.Page, parentID ...string) (
	[]string, error) {

	if err := page.Validate(); err != nil {
		return nil, err
	}

	filter := base.AliveFilter()
	filter = WithParentOperInstID(parentID...)(filter)

	datas, err := h.dao.List(nCtx, filter, base.ParsePage(page), FieldKeyOperationID)
	if err != nil {
		return nil, err
	}

	operationIDs := make([]string, len(datas))
	for idx, data := range datas {
		operationIDs[idx] = data.OperationID
	}

	return operationIDs, nil
}

// Count count operation by conditions.
func (h *handler) Count(nCtx contextx.IContext, opts ...OptFn) (int64, error) {
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.dao.Count(nCtx, filter)
}

// Exist check operation whether exists by opts.
func (h *handler) Exist(nCtx contextx.IContext, opts ...OptFn) (bool, error) {
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.dao.Exist(nCtx, filter)
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

// UpdateLatestInstBriefData updates an operation's latest instance brief data.
func (h *handler) UpdateLatestInstBriefData(nCtx contextx.IContext, operationID string, briefData *operation.InstanceBriefData) error {
	if nCtx == nil {
		return base.ErrEmptyParamData()
	}

	if operationID == "" {
		return base.ErrEmptyParamData()
	}

	if briefData == nil {
		return base.ErrEmptyParamData()
	}

	filter := base.AliveFilter()
	filter = WithOperationID(operationID)(filter)

	err := h.dao.UpdateField(nCtx, filter, FieldKeyLatestInstBriefData, convertInstBriefDataToDB(briefData))
	if err != nil {
		return err
	}

	return nil
}

// GetLatestOperationInstanceStatusDistributionByTriggerID gets the latest operation instance status distribution by trigger id.
func (h *handler) GetLatestOperationInstanceStatusDistributionByTriggerID(nCtx contextx.IContext, triggerID ...string) (
	map[string]*operation.InstanceStatusDistribution, error) {

	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if len(triggerID) == 0 {
		return make(map[string]*operation.InstanceStatusDistribution), nil
	}

	filter := base.AliveFilter()
	filter = WithTriggerID(triggerID...)(filter)

	results, err := h.dao.getLatestOperationInstStatusDistribution(nCtx, filter)
	if err != nil {
		return nil, err
	}

	// build the result map: trigger_id -> distribution
	distribution := make(map[string]*operation.InstanceStatusDistribution)
	for _, result := range results {
		triggerID := result.GroupKey.TriggerID
		status := result.GroupKey.Status
		count := result.Count

		if _, exists := distribution[triggerID]; !exists {
			distribution[triggerID] = &operation.InstanceStatusDistribution{
				NotInitedCount: 0,
				StatusMap:      make(map[string]int64),
			}
		}

		if status == nil {
			distribution[triggerID].NotInitedCount += count
			continue
		}

		distribution[triggerID].StatusMap[*status] = count
	}

	return distribution, nil
}

// DistinctLatestInstState distincts with field latest_inst_state.
func (h *handler) DistinctLatestInstState(nCtx contextx.IContext, opts ...OptFn) ([]operation.State, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	result, err := h.distinctString(nCtx, FieldKeyLatestInstState, opts...)
	if err != nil {
		return nil, err
	}

	stateList, err := conv.SliceToSliceWithError[string, operation.State](result, func(s string) (operation.State, error) {
		state := operation.State(s)
		if err := state.Validate(); err != nil {
			return "", err
		}

		return state, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get distinct latest inst state: %w", err)
	}

	return stateList, nil
}

func (h *handler) distinctString(nCtx contextx.IContext, key string, opts ...OptFn) ([]string, error) {
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.dao.DistinctString(nCtx, key, filter, nil)
}

func convertOperationFromDB(dbOp *Operation) *operation.Operation {
	if dbOp == nil {
		return nil
	}

	defSnapshot := convertDefFromDB(dbOp.DefSnapshot)
	param := convertParamFromDB(dbOp.Parameters)
	retryFlags := convertRetryFlagsFromDB(dbOp.RetryFlags)
	latestInstBriefData := convertInstBriefDataFromDB(dbOp.LatestInstBriefData)

	return &operation.Operation{
		TriggerID:           dbOp.TriggerID,
		OperationID:         dbOp.OperationID,
		Definition:          defSnapshot,
		InstanceIDs:         dbOp.OperInstIDs,
		Param:               param,
		RetryFlags:          retryFlags,
		LatestInstBriefData: latestInstBriefData,
		CreateTime:          dbOp.CreateTime,
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
		OperationID:         oper.OperationID,
		TriggerID:           oper.TriggerID,
		OperInstIDs:         oper.InstanceIDs,
		DefSnapshot:         defSnapshot,
		Parameters:          convertParamToDB(oper.Param),
		RetryFlags:          convertRetryFlagsToDB(oper.RetryFlags),
		LatestInstBriefData: convertInstBriefDataToDB(oper.LatestInstBriefData),
		CreateTime:          oper.CreateTime,
	}

	if len(oper.InstanceIDs) == 0 || (len(oper.RetryFlags) > 0 && oper.RetryFlags[len(oper.RetryFlags)-1].RetryInstanceID == "") {
		operation.Instantiated = true
	}

	return operation
}

func convertParamFromDB(param Parameters) operation.Param {
	return operation.Param{
		ParentOperationID: param.ParentOperationID,
		ParentOperInstID:  param.ParentOperInstID,
		Timeout:           param.Timeout,
		InitContent:       param.InitContent,
		RetryStartPoint:   param.RetryStartPoint,
	}
}

func convertParamToDB(param operation.Param) Parameters {
	return Parameters{
		ParentOperationID: param.ParentOperationID,
		ParentOperInstID:  param.ParentOperInstID,
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

// convertInstBriefDataToDB converts operation instance brief data to db format.
func convertInstBriefDataToDB(briefData *operation.InstanceBriefData) *InstBriefData {
	if briefData == nil {
		return nil
	}

	return &InstBriefData{
		OperInstID:                briefData.Metadata.OperationInstanceID,
		LifeCycle:                 convertLifeCycleToDB(briefData.Lifecycle),
		LatestActionInstBriefData: convertActionInstBriefDataToDB(briefData.LatestActionInstBriefData),
	}
}

// convertLifeCycleToDB converts lifecycle to db format.
func convertLifeCycleToDB(lifecycle *operation.Lifecycle) *LifeCycle {
	if lifecycle == nil {
		return nil
	}

	return &LifeCycle{
		State:     string(lifecycle.State),
		CreatedAt: lifecycle.CreatedAt,
		StartedAt: lifecycle.StartedAt,
		EndedAt:   lifecycle.EndedAt,
		StoppedAt: lifecycle.StoppedAt,
	}
}

func convertInstBriefDataFromDB(briefData *InstBriefData) *operation.InstanceBriefData {
	if briefData == nil {
		return nil
	}

	return &operation.InstanceBriefData{
		Metadata: &operation.InstanceMetadata{
			OperationInstanceID: briefData.OperInstID,
		},
		Lifecycle:                 convertLifeCycleFromDB(briefData.LifeCycle),
		LatestActionInstBriefData: convertActionInstBriefDataFromDB(briefData.LatestActionInstBriefData),
	}
}

func convertLifeCycleFromDB(lifecycle *LifeCycle) *operation.Lifecycle {
	if lifecycle == nil {
		return nil
	}

	return &operation.Lifecycle{
		State:     operation.State(lifecycle.State),
		CreatedAt: lifecycle.CreatedAt,
		StartedAt: lifecycle.StartedAt,
		EndedAt:   lifecycle.EndedAt,
		StoppedAt: lifecycle.StoppedAt,
	}
}

// convertActionInstBriefDataToDB converts action instance brief data to db format.
func convertActionInstBriefDataToDB(briefData *action.InstanceBriefData) *ActionInstBriefData {
	if briefData == nil {
		return nil
	}

	return &ActionInstBriefData{
		Name: briefData.Name,
		Tags: conv.SliceToSlice(briefData.Tags, func(tag action.Tag) string {
			return string(tag)
		}),
	}
}

// convertActionInstBriefDataFromDB converts action instance brief data from db format.
func convertActionInstBriefDataFromDB(briefData *ActionInstBriefData) *action.InstanceBriefData {
	if briefData == nil {
		return nil
	}

	return &action.InstanceBriefData{
		Name: briefData.Name,
		Tags: conv.SliceToSlice(briefData.Tags, func(tag string) action.Tag {
			return action.Tag(tag)
		}),
	}
}
