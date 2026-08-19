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

package credit

import (
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler credit Handler interface.
type IHandler interface {
	// CheckValid check creditIDs is valid or not.
	CheckValid(nCtx contextx.IContext, creditIDs ...string) (map[string]bool, error)

	// Get get credit by creditID.
	Get(nCtx contextx.IContext, creditID string) ([]byte, error)

	// Upsert a credit.
	Upsert(nCtx contextx.IContext, creditID string, creditData []byte, expireAt time.Time) error
}

// Handler credit Handler.
type Handler struct {
	client *mongo.Database

	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

// nolint: forcetypeassert
func (h *Handler) tenantDao(tenantID string) *dao {
	if d, ok := h.daoMap.Load(tenantID); ok {
		return d.(*dao)
	}

	newDaoClient := newDao(tenantID, h.client)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).Warn("failed to ensure credit indexes")
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao)
}

// New create a new credit Handler.
func New(client *mongo.Database) IHandler {
	return &Handler{
		client: client,
		daoMap: sync.Map{},
	}
}

// CheckValid check creditIDs is valid or not.
func (h *Handler) CheckValid(nCtx contextx.IContext, creditIDs ...string) (map[string]bool, error) {
	filter := base.AliveFilter()
	filter = WithCreditID(creditIDs...)(filter)

	data, err := h.tenantDao(nCtx.TenantID()).List(nCtx, filter, nil)
	if err != nil {
		return nil, err
	}

	validMap := make(map[string]bool)
	for _, id := range creditIDs {
		validMap[id] = false
	}
	for _, item := range data {
		validMap[item.CreditID] = true
	}

	return validMap, nil
}

// Get get credit by creditID.
func (h *Handler) Get(nCtx contextx.IContext, creditID string) ([]byte, error) {
	filter := base.AliveFilter()
	filter = WithCreditID(creditID)(filter)

	data, err := h.tenantDao(nCtx.TenantID()).Get(nCtx, filter)
	if err != nil {
		return nil, err
	}

	return data.CreditData, nil
}

// Upsert a credit.
func (h *Handler) Upsert(nCtx contextx.IContext, creditID string, creditData []byte, expireAt time.Time) error {
	if creditID == "" {
		return ErrInvalidCreditID()
	}

	if expireAt.Before(time.Now()) {
		return ErrInvalidExpireAt()
	}

	if len(creditData) == 0 {
		return ErrEmptyCreditData()
	}

	credit := &Credit{
		TenantID:   nCtx.TenantID(),
		CreditID:   creditID,
		CreditData: creditData,
		ExpireAt:   expireAt,
	}

	return h.tenantDao(nCtx.TenantID()).upsert(nCtx, credit)
}
