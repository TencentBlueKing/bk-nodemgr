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

// Package trigger ...
package trigger

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler trigger handler interface.
type IHandler interface {
	// Get get a specified trigger.
	Get(nCtx contextx.IContext, triggerID string) (*trigger.Trigger, error)

	// Count counts triggers with options.
	Count(nCtx contextx.IContext, opts ...OptFn) (int64, error)

	// List list triggers with options.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*trigger.Trigger, int64, error)

	// Create creates a trigger.
	Create(nCtx contextx.IContext, trig *trigger.Trigger) error

	// Update updates trigger.
	Update(nCtx contextx.IContext, trig *trigger.Trigger) error

	// SwitchActive switches trigger active status.
	SwitchActive(nCtx contextx.IContext, triggerID string, active bool) error

	// Delete deletes triggers.
	Delete(nCtx contextx.IContext, triggerIDs ...string) error
}

type handler struct {
	dao *dao
}

// New create a new trigger handler.
func New(client *mongo.Database) IHandler {
	h := &handler{
		dao: newDao(client),
	}

	if err := h.dao.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).Warn("failed to ensure trigger indexes")
	}

	return h
}

// Get get a specified trigger.
func (h *handler) Get(nCtx contextx.IContext, triggerID string) (*trigger.Trigger, error) {
	if nCtx == nil {
		return nil, errors.New("nCtx is nil")
	}

	if triggerID == "" {
		return nil, errors.New("trigger-id is empty")
	}

	filter := base.AliveFilter()
	opt := base.WithStringValues(FieldKeyTriggerID, triggerID)
	filter = opt(filter)

	data, err := h.dao.Get(nCtx, filter)
	if err != nil {
		return nil, err
	}

	return convertTriggerToTypes(data), nil
}

// Count counts triggers with options.
func (h *handler) Count(nCtx contextx.IContext, opts ...OptFn) (int64, error) {
	if nCtx == nil {
		return 0, errors.New("nCtx is nil")
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.dao.Count(nCtx, filter)
	if err != nil {
		return 0, err
	}

	return num, nil
}

// List list triggers with options.
func (h *handler) List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*trigger.Trigger, int64, error) {
	if nCtx == nil {
		return nil, 0, errors.New("nCtx is nil")
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

	trigs, err := h.dao.List(nCtx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}

	data := make([]*trigger.Trigger, len(trigs))
	for idx, trig := range trigs {
		data[idx] = convertTriggerToTypes(trig)
	}

	return data, num, nil
}

// Create creates a trigger.
func (h *handler) Create(nCtx contextx.IContext, trig *trigger.Trigger) error {
	if nCtx == nil {
		return errors.New("nCtx is nil")
	}

	if trig == nil {
		return errors.New("trigger is nil")
	}

	if err := trig.Validate(); err != nil {
		return err
	}

	return h.dao.Create(nCtx, convertTriggerFromTypes(trig))
}

// Update updates a trigger.
func (h *handler) Update(nCtx contextx.IContext, trig *trigger.Trigger) error {
	if nCtx == nil {
		return errors.New("nCtx is nil")
	}

	if trig == nil {
		return errors.New("trigger is nil")
	}

	if err := trig.Validate(); err != nil {
		return err
	}

	return h.dao.update(nCtx, convertTriggerFromTypes(trig))
}

// SwitchActive switches trigger active status.
func (h *handler) SwitchActive(nCtx contextx.IContext, triggerID string, active bool) error {
	if nCtx == nil {
		return errors.New("nCtx is nil")
	}

	if triggerID == "" {
		return errors.New("trigger-id is empty")
	}

	filter := base.AliveFilter()
	filter = WithTriggerID(triggerID)(filter)

	return h.dao.UpdateField(nCtx, filter, FieldKeyActive, active)
}

// Delete deletes triggers.
func (h *handler) Delete(nCtx contextx.IContext, triggerIDs ...string) error {
	if nCtx == nil {
		return errors.New("nCtx is nil")
	}

	if len(triggerIDs) == 0 {
		return errors.New("trigger-ids is empty")
	}

	return h.dao.delete(nCtx, triggerIDs...)
}

func convertTriggerFromTypes(trig *trigger.Trigger) *Trigger {
	t := &Trigger{
		TriggerID:       trig.TriggerID,
		Category:        string(trig.Category),
		Active:          trig.Active,
		CreatedAt:       trig.CreatedAt,
		UpdatedAt:       trig.UpdatedAt,
		LastTriggeredAt: trig.LastTriggeredAt,
	}

	switch trig.Category {
	case trigger.CategoryOnce:
		meta, ok := trig.Metadata.(*trigger.MetadataOnce)
		if ok {
			t.MetadataOnce = MetadataOnce{
				CleanPolicy: CleanPolicy{
					MaxDays: meta.CleanPolicy.MaxDays,
				},
			}
		}

	case trigger.CategoryOrdered:
		meta, ok := trig.Metadata.(*trigger.MetadataOrdered)
		if ok {
			t.MetadataOrdered = MetadataOrdered{
				MaxConcurrencyNum: meta.MaxConcurrencyNum,
				CleanPolicy: CleanPolicy{
					MaxDays: meta.CleanPolicy.MaxDays,
				},
			}
		}

	case trigger.CategoryPeriodic:
		meta, ok := trig.Metadata.(*trigger.MetadataPeriodic)
		if ok {
			t.MetadataPeriodic = MetadataPeriodic{
				Interval:           meta.Interval,
				AllowedConcurrency: meta.AllowedConcurrency,
				CleanPolicy: PeriodicCleanPolicy{
					MaxOperInstNum: meta.CleanPolicy.MaxOperInstNum,
				},
			}
		}
	}

	return t
}

func convertTriggerToTypes(trig *Trigger) *trigger.Trigger {
	typeTrigger := &trigger.Trigger{
		TriggerID:       trig.TriggerID,
		Category:        trigger.Category(trig.Category),
		Active:          trig.Active,
		CreatedAt:       trig.CreatedAt,
		UpdatedAt:       trig.UpdatedAt,
		LastTriggeredAt: trig.LastTriggeredAt,
	}

	switch trigger.Category(trig.Category) {
	case trigger.CategoryOnce:
		typeTrigger.Metadata = &trigger.MetadataOnce{
			CleanPolicy: trigger.MetadataCleanPolicy{
				MaxDays: trig.MetadataOnce.CleanPolicy.MaxDays,
			},
		}

	case trigger.CategoryOrdered:
		typeTrigger.Metadata = &trigger.MetadataOrdered{
			MaxConcurrencyNum: trig.MetadataOrdered.MaxConcurrencyNum,
			CleanPolicy: trigger.MetadataCleanPolicy{
				MaxDays: trig.MetadataOrdered.CleanPolicy.MaxDays,
			},
		}

	case trigger.CategoryPeriodic:
		typeTrigger.Metadata = &trigger.MetadataPeriodic{
			Interval:           trig.MetadataPeriodic.Interval,
			AllowedConcurrency: trig.MetadataPeriodic.AllowedConcurrency,
			CleanPolicy: trigger.PeriodicMetadataCleanPolicy{
				MaxOperInstNum: trig.MetadataPeriodic.CleanPolicy.MaxOperInstNum,
			},
		}
	}

	return typeTrigger
}
