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

// Package process this package is used to store the need data for process.
package process

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(client *mongo.Database, tableName string) *dao {
	d := &dao{
		client:    client.Collection(tableName),
		tableName: tableName,
		IOrm:      nil,
	}
	d.IOrm = base.NewOrm[*Process, Process](d)

	return d
}

type dao struct {
	client    *mongo.Collection
	tableName string

	base.IOrm[*Process, Process]
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

func (d *dao) getProcessDistributionByHostID(nCtx contextx.IContext, filter bson.D, aggregateOptions ...*options.AggregateOptions) (
	[]processDistributionByHostID, error) {

	pipeline := mongo.Pipeline{}

	if filter != nil && len(filter) > 0 {
		pipeline = append(pipeline, bson.D{
			{"$match", filter},
		})
	}

	pipeline = append(pipeline,
		bson.D{
			{"$group", bson.D{
				{"_id", "$" + FieldKeyHostID},
				{"count", bson.D{{"$sum", 1}}},
			}},
		},
		bson.D{
			{"$sort", bson.D{{"_id", 1}}},
		},
	)

	cursor, err := d.client.Aggregate(nCtx, pipeline, aggregateOptions...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := cursor.Close(nCtx); closeErr != nil {
			logger.G.Sys().WithErr(closeErr).With("filter", filter).Error("failed to close cursor of process distribution by host id")
		}
	}()

	var results []processDistributionByHostID
	if err = cursor.All(nCtx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// processDistributionByHostID process aggregate result.
type processDistributionByHostID struct {
	HostID       int64 `bson:"_id"`
	ProcessCount int64 `bson:"count"`
}

func (d *dao) getProcessDistributionByPluginName(nCtx contextx.IContext, filter bson.D, aggregateOptions ...*options.AggregateOptions) (
	[]processDistributionByPluginName, error) {

	pipeline := mongo.Pipeline{}

	if filter != nil && len(filter) > 0 {
		pipeline = append(pipeline, bson.D{
			{"$match", filter},
		})
	}

	pipeline = append(pipeline,
		bson.D{
			{"$group", bson.D{
				{"_id", "$" + FieldKeyPluginName},
				{"count", bson.D{{"$sum", 1}}},
			}},
		},
		bson.D{
			{"$sort", bson.D{{"_id", 1}}},
		},
	)

	cursor, err := d.client.Aggregate(nCtx, pipeline, aggregateOptions...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := cursor.Close(nCtx); closeErr != nil {
			logger.G.Sys().WithErr(closeErr).With("filter", filter).Error("failed to close cursor of process distribution by plugin name")
		}
	}()

	var results []processDistributionByPluginName
	if err = cursor.All(nCtx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// processDistributionByPluginName process aggregate result.
type processDistributionByPluginName struct {
	PluginName   string `bson:"_id"`
	ProcessCount int64  `bson:"count"`
}
