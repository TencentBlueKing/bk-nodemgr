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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/counter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func newDao(client *mongo.Database) *dao {
	d := &dao{
		client:  client.Collection(TableName()),
		counter: counter.New(client)}

	d.IOrm = base.NewOrm[*AccessPoint, AccessPoint](d)

	return d
}

type dao struct {
	client *mongo.Collection

	counter counter.Handler
	base.IOrm[*AccessPoint, AccessPoint]
}

// GetClient get the dao's client.
func (d *dao) GetClient() *mongo.Collection {
	return d.client
}

// GetTableName get the dao's table name.
func (d *dao) GetTableName() string {
	return TableName()
}

// nolint:contextcheck
// ensureIndexes ensures the required indexes for the collection.
func (d *dao) GetIndexes() []mongo.IndexModel {
	var indexes []mongo.IndexModel

	return indexes
}

func (d *dao) create(nCtx contextx.IContext, accessPoint *AccessPoint) (int64, error) {
	newSequence, err := d.counter.Generate(nCtx, TableName())
	if err != nil {
		return 0, err
	}

	accessPoint.AccessPointID = newSequence
	table := &TableAccessPoint{
		Data: accessPoint,
	}

	if _, err = d.client.InsertOne(nCtx, table); err != nil {
		return 0, err
	}

	return newSequence, nil
}

func (d *dao) createMany(nCtx contextx.IContext, accessPoints []*AccessPoint) ([]int64, error) {
	sequences := make([]int64, len(accessPoints))
	for idx, accessPoint := range accessPoints {
		newSequences, err := d.counter.Generate(nCtx, TableName())
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

	_, err := d.client.InsertMany(nCtx, tables)
	if err != nil {
		return nil, err
	}

	return sequences, nil
}

func (d *dao) updateMany(nCtx contextx.IContext, tenantID string, accessPoints []*AccessPoint) error {
	models := buildUpdateManyParams(tenantID, accessPoints)

	result, err := d.client.BulkWrite(nCtx, models)
	if err != nil {
		return err
	}

	if result.MatchedCount > 0 {
		logger.G.Sys().With("matched-count", result.MatchedCount).Info("upserted accesspoints")
	}

	return nil
}

func (d *dao) deleteMany(nCtx contextx.IContext, tenantID string, accessPointIDs ...int64) error {
	models := buildDeleteManyParams(tenantID, accessPointIDs...)

	result, err := d.client.BulkWrite(nCtx, models)
	if err != nil {
		return err
	}

	if result.MatchedCount > 0 {
		logger.G.Sys().With("deleted-count", result.MatchedCount).Info("deleted accesspoints")
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

func tenantFilter(tenantID string) bson.E {
	return bson.E{Key: FieldKeyTenantID, Value: tenantID}
}
