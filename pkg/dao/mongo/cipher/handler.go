/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cipher

import (
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler asymmetric encryption handler interface.
type IHandler interface {
	// Get gets asymmetric encryption by name and key-type.
	Get(nCtx contextx.IContext, name string, keyType types.CipherKeyType) (*types.Cipher, error)

	// Exist checks if asymmetric encryption exists by name and key-type.
	Exist(nCtx contextx.IContext, name string, keyType types.CipherKeyType) (bool, error)

	// List lists asymmetric encryption by options.
	List(nCtx contextx.IContext, page types.Page, opts ...base.OptFn) ([]*types.Cipher, int64, error)

	// ListWithoutCount lists asymmetric encryption by options, without count.
	ListWithoutCount(nCtx contextx.IContext, page types.Page, opts ...base.OptFn) ([]*types.Cipher, error)

	// Count counts asymmetric encryption by options.
	Count(nCtx contextx.IContext, opts ...base.OptFn) (int64, error)

	// Create creates asymmetric encryption.
	Create(nCtx contextx.IContext, asymmetricEncryption ...*types.Cipher) error

	// Delete deletes asymmetric encryption by name and key-type.
	Delete(nCtx contextx.IContext, name string, keyType types.CipherKeyType) error
}

type handler struct {
	client *mongo.Database
	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

func (h *handler) tenantDao(tenantID string) *dao {
	if d, ok := h.daoMap.Load(tenantID); ok {
		return d.(*dao) // nolint: forcetypeassert
	}

	newDaoClient := newDao(tenantID, h.client)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).With("tenant-id", tenantID).Warn("failed to ensure cipher indexes")
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao) // nolint: forcetypeassert
}

// New create a new host handler.
func New(client *mongo.Database) IHandler {
	return &handler{
		client: client,
		daoMap: sync.Map{},
	}
}

// Get gets asymmetric encryption by name and key-type.
func (h *handler) Get(nCtx contextx.IContext, name string, keyType types.CipherKeyType) (*types.Cipher, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	filter = WithName(name)(filter)
	filter = WithKeyType(string(keyType))(filter)

	data, err := h.tenantDao(tenantID).Get(nCtx, filter)
	if err != nil {
		return nil, err
	}

	return convertCipherToType(data), nil
}

// Exist checks if asymmetric encryption exists by name and key-type.
func (h *handler) Exist(nCtx contextx.IContext, name string, keyType types.CipherKeyType) (bool, error) {
	if nCtx == nil {
		return false, base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return false, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	filter = WithName(name)(filter)
	filter = WithKeyType(string(keyType))(filter)

	return h.tenantDao(tenantID).Exist(nCtx, filter)
}

// List lists asymmetric encryption by options.
func (h *handler) List(nCtx contextx.IContext, page types.Page, opts ...base.OptFn) ([]*types.Cipher, int64, error) {
	if nCtx == nil {
		return nil, 0, base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.tenantDao(tenantID).Count(nCtx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := base.ParsePage(page)

	encryptions, err := h.tenantDao(tenantID).List(nCtx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}

	data := make([]*types.Cipher, len(encryptions))
	for idx, encryption := range encryptions {
		data[idx] = convertCipherToType(encryption)
	}

	return data, num, nil
}

// ListWithoutCount lists asymmetric encryption by options, without count.
func (h *handler) ListWithoutCount(nCtx contextx.IContext, page types.Page, opts ...base.OptFn) ([]*types.Cipher, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	findOpt := base.ParsePage(page)

	encryptions, err := h.tenantDao(tenantID).List(nCtx, filter, findOpt)
	if err != nil {
		return nil, err
	}

	data := make([]*types.Cipher, len(encryptions))
	for idx, encryption := range encryptions {
		data[idx] = convertCipherToType(encryption)
	}

	return data, nil
}

// Count counts asymmetric encryption by options.
func (h *handler) Count(nCtx contextx.IContext, opts ...base.OptFn) (int64, error) {
	if nCtx == nil {
		return 0, base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return 0, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.tenantDao(tenantID).Count(nCtx, filter)
}

// Create creates asymmetric encryption.
func (h *handler) Create(nCtx contextx.IContext, asymmetricEncryption ...*types.Cipher) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()

	data := make([]*Cipher, 0, len(asymmetricEncryption))
	for _, item := range asymmetricEncryption {
		data = append(data, convertCipherFromType(item))
	}

	return h.tenantDao(tenantID).CreateMany(nCtx, data)
}

// Delete deletes asymmetric encryption by name and key-type.
func (h *handler) Delete(nCtx contextx.IContext, name string, keyType types.CipherKeyType) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	filter = WithName(name)(filter)
	filter = WithKeyType(string(keyType))(filter)

	return h.tenantDao(tenantID).DeleteMany(nCtx, filter)
}

func convertCipherToType(data *Cipher) *types.Cipher {
	if data == nil {
		return nil
	}

	return &types.Cipher{
		Name:        data.Name,
		KeyType:     types.CipherKeyType(data.KeyType),
		Description: data.Description,
		PrivateKey:  data.PrivateKey,
		PublicKey:   data.PublicKey,
	}
}

func convertCipherFromType(data *types.Cipher) *Cipher {
	if data == nil {
		return nil
	}

	return &Cipher{
		Name:        data.Name,
		KeyType:     string(data.KeyType),
		Description: data.Description,
		PrivateKey:  data.PrivateKey,
		PublicKey:   data.PublicKey,
	}
}
