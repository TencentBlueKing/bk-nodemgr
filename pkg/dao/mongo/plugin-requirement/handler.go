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

// Package pluginrequirement stores independent tenant-scoped requirement file records.
package pluginrequirement

import (
	"errors"
	"fmt"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler defines requirement file metadata persistence.
type IHandler interface {
	// Create inserts a new requirement file record.
	Create(nCtx contextx.IContext, record *types.PluginRequirementFileRecord) error
	// Update replaces names and content metadata without inserting or changing the plugin or applicability.
	Update(nCtx contextx.IContext, record *types.PluginRequirementFileRecord) error
	// UpdateScope updates only applicability and audit metadata.
	UpdateScope(nCtx contextx.IContext, record *types.PluginRequirementFileRecord) error
	// Get retrieves a requirement file record by ID within the current tenant.
	Get(nCtx contextx.IContext, id string) (*types.PluginRequirementFileRecord, error)
	// List retrieves a page of requirement records and the total matching count.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.PluginRequirementFileRecord, int64, error)
}

var _ IHandler = &Handler{}

// Handler implements IHandler.
type Handler struct {
	client *mongo.Database
	// daoMap is written only by tenantDao and contains tenant-scoped DAOs.
	daoMap sync.Map
}

// New creates a requirement file metadata handler.
func New(client *mongo.Database) *Handler {
	return &Handler{client: client}
}

func (h *Handler) tenantDao(nCtx contextx.IContext) (*dao, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, fmt.Errorf("invalid requirement tenant: %w", err)
	}

	tableName := TableName(nCtx.TenantID())
	if d, ok := h.daoMap.Load(tableName); ok {
		return d.(*dao), nil // nolint:forcetypeassert
	}

	newDaoClient := newDao(tableName, h.client)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		return nil, fmt.Errorf("failed to ensure requirement indexes: %w", err)
	}
	d, _ := h.daoMap.LoadOrStore(tableName, newDaoClient)

	return d.(*dao), nil // nolint:forcetypeassert
}

// Create inserts a new requirement file record.
func (h *Handler) Create(nCtx contextx.IContext, record *types.PluginRequirementFileRecord) error {
	if record == nil {
		return errors.New("plugin requirement file record is nil")
	}
	if record.RequirementFileID == "" {
		return errors.New("plugin requirement file ID is empty")
	}

	if err := record.GetUniqueKey().Validate(); err != nil {
		return fmt.Errorf("invalid plugin requirement identity: %w", err)
	}
	if err := types.ValidatePluginRequirementFileName(record.TargetFileName); err != nil {
		return fmt.Errorf("invalid plugin requirement filename: %w", err)
	}
	scope := types.PluginRequirementFileScope{OS: record.OS, Arch: record.Arch, Version: record.Version}
	if err := scope.Validate(); err != nil {
		return fmt.Errorf("invalid plugin requirement scope: %w", err)
	}
	d, err := h.tenantDao(nCtx)
	if err != nil {
		return fmt.Errorf("failed to get requirement dao: %w", err)
	}
	if err := d.Create(nCtx, convertFromTypes(record)); err != nil {
		return fmt.Errorf("failed to create requirement file record: %w", err)
	}

	return nil
}

// Update replaces names and content metadata without inserting or changing the plugin or applicability.
func (h *Handler) Update(nCtx contextx.IContext, record *types.PluginRequirementFileRecord) error {
	if record == nil {
		return errors.New("plugin requirement file record is nil")
	}
	if record.RequirementFileID == "" {
		return errors.New("plugin requirement file ID is empty")
	}

	if err := types.ValidatePluginRequirementName(record.Name); err != nil {
		return fmt.Errorf("invalid plugin requirement name: %w", err)
	}
	if err := types.ValidatePluginRequirementFileName(record.TargetFileName); err != nil {
		return fmt.Errorf("invalid plugin requirement filename: %w", err)
	}
	d, err := h.tenantDao(nCtx)
	if err != nil {
		return fmt.Errorf("failed to get requirement dao: %w", err)
	}
	if err := d.update(nCtx, convertFromTypes(record)); err != nil {
		return fmt.Errorf("failed to update requirement file record: %w", err)
	}

	return nil
}

