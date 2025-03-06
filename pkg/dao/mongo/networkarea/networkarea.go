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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(tenantID string, client *mongo.Database, logger logger.Logger) *dao {
	return &dao{tenantID: tenantID, client: client.Collection(TableName()), logger: logger}
}

type dao struct {
	tenantID string
	client   *mongo.Collection
	logger   logger.Logger
}

// ensureIndexes ensures the required indexes for the collection.
func (d *dao) ensureIndexes() error {
	var indexes []mongo.IndexModel

	opts := new(options.IndexOptions)
	indexes = append(indexes, mongo.IndexModel{
		Keys:    bson.D{{Key: "data.networkarea_id", Value: 1}},
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

func (d *dao) list(ctx context.Context, filter bson.D, findOpt *options.FindOptions) ([]*NetworkArea, error) {
	filter = append(filter, tenantFilter(d.tenantID))

	result, err := d.client.Find(ctx, filter, findOpt)
	if err != nil {
		return nil, err
	}

	networkAreas := make([]*NetworkArea, 0)
	for result.Next(ctx) {
		table := &TableNetworkArea{}
		if err := result.Decode(table); err != nil {
			d.logger.Warnf("failed to decode networkarea, err %v", err)

			continue
		}
		networkAreas = append(networkAreas, table.Data)
	}

	return networkAreas, nil
}

func (d *dao) get(ctx context.Context, filter bson.D) (*NetworkArea, error) {
	filter = append(filter, tenantFilter(d.tenantID))

	result := &TableNetworkArea{}
	err := d.client.FindOne(ctx, filter).Decode(result)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (d *dao) upsertMany(ctx context.Context, networkAreas []*NetworkArea) error {
	models := buildUpsertManyParams(d.tenantID, networkAreas)

	result, err := d.client.BulkWrite(ctx, models)
	if err != nil {
		return err
	}

	if result.UpsertedCount > 0 {
		d.logger.Infof("successfully inserted networkareas, inserted-count(%v)", result.UpsertedCount)
	}

	if result.MatchedCount > 0 {
		d.logger.Infof("successfully updated networkareas, update-count(%v)", result.MatchedCount)
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
		d.logger.Infof("successfully updated networkareas, update-count(%v)", result.MatchedCount)
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
		d.logger.Infof("successfully deleted networkareas, deleted-count(%v)", result.MatchedCount)
	}

	return nil
}

// buildUpsertManyParams build upsert many params.
func buildUpsertManyParams(tenantID string, networkAreas []*NetworkArea) []mongo.WriteModel {
	models := make([]mongo.WriteModel, 0)

	for _, networkarea := range networkAreas {
		filter := bson.D{
			bson.E{Key: "data.networkarea_id", Value: networkarea.NetworkAreaID},
			bson.E{Key: "data.tenant_id", Value: tenantID},
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
			bson.E{Key: "data.networkarea_id", Value: networkarea.NetworkAreaID},
			bson.E{Key: "data.tenant_id", Value: tenantID})

		update := base.BuildUpsertParam(networkarea)

		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(false))
	}

	return models
}

// buildDeleteManyParams build delete many params.
func buildDeleteManyParams(tenantID string, networkAreaIDs ...int64) []mongo.WriteModel {
	filter := bson.D{
		bson.E{Key: "data.networkarea_id", Value: bson.D{{Key: "$in", Value: networkAreaIDs}}},
		bson.E{Key: "data.tenant_id", Value: tenantID}}

	update := base.BuildDeleteParam()

	return []mongo.WriteModel{mongo.NewUpdateManyModel().SetFilter(filter).SetUpdate(update).SetUpsert(false)}
}

// tenantFilter additional tenant filter.
// all tenants can filter global networkarea in query methods.
// and the networkareas in their own tenant.
func tenantFilter(tenantID string) bson.E {
	return bson.E{
		Key: "$or",
		Value: bson.A{
			bson.D{{"data.tenant_id", tenantID}},
			bson.D{{"data.networkarea_id", base.GlobalNetworkAreaID}},
		}}
}
