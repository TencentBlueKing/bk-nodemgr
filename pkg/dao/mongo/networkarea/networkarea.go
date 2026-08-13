/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package networkarea provides network-area dao operations.
package networkarea

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func newDao(tenantID string, client *mongo.Database) *dao {
	tableName := TableName(tenantID)
	d := &dao{
		client:    client.Collection(tableName),
		tableName: tableName,
	}

	d.IOrm = base.NewOrm[*NetworkArea, NetworkArea](d)

	return d
}

type dao struct {
	client    *mongo.Collection
	tableName string
	base.IOrm[*NetworkArea, NetworkArea]
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

func (d *dao) upsertMany(nCtx contextx.IContext, networkAreas []*NetworkArea) error {
	models := buildUpsertManyParams(networkAreas)

	result, err := d.client.BulkWrite(nCtx, models)
	if err != nil {
		return err
	}

	if result.UpsertedCount > 0 {
		logger.G.Sys().With("inserted-count", result.UpsertedCount).Info("upserted networkareas")
	}

	if result.MatchedCount > 0 {
		logger.G.Sys().With("matched-count", result.MatchedCount).Info("upserted networkareas")
	}

	return nil
}

func (d *dao) updateMany(nCtx contextx.IContext, networkAreas []*NetworkArea) error {
	models := buildUpdateManyParams(networkAreas)

	result, err := d.client.BulkWrite(nCtx, models)
	if err != nil {
		return err
	}

	if result.MatchedCount > 0 {
		logger.G.Sys().With("matched-count", result.MatchedCount).Info("updated networkareas")
	}

	return nil
}

func (d *dao) deleteMany(nCtx contextx.IContext, networkAreaIDs ...int64) error {
	models := buildDeleteManyParams(networkAreaIDs...)

	result, err := d.client.BulkWrite(nCtx, models)
	if err != nil {
		return err
	}

	if result.MatchedCount > 0 {
		logger.G.Sys().With("deleted-count", result.MatchedCount).Info("deleted networkareas")
	}

	return nil
}

// buildUpsertManyParams build upsert many params.
func buildUpsertManyParams(networkAreas []*NetworkArea) []mongo.WriteModel {
	models := make([]mongo.WriteModel, 0)

	for _, networkarea := range networkAreas {
		filter := bson.D{
			bson.E{Key: FieldKeyNetworkAreaID, Value: networkarea.NetworkAreaID},
		}

		update := base.BuildUpsertParam(networkarea)

		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	}

	return models
}

// buildUpdateManyParams build update many params.
func buildUpdateManyParams(networkAreas []*NetworkArea) []mongo.WriteModel {
	models := make([]mongo.WriteModel, 0)

	for _, networkarea := range networkAreas {
		filter := append(base.AliveFilter(),
			bson.E{Key: FieldKeyNetworkAreaID, Value: networkarea.NetworkAreaID})

		update := base.BuildUpsertParam(networkarea)

		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(false))
	}

	return models
}

// buildDeleteManyParams build delete many params.
func buildDeleteManyParams(networkAreaIDs ...int64) []mongo.WriteModel {
	filter := bson.D{
		bson.E{Key: FieldKeyNetworkAreaID, Value: bson.D{{Key: "$in", Value: networkAreaIDs}}}}

	update := base.BuildDeleteParam()

	return []mongo.WriteModel{mongo.NewUpdateManyModel().SetFilter(filter).SetUpdate(update).SetUpsert(false)}
}
