/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package stopoperinst define the table to store stopping operation instance.
package stopoperinst

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(client *mongo.Database) *dao {
	d := &dao{
		client:    client.Collection(TableName),
		tableName: TableName,
	}

	d.IOrm = base.NewOrm[*StopOperInst, StopOperInst](d)

	return d
}

type dao struct {
	client    *mongo.Collection
	tableName string
	base.IOrm[*StopOperInst, StopOperInst]
}

// GetClient gets the MongoDB collection.
func (d *dao) GetClient() *mongo.Collection {
	return d.client
}

// GetTableName gets the MongoDB collection name.
func (d *dao) GetTableName() string {
	return d.tableName
}

// GetIndexes gets the stopping operation instance indexes.
func (d *dao) GetIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys: bson.D{{Key: FieldKeyExpireAt, Value: 1}},
			Options: mongoOptions.Index().
				SetExpireAfterSeconds(0),
		},
	}
}

// upsert updates or inserts an operation instance data.
func (d *dao) upsert(nCtx contextx.IContext, inst *StopOperInst) error {
	filter, upsert, opts := buildUpsertParams(inst)
	result, err := d.client.UpdateOne(nCtx, filter, upsert, opts)
	if err != nil {
		return err
	}

	switch {
	case result.UpsertedCount > 0:
		logger.G.Sys().With("unique-key", inst.UniqueKey()).Info("inserted stop operation instance")

	case result.MatchedCount > 0:
		logger.G.Sys().With("unique-key", inst.UniqueKey()).Info("updated stop operation instance")

	default:
		logger.G.Sys().With("unique-key", inst.UniqueKey()).Info("try to upsert stop operation instance but no changes made")
	}

	return nil
}

// buildUpsertParams build update params.
func buildUpsertParams(inst *StopOperInst) (bson.D, bson.D, *mongoOptions.UpdateOptions) {
	// update data by operation_inst_data_id.
	filter := bson.D{{Key: FieldKeyOperInstID, Value: inst.OperInstID}}

	// upsert as creation or update data only.
	update := base.BuildUpsertParam(inst)

	// do upsert.
	opts := mongoOptions.Update().SetUpsert(true)

	return filter, update, opts
}
