/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package host ...
package host

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(tenantID string, client *mongo.Database) *dao {
	tableName := TableName(tenantID)
	d := &dao{
		client:    client.Collection(tableName),
		tableName: tableName,
	}

	d.IOrm = base.NewOrm[*Host, Host](d)

	return d
}

type dao struct {
	client    *mongo.Collection
	tableName string
	base.IOrm[*Host, Host]
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

// upsertMany upsert many hosts.
func (d *dao) upsertMany(nCtx contextx.IContext, hosts []*Host) error {
	models := buildUpsertManyParams(hosts)

	result, err := d.client.BulkWrite(nCtx, models)
	if err != nil {
		return err
	}

	if result.UpsertedCount > 0 {
		logger.G.Sys().With("inserted-count", result.UpsertedCount).Info("upserted hosts")
	}

	if result.MatchedCount > 0 {
		logger.G.Sys().With("matched-count", result.MatchedCount).Info("upserted hosts")
	}

	return nil
}

// upsertStaticMany upsert many host statics.
func (d *dao) upsertStaticMany(nCtx contextx.IContext, hosts []*Host) error {
	models := buildUpsertStaticManyParams(hosts)

	result, err := d.client.BulkWrite(nCtx, models)
	if err != nil {
		return err
	}

	if result.UpsertedCount > 0 {
		logger.G.Sys().With("inserted-count", result.UpsertedCount).Info("upserted host statics")
	}

	if result.MatchedCount > 0 {
		logger.G.Sys().With("matched-count", result.MatchedCount).Info("upserted host statics")
	}

	return nil
}

// updateDynamicMany update many host dynamic.
func (d *dao) updateDynamicMany(nCtx contextx.IContext, hosts []*Host) error {
	models := buildUpdateDynamicManyParams(hosts)

	result, err := d.client.BulkWrite(nCtx, models)
	if err != nil {
		return err
	}

	if result.UpsertedCount > 0 {
		logger.G.Sys().With("inserted-count", result.UpsertedCount).Info("upserted host dynamics")
	}

	if result.MatchedCount > 0 {
		logger.G.Sys().With("matched-count", result.MatchedCount).Info("upserted host dynamics")
	}

	return nil
}

// distinctString distinct string field.
func (d *dao) distinctString(
	nCtx contextx.IContext, key string, filter bson.D, distinctOpt *options.DistinctOptions) ([]string, error) {

	return d.DistinctString(nCtx, key, filter, distinctOpt)
}

// distinctInt64 distinct int64 field.
func (d *dao) distinctInt64(
	nCtx contextx.IContext, key string, filter bson.D, distinctOpt *options.DistinctOptions) ([]int64, error) {

	return d.DistinctInt64(nCtx, key, filter, distinctOpt)
}

// buildUpsertManyParams build upsert many params.
func buildUpsertManyParams(hosts []*Host) []mongo.WriteModel {
	models := make([]mongo.WriteModel, 0)
	for _, host := range hosts {
		filter := bson.D{{Key: FieldKeyHostID, Value: host.HostID}}

		update := base.BuildUpsertParam(host)

		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	}

	return models
}

// buildUpsertStaticManyParams build upsert static many params.
func buildUpsertStaticManyParams(hosts []*Host) []mongo.WriteModel {
	models := make([]mongo.WriteModel, 0)
	for _, host := range hosts {
		filter := bson.D{{Key: FieldKeyHostID, Value: host.HostID}}

		nowTime := time.Now()
		update := bson.D{
			{
				Key: "$set",
				Value: bson.M{
					"basic.is_deleted": false,
					"basic.updated_at": nowTime,
					"data.tenant_id":   host.TenantID,
					"data.static":      host.Static,
				},
			},
			{
				Key: "$setOnInsert",
				Value: bson.M{
					"basic.created_at": nowTime,
					"data.dynamic":     host.Dynamic,
				},
			},
		}

		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	}

	return models
}

// buildUpdateDynamicManyParams build upsert dynamic many params.
func buildUpdateDynamicManyParams(hosts []*Host) []mongo.WriteModel {
	models := make([]mongo.WriteModel, 0)
	for _, host := range hosts {
		filter := append(base.AliveFilter(), bson.E{Key: FieldKeyHostID, Value: host.HostID})

		nowTime := time.Now()
		update := bson.D{
			{
				Key: "$set",
				Value: bson.M{
					"basic.is_deleted": false,
					"basic.updated_at": nowTime,
					"data.dynamic":     host.Dynamic,
				},
			},
		}

		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(false))
	}

	return models
}
