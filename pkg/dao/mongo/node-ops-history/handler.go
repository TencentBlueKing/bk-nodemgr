/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package nodeopshistory

import (
	"fmt"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler provides operation history persistence and queries.
type IHandler interface {
	// Create inserts an operation and all its snapshots atomically as one document.
	Create(nCtx contextx.IContext, history *types.NodeOpsHistory) error

	// GetWithFields retrieves selected fields of an operation by ID.
	GetWithFields(nCtx contextx.IContext, operateID string, selection *types.NodeOpsHistoryFieldSelection) (*types.NodeOpsHistory, error)

	// ListWithFields lists selected fields matching the supplied filters and page.
	ListWithFields(nCtx contextx.IContext, page types.Page, selection *types.NodeOpsHistoryFieldSelection, opts ...OptFn) (
		[]*types.NodeOpsHistory, int64, error)

	// Count counts operations matching the supplied filters, without pagination.
	Count(nCtx contextx.IContext, opts ...OptFn) (int64, error)
}

var _ IHandler = &Handler{}

// Handler operates the tenant-specific operation history collection.
type Handler struct {
	client *mongo.Database

	// Only tenantDao writes daoMap.
	daoMap sync.Map
}

// New creates an operation history handler.
func New(client *mongo.Database) *Handler {
	return &Handler{
		client: client,
		daoMap: sync.Map{},
	}
}

func (h *Handler) tenantDao(tenantID string) *dao {
	if d, ok := h.daoMap.Load(tenantID); ok {
		return d.(*dao) // nolint: forcetypeassert
	}

	newDaoClient := newDao(tenantID, h.client)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).With("tenant-id", tenantID).Warn("failed to ensure node operation history indexes")
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	return d.(*dao) // nolint: forcetypeassert
}

// Create inserts an operation and all its snapshots atomically as one document.
func (h *Handler) Create(nCtx contextx.IContext, history *types.NodeOpsHistory) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}
	if history == nil || len(history.Snapshots) == 0 {
		return base.ErrEmptyParamData()
	}
	if err := base.CheckTenantIDMatched(nCtx.TenantID(), history.TenantID); err != nil {
		return err
	}

	for _, snapshot := range history.Snapshots {
		if snapshot == nil {
			return base.ErrInvalidItemInParamList()
		}
	}

	data := convertNodeOpsHistoryFromTypes(history)
	if err := h.tenantDao(nCtx.TenantID()).Create(nCtx, data); err != nil {
		return fmt.Errorf("failed to create node operation history: %w", err)
	}

	return nil
}

// GetWithFields retrieves selected fields of an operation by ID.
// Nil or an empty selection returns the complete operation.
func (h *Handler) GetWithFields(nCtx contextx.IContext, operateID string, selection *types.NodeOpsHistoryFieldSelection) (
	*types.NodeOpsHistory, error) {

	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}
	if operateID == "" {
		return nil, base.ErrEmptyParamData()
	}

	filter := base.AliveFilter()
	filter = WithOperateID(operateID)(filter)
	fields := convertNodeOpsHistoryFieldSelectionToFields(selection)
	data, err := h.tenantDao(nCtx.TenantID()).Get(nCtx, filter, fields...)
	if err != nil {
		return nil, fmt.Errorf("failed to get node operation history: %w", err)
	}

	return convertNodeOpsHistoryToTypes(data), nil
}

// ListWithFields lists selected fields matching the supplied filters and page.
// Nil or an empty selection returns complete operations.
func (h *Handler) ListWithFields(nCtx contextx.IContext, page types.Page, selection *types.NodeOpsHistoryFieldSelection, opts ...OptFn) (
	[]*types.NodeOpsHistory, int64, error) {

	if nCtx == nil {
		return nil, 0, base.ErrInvalidContext()
	}
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}
	if err := page.Validate(); err != nil {
		return nil, 0, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	d := h.tenantDao(nCtx.TenantID())
	total, err := d.Count(nCtx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count node operation history: %w", err)
	}
	findOpt := base.ParsePage(page).SetSort(bson.D{
		{
			Key:   FieldKeyOperateTime,
			Value: -1,
		},
		{
			Key:   FieldKeyOperateID,
			Value: 1,
		},
	})
	fields := convertNodeOpsHistoryFieldSelectionToFields(selection)
	rows, err := d.List(nCtx, filter, findOpt, fields...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list node operation history: %w", err)
	}
	items := make([]*types.NodeOpsHistory, len(rows))
	for i, row := range rows {
		items[i] = convertNodeOpsHistoryToTypes(row)
	}

	return items, total, nil
}

