/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package cipher defines the interface for cipher storage handlers.
package cipher

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	daoCipher "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/cipher"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// constants ...
const (
	// StorageName defines the storage name.
	StorageName = "cipher"

	metricOperationGetCipher    = "get_cipher"
	metricOperationCreateCipher = "create_cipher"
	metricOperationExistCipher  = "exist_cipher"
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

	daoCipher daoCipher.IHandler
}

func (s *Storage) initDao() error {
	s.daoCipher = daoCipher.New(s.Database)

	return nil
}

func (s *Storage) check() error {
	if s.daoCipher == nil {
		return errors.New("asymmetricencryption dao is nil")
	}

	return nil
}

// GetCipher gets cipher by name and key-type.
func (s *Storage) GetCipher(nCtx contextx.IContext, name string, keyType types.CipherKeyType) (*types.Cipher, error) {
	var ae *types.Cipher

	err := s.WrapFn(nCtx, metricOperationGetCipher, func(contextx.IContext) error {
		var err error
		ae, err = s.getCipher(nCtx, name, keyType)

		return err
	})

	return ae, err
}

// ExistCipher checks if cipher exists by name and key-type.
func (s *Storage) ExistCipher(nCtx contextx.IContext, name string, keyType types.CipherKeyType) (
	bool, error) {

	var exist bool

	err := s.WrapFn(nCtx, metricOperationExistCipher, func(contextx.IContext) error {
		var err error
		exist, err = s.existCipher(nCtx, name, keyType)

		return err
	})

	return exist, err
}

// CreateCipher creates cipher.
func (s *Storage) CreateCipher(nCtx contextx.IContext, encryption ...*types.Cipher) error {
	return s.WrapFn(nCtx, metricOperationCreateCipher, func(contextx.IContext) error {
		var err error
		err = s.createCipher(nCtx, encryption...)

		return err
	})
}
