/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package business provides business dao operations.
package business

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(tenantID string, client *mongo.Database) *dao {
	tableName := TableName(tenantID)
	d := &dao{
		client:    client.Collection(tableName),
		tableName: tableName,
	}

	d.IOrm = base.NewOrm[*Business, Business](d)

	return d
}

type dao struct {
	client    *mongo.Collection
	tableName string

	base.IOrm[*Business, Business]
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

// upsert updates or inserts a business.
func (d *dao) upsert(nCtx contextx.IContext, biz *Business) error {
	filter, upsert, opts := buildUpsertParams(biz)
	result, err := d.client.UpdateOne(nCtx, filter, upsert, opts)
	if err != nil {
		return err
	}

	switch {
	case result.UpsertedCount > 0:
		logger.G.Sys().With("unique-key", biz.UniqueKey()).Info("upserted business")

	case result.MatchedCount > 0:
		logger.G.Sys().With("unique-key", biz.UniqueKey()).Info("updated business")

	default:
		logger.G.Sys().With("unique-key", biz.UniqueKey()).Warn("try to upsert business but no changes made")
	}

	return nil
}

// buildUpsertParams build update params.
func buildUpsertParams(biz *Business) (bson.D, bson.D, *mongoOptions.UpdateOptions) {
	// update business by biz_id.
	filter := bson.D{{Key: FieldKeyBizID, Value: biz.BizID}}

	// insert as creation or update data only.
	update := base.BuildUpsertParam(biz)

	// do upsert.
	opts := mongoOptions.Update().SetUpsert(true)

	return filter, update, opts
}

// upsertMany upsert many business.
func (d *dao) upsertMany(nCtx contextx.IContext, bizs []*Business) error {
	models := buildUpsertManyParams(bizs)

	result, err := d.client.BulkWrite(nCtx, models)
	if err != nil {
		return err
	}

	if result.UpsertedCount > 0 {
		logger.G.Sys().With("inserted-count", result.UpsertedCount).Info("upserted business")
	}

	if result.MatchedCount > 0 {
		logger.G.Sys().With("matched-count", result.MatchedCount).Info("upserted business")
	}

	return nil
}

// buildUpsertManyParams build upsert many params.
func buildUpsertManyParams(bizs []*Business) []mongo.WriteModel {
	models := make([]mongo.WriteModel, 0, len(bizs))

	for _, biz := range bizs {
		filter := bson.D{{Key: FieldKeyBizID, Value: biz.BizID}}

		update := base.BuildUpsertParam(biz)

		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	}

	return models
}
