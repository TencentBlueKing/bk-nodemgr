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
		Keys:    bson.D{{Key: "data.networkunit_id", Value: 1}},
		Options: opts.SetUnique(true),
	})

	_, err := d.client.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return err
	}

	d.logger.Infof("successfully created required indexes")

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

func (d *dao) list(ctx context.Context, filter bson.D, findOpt *options.FindOptions) ([]*NetworkUnit, error) {
	filter = append(filter, tenantFilter(d.tenantID))

	result, err := d.client.Find(ctx, filter, findOpt)
	if err != nil {
		return nil, err
	}

	networkUnits := make([]*NetworkUnit, 0)
	for result.Next(ctx) {
		table := &TableNetworkUnit{}
		if err := result.Decode(table); err != nil {
			d.logger.Warnf("failed to decode networkunit, err %v", err)

			continue
		}
		networkUnits = append(networkUnits, table.Data)
	}

	return networkUnits, nil
}

func (d *dao) get(ctx context.Context, filter bson.D) (*NetworkUnit, error) {
	filter = append(filter, tenantFilter(d.tenantID))

	result := &TableNetworkUnit{}
	err := d.client.FindOne(ctx, filter).Decode(result)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
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

func (d *dao) updateMany(ctx context.Context, networkunits []*NetworkUnit) error {
	models := buildUpdateManyParams(d.tenantID, networkunits)

	result, err := d.client.BulkWrite(ctx, models)
	if err != nil {
		return err
	}

	if result.MatchedCount > 0 {
		d.logger.Infof("successfully updated networkunits, update-count(%v)", result.MatchedCount)
	}

	return nil
}

func (d *dao) deleteMany(ctx context.Context, networkUnitIDs ...int64) error {
	models := buildDeleteManyParams(d.tenantID, networkUnitIDs...)

	result, err := d.client.BulkWrite(ctx, models)
	if err != nil {
		return err
	}

	if result.MatchedCount > 0 {
		d.logger.Infof("successfully deleted networkunits, deleted-count(%v)", result.MatchedCount)
	}

	return nil
}

// buildUpdateManyParams build update many params.
func buildUpdateManyParams(tenantID string, networkUnits []*NetworkUnit) []mongo.WriteModel {
	models := make([]mongo.WriteModel, 0)

	for _, networkUnit := range networkUnits {
		filter := append(base.AliveFilter(),
			bson.E{Key: "data.networkunit_id", Value: networkUnit.NetworkUnitID},
			bson.E{Key: "data.tenant_id", Value: tenantID})

		update := base.BuildUpsertParam(networkUnit)

		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(false))
	}

	return models
}

// buildDeleteManyParams build delete many params.
func buildDeleteManyParams(tenantID string, networkUnitIDs ...int64) []mongo.WriteModel {
	filter := bson.D{
		bson.E{Key: "data.networkunit_id", Value: bson.D{{Key: "$in", Value: networkUnitIDs}}},
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
