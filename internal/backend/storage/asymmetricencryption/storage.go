/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package asymmetricencryption defines the interface for asymmetric encryption storage handlers.
package asymmetricencryption

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	daoAsymmetricEncryption "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/asymmetric-encryption"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// constants ...
const (
	// StorageName defines the storage name.
	StorageName = "asymmetricencryption"

	metricOperationGetAsymmetricEncryption    = "get_asymmetric_encryption"
	metricOperationCreateAsymmetricEncryption = "create_asymmetric_encryption"
	metricOperationExistAsymmetricEncryption  = "exist_asymmetric_encryption"
	metricOperationDeleteAsymmetricEncryption = "delete_asymmetric_encryption"
)

// NewStorage creates a new workflow storage.
func NewStorage(client *mongo.Client, database string) (IStorage, error) {
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

// Storage implements IStorage.
type Storage struct {
	basestorage.Storage

	daoAsymmetricEncryption daoAsymmetricEncryption.IHandler
}

func (s *Storage) initDao() error {
	s.daoAsymmetricEncryption = daoAsymmetricEncryption.New(s.Database)
	return nil
}

func (s *Storage) check() error {
	if s.daoAsymmetricEncryption == nil {
		return errors.New("asymmetricencryption dao is nil")
	}

	return nil
}

// GetAsymmetricEncryption gets asymmetric encryption.
func (s *Storage) GetAsymmetricEncryption(nCtx contextx.IContext, keyType types.AsymmetricKeyType, cipherType types.AsymmetricCipherType) (
	*types.AsymmetricEncryption, error) {

	var (
		ae  *types.AsymmetricEncryption
		err error
	)

	err = s.WrapFn(nCtx, metricOperationGetAsymmetricEncryption, func(contextx.IContext) error {
		ae, err = s.getAsymmetricEncryption(nCtx, keyType, cipherType)
		if err != nil {
			return err
		}

		return nil
	})

	return ae, err
}

// CreateAsymmetricEncryption creates asymmetric encryption.
func (s *Storage) CreateAsymmetricEncryption(nCtx contextx.IContext, encryption ...*types.AsymmetricEncryption) error {
	var err error

	err = s.WrapFn(nCtx, metricOperationCreateAsymmetricEncryption, func(contextx.IContext) error {
		err = s.createAsymmetricEncryption(nCtx, encryption...)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

// UpsertAsymmetricEncryption updates or inserts asymmetric encryption.
func (s *Storage) UpsertAsymmetricEncryption(nCtx contextx.IContext, encryption *types.AsymmetricEncryption) error {
	var err error

	err = s.WrapFn(nCtx, metricOperationCreateAsymmetricEncryption, func(contextx.IContext) error {
		err = s.upsertAsymmetricEncryption(nCtx, encryption)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

// ExistAsymmetricEncryption checks if asymmetric encryption exists.
func (s *Storage) ExistAsymmetricEncryption(nCtx contextx.IContext, keyType types.AsymmetricKeyType, cipherType types.AsymmetricCipherType) (
	bool, error) {

	var (
		exist bool
		err   error
	)

	err = s.WrapFn(nCtx, metricOperationExistAsymmetricEncryption, func(contextx.IContext) error {
		exist, err = s.existAsymmetricEncryption(nCtx, keyType, cipherType)
		if err != nil {
			return err
		}

		return nil
	})

	return exist, err
}

// DeleteAsymmetricEncryption deletes asymmetric encryption.
func (s *Storage) DeleteAsymmetricEncryption(nCtx contextx.IContext, keyType types.AsymmetricKeyType, cipherType types.AsymmetricCipherType) error {
	var err error

	err = s.WrapFn(nCtx, metricOperationDeleteAsymmetricEncryption, func(contextx.IContext) error {
		err = s.deleteAsymmetricEncryption(nCtx, keyType, cipherType)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}
