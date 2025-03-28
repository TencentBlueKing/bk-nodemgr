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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(client *mongo.Database, logger logger.Logger) *dao {
	d := &dao{client: client.Collection(TableName), logger: logger}
	d.baseOrm = base.NewOrm[*OperInstData, OperInstData](d)

	return d
}

type dao struct {
	client  *mongo.Collection
	logger  logger.Logger
	baseOrm base.IOrm[*OperInstData, OperInstData]
}

// GetClient get client.
func (d *dao) GetClient() *mongo.Collection {
	return d.client
}

// GetLogger get logger.
func (d *dao) GetLogger() logger.Logger {
	return d.logger
}

// GetTableName get table name.
func (d *dao) GetTableName() string {
	return TableName
}

// GetIndexes get indexes.
func (d *dao) GetIndexes() []mongo.IndexModel {
	return nil
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
			d.logger.Infof("successfully inserted data, unique-key(%s)", data.UniqueKey())
		}
	case result.MatchedCount > 0:
		{
			d.logger.Infof("successfully updated data, unique-key(%s)", data.UniqueKey())
		}
	default:
		d.logger.Warnf("try to upsert data but no changes made, unique-key(%s)", data.UniqueKey())
	}

	return nil
}

// buildUpsertParams build update params.
func buildUpsertParams(data *OperInstData) (bson.D, bson.D, *mongoOptions.UpdateOptions) {
	// update data by operation_inst_data_id.
	filter := bson.D{{Key: FieldOperInstID.String(), Value: data.OperInstID}}

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
func (d *dao) findWithoutFields(ctx context.Context, filter bson.D, fields ...string) ([]*OperInstData, error) {
	projection := bson.D{}
	for _, field := range fields {
		projection = append(projection, bson.E{Key: field, Value: 0})
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

	d.logger.Infof("successfully push, field(%v), value(%v), updated-count(%d)", field, value, result.MatchedCount)

	return nil
}

func (d *dao) get(ctx context.Context, filter bson.D, fields ...string) (*OperInstData, error) {
	return d.baseOrm.Get(ctx, filter, fields...)
}
