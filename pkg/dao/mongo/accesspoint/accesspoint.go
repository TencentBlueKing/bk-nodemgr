/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package accesspoint provides access-point dao operations.
package accesspoint

import (
	"context"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/counter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(client *mongo.Database, logger logger.Logger) *dao {
	d := &dao{
		client:  client.Collection(TableName()),
		logger:  logger,
		counter: counter.New(client, logger)}

	d.IOrm = base.NewOrm[*AccessPoint, AccessPoint](d)

	return d
}

type dao struct {
	client  *mongo.Collection
	logger  logger.Logger
	counter counter.Handler
	base.IOrm[*AccessPoint, AccessPoint]
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

// nolint:contextcheck
// ensureIndexes ensures the required indexes for the collection.
func (d *dao) GetIndexes() []mongo.IndexModel {
	opts := new(options.IndexOptions)
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: FieldKeyAccessPointID, Value: 1}},
			Options: opts.SetUnique(true),
		},
	}

	return indexes
}

func (d *dao) create(ctx context.Context, accessPoint *AccessPoint) (int64, error) {
	newSequence, err := d.counter.Generate(ctx, TableName())
	if err != nil {
		return 0, err
	}

	accessPoint.AccessPointID = newSequence
	table := &TableAccessPoint{
		Data: accessPoint,
	}

	if _, err = d.client.InsertOne(ctx, table); err != nil {
		return 0, err
	}

	return newSequence, nil
}

func (d *dao) createMany(ctx context.Context, accessPoints []*AccessPoint) ([]int64, error) {
	sequences := make([]int64, len(accessPoints))
	for idx, accessPoint := range accessPoints {
		newSequences, err := d.counter.Generate(ctx, TableName())
		if err != nil {
			return nil, err
		}

		accessPoint.AccessPointID = newSequences
		sequences[idx] = newSequences
	}

	tables := make([]interface{}, 0)
	for _, accessPoint := range accessPoints {
		table := &TableAccessPoint{
			Data: accessPoint,
			BasicInfo: base.BasicInfo{
				CreatedAt: time.Now(),
			},
		}
		tables = append(tables, table)
	}

	_, err := d.client.InsertMany(ctx, tables)
	if err != nil {
		return nil, err
	}

	return sequences, nil
}

func (d *dao) updateMany(ctx context.Context, tenantID string, accessPoints []*AccessPoint) error {
	models := buildUpdateManyParams(tenantID, accessPoints)

	result, err := d.client.BulkWrite(ctx, models)
	if err != nil {
		return err
	}

	if result.MatchedCount > 0 {
		d.logger.Infof("successfully updated accesspoints, update-count(%v)", result.MatchedCount)
	}

	return nil
}

func (d *dao) deleteMany(ctx context.Context, tenantID string, accessPointIDs ...int64) error {
	models := buildDeleteManyParams(tenantID, accessPointIDs...)

	result, err := d.client.BulkWrite(ctx, models)
	if err != nil {
		return err
	}

	if result.MatchedCount > 0 {
		d.logger.Infof("successfully deleted accesspoints, deleted-count(%v)", result.MatchedCount)
	}

	return nil
}

// buildUpdateManyParams build update many params.
func buildUpdateManyParams(tenantID string, accessPoints []*AccessPoint) []mongo.WriteModel {
	models := make([]mongo.WriteModel, 0)

	for _, accessPoint := range accessPoints {
		filter := append(base.AliveFilter(),
			bson.E{Key: FieldKeyAccessPointID, Value: accessPoint.AccessPointID},
			bson.E{Key: FieldKeyTenantID, Value: tenantID})

		update := base.BuildUpsertParam(accessPoint)

		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(false))
	}

	return models
}

// buildDeleteManyParams build delete many params.
func buildDeleteManyParams(tenantID string, accessPointIDs ...int64) []mongo.WriteModel {
	filter := bson.D{
		bson.E{Key: FieldKeyAccessPointID, Value: bson.D{{Key: "$in", Value: accessPointIDs}}},
		bson.E{Key: FieldKeyTenantID, Value: tenantID}}

	update := base.BuildDeleteParam()

	return []mongo.WriteModel{mongo.NewUpdateManyModel().SetFilter(filter).SetUpdate(update).SetUpsert(false)}
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
