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

// Package pluginrequirement provides independent requirement file metadata storage.
package pluginrequirement

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/plugin-requirement"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// StorageName defines the storage name.
	StorageName = "pluginrequirement"

	metricOperationCreatePluginRequirement = "create_plugin_requirement"
	metricOperationUpdatePluginRequirement = "update_plugin_requirement"
	metricOperationGetPluginRequirement    = "get_plugin_requirement"
)

// NewStorage creates requirement file metadata storage.
func NewStorage(client *mongo.Client, database string) (*Storage, error) {
	if client == nil {
		return nil, errors.New("mongo client is nil")
	}

	s := &Storage{
		Storage: basestorage.Storage{
			Name:     StorageName,
			Database: client.Database(database),
		},
	}
	if err := basestorage.InitStorage(&s.Storage,
		basestorage.WithStartFunc(s.initDao),
		basestorage.WithCheckFunc(s.check)); err != nil {
		return nil, fmt.Errorf("failed to initialize requirement storage: %w", err)
	}

	return s, nil
}

var _ IStorage = &Storage{}

// Storage implements IStorage.
type Storage struct {
	basestorage.Storage

	daoPluginRequirement pluginrequirement.IHandler
}

func (s *Storage) initDao() error {
	s.daoPluginRequirement = pluginrequirement.New(s.Database)

	return nil
}

func (s *Storage) check() error {
	if s.daoPluginRequirement == nil {
		return errors.New("dao plugin requirement is nil")
	}

	return nil
}

// CreatePluginRequirement creates an independent requirement file record.
func (s *Storage) CreatePluginRequirement(nCtx contextx.IContext, record *types.PluginRequirementFileRecord) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	return s.WrapFn(nCtx, metricOperationCreatePluginRequirement, func(ctx contextx.IContext) error {
		return s.createPluginRequirement(ctx, record)
	})
}

// UpdatePluginRequirement updates an existing record and never inserts one.
func (s *Storage) UpdatePluginRequirement(nCtx contextx.IContext, record *types.PluginRequirementFileRecord) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	return s.WrapFn(nCtx, metricOperationUpdatePluginRequirement, func(ctx contextx.IContext) error {
		return s.updatePluginRequirement(ctx, record)
	})
}

// GetPluginRequirement retrieves a record by ID within the current tenant.
func (s *Storage) GetPluginRequirement(nCtx contextx.IContext, id string) (*types.PluginRequirementFileRecord, error) {
	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	var record *types.PluginRequirementFileRecord
	err := s.WrapFn(nCtx, metricOperationGetPluginRequirement, func(ctx contextx.IContext) error {
		var err error
		record, err = s.getPluginRequirement(ctx, id)

		return err
	})

	return record, err
}
