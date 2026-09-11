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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	daoPackageEvent "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/packageevent"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// StorageName defines the storage name.
	StorageName = "packageevent"

	metricOperationCountPackageEvent      = "count_package_event"
	metricOperationListPackageEvent       = "list_package_event"
	metricOperationCreateManyPackageEvent = "create_many_package_event"
	metricOperationDistinctPackageEvent   = "distinct_package_event"
)

// NewStorage creates a new package event storage.
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

// CountPackageEvent counts package events by conditions.
func (s *Storage) CountPackageEvent(nCtx contextx.IContext, conditions ...*types.PackageEventCondition) (int64, error) {
	if nCtx == nil {
		return 0, base.ErrInvalidContext()
	}

	var count int64
	err := s.WrapFn(nCtx, metricOperationCountPackageEvent, func(nCtx contextx.IContext) error {
		var err error
		count, err = s.countPackageEvent(nCtx, conditions...)

		return err
	})

	return count, err
}

// ListPackageEvent lists package events by page and conditions.
func (s *Storage) ListPackageEvent(nCtx contextx.IContext, page types.Page,
	conditions ...*types.PackageEventCondition) ([]*types.PackageEvent, int64, error) {

	if nCtx == nil {
		return nil, 0, base.ErrInvalidContext()
	}

	var events []*types.PackageEvent
	var count int64
	err := s.WrapFn(nCtx, metricOperationListPackageEvent, func(nCtx contextx.IContext) error {
		var err error
		events, count, err = s.listPackageEvent(nCtx, page, conditions...)

		return err
	})

	return events, count, err
}

// CreateManyPackageEvent creates package events.
func (s *Storage) CreateManyPackageEvent(nCtx contextx.IContext, events ...*types.PackageEvent) error {
	return s.WrapFn(nCtx, metricOperationCreateManyPackageEvent, func(nCtx contextx.IContext) error {
		return s.createManyPackageEvent(nCtx, events...)
	})
}

// DistinctPackageEvent distincts package event fields.
func (s *Storage) DistinctPackageEvent(nCtx contextx.IContext, request types.PackageEventDistinctRequest,
	conditions ...*types.PackageEventCondition) (*types.PackageEventDistinctResult, error) {

	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	var result *types.PackageEventDistinctResult
	err := s.WrapFn(nCtx, metricOperationDistinctPackageEvent, func(nCtx contextx.IContext) error {
		var err error
		result, err = s.distinctPackageEvent(nCtx, request, conditions...)

		return err
	})

	return result, err
}
