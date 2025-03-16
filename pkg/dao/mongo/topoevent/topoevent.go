/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package topoevent provides topo-event dao operations.
package topoevent

import (
	"context"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(tenantID string, client *mongo.Database, logger logger.Logger) *dao {
	return &dao{client: client.Collection(TableName(tenantID)), logger: logger}
}

type dao struct {
	client *mongo.Collection
	logger logger.Logger
}

// nolint:contextcheck
// ensureIndexes ensures the required indexes for the collection.
func (d *dao) ensureIndexes() error {
	return nil
}

func (d *dao) count(ctx context.Context, filter bson.D) (int64, error) {
	num, err := d.client.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	if num < 0 {
		return 0, fmt.Errorf("count documents get unexpected result: %d", num)
	}

	return num, nil
}

func (d *dao) list(ctx context.Context, filter bson.D, findOpt *options.FindOptions) ([]*TopoEvent, error) {
	result, err := d.client.Find(ctx, filter, findOpt)
	if err != nil {
		return nil, err
	}

	events := make([]*TopoEvent, 0)
	for result.Next(ctx) {
		table := &TableTopoEvent{}
		if err := result.Decode(table); err != nil {
			d.logger.Warnf("failed to decode topoevent, err %v", err)

			continue
		}
		events = append(events, table.Data)
	}

	return events, nil
}

func (d *dao) createMany(ctx context.Context, events []*TopoEvent) error {
	now := time.Now()

	tables := make([]interface{}, 0)
	for _, event := range events {
		tables = append(tables, &TableTopoEvent{
			BasicInfo: base.BasicInfo{
				CreatedAt: now,
				IsDeleted: false,
			},
			Data: event,
		})
	}

	_, err := d.client.InsertMany(ctx, tables)
	if err != nil {
		return err
	}

	return nil
}
