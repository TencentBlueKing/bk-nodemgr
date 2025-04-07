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
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/counter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(tenantID string, client *mongo.Database, logger logger.Logger) *dao {
	return &dao{
		tenantID: tenantID,
		client:   client.Collection(TableName()),
		logger:   logger,
		counter:  counter.New(client, logger)}
}

type dao struct {
	tenantID string
	client   *mongo.Collection
	logger   logger.Logger
	counter  counter.Handler
}

// nolint:contextcheck
// ensureIndexes ensures the required indexes for the collection.
func (d *dao) ensureIndexes() error {
	var indexes []mongo.IndexModel

	opts := new(options.IndexOptions)
	indexes = append(indexes, mongo.IndexModel{
		Keys:    bson.D{{Key: "data.accesspoint_id", Value: 1}},
		Options: opts.SetUnique(true),
	})

	_, err := d.client.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return err
	}

	d.logger.Infof("created required indexes")

	return nil
}

func (d *dao) count(ctx context.Context, filter bson.D) (int64, error) {
	filter = append(filter, tenantFilter(d.tenantID))

	num, err := d.client.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	if num < 0 {
		return 0, fmt.Errorf("count documents get unexpected result: %d", num)
	}

	return num, nil
}

func (d *dao) list(ctx context.Context, filter bson.D, findOpt *options.FindOptions) ([]*AccessPoint, error) {
	filter = append(filter, tenantFilter(d.tenantID))

	result, err := d.client.Find(ctx, filter, findOpt)
	if err != nil {
		return nil, err
	}

	accesspoints := make([]*AccessPoint, 0)
	for result.Next(ctx) {
		table := &TableAccessPoint{}
		if err := result.Decode(table); err != nil {
			d.logger.Warnf("failed to decode accesspoint, err %v", err)

			continue
		}
		accesspoints = append(accesspoints, table.Data)
	}

	return accesspoints, nil
}

func (d *dao) get(ctx context.Context, filter bson.D) (*AccessPoint, error) {
	filter = append(filter, tenantFilter(d.tenantID))

	result := &TableAccessPoint{}
	err := d.client.FindOne(ctx, filter).Decode(result)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
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

func (d *dao) updateMany(ctx context.Context, accessPoints []*AccessPoint) error {
	models := buildUpdateManyParams(d.tenantID, accessPoints)

	result, err := d.client.BulkWrite(ctx, models)
	if err != nil {
		return err
	}

	if result.MatchedCount > 0 {
		d.logger.Infof("successfully updated accesspoints, update-count(%v)", result.MatchedCount)
	}

	return nil
}

func (d *dao) deleteMany(ctx context.Context, accessPointIDs ...int64) error {
	models := buildDeleteManyParams(d.tenantID, accessPointIDs...)

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
			bson.E{Key: "data.accesspoint_id", Value: accessPoint.AccessPointID},
			bson.E{Key: "data.tenant_id", Value: tenantID})

		update := base.BuildUpsertParam(accessPoint)

		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(false))
	}

	return models
}

// buildDeleteManyParams build delete many params.
func buildDeleteManyParams(tenantID string, accessPointIDs ...int64) []mongo.WriteModel {
	filter := bson.D{
		bson.E{Key: "data.accesspoint_id", Value: bson.D{{Key: "$in", Value: accessPointIDs}}},
		bson.E{Key: "data.tenant_id", Value: tenantID}}

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
			bson.D{{"data.tenant_id", tenantID}},
			bson.D{{"data.networkarea_id", base.GlobalNetworkAreaID}},
		}}
}
