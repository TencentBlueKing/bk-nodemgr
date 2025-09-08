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
// nolint: nonamedreturns
package trigger

import (
	"context"
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage"
	daoTrigger "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/trigger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
	"go.mongodb.org/mongo-driver/mongo"
)

// StorageName defines the storage name.
const StorageName = "trigger"

// NewStorage creates a new trigger storage.
func NewStorage(client *mongo.Client, database string, logger logger.ILogger) (IStorage, error) {
	if client == nil {
		return nil, errors.New("mongo client is nil")
	}

	s := &Storage{
		Storage: basestorage.Storage{
			Name:     StorageName,
			Database: client.Database(database),
			Logger:   logger,
		},
	}
	err := basestorage.InitStorage(&s.Storage,
		basestorage.WithStartFunc(s.initDao),
		basestorage.WithCheckFunc(s.check))
	if err != nil {
		s.Logger.Errorf("new storage failed, err: %v", err)
		return nil, err
	}

	return s, nil
}

// storage implements IStorage.
type Storage struct {
	basestorage.Storage

	// dao
	triggerDao daoTrigger.IHandler
}

func (s *Storage) initDao() error {
	s.triggerDao = daoTrigger.New(s.Database, s.Logger)

	return nil
}

func (s *Storage) check() error {
	if s.triggerDao == nil {
		return errors.New("trigger dao is nil")
	}

	return nil
}

// CreateTrigger creates a new trigger.
func (s *Storage) CreateTrigger(ctx context.Context, trig *trigger.Trigger) (err error) {
	// record metric.
	metric := s.metric().Start("create")
	defer metric.End(err)

	if trig == nil {
		return errors.New("trigger is nil")
	}

	if err = s.triggerDao.Create(ctx, trig); err != nil {
		s.Logger.Errorf("failed to create trigger. trigger-id: %s, err: %v", trig.TriggerID, err)
		return err
	}

	return nil
}

// UpdateTrigger updates a trigger.
func (s *Storage) UpdateTrigger(ctx context.Context, trig *trigger.Trigger) (err error) {
	// record metric.
	metric := s.metric().Start("update")
	defer metric.End(err)

	if trig == nil {
		return errors.New("trigger is nil")
	}

	if err = s.triggerDao.Update(ctx, trig); err != nil {
		s.Logger.Errorf("failed to update trigger. trigger-id: %s, err: %v", trig.TriggerID, err)
		return err
	}

	return nil
}

// UpdateTriggerState updates a trigger's state.
func (s *Storage) UpdateTriggerState(ctx context.Context, triggerID string, state trigger.State) (err error) {
	// record metric.
	metric := s.metric().Start("update_state")
	defer metric.End(err)

	if err = s.triggerDao.UpdateState(ctx, triggerID, state); err != nil {
		s.Logger.Errorf("failed to update trigger state. trigger-id: %s, err: %v", triggerID, err)
		return err
	}

	return nil
}

// GetTrigger gets a trigger by triggerID.
func (s *Storage) GetTrigger(ctx context.Context, triggerID string) (data *trigger.Trigger, err error) {
	// record metric.
	metric := s.metric().Start("get")
	defer metric.End(err)

	if data, err = s.triggerDao.Get(ctx, triggerID); err != nil {
		s.Logger.Errorf("failed to get trigger. trigger-id: %s, err: %v", triggerID, err)
		return nil, err
	}

	return data, nil
}

// ListAliveTrigger lists alive triggers by category.
func (s *Storage) ListAliveTrigger(ctx context.Context, category trigger.Category) (results []*trigger.Trigger, err error) {
	// record metric.
	metric := s.metric().Start("list_alive")
	defer metric.End(err)

	if results, _, err = s.triggerDao.List(ctx, types.UnlimitedPage(),
		daoTrigger.WithState(trigger.StateInit, trigger.StateRunning),
		daoTrigger.WithCategory(category)); err != nil {
		s.Logger.Errorf("failed to list alive triggers. category: %s, err: %v", category, err)
		return nil, err
	}

	return results, nil
}

// DeleteTriggers deletes triggers by given trigger IDs.
func (s *Storage) DeleteTriggers(ctx context.Context, triggerIDs ...string) (err error) {
	// record metric.
	metric := s.metric().Start("delete_many")
	defer metric.End(err)

	if len(triggerIDs) == 0 {
		return nil
	}

	if err = s.triggerDao.Delete(ctx, triggerIDs...); err != nil {
		return err
	}

	return nil
}

func (s *Storage) metric() *storage.MetricData {
	return storage.Metric(StorageName)
}
