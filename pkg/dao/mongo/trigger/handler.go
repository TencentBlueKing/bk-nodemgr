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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigengine"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Handler trigger handler interface.
type Handler interface {
	// Upsert update or insert a trigger
	Upsert(ctx context.Context, trigger *trigengine.Trigger) error

	// FindAll find all triggers
	FindAll(ctx context.Context) ([]*trigengine.Trigger, error)

	// Find trigger by options.
	Find(ctx context.Context, opts ...OptFn) ([]*trigengine.Trigger, error)

	// FindOne find one trigger by options.
	FindOne(ctx context.Context, opts ...OptFn) (*trigengine.Trigger, error)
}

type handler struct {
	dao *dao
}

// New create a new host handler.
func New(client *mongo.Database, logger logger.Logger) Handler {
	return &handler{
		dao: newDao(client, logger),
	}
}

// Upsert ...
func (h handler) Upsert(ctx context.Context, trigger *trigengine.Trigger) error {
	if trigger == nil {
		return errors.New("trigger is nil")
	}

	if err := trigger.Validate(); err != nil {
		return err
	}

	data := &Trigger{
		TriggerID: trigger.TriggerID,
		Category:  string(trigger.Category),
		MetadataPeriodic: MetadataPeriodic{
			IntervalSecond: trigger.MetadataPeriodic.IntervalSecond,
		},
		State:     string(trigger.State),
		CreatedAt: trigger.CreatedAt,
		UpdatedAt: trigger.UpdatedAt,
	}

	if err := h.dao.upsert(ctx, data); err != nil {
		return err
	}
	return nil
}

// FindAll ...
func (h handler) FindAll(ctx context.Context) ([]*trigengine.Trigger, error) {
	triggers, err := h.dao.find(ctx, bson.D{{Key: "basic.is_deleted", Value: false}})
	if err != nil {
		return nil, err
	}

	data := make([]*trigengine.Trigger, len(triggers))
	for idx, trigger := range triggers {
		data[idx] = &trigengine.Trigger{
			TriggerID:        trigger.TriggerID,
			Category:         trigengine.Category(trigger.Category),
			MetadataOnce:     trigengine.MetadataOnce(trigger.MetadataOnce),
			MetadataPeriodic: trigengine.MetadataPeriodic(trigger.MetadataPeriodic),
			MetadataOrdered:  trigengine.MetadataOrdered(trigger.MetadataOrdered),
			State:            trigengine.State(trigger.State),
			CreatedAt:        trigger.CreatedAt,
			UpdatedAt:        trigger.UpdatedAt,
		}
	}

	return data, nil
}

// Find ...
func (h handler) Find(ctx context.Context, opts ...OptFn) ([]*trigengine.Trigger, error) {
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	triggers, err := h.dao.find(ctx, filter)
	if err != nil {
		return nil, err
	}

	data := make([]*trigengine.Trigger, len(triggers))
	for idx, trigger := range triggers {
		data[idx] = &trigengine.Trigger{
			TriggerID:        trigger.TriggerID,
			Category:         trigengine.Category(trigger.Category),
			MetadataOnce:     trigengine.MetadataOnce(trigger.MetadataOnce),
			MetadataPeriodic: trigengine.MetadataPeriodic(trigger.MetadataPeriodic),
			MetadataOrdered:  trigengine.MetadataOrdered(trigger.MetadataOrdered),
			State:            trigengine.State(trigger.State),
			CreatedAt:        trigger.CreatedAt,
			UpdatedAt:        trigger.UpdatedAt,
		}
	}

	return data, nil
}

// FindOne ...
func (h handler) FindOne(ctx context.Context, opts ...OptFn) (*trigengine.Trigger, error) {
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	triggers, err := h.dao.find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(triggers) == 0 {
		return nil, nil
	}

	trigger := triggers[0]

	data := &trigengine.Trigger{
		TriggerID:        trigger.TriggerID,
		Category:         trigengine.Category(trigger.Category),
		MetadataOnce:     trigengine.MetadataOnce(trigger.MetadataOnce),
		MetadataPeriodic: trigengine.MetadataPeriodic(trigger.MetadataPeriodic),
		MetadataOrdered:  trigengine.MetadataOrdered(trigger.MetadataOrdered),
		State:            trigengine.State(trigger.State),
		CreatedAt:        trigger.CreatedAt,
		UpdatedAt:        trigger.UpdatedAt,
	}

	return data, nil
}
