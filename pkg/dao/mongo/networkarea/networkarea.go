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
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(tenantID string, client *mongo.Database, logger logger.Logger) *dao {
	tableName := TableName()
	d := &dao{
		tenantID:  tenantID,
		client:    client.Collection(tableName),
		logger:    logger,
		tableName: tableName,
	}

	d.IOrm = base.NewOrm[*NetworkArea, NetworkArea](d)

	return d
}

type dao struct {
	tenantID  string
	client    *mongo.Collection
	tableName string
	logger    logger.Logger
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
	return d.tableName
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

func (d *dao) count(ctx context.Context, filter bson.D) (int64, error) {
	return d.Count(ctx, append(filter, d.tenantFilter()))
}

func (d *dao) list(ctx context.Context, filter bson.D, findOpt *options.FindOptions) ([]*NetworkArea, error) {
	return d.List(ctx, append(filter, d.tenantFilter()), findOpt)
}

func (d *dao) get(ctx context.Context, filter bson.D) (*NetworkArea, error) {
	return d.Get(ctx, append(filter, d.tenantFilter()))
}

func (d *dao) upsertMany(ctx context.Context, networkAreas []*NetworkArea) error {
	models := buildUpsertManyParams(d.tenantID, networkAreas)

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

func (d *dao) updateMany(ctx context.Context, networkAreas []*NetworkArea) error {
	models := buildUpdateManyParams(d.tenantID, networkAreas)

	result, err := d.client.BulkWrite(ctx, models)
	if err != nil {
		return err
	}

	if result.MatchedCount > 0 {
		d.logger.Infof("updated networkareas, update-count(%v)", result.MatchedCount)
	}

	return nil
}

func (d *dao) deleteMany(ctx context.Context, networkAreaIDs ...int64) error {
	models := buildDeleteManyParams(d.tenantID, networkAreaIDs...)

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
func (d *dao) tenantFilter() bson.E {
	return bson.E{
		Key: "$or",
		Value: bson.A{
			bson.D{{FieldKeyTenantID, d.tenantID}},
			bson.D{{FieldKeyNetworkAreaID, base.GlobalNetworkAreaID}},
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
