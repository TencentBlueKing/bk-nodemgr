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

package packageexport

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/package-export"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// StorageName defines the storage name.
	StorageName = "packageexport"

	metricOperationListPackageExport   = "list_package_export"
	metricOperationCreatePackageExport = "create_package_export"
	metricOperationGetPackageExport    = "get_package_export"
	metricOperationDeletePackageExport = "delete_package_export"
)

// NewStorage creates a package export storage.
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
		logger.G.Sys().WithErr(err).Error("failed to new storage")

		return nil, err
	}

	return s, nil
}

var _ IStorage = &Storage{}

// Storage implements IStorage.
type Storage struct {
	basestorage.Storage

	daoPackageExport packageexport.IHandler
}

func (s *Storage) initDao() error {
	s.daoPackageExport = packageexport.New(s.Database)

	return nil
}

func (s *Storage) check() error {
	if s.daoPackageExport == nil {
		return errors.New("dao package export is nil")
	}

	return nil
}

// ListPackageExport lists package export records.
func (s *Storage) ListPackageExport(
	nCtx contextx.IContext, page types.Page, conditions ...*types.PackageExportCondition) (
	[]*types.PackageExport, int64, error) {

	var exports []*types.PackageExport
	var count int64
	err := s.WrapFn(nCtx, metricOperationListPackageExport, func(nCtx contextx.IContext) error {
		var err error
		exports, count, err = s.listPackageExport(nCtx, page, conditions...)

		return err
	})

	return exports, count, err
}

// CreatePackageExport creates a package export record.
func (s *Storage) CreatePackageExport(nCtx contextx.IContext, exportData *types.PackageExport) error {
	return s.WrapFn(nCtx, metricOperationCreatePackageExport, func(nCtx contextx.IContext) error {
		return s.createPackageExport(nCtx, exportData)
	})
}

// GetPackageExport gets a package export by export ID.
func (s *Storage) GetPackageExport(nCtx contextx.IContext, exportID string) (*types.PackageExport, error) {
	var exportData *types.PackageExport
	err := s.WrapFn(nCtx, metricOperationGetPackageExport, func(nCtx contextx.IContext) error {
		var err error
		exportData, err = s.getPackageExport(nCtx, exportID)

		return err
	})

	return exportData, err
}

// DeletePackageExport deletes a package export by export ID.
func (s *Storage) DeletePackageExport(nCtx contextx.IContext, exportID string) error {
	return s.WrapFn(nCtx, metricOperationDeletePackageExport, func(nCtx contextx.IContext) error {
		return s.deletePackageExport(nCtx, exportID)
	})
}
