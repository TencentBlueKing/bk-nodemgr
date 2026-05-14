/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operation ...
package operation

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

// TableName the operation table name.
func newDao(client *mongo.Database) *dao {
	d := &dao{
		client:    client.Collection(TableName),
		tableName: TableName,
	}

	d.IOrm = base.NewOrm[*Operation, Operation](d)

	return d
}

type dao struct {
	client    *mongo.Collection
	tableName string

	base.IOrm[*Operation, Operation]
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
			Keys: bson.D{{Key: FieldKeyTriggerID, Value: 1}},
		},
		{
			Keys: bson.D{{Key: FieldKeyInitContentToken, Value: 1}},
		},
		{
			Keys: bson.D{{Key: FieldKeyParentOperInstID, Value: 1}},
		},
		{
			Keys: bson.D{{Key: FieldKeyParentOperationID, Value: 1}},
		},
	}
}

// upsert updates or inserts an operation.
func (d *dao) upsert(nCtx contextx.IContext, operation *Operation) error {
	filter, upsert, opts := buildUpsertParams(operation)
	result, err := d.client.UpdateOne(nCtx, filter, upsert, opts)
	if err != nil {
		return err
	}

	switch {
	case result.UpsertedCount > 0:
		logger.G.Sys().With("unique-key", operation.UniqueKey(), "table", d.tableName).Info("upserted operation")

	case result.MatchedCount > 0:
		logger.G.Sys().With("unique-key", operation.UniqueKey(), "table", d.tableName).Info("updated operation")

	default:
		logger.G.Sys().With("unique-key", operation.UniqueKey(), "table", d.tableName).Warn("try to upsert operation but no changes made")
	}

	return nil
}

// buildUpsertParams build update params.
func buildUpsertParams(operation *Operation) (bson.D, bson.D, *mongoOptions.UpdateOptions) {
	// update operation by operation_id
	filter := bson.D{{Key: FieldKeyOperationID, Value: operation.OperationID}}

	// insert as creation or update data only.
	update := base.BuildUpsertParam(operation)

	// do upsert.
	opts := mongoOptions.Update().SetUpsert(true)

	return filter, update, opts
}

// delete deletes operations.
func (d *dao) delete(nCtx contextx.IContext, operIDs ...string) error {
	filter := base.AliveFilter()
	filter = WithOperationID(operIDs...)(filter)

	result, err := d.client.DeleteMany(nCtx, filter)
	if err != nil {
		return err
	}

	logger.G.Sys().With("operation-ids", operIDs, "deleted-count", result.DeletedCount).Info("deleted operations")

	return nil
}

// delete deletes operations by trigger id.
func (d *dao) deleteByTriggerID(nCtx contextx.IContext, triggerID ...string) error {
	filter := base.AliveFilter()
	filter = WithTriggerID(triggerID...)(filter)

	result, err := d.client.DeleteMany(nCtx, filter)
	if err != nil {
		return err
	}

	logger.G.Sys().With("trigger-ids", triggerID, "deleted-count", result.DeletedCount).Info("deleted operations")

	return nil
}

// pullField pull value from array field.
func (d *dao) pullField(nCtx contextx.IContext, operID, field string, value any) error {
	filter := base.AliveFilter()
	filter = WithOperationID(operID)(filter)

	update := base.BuildPullField(field, value)
	result, err := d.client.UpdateOne(nCtx, filter, update)
	if err != nil {
		return err
	}

	logger.G.Sys().With("field", field, "table", d.tableName, "value", value, "matched-count", result.MatchedCount).Debug("pulled operation field")

	return nil
}

// getLatestOperationInstStatusDistribution gets the operation latest instance status distribution.
func (d *dao) getLatestOperationInstStatusDistribution(nCtx contextx.IContext, filter bson.D, aggregateOptions ...*mongoOptions.AggregateOptions) (
	[]operationLatestInstStatusDistribution, error) {

	pipeline := mongo.Pipeline{}

	if len(filter) > 0 {
		pipeline = append(pipeline, bson.D{
			{Key: "$match", Value: filter},
		})
	}

	// group by trigger_id and latest instance status
	pipeline = append(pipeline,
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: bson.D{
					{Key: "trigger_id", Value: "$" + FieldKeyTriggerID},
					{Key: "status", Value: "$" + FieldKeyLatestInstState},
				}},
				{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
			}},
		},
	)

	cursor, err := d.client.Aggregate(nCtx, pipeline, aggregateOptions...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := cursor.Close(nCtx); closeErr != nil {
			logger.G.Sys().WithErr(closeErr).With("filter", filter).
				Error("failed to close cursor of status distribution")
		}
	}()

	var results []operationLatestInstStatusDistribution
	if err := cursor.All(nCtx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

type operationLatestInstStatusDistribution struct {
	GroupKey struct {
		TriggerID string  `bson:"trigger_id"`
		Status    *string `bson:"status"`
	} `bson:"_id"`
	Count int64 `bson:"count"`
}

// findWithoutFields find without fields.
func (d *dao) findWithoutFields(nCtx contextx.IContext, filter bson.D, page types.Page, fields ...string) (
	[]*Operation, error) {

	projection := bson.D{}
	for _, field := range fields {
		projection = append(projection, bson.E{Key: field, Value: 0})
	}
	findOptions := base.ParsePage(page)
	findOptions = findOptions.SetProjection(projection)

	result, err := d.client.Find(nCtx, filter, findOptions)
	if err != nil {
		return nil, err
	}

	datas := make([]*Operation, 0)
	for result.Next(nCtx) {
		table := &TableOperation{}
		if err := result.Decode(table); err != nil {
			logger.G.Sys().WithErr(err).Warn("failed to decode operation")

			continue
		}
		datas = append(datas, table.Data)
	}

	return datas, nil
}
