/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package networkunit provides network-unit dao operations.
package networkunit

import (
	"context"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/counter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func newDao(client *mongo.Database, logger logger.ILogger) *dao {
	tableName := TableName()
	d := &dao{
		client:    client.Collection(tableName),
		logger:    logger,
		tableName: tableName,
		counter:   counter.New(client, logger),
	}

	d.IOrm = base.NewOrm[*NetworkUnit, NetworkUnit](d)

	return d
}

type dao struct {
	client    *mongo.Collection
	tableName string
	logger    logger.ILogger
	counter   counter.Handler

	base.IOrm[*NetworkUnit, NetworkUnit]
}

// GetClient get the dao's client.
func (d *dao) GetClient() *mongo.Collection {
	return d.client
}

// GetLogger get the dao's logger.
func (d *dao) GetLogger() logger.ILogger {
	return d.logger
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

func (d *dao) create(ctx context.Context, networkUnit *NetworkUnit) (int64, error) {
	newSequence, err := d.counter.Generate(ctx, TableName())
	if err != nil {
		return 0, err
	}

	networkUnit.NetworkUnitID = newSequence
	table := &TableNetworkUnit{
		BasicInfo: base.BasicInfo{
			CreatedAt: time.Now(),
			IsDeleted: false,
		},
		Data: networkUnit,
	}

	if _, err = d.client.InsertOne(ctx, table); err != nil {
		return 0, err
	}

	return newSequence, nil
}

func (d *dao) updateMany(ctx context.Context, tenantID string, networkunits []*NetworkUnit) error {
	models := buildUpdateManyParams(tenantID, networkunits)

	result, err := d.client.BulkWrite(ctx, models)
	if err != nil {
		return err
	}

	if result.MatchedCount > 0 {
		d.logger.Infof("successfully updated networkunits, update-count(%v)", result.MatchedCount)
	}

	return nil
}

func (d *dao) deleteMany(ctx context.Context, tenantID string, networkUnitIDs ...int64) error {
	models := buildDeleteManyParams(tenantID, networkUnitIDs...)

	result, err := d.client.BulkWrite(ctx, models)
	if err != nil {
		return err
	}

	if result.MatchedCount > 0 {
		d.logger.Infof("successfully deleted networkunits, deleted-count(%v)", result.MatchedCount)
	}

	return nil
}

// tenantFilter additional tenant filter.
// global networkarea is a special networkarea, it belongs to system tenant, but it can be seen by all tenants.
// this scene is also ensured in CMDB.
// a query from a tenant, should be filtered in its own tenant, and plus the global networkarea.
func tenantFilter(tenantID string) bson.E {
	return bson.E{
		Key: "$or",
		Value: bson.A{
			bson.D{{Key: FieldKeyTenantID, Value: tenantID}},
			bson.D{{Key: FieldKeyNetworkAreaID, Value: base.GlobalNetworkAreaID}},
		}}
}

// buildUpdateManyParams build update many params.
func buildUpdateManyParams(tenantID string, networkUnits []*NetworkUnit) []mongo.WriteModel {
	models := make([]mongo.WriteModel, 0)

	for _, networkUnit := range networkUnits {
		filter := append(base.AliveFilter(),
			bson.E{Key: FieldKeyNetworkUnitID, Value: networkUnit.NetworkUnitID},
			bson.E{Key: FieldKeyTenantID, Value: tenantID})

		update := base.BuildUpsertParam(networkUnit)

		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(false))
	}

	return models
}

// buildDeleteManyParams build delete many params.
func buildDeleteManyParams(tenantID string, networkUnitIDs ...int64) []mongo.WriteModel {
	filter := bson.D{
		bson.E{Key: FieldKeyNetworkUnitID, Value: bson.D{{Key: "$in", Value: networkUnitIDs}}},
		bson.E{Key: FieldKeyTenantID, Value: tenantID}}

	update := base.BuildDeleteParam()

	return []mongo.WriteModel{mongo.NewUpdateManyModel().SetFilter(filter).SetUpdate(update).SetUpsert(false)}
}