// Count counts operations matching the supplied filters, without pagination.
func (h *Handler) Count(nCtx contextx.IContext, opts ...OptFn) (int64, error) {
	if nCtx == nil {
		return 0, base.ErrInvalidContext()
	}
	if err := nCtx.CheckTenantID(); err != nil {
		return 0, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	total, err := h.tenantDao(nCtx.TenantID()).Count(nCtx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count node operation history: %w", err)
	}

	return total, nil
}

func convertNodeOpsHistoryFieldSelectionToFields(selection *types.NodeOpsHistoryFieldSelection) []string {
	fields := make([]string, 0)
	if selection == nil {
		return fields
	}
	if selection.TenantID {
		fields = append(fields, FieldKeyTenantID)
	}
	if selection.OperateID {
		fields = append(fields, FieldKeyOperateID)
	}
	if selection.Operator {
		fields = append(fields, FieldKeyOperator)
	}
	if selection.OperateTime {
		fields = append(fields, FieldKeyOperateTime)
	}
	if selection.NodeRole {
		fields = append(fields, FieldKeyNodeRole)
	}
	if selection.BizIDs {
		fields = append(fields, FieldKeyBizIDs)
	}
	if selection.Snapshots {
		fields = append(fields, FieldKeySnapshots)
	}

	return fields
}

func convertHostSnapshotFromTypes(snapshot *types.NodeOpsHistoryHostSnapshot) HostSnapshot {
	return HostSnapshot{
		HostID:                  snapshot.HostID,
		OpsConsoleHostID:        snapshot.OpsConsoleHostID,
		OpsConsoleHostInnerIP:   snapshot.OpsConsoleHostInnerIP,
		OpsConsoleHostInnerIPV6: snapshot.OpsConsoleHostInnerIPV6,
		OpsOutBandIP:            snapshot.OpsOutBandIP,
		OpsOutBandPort:          snapshot.OpsOutBandPort,
		OpsOutBandProtocol:      snapshot.OpsOutBandProtocol,
		OpsOutBandType:          snapshot.OpsOutBandType,
	}
}

func convertHostSnapshotToTypes(snapshot HostSnapshot) *types.NodeOpsHistoryHostSnapshot {
	return &types.NodeOpsHistoryHostSnapshot{
		HostID:                  snapshot.HostID,
		OpsConsoleHostID:        snapshot.OpsConsoleHostID,
		OpsConsoleHostInnerIP:   snapshot.OpsConsoleHostInnerIP,
		OpsConsoleHostInnerIPV6: snapshot.OpsConsoleHostInnerIPV6,
		OpsOutBandIP:            snapshot.OpsOutBandIP,
		OpsOutBandPort:          snapshot.OpsOutBandPort,
		OpsOutBandProtocol:      snapshot.OpsOutBandProtocol,
		OpsOutBandType:          snapshot.OpsOutBandType,
	}
}

func convertNodeOpsHistoryToTypes(data *Data) *types.NodeOpsHistory {
	snapshots := make([]*types.NodeOpsHistoryHostSnapshot, len(data.Snapshots))
	for idx, snapshot := range data.Snapshots {
		snapshots[idx] = convertHostSnapshotToTypes(snapshot)
	}

	return &types.NodeOpsHistory{
		TenantID: data.TenantID,
		Info: types.NodeOpsHistoryInfo{
			OperateID:   data.OperateID,
			Operator:    data.Operator,
			OperateTime: data.OperateTime,
			NodeRole:    types.NodeRole(data.NodeRole),
			BizIDs:      data.BizIDs,
		},
		Snapshots: snapshots,
	}
}

func convertNodeOpsHistoryFromTypes(history *types.NodeOpsHistory) *Data {
	snapshots := make([]HostSnapshot, len(history.Snapshots))
	for idx, snapshot := range history.Snapshots {
		snapshots[idx] = convertHostSnapshotFromTypes(snapshot)
	}

	return &Data{
		TenantID:    history.TenantID,
		OperateID:   history.Info.OperateID,
		Operator:    history.Info.Operator,
		OperateTime: history.Info.OperateTime,
		NodeRole:    string(history.Info.NodeRole),
		BizIDs:      history.Info.BizIDs,
		Snapshots:   snapshots,
	}
}
