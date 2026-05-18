/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package networkunit provides network-unit dao operations.
package networkunit

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/counter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(client *mongo.Database) *dao {
	tableName := TableName()
	d := &dao{
		client:    client.Collection(tableName),
		tableName: tableName,
		counter:   counter.New(client),
	}

	d.IOrm = base.NewOrm[*NetworkUnit, NetworkUnit](d)

	return d
}

type dao struct {
	client    *mongo.Collection
	tableName string
	counter   counter.Handler

	base.IOrm[*NetworkUnit, NetworkUnit]
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
	return []mongo.IndexModel{
		{
			Keys: bson.D{{Key: FieldKeyNetworkUnitID, Value: 1}},
			Options: options.Index().SetPartialFilterExpression(bson.D{
				{Key: base.FieldKeyIsDeleted, Value: false},
			}),
		},
	}
}

func (d *dao) create(nCtx contextx.IContext, networkUnit *NetworkUnit) (int64, error) {
	newSequence, err := d.counter.Generate(nCtx, TableName())
	if err != nil {
		return 0, err
	}

	networkUnit.NetworkUnitID = newSequence
	table := &TableNetworkUnit{
		BasicInfo: base.BasicInfo{
			CreatedAt: time.Now(),
			IsDeleted: false,
		},
		Data: networkUnit,
	}

	if _, err = d.client.InsertOne(nCtx, table); err != nil {
		return 0, err
	}

	return newSequence, nil
}

func (d *dao) deleteMany(nCtx contextx.IContext, tenantID string, networkUnitIDs ...int64) error {
	models := buildDeleteManyParams(tenantID, networkUnitIDs...)

	result, err := d.client.BulkWrite(nCtx, models)
	if err != nil {
		return err
	}

	if result.MatchedCount > 0 {
		logger.G.Sys().With("deleted-count", result.MatchedCount).Info("deleted networkunits")
	}

	return nil
}

// tenantFilter additional tenant filter.
// global networkarea is a special networkarea, it belongs to system tenant, but it can be seen by all tenants.
// this scene is also ensured in CMDB.
// a query from a tenant, should be filtered in its own tenant, and plus the global networkarea.
func tenantFilter(tenantID string) bson.E {
	return bson.E{
		Key: "$or",
		Value: bson.A{
			bson.D{{Key: FieldKeyTenantID, Value: tenantID}},
			bson.D{{Key: FieldKeyNetworkAreaID, Value: base.GlobalNetworkAreaID}},
		}}
}

// buildDeleteManyParams build delete many params.
func buildDeleteManyParams(tenantID string, networkUnitIDs ...int64) []mongo.WriteModel {
	filter := bson.D{
		bson.E{Key: FieldKeyNetworkUnitID, Value: bson.D{{Key: "$in", Value: networkUnitIDs}}},
		bson.E{Key: FieldKeyTenantID, Value: tenantID}}

	update := base.BuildDeleteParam()

	return []mongo.WriteModel{mongo.NewUpdateManyModel().SetFilter(filter).SetUpdate(update).SetUpsert(false)}
}

// getNetworkUnitDistributionByNetworkAreaID get networkunit distribution by network area id.
func (d *dao) getNetworkUnitDistributionByNetworkAreaID(nCtx contextx.IContext, filter bson.D, aggregateOptions ...*options.AggregateOptions) (
	[]networkUnitDistributionByNetworkAreaID, error) {

	pipeline := mongo.Pipeline{}

	if len(filter) > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: filter}})
	}

	pipeline = append(pipeline,
		bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$" + FieldKeyNetworkAreaID}, {Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}}}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
	)

	cursor, err := d.client.Aggregate(nCtx, pipeline, aggregateOptions...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := cursor.Close(nCtx); closeErr != nil {
			logger.G.Sys().WithErr(closeErr).With("filter", filter).Error("failed to close cursor of networkunit distribution by network area id")
		}
	}()

	var results []networkUnitDistributionByNetworkAreaID
	if err = cursor.All(nCtx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// networkUnitDistributionByNetworkAreaID network unit aggregate result.
type networkUnitDistributionByNetworkAreaID struct {
	NetworkAreaID    int64 `bson:"_id"`
	NetworkUnitCount int64 `bson:"count"`
}
