/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package business provides business dao operations.
package business

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(tenantID string, client *mongo.Database, logger logger.Logger) *dao {
	return &dao{client: client.Collection(TableName(tenantID)), logger: logger}
}

type dao struct {
	client *mongo.Collection
	logger logger.Logger
}

// upsert updates or inserts a business.
func (d *dao) upsert(ctx context.Context, biz *Business) error {
	filter, upsert, opts := buildUpsertParams(biz)
	result, err := d.client.UpdateOne(ctx, filter, upsert, opts)
	if err != nil {
		return err
	}

	switch {
	case result.UpsertedCount > 0:
		{
			d.logger.Infof("successfully upserted business, unique-key(%s)", biz.UniqueKey())
		}
	case result.MatchedCount > 0:
		{
			d.logger.Infof("successfully updated business, unique-key(%s)", biz.UniqueKey())
		}
	default:
		d.logger.Warnf("try to upsert business but no changes made, unique-key(%s", biz.UniqueKey())
	}

	return nil
}

// buildUpsertParams build update params.
func buildUpsertParams(biz *Business) (bson.D, bson.D, *mongoOptions.UpdateOptions) {
	// update business by biz_id.
	filter := bson.D{{Key: "data.biz_id", Value: biz.BizID}}

	// insert as creation or update data only.
	update := base.BuildUpsertParam(biz)

	// do upsert.
	opts := mongoOptions.Update().SetUpsert(true)

	return filter, update, opts
}

// ListAll list all business.
func (d *dao) listAll(ctx context.Context) ([]*Business, error) {
	result, err := d.client.Find(ctx, bson.D{{Key: "basic.is_deleted", Value: false}})
	if err != nil {
		return nil, err
	}

	bizs := make([]*Business, 0)
	for result.Next(ctx) {
		table := &TableBusiness{}
		if err := result.Decode(table); err != nil {
			d.logger.Warnf("failed to decode business, err %v", err)

			continue
		}
		bizs = append(bizs, table.Data)
	}

	return bizs, nil
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

func (d *dao) list(ctx context.Context, filter bson.D, findOpt *options.FindOptions) ([]*Business, error) {
	result, err := d.client.Find(ctx, filter, findOpt)
	if err != nil {
		return nil, err
	}

	hosts := make([]*Business, 0)
	for result.Next(ctx) {
		table := &TableBusiness{}
		if err := result.Decode(table); err != nil {
			d.logger.Warnf("failed to decode host, err %v", err)

			continue
		}
		hosts = append(hosts, table.Data)
	}

	return hosts, nil
}

// upsertMany upsert many business.
func (d *dao) upsertMany(ctx context.Context, bizs []*Business) error {
	models := buildUpsertManyParams(bizs)

	result, err := d.client.BulkWrite(ctx, models)
	if err != nil {
		return err
	}

	if result.UpsertedCount > 0 {
		d.logger.Infof("successfully inserted bizs, inserted-count(%v)", result.UpsertedCount)
	}

	if result.MatchedCount > 0 {
		d.logger.Infof("successfully updated bizs, update-count(%v)", result.MatchedCount)
	}

	return nil
}

// buildUpsertManyParams build upsert many params.
func buildUpsertManyParams(bizs []*Business) []mongo.WriteModel {
	models := make([]mongo.WriteModel, 0, len(bizs))

	for _, biz := range bizs {
		filter := bson.D{{Key: "data.biz_id", Value: biz.BizID}}

		update := base.BuildUpsertParam(biz)

		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	}

	return models
}
