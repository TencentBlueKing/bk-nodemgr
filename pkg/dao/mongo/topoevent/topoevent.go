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

// Package topoevent provides topo-event dao operations.
package topoevent

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(tenantID string, client *mongo.Database) *dao {
	tableName := TableName(tenantID)
	d := &dao{
		client:    client.Collection(tableName),
		tableName: tableName,
	}

	d.IOrm = base.NewOrm[*TopoEvent, TopoEvent](d)

	return d
}

type dao struct {
	client    *mongo.Collection
	tableName string
	base.IOrm[*TopoEvent, TopoEvent]
}

// GetClient get the dao's client.
func (d *dao) GetClient() *mongo.Collection {
	return d.client
}

// GetTableName get the dao's table name.
func (d *dao) GetTableName() string {
	return d.tableName
}

// GetIndexes get the dao's indexes.
func (d *dao) GetIndexes() []mongo.IndexModel {
	var indexes []mongo.IndexModel

	return indexes
}

// nolint:contextcheck
// ensureIndexes ensures the required indexes for the collection.
func (d *dao) ensureIndexes() error {
	return nil
}

func (d *dao) count(nCtx contextx.IContext, filter bson.D) (int64, error) {
	num, err := d.client.CountDocuments(nCtx, filter)
	if err != nil {
		return 0, err
	}

	if num < 0 {
		return 0, fmt.Errorf("count documents get unexpected result: %d", num)
	}

	return num, nil
}

func (d *dao) list(nCtx contextx.IContext, filter bson.D, findOpt *options.FindOptions) ([]*TopoEvent, error) {
	result, err := d.client.Find(nCtx, filter, findOpt)
	if err != nil {
		return nil, err
	}

	events := make([]*TopoEvent, 0)
	for result.Next(nCtx) {
		table := &TableTopoEvent{}
		if err := result.Decode(table); err != nil {
			logger.G.Sys().WithErr(err).Warn("failed to decode topoevent")

			continue
		}
		events = append(events, table.Data)
	}

	return events, nil
}

func (d *dao) createMany(nCtx contextx.IContext, events []*TopoEvent) error {
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

	_, err := d.client.InsertMany(nCtx, tables)
	if err != nil {
		return err
	}

	return nil
}

// distinctString distinct string field.
func (d *dao) distinctString(
	nCtx contextx.IContext, key string, filter bson.D, distinctOpt *options.DistinctOptions) ([]string, error) {

	return d.DistinctString(nCtx, key, filter, distinctOpt)
}

// distinctInt64 distinct int64 field.
func (d *dao) distinctInt64(
	nCtx contextx.IContext, key string, filter bson.D, distinctOpt *options.DistinctOptions) ([]int64, error) {

	return d.DistinctInt64(nCtx, key, filter, distinctOpt)
}
