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
	"context"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(tenantID string, client *mongo.Database, logger logger.Logger) *dao {
	return &dao{client: client.Collection(TableName(tenantID)), logger: logger}
}

type dao struct {
	client *mongo.Collection
	logger logger.Logger
}

// ensureIndexes ensures the required indexes for the collection.
func (d *dao) ensureIndexes(ctx context.Context) error {
	var indexes []mongo.IndexModel

	indexes = append(indexes, mongo.IndexModel{
		Keys: bson.D{{Key: "data.biz_id", Value: 1}},
	})

	_, err := d.client.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return err
	}

	d.logger.Infof("successfully created required indexes")

	return nil
}

// ListAll list all host.
func (d *dao) listAll(ctx context.Context) ([]*Host, error) {
	result, err := d.client.Find(ctx, bson.D{{Key: "basic.is_deleted", Value: false}})
	if err != nil {
		return nil, err
	}

	hosts := make([]*Host, 0)
	for result.Next(ctx) {
		table := &TableHost{}
		if err := result.Decode(table); err != nil {
			d.logger.Warnf("failed to decode host, err %v", err)

			continue
		}
		hosts = append(hosts, table.Data)
	}

	return hosts, nil
}

func (d *dao) count(ctx context.Context, filter bson.D) (int64, error) {
	num, err := d.client.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	if num < 0 {
		return 0, fmt.Errorf("count documents get unexpected result: %d", num)
	}

	return num, nil
}

func (d *dao) list(ctx context.Context, filter bson.D, findOpt *options.FindOptions) ([]*Host, error) {
	result, err := d.client.Find(ctx, filter, findOpt)
	if err != nil {
		return nil, err
	}

	bizs := make([]*Host, 0)
	for result.Next(ctx) {
		table := &TableHost{}
		if err := result.Decode(table); err != nil {
			d.logger.Warnf("failed to decode business, err %v", err)

			continue
		}
		bizs = append(bizs, table.Data)
	}

	return bizs, nil
}

// upsertMany upsert many hosts.
func (d *dao) upsertMany(ctx context.Context, hosts []*Host) error {
	models := buildUpsertManyParams(hosts)

	result, err := d.client.BulkWrite(ctx, models)
	if err != nil {
		return err
	}

	if result.UpsertedCount > 0 {
		d.logger.Infof("successfully inserted hosts, inserted-count(%v)", result.UpsertedCount)
	}

	if result.MatchedCount > 0 {
		d.logger.Infof("successfully updated hosts, update-count(%v)", result.MatchedCount)
	}

	return nil
}

// upsertStaticMany upsert many host statics.
func (d *dao) upsertStaticMany(ctx context.Context, hosts []*Host) error {
	models := buildUpsertStaticManyParams(hosts)

	result, err := d.client.BulkWrite(ctx, models)
	if err != nil {
		return err
	}

	if result.UpsertedCount > 0 {
		d.logger.Infof("successfully inserted host statics, inserted-count(%v)", result.UpsertedCount)
	}

	if result.MatchedCount > 0 {
		d.logger.Infof("successfully updated host statics, update-count(%v)", result.MatchedCount)
	}

	return nil
}

// updateDynamicMany update many host dynamic.
func (d *dao) updateDynamicMany(ctx context.Context, hosts []*Host) error {
	models := buildUpdateDynamicManyParams(hosts)

	result, err := d.client.BulkWrite(ctx, models)
	if err != nil {
		return err
	}

	if result.UpsertedCount > 0 {
		d.logger.Infof("successfully inserted host statics, inserted-count(%v)", result.UpsertedCount)
	}

	if result.MatchedCount > 0 {
		d.logger.Infof("successfully updated host statics, update-count(%v)", result.MatchedCount)
	}

	return nil
}

// buildUpsertManyParams build upsert many params.
func buildUpsertManyParams(hosts []*Host) []mongo.WriteModel {
	models := make([]mongo.WriteModel, 0)
	for _, host := range hosts {
		filter := bson.D{{Key: "data.host_id", Value: host.HostID}}

		update := base.BuildUpsertParam(host)

		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	}

	return models
}

// buildUpsertStaticManyParams build upsert static many params.
func buildUpsertStaticManyParams(hosts []*Host) []mongo.WriteModel {
	models := make([]mongo.WriteModel, 0)
	for _, host := range hosts {
		filter := bson.D{{Key: "data.host_id", Value: host.HostID}}

		nowTime := time.Now()
		update := bson.D{
			{
				Key: "$set",
				Value: bson.M{
					"basic.is_deleted": false,
					"basic.updated_at": nowTime,
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
		filter := bson.D{{Key: "data.host_id", Value: host.HostID}}

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
