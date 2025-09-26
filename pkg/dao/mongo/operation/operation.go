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
	var indexes []mongo.IndexModel

	return indexes
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

	logger.G.Sys().With("trigger-ids", operIDs, "deleted-count", result.DeletedCount).Info("deleted triggers")

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
