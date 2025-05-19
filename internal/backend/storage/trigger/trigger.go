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

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/base"
	daoTrigger "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/trigger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
	"go.mongodb.org/mongo-driver/mongo"
)

// StorageName defines the storage name.
const StorageName = "trigger"

// NewStorage creates a new trigger storage.
func NewStorage(client *mongo.Client, database string, logger logger.Logger) (IStorage, error) {
	if client == nil {
		return nil, errors.New("mongo client is nil")
	}

	s := &storage{
		Storage: base.Storage{
			Name:     StorageName,
			Database: client.Database(database),
			Logger:   logger,
		},
	}
	err := base.InitStorage(&s.Storage,
		base.WithStartFunc(s.initDao),
		base.WithCheckFunc(s.check))
	if err != nil {
		s.Logger.Errorf("new storage failed, err: %v", err)
		return nil, err
	}

	return s, nil
}

type storage struct {
	base.Storage

	// dao
	triggerDao daoTrigger.IHandler
}

func (s *storage) initDao() error {
	s.triggerDao = daoTrigger.New(s.Database, s.Logger)

	return nil
}

func (s *storage) check() error {
	if s.triggerDao == nil {
		return errors.New("trigger dao is nil")
	}

	return nil
}

// CreateTrigger creates a new trigger.
func (s *storage) CreateTrigger(ctx context.Context, trig *trigger.Trigger) error {
	if trig == nil {
		return errors.New("trigger is nil")
	}

	if err := s.triggerDao.Create(ctx, trig); err != nil {
		s.Logger.Errorf("failed to create trigger. trigger-id: %s, err: %v", trig.TriggerID, err)
		return err
	}

	return nil
}

// UpdateTrigger updates a trigger.
func (s *storage) UpdateTrigger(ctx context.Context, trig *trigger.Trigger) error {
	if trig == nil {
		return errors.New("trigger is nil")
	}

	if err := s.triggerDao.Update(ctx, trig); err != nil {
		s.Logger.Errorf("failed to update trigger. trigger-id: %s, err: %v", trig.TriggerID, err)
		return err
	}

	return nil
}

// GetTrigger gets a trigger by triggerID.
func (s *storage) GetTrigger(ctx context.Context, triggerID string) (*trigger.Trigger, error) {
	data, err := s.triggerDao.Get(ctx, triggerID)
	if err != nil {
		s.Logger.Errorf("failed to get trigger. trigger-id: %s, err: %v", triggerID, err)
		return nil, err
	}

	return data, nil
}

// ListAliveTrigger lists alive triggers by category.
func (s *storage) ListAliveTrigger(ctx context.Context, category trigger.Category) ([]*trigger.Trigger, error) {
	trigs, _, err := s.triggerDao.List(ctx, types.UnlimitedPage(),
		daoTrigger.WithState(trigger.StateInit, trigger.StateRunning),
		daoTrigger.WithCategory(category))
	if err != nil {
		s.Logger.Errorf("failed to list alive triggers. category: %s, err: %v", category, err)
		return nil, err
	}

	return trigs, nil
}
