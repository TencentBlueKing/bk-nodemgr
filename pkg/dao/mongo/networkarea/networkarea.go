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
	"context"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func newDao(client *mongo.Database, logger logger.Logger) *dao {
	d := &dao{
		client: client.Collection(TableName()),
		logger: logger,
	}

	d.IOrm = base.NewOrm[*NetworkArea, NetworkArea](d)

	return d
}

type dao struct {
	client *mongo.Collection
	logger logger.Logger
	base.IOrm[*NetworkArea, NetworkArea]
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
	return TableName()
}

// GetIndexes get the dao's indexes.
func (d *dao) GetIndexes() []mongo.IndexModel {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: FieldKeyNetworkAreaID, Value: 1}},
		},
	}

	return indexes
}

func (d *dao) upsertMany(ctx context.Context, tenantID string, networkAreas []*NetworkArea) error {
	models := buildUpsertManyParams(tenantID, networkAreas)

	result, err := d.client.BulkWrite(ctx, models)
	if err != nil {
		return err
	}

	if result.UpsertedCount > 0 {
		d.logger.Infof("inserted networkareas, inserted-count(%v)", result.UpsertedCount)
	}

	if result.MatchedCount > 0 {
		d.logger.Infof("updated networkareas, update-count(%v)", result.MatchedCount)
	}

	return nil
}

func (d *dao) updateMany(ctx context.Context, tenantID string, networkAreas []*NetworkArea) error {
	models := buildUpdateManyParams(tenantID, networkAreas)

	result, err := d.client.BulkWrite(ctx, models)
	if err != nil {
		return err
	}

	if result.MatchedCount > 0 {
		d.logger.Infof("updated networkareas, update-count(%v)", result.MatchedCount)
	}

	return nil
}

func (d *dao) deleteMany(ctx context.Context, tenantID string, networkAreaIDs ...int64) error {
	models := buildDeleteManyParams(tenantID, networkAreaIDs...)

	result, err := d.client.BulkWrite(ctx, models)
	if err != nil {
		return err
	}

	if result.MatchedCount > 0 {
		d.logger.Infof("deleted networkareas, deleted-count(%v)", result.MatchedCount)
	}

	return nil
}

// tenantFilter additional tenant filter.
// all tenants can filter global networkarea in query methods.
// and the networkareas in their own tenant.
func tenantFilter(tenantID string) bson.E {
	return bson.E{
		Key: "$or",
		Value: bson.A{
			bson.D{{Key: FieldKeyTenantID, Value: tenantID}},
			bson.D{{Key: FieldKeyNetworkAreaID, Value: base.GlobalNetworkAreaID}},
		}}
}

// buildUpsertManyParams build upsert many params.
func buildUpsertManyParams(tenantID string, networkAreas []*NetworkArea) []mongo.WriteModel {
	models := make([]mongo.WriteModel, 0)

	for _, networkarea := range networkAreas {
		filter := bson.D{
			bson.E{Key: FieldKeyNetworkAreaID, Value: networkarea.NetworkAreaID},
			bson.E{Key: FieldKeyTenantID, Value: tenantID},
		}

		update := base.BuildUpsertParam(networkarea)

		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	}

	return models
}

// buildUpdateManyParams build update many params.
func buildUpdateManyParams(tenantID string, networkAreas []*NetworkArea) []mongo.WriteModel {
	models := make([]mongo.WriteModel, 0)

	for _, networkarea := range networkAreas {
		filter := append(base.AliveFilter(),
			bson.E{Key: FieldKeyNetworkAreaID, Value: networkarea.NetworkAreaID},
			bson.E{Key: FieldKeyTenantID, Value: tenantID})

		update := base.BuildUpsertParam(networkarea)

		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(false))
	}

	return models
}

// buildDeleteManyParams build delete many params.
func buildDeleteManyParams(tenantID string, networkAreaIDs ...int64) []mongo.WriteModel {
	filter := bson.D{
		bson.E{Key: FieldKeyNetworkAreaID, Value: bson.D{{Key: "$in", Value: networkAreaIDs}}},
		bson.E{Key: FieldKeyTenantID, Value: tenantID}}

	update := base.BuildDeleteParam()

	return []mongo.WriteModel{mongo.NewUpdateManyModel().SetFilter(filter).SetUpdate(update).SetUpsert(false)}
}
