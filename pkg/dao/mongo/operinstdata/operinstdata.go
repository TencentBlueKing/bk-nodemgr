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
	"log"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

// TableName the operinstdata table name.
func newDao(client *mongo.Database, logger logger.ILogger) *dao {
	d := &dao{
		client:    client.Collection(TableName),
		logger:    logger,
		tableName: TableName,
	}
	d.baseOrm = base.NewOrm[*OperInstData, OperInstData](d)

	return d
}

type dao struct {
	client    *mongo.Collection
	logger    logger.ILogger
	tableName string
	baseOrm   base.IOrm[*OperInstData, OperInstData]
}

// GetClient get client.
func (d *dao) GetClient() *mongo.Collection {
	return d.client
}

// GetLogger get logger.
func (d *dao) GetLogger() logger.ILogger {
	return d.logger
}

// GetTableName get table name.
func (d *dao) GetTableName() string {
	return d.tableName
}

// GetIndexes get indexes.
func (d *dao) GetIndexes() []mongo.IndexModel {
	var indexes []mongo.IndexModel

	return indexes
}

// upsert updates or inserts a operation_inst_data.
func (d *dao) upsert(ctx context.Context, data *OperInstData) error {
	filter, upsert, opts := buildUpsertParams(data)
	result, err := d.client.UpdateOne(ctx, filter, upsert, opts)
	if err != nil {
		return err
	}

	switch {
	case result.UpsertedCount > 0:
		{
			d.logger.Infof("inserted data, unique-key(%s), table(%s)", data.UniqueKey(), TableName)
		}
	case result.MatchedCount > 0:
		{
			d.logger.Infof("updated data, unique-key(%s), table(%s)", data.UniqueKey(), TableName)
		}
	default:
		d.logger.Warnf("try to upsert data but no changes made. unique-key(%s), table(%s)",
			data.UniqueKey(), TableName)
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
func (d *dao) find(ctx context.Context, filter bson.D, fields ...string) ([]*OperInstData, error) {
	projection := bson.D{}
	for _, field := range fields {
		projection = append(projection, bson.E{Key: field, Value: 1})
	}
	findOptions := mongoOptions.Find().SetProjection(projection)
	result, err := d.client.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}

	datas := make([]*OperInstData, 0)
	for result.Next(ctx) {
		table := &TableOperInstData{}
		if err := result.Decode(table); err != nil {
			d.logger.Warnf("failed to decode operinstdata, err %v", err)

			continue
		}

		datas = append(datas, table.Data)
	}

	return datas, nil
}

// findWithoutFields find without fields.
func (d *dao) findWithoutFields(ctx context.Context, filter bson.D, page types.Page, fields ...string) (
	[]*OperInstData, error) {

	projection := bson.D{}
	for _, field := range fields {
		projection = append(projection, bson.E{Key: field, Value: 0})
	}
	findOptions := base.ParsePage(page)
	findOptions = findOptions.SetProjection(projection)

	result, err := d.client.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}

	datas := make([]*OperInstData, 0)
	for result.Next(ctx) {
		table := &TableOperInstData{}
		if err := result.Decode(table); err != nil {
			d.logger.Warnf("failed to decode operinstdata, err %v", err)

			continue
		}
		datas = append(datas, table.Data)
	}

	return datas, nil
}

// updateField update field.
func (d *dao) updateField(ctx context.Context, filter bson.D, field string, value any) error {
	return d.baseOrm.UpdateField(ctx, filter, field, value)
}

// pushField push field.
func (d *dao) pushField(ctx context.Context, filter bson.D, field string, value any) error {
	update := base.BuildPushField(field, value)
	result, err := d.client.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	d.logger.Debugf("pushed oper-inst-data field(%v), table(%s), value(%v), updated-count(%d)",
		field, d.tableName, value, result.MatchedCount)

	return nil
}

func (d *dao) get(ctx context.Context, filter bson.D, fields ...string) (*OperInstData, error) {
	return d.baseOrm.Get(ctx, filter, fields...)
}

// listALLLastOperInst lists all last operation instances base on operation id.
func (d *dao) listALLLastOperInst(ctx context.Context, filter bson.D) ([]*OperInstData, error) {
	pipeline := mongo.Pipeline{
		{{"$match", filter}},
		{{"$sort", bson.D{{base.FieldKeyCreatedAt, -1}}}},
		{{"$group", bson.D{
			{"_id", "$" + FieldKeyOperationID},
			{"doc", bson.D{{"$first", "$$ROOT"}}},
		}}},
		{{"$replaceRoot", bson.D{{"newRoot", "$doc"}}}},
	}

	opts := mongoOptions.Aggregate().SetAllowDiskUse(true)
	cursor, err := d.client.Aggregate(ctx, pipeline, opts)
	if err != nil {
		log.Fatal(err)
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			d.logger.Errorf("failed to close cursor, err %v", err)
		}
	}(cursor, ctx)

	datas := make([]*OperInstData, 0)
	for cursor.Next(ctx) {
		table := &TableOperInstData{}
		if err := cursor.Decode(table); err != nil {
			d.logger.Errorf("failed to decode table, err %v", err)
			continue
		}
		datas = append(datas, table.Data)
	}

	return datas, nil
}

// delete deletes operinstdata by given operInstIDs.
func (d *dao) delete(ctx context.Context, operInstIDs ...string) error {
	filter := base.AliveFilter()
	filter = append(filter, bson.E{Key: FieldKeyOperInstID, Value: bson.D{{Key: "$in", Value: operInstIDs}}})

	result, err := d.client.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}

	d.logger.Infof("deleted oper-inst-data, table(%s), oper-inst-ids(%v), deleted-count(%d)",
		d.tableName, operInstIDs, result.DeletedCount)

	return nil
}
