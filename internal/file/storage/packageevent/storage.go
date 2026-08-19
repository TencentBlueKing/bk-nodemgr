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

// Package packageevent provides the event storage.
package packageevent

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/internal/file/storage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	daoPackageEvent "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/packageevent"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// StorageName defines the storage name.
const StorageName = "packageevent"

// NewStorage creates a new release storage.
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
	err := basestorage.InitStorage(&s.Storage,
		basestorage.WithStartFunc(s.initDao),
		basestorage.WithCheckFunc(s.check))
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to new storage")

		return nil, err
	}

	return s, nil
}

var _ IStorage = &Storage{}

// Storage implements IStorage.
type Storage struct {
	basestorage.Storage

	daoPackageEvent daoPackageEvent.IHandler
}

func (s *Storage) initDao() error {
	s.daoPackageEvent = daoPackageEvent.New(s.Database)

	return nil
}

func (s *Storage) check() error {
	if s.daoPackageEvent == nil {
		return errors.New("dao package event is nil")
	}

	return nil
}

func (s *Storage) metric() *storage.MetricData {
	return storage.Metric(StorageName)
}

// CreateManyPackageEvent creates package events.
// nolint: nonamedreturns
func (s *Storage) CreateManyPackageEvent(nCtx contextx.IContext, events ...*types.PackageEvent) error {
	var err error

	// record metric.
	metric := s.metric().Start("create_many_package_event")
	defer metric.End(err)

	if err = s.daoPackageEvent.CreateMany(nCtx, events...); err != nil {
		return err
	}

	return nil
}
