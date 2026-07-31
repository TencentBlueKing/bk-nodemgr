/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package stopoperinst ...
package stopoperinst

import (
	"errors"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler provides stopping operation instance operations.
type IHandler interface {
	// Upsert updates or inserts a stopping operation instance.
	Upsert(nCtx contextx.IContext, operInstID string) error

	// FindAll find all stopping operation instances.
	FindAll(nCtx contextx.IContext) ([]string, error)

	// FindByIDs finds stopping operation instances by operation instance IDs.
	FindByIDs(nCtx contextx.IContext, operInstIDs ...string) ([]string, error)
}

const stopOperInstTTL = 30 * time.Second

type handler struct {
	dao *dao
}

// New create a new host handler.
func New(client *mongo.Database) IHandler {
	h := &handler{
		dao: newDao(client),
	}
	if err := h.dao.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).Warn("failed to ensure stopping operation instance indexes")
	}

	return h
}

// Upsert ...
func (h *handler) Upsert(nCtx contextx.IContext, operInstID string) error {
	if nCtx == nil {
		return errors.New("empty context")
	}

	if operInstID == "" {
		return errors.New("empty oper_inst_id")
	}

	data := &StopOperInst{
		OperInstID: operInstID,
		ExpireAt:   time.Now().Add(stopOperInstTTL),
	}
	if err := h.dao.upsert(nCtx, data); err != nil {
		return err
	}

	return nil
}

// FindAll ...
func (h *handler) FindAll(nCtx contextx.IContext) ([]string, error) {
	stopOperInsts, err := h.dao.List(nCtx, base.AliveFilter(), nil)
	if err != nil {
		return nil, err
	}

	return conv.SliceToSlice(stopOperInsts, convertStopOperInstToID), nil
}

// FindByIDs ...
func (h *handler) FindByIDs(nCtx contextx.IContext, operInstIDs ...string) ([]string, error) {
	if len(operInstIDs) == 0 {
		return nil, nil
	}

	filter := base.AliveFilter()
	filter = base.WithValues(FieldKeyOperInstID, operInstIDs...)(filter)
	stopOperInsts, err := h.dao.List(nCtx, filter, nil)
	if err != nil {
		return nil, err
	}

	return conv.SliceToSlice(stopOperInsts, convertStopOperInstToID), nil
}

func convertStopOperInstToID(stopOperInst *StopOperInst) string {
	return stopOperInst.OperInstID
}
