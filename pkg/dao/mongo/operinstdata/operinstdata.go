/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operinstdata ...
package operinstdata

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

// TableName the operinstdata table name.
func newDao(client *mongo.Database) *dao {
	d := &dao{
		client:    client.Collection(TableName),
		tableName: TableName,
	}
	d.IOrm = base.NewOrm[*OperInstData, OperInstData](d)

	return d
}

type dao struct {
	client    *mongo.Collection
	tableName string
	base.IOrm[*OperInstData, OperInstData]
}

// GetClient get client.
func (d *dao) GetClient() *mongo.Collection {
	return d.client
}

// GetTableName get table name.
func (d *dao) GetTableName() string {
	return d.tableName
}

// GetIndexes get indexes.
func (d *dao) GetIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys: bson.D{{Key: FieldKeyOperationID, Value: 1}},
		},
	}
}

// upsert updates or inserts a operation_inst_data.
func (d *dao) upsert(nCtx contextx.IContext, data *OperInstData) error {
	filter, upsert, opts := buildUpsertParams(data)
	result, err := d.client.UpdateOne(nCtx, filter, upsert, opts)
	if err != nil {
		return err
	}

	switch {
	case result.UpsertedCount > 0:
		logger.G.Sys().With("unique-key", data.UniqueKey(), "table", d.tableName).Info("inserted operation instance data")

	case result.MatchedCount > 0:
		logger.G.Sys().With("unique-key", data.UniqueKey(), "table", d.tableName).Info("updated operation instance data")

	default:
		logger.G.Sys().With("unique-key", data.UniqueKey(), "table", d.tableName).Info("try to upsert data but no changes made")
	}

	return nil
}

// buildUpsertParams build update params.
func buildUpsertParams(data *OperInstData) (bson.D, bson.D, *mongoOptions.UpdateOptions) {
	// update data by operation_inst_data_id.
	filter := bson.D{{Key: FieldKeyOperInstID, Value: data.OperInstID}}

	// insert as creation or update data only.
	update := base.BuildUpsertParam(data)

	// do upsert.
	opts := mongoOptions.Update().SetUpsert(true)

	return filter, update, opts
}

// find all operinstdata.
func (d *dao) find(nCtx contextx.IContext, filter bson.D, fields ...string) ([]*OperInstData, error) {
	projection := bson.D{}
	for _, field := range fields {
		projection = append(projection, bson.E{Key: field, Value: 1})
	}
	findOptions := mongoOptions.Find().SetProjection(projection)
	result, err := d.client.Find(nCtx, filter, findOptions)
	if err != nil {
		return nil, err
	}

	datas := make([]*OperInstData, 0)
	for result.Next(nCtx) {
		table := &TableOperInstData{}
		if err := result.Decode(table); err != nil {
			logger.G.Sys().WithErr(err).Warn("failed to decode operation instance data")

			continue
		}

		datas = append(datas, table.Data)
	}

	return datas, nil
}

// findWithoutFields find without fields.
func (d *dao) findWithoutFields(nCtx contextx.IContext, filter bson.D, page types.Page, fields ...string) (
	[]*OperInstData, error) {

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

	datas := make([]*OperInstData, 0)
	for result.Next(nCtx) {
		table := &TableOperInstData{}
		if err := result.Decode(table); err != nil {
			logger.G.Sys().WithErr(err).Warn("failed to decode operation instance data")

			continue
		}
		datas = append(datas, table.Data)
	}

	return datas, nil
}

// updateField update field.
func (d *dao) updateField(nCtx contextx.IContext, filter bson.D, field string, value any) error {
	return d.UpdateField(nCtx, filter, field, value)
}

// pushField push field.
func (d *dao) pushField(nCtx contextx.IContext, filter bson.D, field string, value any) error {
	update := base.BuildPushField(field, value)
	result, err := d.client.UpdateOne(nCtx, filter, update)
	if err != nil {
		return err
	}

	logger.G.Sys().
		With("field", field, "table", d.tableName, "value", value, "matched-count", result.MatchedCount).
		Debug("pushed operation instance data field")

	return nil
}

func (d *dao) get(nCtx contextx.IContext, filter bson.D, fields ...string) (*OperInstData, error) {
	return d.Get(nCtx, filter, fields...)
}

// listALLLastOperInst lists all last operation instances base on operation id.
func (d *dao) listALLLastOperInst(nCtx contextx.IContext, filter bson.D) ([]*OperInstData, error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: base.FieldKeyCreatedAt, Value: -1}}}},
		bson.D{
			{
				Key: "$group",
				Value: bson.D{
					{Key: "_id", Value: "$" + FieldKeyOperationID},
					{Key: "doc", Value: bson.D{{Key: "$first", Value: "$$ROOT"}}},
				},
			},
		},
		bson.D{{Key: "$replaceRoot", Value: bson.D{{Key: "newRoot", Value: "$doc"}}}},
	}

	opts := mongoOptions.Aggregate().SetAllowDiskUse(true)
	cursor, err := d.client.Aggregate(nCtx, pipeline, opts)
	if err != nil {
		return nil, fmt.Errorf("aggregate last oper inst failed: %w", err)
	}

	// Use a background context for cursor iteration and close to avoid the cursor
	// being interrupted when the caller's context deadline expires mid-stream.
	// The aggregate query itself is already bound to nCtx above.
	cursorCtx := context.Background()
	defer func() {
		if err := cursor.Close(cursorCtx); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to close cursor")
		}
	}()

	datas := make([]*OperInstData, 0)
	for cursor.Next(cursorCtx) {
		table := &TableOperInstData{}
		if err := cursor.Decode(table); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to decode table")

			continue
		}
		datas = append(datas, table.Data)
	}

	return datas, nil
}

// delete deletes operinstdata by given operInstIDs.
func (d *dao) delete(nCtx contextx.IContext, operInstIDs ...string) error {
	filter := base.AliveFilter()
	filter = WithOperInstID(operInstIDs...)(filter)

	result, err := d.client.DeleteMany(nCtx, filter)
	if err != nil {
		return err
	}

	logger.G.Sys().
		With("table", d.tableName, "oper-inst-ids", operInstIDs, "deleted-count", result.DeletedCount).
		Info("deleted operation instance data")

	return nil
}

// deleteByTriggerID deletes operinstdata by given operInstIDs.
func (d *dao) deleteByTriggerID(nCtx contextx.IContext, triggerID ...string) error {
	filter := base.AliveFilter()
	filter = WithTriggerID(triggerID...)(filter)

	result, err := d.client.DeleteMany(nCtx, filter)
	if err != nil {
		return err
	}

	logger.G.Sys().With("table", d.tableName, "trigger-ids", triggerID, "deleted-count", result.DeletedCount).Info("deleted operation instance data")

	return nil
}
