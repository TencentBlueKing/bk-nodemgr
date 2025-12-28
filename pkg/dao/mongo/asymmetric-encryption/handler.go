/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package asymmetricencryption

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler asymmetric encryption handler interface.
type IHandler interface {
	// Get gets asymmetric encryption by key-type and cipher-type.
	Get(nCtx contextx.IContext, keyType types.AsymmetricKeyType, cipherType types.AsymmetricCipherType) (*types.AsymmetricEncryption, error)

	// Exist checks if asymmetric encryption exists by key-type and cipher-type.
	Exist(nCtx contextx.IContext, keyType types.AsymmetricKeyType, cipherType types.AsymmetricCipherType) (bool, error)

	// Create creates asymmetric encryption.
	Create(nCtx contextx.IContext, asymmetricEncryption ...*types.AsymmetricEncryption) error

	// Upsert updates or inserts asymmetric encryption.
	Upsert(nCtx contextx.IContext, asymmetricEncryption *types.AsymmetricEncryption) error

	// Delete deletes asymmetric encryption by key-type and cipher-type.
	Delete(nCtx contextx.IContext, keyType types.AsymmetricKeyType, cipherType types.AsymmetricCipherType) error
}

// Handler this is a Handler to manage asymmetric encryption related db operations.
type Handler struct {
	dao *dao
}

// New create a new asymmetric encryption handler instance.
func New(client *mongo.Database) *Handler {
	h := &Handler{
		dao: newDao(client),
	}

	if err := h.dao.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).Warn("failed to ensure asymmetric encryption indexes")
	}

	return h
}

// Get gets asymmetric encryption by key-type and cipher-type.
func (h *Handler) Get(nCtx contextx.IContext, keyType types.AsymmetricKeyType, cipherType types.AsymmetricCipherType) (
	*types.AsymmetricEncryption, error) {

	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	filter := base.AliveFilter()
	filter = WithKeyType(string(keyType))(filter)
	filter = WithCipherType(string(cipherType))(filter)

	data, err := h.dao.Get(nCtx, filter)
	if err != nil {
		return nil, err
	}

	return convertAsymmetricEncryptionToType(data), nil
}

// Exist checks if asymmetric encryption exists by key-type and cipher-type.
func (h *Handler) Exist(nCtx contextx.IContext, keyType types.AsymmetricKeyType, cipherType types.AsymmetricCipherType) (bool, error) {
	if nCtx == nil {
		return false, base.ErrInvalidContext()
	}

	filter := base.AliveFilter()
	filter = WithKeyType(string(keyType))(filter)
	filter = WithCipherType(string(cipherType))(filter)

	return h.dao.Exist(nCtx, filter)
}

// Create creates asymmetric encryption.
func (h *Handler) Create(nCtx contextx.IContext, asymmetricEncryption ...*types.AsymmetricEncryption) error {
	if nCtx == nil {
		return errors.New("nCtx is nil")
	}

	data := make([]*AsymmetricEncryption, 0, len(asymmetricEncryption))
	for _, item := range asymmetricEncryption {
		data = append(data, convertAsymmetricEncryptionFromType(item))
	}

	return h.dao.CreateMany(nCtx, data)
}

// Upsert updates or inserts asymmetric encryption.
func (h *Handler) Upsert(nCtx contextx.IContext, asymmetricEncryption *types.AsymmetricEncryption) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	data := convertAsymmetricEncryptionFromType(asymmetricEncryption)

	return h.dao.upsert(nCtx, data)
}

// Delete deletes asymmetric encryption by key-type and cipher-type.
func (h *Handler) Delete(nCtx contextx.IContext, keyType types.AsymmetricKeyType, cipherType types.AsymmetricCipherType) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	filter := base.AliveFilter()
	filter = WithKeyType(string(keyType))(filter)
	filter = WithCipherType(string(cipherType))(filter)

	return h.dao.DeleteMany(nCtx, filter)
}

func convertAsymmetricEncryptionToType(data *AsymmetricEncryption) *types.AsymmetricEncryption {
	if data == nil {
		return nil
	}

	return &types.AsymmetricEncryption{
		CipherType:  types.AsymmetricCipherType(data.CipherType),
		KeyType:     types.AsymmetricKeyType(data.KeyType),
		Description: data.Description,
		Content:     data.Content,
	}
}

func convertAsymmetricEncryptionFromType(data *types.AsymmetricEncryption) *AsymmetricEncryption {
	if data == nil {
		return nil
	}

	return &AsymmetricEncryption{
		CipherType:  string(data.CipherType),
		KeyType:     string(data.KeyType),
		Description: data.Description,
		Content:     data.Content,
	}
}
