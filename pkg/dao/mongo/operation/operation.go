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
	"context"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

// TableName the operation table name.
func newDao(tenantID string, client *mongo.Database, logger logger.Logger) *dao {
	tableName := TableName(tenantID)
	d := &dao{
		client:    client.Collection(tableName),
		logger:    logger,
		tableName: tableName,
	}

	d.IOrm = base.NewOrm[*Operation, Operation](d)

	return d
}

type dao struct {
	client    *mongo.Collection
	tableName string
	logger    logger.Logger
	base.IOrm[*Operation, Operation]
}

// GetClient get the dao's client.
func (d *dao) GetClient() *mongo.Collection {
	return d.client
}

// GetLogger get the dao's logger.
func (d *dao) GetLogger() logger.Logger {
	return d.logger
}

// GetTableName get the dao's table name.
func (d *dao) GetTableName() string {
	return d.tableName
}

// GetIndexes get the dao's indexes.
func (d *dao) GetIndexes() []mongo.IndexModel {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: FieldKeyOperationID, Value: 1}},
			Options: mongoOptions.Index().SetUnique(true),
		},
	}

	return indexes
}

// upsert updates or inserts an operation.
func (d *dao) upsert(ctx context.Context, operation *Operation) error {
	filter, upsert, opts := buildUpsertParams(operation)
	result, err := d.client.UpdateOne(ctx, filter, upsert, opts)
	if err != nil {
		return err
	}

	switch {
	case result.UpsertedCount > 0:
		{
			d.logger.Infof("upserted operation, unique-key(%s), table(%s)",
				operation.UniqueKey(), d.tableName)
		}
	case result.MatchedCount > 0:
		{
			d.logger.Infof("updated operation, unique-key(%s), table(%s)",
				operation.UniqueKey(), d.tableName)
		}
	default:
		d.logger.Warnf("try to upsert operation but no changes made. unique-key(%s), table(%s)",
			operation.UniqueKey(), d.tableName)
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

// find all operations.
func (d *dao) find(ctx context.Context, filter bson.D) ([]*Operation, error) {
	result, err := d.client.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	operations := make([]*Operation, 0)
	for result.Next(ctx) {
		table := &TableOperation{}
		if err := result.Decode(table); err != nil {
			d.logger.Warnf("failed to decode trigger, err %v", err)

			continue
		}
		operations = append(operations, table.Data)
	}

	return operations, nil
}
