/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package trigger ...
package trigger

import (
	"context"
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler trigger handler interface.
type IHandler interface {
	// Get get a specified trigger.
	Get(ctx context.Context, triggerID string) (*trigger.Trigger, error)

	// List list triggers with options.
	List(ctx context.Context, page types.Page, opts ...OptFn) ([]*trigger.Trigger, int64, error)

	// Create creates a trigger.
	Create(ctx context.Context, trig *trigger.Trigger) error

	// Update updates trigger.
	Update(ctx context.Context, trig *trigger.Trigger) error

	// UpdateState updates trigger state.
	UpdateState(ctx context.Context, triggerID string, state trigger.State) error
}

type handler struct {
	logger logger.Logger
	dao    *dao
}

// New create a new trigger handler.
func New(client *mongo.Database, logger logger.Logger) IHandler {
	return &handler{
		logger: logger,
		dao:    newDao(client, logger),
	}
}

// Get get a specified trigger.
func (h *handler) Get(ctx context.Context, triggerID string) (*trigger.Trigger, error) {
	if ctx == nil {
		return nil, errors.New("ctx is nil")
	}

	if triggerID == "" {
		return nil, errors.New("trigger-id is empty")
	}

	filter := base.AliveFilter()
	opt := base.WithStringValues(FieldKeyTriggerID, triggerID)
	filter = opt(filter)

	data, err := h.dao.Get(ctx, filter)
	if err != nil {
		return nil, err
	}

	return convertTriggerToTypes(data), nil
}

// List list triggers with options.
func (h *handler) List(ctx context.Context, page types.Page, opts ...OptFn) ([]*trigger.Trigger, int64, error) {
	if ctx == nil {
		return nil, 0, errors.New("ctx is nil")
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

	trigs, err := h.dao.List(ctx, filter, findOpt)
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
func (h *handler) Create(ctx context.Context, trig *trigger.Trigger) error {
	if ctx == nil {
		return errors.New("ctx is nil")
	}

	if trig == nil {
		return errors.New("trigger is nil")
	}

	if err := trig.Validate(); err != nil {
		return err
	}

	return h.dao.Create(ctx, convertTriggerFromTypes(trig))
}

// Update updates a trigger.
func (h *handler) Update(ctx context.Context, trig *trigger.Trigger) error {
	if ctx == nil {
		return errors.New("ctx is nil")
	}

	if trig == nil {
		return errors.New("trigger is nil")
	}

	if err := trig.Validate(); err != nil {
		return err
	}

	return h.dao.update(ctx, convertTriggerFromTypes(trig))
}

// UpdateState updates a trigger's state.
func (h *handler) UpdateState(ctx context.Context, triggerID string, state trigger.State) error {
	if ctx == nil {
		return errors.New("ctx is nil")
	}

	if triggerID == "" {
		return errors.New("trigger-id is empty")
	}

	filter := base.AliveFilter()
	filter = WithTriggerID(triggerID)(filter)

	return h.dao.UpdateField(ctx, filter, FieldKeyState, string(state))
}

func convertTriggerFromTypes(trig *trigger.Trigger) *Trigger {
	t := &Trigger{
		TriggerID:       trig.TriggerID,
		Category:        string(trig.Category),
		State:           string(trig.State),
		CreatedAt:       trig.CreatedAt,
		UpdatedAt:       trig.UpdatedAt,
		LastTriggeredAt: trig.LastTriggeredAt,
	}

	switch trig.Category {
	case trigger.CategoryOnce:
		t.MetadataOnce = MetadataOnce{}

	case trigger.CategoryOrdered:
		meta, ok := trig.Metadata.(*trigger.MetadataOrdered)
		if ok {
			t.MetadataOrdered = MetadataOrdered{
				MaxConcurrencyNum: meta.MaxConcurrencyNum,
			}
		}

	case trigger.CategoryPeriodic:
		meta, ok := trig.Metadata.(*trigger.MetadataPeriodic)
		if ok {
			t.MetadataPeriodic = MetadataPeriodic{
				Interval:           meta.Interval,
				AllowedConcurrency: meta.AllowedConcurrency,
			}
		}
	}

	return t
}

func convertTriggerToTypes(trig *Trigger) *trigger.Trigger {
	typeTrigger := &trigger.Trigger{
		TriggerID:       trig.TriggerID,
		Category:        trigger.Category(trig.Category),
		State:           trigger.State(trig.State),
		CreatedAt:       trig.CreatedAt,
		UpdatedAt:       trig.UpdatedAt,
		LastTriggeredAt: trig.LastTriggeredAt,
	}

	switch trigger.Category(trig.Category) {
	case trigger.CategoryOnce:
		typeTrigger.Metadata = &trigger.MetadataOnce{}

	case trigger.CategoryOrdered:
		typeTrigger.Metadata = &trigger.MetadataOrdered{
			MaxConcurrencyNum: trig.MetadataOrdered.MaxConcurrencyNum,
		}

	case trigger.CategoryPeriodic:
		typeTrigger.Metadata = &trigger.MetadataPeriodic{
			Interval:           trig.MetadataPeriodic.Interval,
			AllowedConcurrency: trig.MetadataPeriodic.AllowedConcurrency,
		}
	}

	return typeTrigger
}