// UpdateScope updates applicability without modifying the stored content reference.
func (h *Handler) UpdateScope(nCtx contextx.IContext, record *types.PluginRequirementFileRecord) error {
	if record == nil {
		return errors.New("plugin requirement file record is nil")
	}
	if record.RequirementFileID == "" {
		return errors.New("plugin requirement file ID is empty")
	}
	scope := types.PluginRequirementFileScope{OS: record.OS, Arch: record.Arch, Version: record.Version}
	if err := scope.Validate(); err != nil {
		return fmt.Errorf("invalid plugin requirement scope: %w", err)
	}
	d, err := h.tenantDao(nCtx)
	if err != nil {
		return fmt.Errorf("failed to get requirement dao: %w", err)
	}
	if err := d.updateScope(nCtx, convertFromTypes(record)); err != nil {
		return fmt.Errorf("failed to update requirement scope: %w", err)
	}

	return nil
}

// List retrieves a page of requirement records and the total matching count.
func (h *Handler) List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.PluginRequirementFileRecord, int64, error) {
	if err := page.Validate(); err != nil {
		return nil, 0, fmt.Errorf("invalid requirement page: %w", err)
	}
	d, err := h.tenantDao(nCtx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get requirement dao: %w", err)
	}
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}
	total, err := d.Count(nCtx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count requirement records: %w", err)
	}
	data, err := d.List(nCtx, filter, base.ParsePage(page))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list requirement records: %w", err)
	}
	records := make([]*types.PluginRequirementFileRecord, len(data))
	for idx, item := range data {
		records[idx] = convertToTypes(item)
	}

	return records, total, nil
}

// Get retrieves a requirement file record by ID within the current tenant.
func (h *Handler) Get(nCtx contextx.IContext, id string) (*types.PluginRequirementFileRecord, error) {
	if id == "" {
		return nil, errors.New("plugin requirement file ID is empty")
	}

	d, err := h.tenantDao(nCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to get requirement dao: %w", err)
	}
	filter := append(base.AliveFilter(), bson.E{Key: FieldKeyRequirementFileID, Value: id})
	data, err := d.Get(nCtx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get requirement file record: %w", err)
	}

	return convertToTypes(data), nil
}

func convertFromTypes(record *types.PluginRequirementFileRecord) *Data {
	if record == nil {
		return nil
	}

	return &Data{
		RequirementFileID: record.RequirementFileID,
		PluginName:        record.PluginName,
		Name:              record.Name,
		TargetFileName:    record.TargetFileName,
		FileName:          record.FileName,
		OS:                conv.NonEmptyOr(record.OS, []criteria.OSType{}),
		Arch:              conv.NonEmptyOr(record.Arch, []criteria.CPUArch{}),
		Version:           conv.NonEmptyOr(record.Version, []string{}),
		MD5:               record.MD5,
		Size:              record.Size,
		UpdatedAt:         record.UpdatedAt,
		Operator:          record.Operator,
	}
}

func convertToTypes(data *Data) *types.PluginRequirementFileRecord {
	if data == nil {
		return nil
	}

	return &types.PluginRequirementFileRecord{
		OS:                conv.NonEmptyOr(data.OS, []criteria.OSType{}),
		Arch:              conv.NonEmptyOr(data.Arch, []criteria.CPUArch{}),
		Version:           conv.NonEmptyOr(data.Version, []string{}),
		RequirementFileID: data.RequirementFileID,
		PluginName:        data.PluginName,
		Name:              data.Name,
		TargetFileName:    data.TargetFileName,
		FileName:          data.FileName,
		MD5:               data.MD5,
		Size:              data.Size,
		UpdatedAt:         data.UpdatedAt,
		Operator:          data.Operator,
	}
}
