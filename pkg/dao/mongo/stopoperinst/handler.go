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
	"context"
	"errors"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Handler provide stopping operation instance handler.
type Handler interface {
	// Upsert updates or inserts a stopping operation instance.
	Upsert(ctx context.Context, operInstID string) error

	// FindAll find all stopping operation instances.
	FindAll(ctx context.Context) ([]string, error)

	// WatchInsert watch upsert event
	WatchInsert(fn func(string))
}

type handler struct {
	dao    *dao
	logger logger.ILogger
}

// New create a new host handler.
func New(client *mongo.Database, logger logger.ILogger) Handler {
	return &handler{
		dao:    newDao(client, logger),
		logger: logger,
	}
}

// Upsert ...
func (h *handler) Upsert(ctx context.Context, operInstID string) error {
	if ctx == nil {
		return errors.New("empty context")
	}

	if operInstID == "" {
		return errors.New("empty oper_inst_id")
	}

	data := &StopOperInst{
		OperInstID: operInstID,
		ExpireAt:   time.Now().Add(time.Second * 5),
	}
	if err := h.dao.upsert(ctx, data); err != nil {
		return err
	}

	return nil
}

// FindAll ...
func (h *handler) FindAll(ctx context.Context) ([]string, error) {
	stopOperInsts, err := h.dao.find(ctx, bson.D{{Key: "basic.is_deleted", Value: false}})
	if err != nil {
		return nil, err
	}

	data := make([]string, len(stopOperInsts))
	for idx, stopOperInst := range stopOperInsts {
		data[idx] = stopOperInst.OperInstID
	}

	return data, nil
}

// WatchInsert ...
func (h *handler) WatchInsert(fn func(string)) {
	h.dao.watchWithRetry(context.Background(),
		bson.D{{Key: "operationType", Value: "insert"}},
		func(inst *StopOperInst) { fn(inst.OperInstID) })
}
