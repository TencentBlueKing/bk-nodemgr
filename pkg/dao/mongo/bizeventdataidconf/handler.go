/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package bizeventdataidconf

import (
	"fmt"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler defines business event data-id config table operations.
type IHandler interface {
	// Get gets business event data-id config by business ID.
	Get(nCtx contextx.IContext, bkBizID int64) (*types.BizEventDataIDConf, error)

	// UpdateAgentBaseAlarmEventDataID updates agent base alarm event data-id by business ID.
	UpdateAgentBaseAlarmEventDataID(nCtx contextx.IContext, bkBizID, eventDataID int64) error
}

// Handler operates business event data-id config table.
type Handler struct {
	client *mongo.Database
	daoMap sync.Map
}

// New creates a new business event data-id config handler.
func New(client *mongo.Database) *Handler {
	return &Handler{
		client: client,
		daoMap: sync.Map{},
	}
}

func (h *Handler) tenantDao(tenantID string) *dao {
	if d, ok := h.daoMap.Load(tenantID); ok {
		return d.(*dao) // nolint: forcetypeassert
	}

	newDaoClient := newDao(h.client, tenantID)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).With("tenant-id", tenantID).Warn("failed to ensure biz event data-id indexes")
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	return d.(*dao) // nolint: forcetypeassert
}

// Get gets business event data-id config by business ID.
func (h *Handler) Get(nCtx contextx.IContext, bkBizID int64) (*types.BizEventDataIDConf, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()

	if bkBizID <= 0 {
		return nil, fmt.Errorf("bk_biz_id must be positive, got %d", bkBizID)
	}

	filter := base.AliveFilter()
	filter = WithBizID(bkBizID)(filter)

	data, err := h.tenantDao(tenantID).Get(nCtx, filter)
	if err != nil {
		return nil, err
	}

	return convertBizEventDataIDConfToTypes(data), nil
}

// UpdateAgentBaseAlarmEventDataID updates agent base alarm event data-id by business ID.
func (h *Handler) UpdateAgentBaseAlarmEventDataID(nCtx contextx.IContext, bkBizID, eventDataID int64) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()

	if bkBizID <= 0 {
		return fmt.Errorf("bk_biz_id must be positive, got %d", bkBizID)
	}

	if eventDataID <= 0 {
		return fmt.Errorf("agentBaseAlarmEventDataID must be positive, got %d", eventDataID)
	}

	if err := h.tenantDao(tenantID).upsertAgentBaseAlarmEventDataID(nCtx, bkBizID, eventDataID); err != nil {
		return err
	}

	return nil
}

func convertBizEventDataIDConfToTypes(data *BizEventDataIDConf) *types.BizEventDataIDConf {
	if data == nil {
		return nil
	}

	return &types.BizEventDataIDConf{
		BizID:                     data.BizID,
		AgentBaseAlarmEventDataID: data.AgentBaseAlarmEventDataID,
		TaskProcEventDataID:       data.TaskProcEventDataID,
	}
}
