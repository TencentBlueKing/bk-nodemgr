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

package packageexport

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler defines package export DAO operations.
type IHandler interface {
	// List lists package exports by page and options.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.PackageExport, int64, error)

	// Create creates a package export record.
	Create(nCtx contextx.IContext, exportData *types.PackageExport) error

	// UpdateFields updates package export fields.
	UpdateFields(nCtx contextx.IContext, fields types.PackageExportFields, exportData ...*types.PackageExport) error

	// Get gets a package export by its export ID.
	Get(nCtx contextx.IContext, exportID string) (*types.PackageExport, error)

	// Delete deletes a package export by its export ID.
	Delete(nCtx contextx.IContext, exportID string) error
}

var _ IHandler = &Handler{}

// Handler operates the fixed package export collection.
type Handler struct {
	dao *dao
}

// New creates a package export handler.
func New(client *mongo.Database) *Handler {
	d := newDao(client)
	if err := d.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).Warn("failed to ensure package export indexes")
	}

	return &Handler{dao: d}
}

// List lists package exports by page and options.
func (h *Handler) List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.PackageExport, int64, error) {
	if nCtx == nil {
		return nil, 0, base.ErrInvalidContext()
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.dao.Count(nCtx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count package exports: %w", err)
	}

	datas, err := h.dao.List(nCtx, filter, base.ParsePage(page))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list package exports: %w", err)
	}

	exports := make([]*types.PackageExport, len(datas))
	for idx, data := range datas {
		exports[idx] = convertToTypes(data)
	}

	return exports, num, nil
}

// Create creates a package export record.
func (h *Handler) Create(nCtx contextx.IContext, exportData *types.PackageExport) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	if exportData == nil {
		return base.ErrEmptyParamData()
	}
	if exportData.ExportID == "" {
		return errors.New("export id should not be empty")
	}

	exportData.TenantID = nCtx.TenantID()
	if err := h.dao.Create(nCtx, convertFromTypes(exportData)); err != nil {
		return fmt.Errorf("failed to create package export: %w", err)
	}

	return nil
}

// UpdateFields updates package export fields.
func (h *Handler) UpdateFields(nCtx contextx.IContext, fields types.PackageExportFields, exportData ...*types.PackageExport) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	if len(exportData) == 0 {
		return base.ErrEmptyParamData()
	}

	if fields == (types.PackageExportFields{}) {
		return base.ErrEmptyParamData()
	}

	docs := make([]*base.DocumentFieldUpdate, 0, len(exportData))
	for _, data := range exportData {
		if data == nil || data.ExportID == "" {
			return base.ErrInvalidItemInParamList()
		}

		dbExport := convertFromTypes(data)
		updates := generatePackageExportUpdates(fields, dbExport)
		filter := base.AliveFilter()
		filter = WithExportID(data.ExportID)(filter)
		docs = append(docs, &base.DocumentFieldUpdate{
			Filter: filter,
			Fields: updates,
		})
	}

	if err := h.dao.UpdateOneFieldBulk(nCtx, docs); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to update package export")

		return err
	}

	return nil
}

func generatePackageExportUpdates(fields types.PackageExportFields, exportData *Data) map[string]any {
	updates := make(map[string]any)
	if fields.WorkflowID {
		updates[FieldKeyWorkflowID] = exportData.WorkflowID
	}
	if fields.StorageKey {
		updates[FieldKeyStorageKey] = exportData.StorageKey
	}
	if fields.DownloadName {
		updates[FieldKeyDownloadName] = exportData.DownloadName
	}
	if fields.MD5 {
		updates[FieldKeyMD5] = exportData.MD5
	}
	if fields.Size {
		updates[FieldKeySize] = exportData.Size
	}
	if fields.Operator {
		updates[FieldKeyOperator] = exportData.Operator
	}

	return updates
}

// Get gets a package export by its export ID.
func (h *Handler) Get(nCtx contextx.IContext, exportID string) (*types.PackageExport, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if exportID == "" {
		return nil, errors.New("export id should not be empty")
	}

	filter := base.AliveFilter()
	filter = WithExportID(exportID)(filter)

	data, err := h.dao.Get(nCtx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get package export: %w", err)
	}

	return convertToTypes(data), nil
}

// Delete deletes a package export by its export ID.
func (h *Handler) Delete(nCtx contextx.IContext, exportID string) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if exportID == "" {
		return errors.New("export id should not be empty")
	}

	filter := base.AliveFilter()
	filter = WithExportID(exportID)(filter)

	if err := h.dao.DeleteMany(nCtx, filter); err != nil {
		return fmt.Errorf("failed to delete package export: %w", err)
	}

	return nil
}

func convertFromTypes(exportData *types.PackageExport) *Data {
	if exportData == nil {
		return nil
	}

	return &Data{
		ExportID:     exportData.ExportID,
		WorkflowID:   exportData.WorkflowID,
		TenantID:     exportData.TenantID,
		StorageKey:   exportData.StorageKey,
		DownloadName: exportData.DownloadName,
		MD5:          exportData.MD5,
		Size:         exportData.Size,
		Operator:     exportData.Operator,
	}
}

func convertToTypes(data *Data) *types.PackageExport {
	if data == nil {
		return nil
	}

	return &types.PackageExport{
		ExportID:     data.ExportID,
		WorkflowID:   data.WorkflowID,
		TenantID:     data.TenantID,
		StorageKey:   data.StorageKey,
		DownloadName: data.DownloadName,
		MD5:          data.MD5,
		Size:         data.Size,
		Operator:     data.Operator,
	}
}
