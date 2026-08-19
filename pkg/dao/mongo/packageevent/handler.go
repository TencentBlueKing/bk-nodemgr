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

package packageevent

import (
	"fmt"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler package event handler interface.
type IHandler interface {
	// List list package events by page and conditions.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.PackageEvent, int64, error)

	// Count count package events by conditions.
	Count(nCtx contextx.IContext, opts ...OptFn) (int64, error)

	// CreateMany create multiple package events.
	CreateMany(nCtx contextx.IContext, events ...*types.PackageEvent) error

	IDistinctor
}

// IDistinctor defines the interface for distinctor.
type IDistinctor interface {
	// DistinctReleaseType distincts with field type.
	DistinctReleaseType(nCtx contextx.IContext, opts ...OptFn) ([]types.ReleaseType, error)

	// DistinctEventType distincts with field type.
	DistinctEventType(nCtx contextx.IContext, opts ...OptFn) ([]types.PackageEventType, error)

	// DistinctOsType distincts with field type.
	DistinctOsType(nCtx contextx.IContext, opts ...OptFn) ([]criteria.OSType, error)

	// DistinctCPUArch distincts cpu archs.
	DistinctCPUArch(nCtx contextx.IContext, opts ...OptFn) ([]criteria.CPUArch, error)

	// DistinctVersion distincts with field operator.
	DistinctVersion(nCtx contextx.IContext, opts ...OptFn) ([]string, error)

	// DistinctOperator distincts with field operator.
	DistinctOperator(nCtx contextx.IContext, opts ...OptFn) ([]string, error)
}

var _ IHandler = &Handler{}

// Handler this is a Handler to operate process table.
type Handler struct {
	client *mongo.Database
	daoMap sync.Map
}

// New create a new package event handler.
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
		logger.G.Sys().WithErr(err).With("tenant-id", tenantID).Warn("failed to ensure package event indexes")
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	return d.(*dao) // nolint: forcetypeassert
}

// List list package events by page and conditions.
func (h *Handler) List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.PackageEvent, int64, error) {
	if nCtx == nil {
		return nil, 0, base.ErrInvalidContext()
	}
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.tenantDao(tenantID).Count(nCtx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count package events: %w", err)
	}

	findOpt := base.ParsePage(page)

	events, err := h.tenantDao(tenantID).List(nCtx, filter, findOpt)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list package events: %w", err)
	}

	data := make([]*types.PackageEvent, len(events))
	for idx, event := range events {
		data[idx] = convertPackageEventToTypes(event)
	}

	return data, num, nil
}

// Count count the number of package events by conditions.
func (h *Handler) Count(nCtx contextx.IContext, opts ...OptFn) (int64, error) {
	if nCtx == nil {
		return 0, base.ErrInvalidContext()
	}
	if err := nCtx.CheckTenantID(); err != nil {
		return 0, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.tenantDao(tenantID).Count(nCtx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count package events: %w", err)
	}

	return num, nil
}

// CreateMany create a new package event.
func (h *Handler) CreateMany(nCtx contextx.IContext, events ...*types.PackageEvent) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()

	if len(events) == 0 {
		return base.ErrEmptyParamData()
	}

	data := make([]*PackageEvent, len(events))
	for idx, event := range events {
		if event == nil {
			return base.ErrInvalidItemInParamList()
		}

		data[idx] = convertPackageEventFromTypes(event)
		if data[idx].TenantID == "" {
			data[idx].TenantID = tenantID
		}
		if err := base.CheckTenantIDMatched(tenantID, data[idx].TenantID); err != nil {
			return fmt.Errorf("failed to create package event: %w", err)
		}

		// generate sequence
		sequence, err := h.tenantDao(tenantID).counter.Generate(nCtx, tableNamePrefix)
		if err != nil {
			return fmt.Errorf("failed to generate sequence: %w", err)
		}
		data[idx].EventID = sequence
	}

	if err := h.tenantDao(tenantID).CreateMany(nCtx, data); err != nil {
		return fmt.Errorf("failed to create package events: %w", err)
	}

	return nil
}

