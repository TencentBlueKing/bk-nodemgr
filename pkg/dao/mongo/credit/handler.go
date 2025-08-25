/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package credit

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler credit Handler interface.
type IHandler interface {
	// Get get credit by creditID.
	Get(ctx context.Context, creditID string) ([]byte, error)

	// Upsert a credit.
	Upsert(ctx context.Context, creditID string, creditData []byte, expireAt time.Time) error
}

// Handler credit Handler.
type Handler struct {
	client *mongo.Database
	logger logger.ILogger
	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

// nolint: forcetypeassert
func (h *Handler) tenantDao(tenantID string) *dao {
	if d, ok := h.daoMap.Load(tenantID); ok {
		return d.(*dao)
	}

	newDaoClient := newDao(tenantID, h.client, h.logger)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		h.logger.Warnf("failed to ensure credit indexes, err: %v", errors.Join(base.ErrEnsureIndexesFailed(), err))
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao)
}

// New create a new credit Handler.
func New(client *mongo.Database, logger logger.ILogger) IHandler {
	return &Handler{
		client: client,
		logger: logger,
		daoMap: sync.Map{},
	}
}

// Get get credit by creditID.
func (h *Handler) Get(ctx context.Context, creditID string) ([]byte, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	filter := base.AliveFilter()
	filter = WithCreditID(creditID)(filter)
	data, err := h.tenantDao(tenantID).Get(ctx, filter)
	if err != nil {
		return nil, err
	}

	return data.CreditData, nil
}

// Upsert a credit.
func (h *Handler) Upsert(ctx context.Context, creditID string, creditData []byte, expireAt time.Time) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

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
		TenantID:   tenantID,
		CreditID:   creditID,
		CreditData: creditData,
		ExpireAt:   expireAt,
	}

	return h.tenantDao(tenantID).upsert(ctx, credit)
}