// DistinctReleaseType returns distinct values of release type field.
func (h *Handler) DistinctReleaseType(nCtx contextx.IContext, opts ...OptFn) ([]types.ReleaseType, error) {
	result, err := h.distinctString(nCtx, FieldKeyReleaseType, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to distinct release type: %w", err)
	}

	return types.StringListToReleaseTypeList(result), nil
}

// DistinctEventType returns distinct values of event type field.
func (h *Handler) DistinctEventType(nCtx contextx.IContext, opts ...OptFn) ([]types.PackageEventType, error) {
	result, err := h.distinctString(nCtx, FieldKeyEventType, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to distinct event type: %w", err)
	}

	return types.StringListToPackageEventTypeList(result), nil
}

// DistinctOsType returns distinct values of os type field.
func (h *Handler) DistinctOsType(nCtx contextx.IContext, opts ...OptFn) ([]criteria.OSType, error) {
	result, err := h.distinctString(nCtx, FieldKeyOSType, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to distinct os type: %w", err)
	}

	osList, err := conv.SliceToSliceWithError[string, criteria.OSType](result, func(s string) (criteria.OSType, error) {
		osType := criteria.OSType(s)
		if err := osType.Validate(); err != nil {
			return "", err
		}

		return osType, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get distinct os type: %w", err)
	}

	return osList, nil
}

// DistinctCPUArch distincts cpu archs.
func (h *Handler) DistinctCPUArch(nCtx contextx.IContext, opts ...OptFn) ([]criteria.CPUArch, error) {
	result, err := h.distinctString(nCtx, FieldKeyCPUArch, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to distinct cpu arch: %w", err)
	}

	archList, err := conv.SliceToSliceWithError[string, criteria.CPUArch](result, func(s string) (criteria.CPUArch, error) {
		arch := criteria.CPUArch(s)
		if err := arch.Validate(); err != nil {
			return "", err
		}

		return arch, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get distinct cpu arch: %w", err)
	}

	return archList, nil
}

// DistinctOperator returns distinct values of operator field.
func (h *Handler) DistinctOperator(nCtx contextx.IContext, opts ...OptFn) ([]string, error) {
	return h.distinctString(nCtx, FieldKeyOperator, opts...)
}

// DistinctVersion returns distinct values of version field.
func (h *Handler) DistinctVersion(nCtx contextx.IContext, opts ...OptFn) ([]string, error) {
	return h.distinctString(nCtx, FieldKeyVersion, opts...)
}

func (h *Handler) distinctString(nCtx contextx.IContext, key string, opts ...OptFn) ([]string, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	result, err := h.tenantDao(tenantID).DistinctString(nCtx, key, filter, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to distinct string: %w", err)
	}

	return result, nil
}

func convertPackageEventToTypes(event *PackageEvent) *types.PackageEvent {
	return &types.PackageEvent{
		TenantID:    event.TenantID,
		Name:        event.Name,
		EventType:   types.PackageEventType(event.EventType),
		Generation:  types.Generation(event.Generation),
		ReleaseType: types.ReleaseType(event.ReleaseType),
		Version:     event.Version,
		CPUArch:     criteria.CPUArch(event.CPUArch),
		OSType:      criteria.OSType(event.OSType),
		Operator:    event.Operator,
		OperateTime: event.OperateTime,
	}
}

func convertPackageEventFromTypes(event *types.PackageEvent) *PackageEvent {
	return &PackageEvent{
		TenantID:    event.TenantID,
		Name:        event.Name,
		EventType:   string(event.EventType),
		Generation:  int64(event.Generation),
		ReleaseType: string(event.ReleaseType),
		Version:     event.Version,
		CPUArch:     string(event.CPUArch),
		OSType:      string(event.OSType),
		Operator:    event.Operator,
		OperateTime: event.OperateTime,
	}
}
