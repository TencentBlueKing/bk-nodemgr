/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

// Package counter provides global counter management for dao.
package counter

import (
	"context"
	"errors"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(client *mongo.Database) *dao {
	return &dao{client: client.Collection(TableName())}
}

type dao struct {
	client *mongo.Collection
}

// nolint:contextcheck
// ensureIndexes ensures the required indexes for the collection.
func (d *dao) ensureIndexes() error {
	var indexes []mongo.IndexModel

	indexes = append(indexes, mongo.IndexModel{
		Keys: bson.D{{Key: FieldKeyKey, Value: 1}},
	})

	_, err := d.client.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return err
	}

	logger.G.Sys().With("table", TableName(), "indexes", indexes).Info("created required indexes")

	return nil
}

func (d *dao) generate(nCtx contextx.IContext, key string) (int64, error) {
	return d.generateN(nCtx, key, 1)
}

// generateN atomically increments the counter by n and returns the value before the increment.
func (d *dao) generateN(nCtx contextx.IContext, key string, n int64) (int64, error) {
	filter := append(base.AliveFilter(), bson.E{Key: FieldKeyKey, Value: key})
	opts := new(options.FindOneAndUpdateOptions)
	opts.SetUpsert(true)

	result := d.client.FindOneAndUpdate(nCtx, filter, buildGenerateParam(n), opts)
	if errors.Is(result.Err(), mongo.ErrNoDocuments) {
		return 0, nil
	}
	if err := result.Err(); err != nil {
		logger.G.Sys().WithErr(err).With("key", key).Info("failed to generate counter")

		return -1, err
	}

	data := &TableCounter{}
	if err := result.Decode(data); err != nil {
		logger.G.Sys().WithErr(err).With("key", key).Info("failed to decode counter")

		return -1, err
	}

	return data.Data.Seqeunce, nil
}

func buildGenerateParam(n int64) bson.D {
	nowTime := time.Now()
	update := bson.D{
		{
			Key: "$set",
			Value: bson.M{
				base.FieldKeyIsDeleted: false,
				base.FieldKeyUpdatedAt: nowTime,
			},
		},
		{
			Key: "$setOnInsert",
			Value: bson.M{
				base.FieldKeyCreatedAt: nowTime,
			},
		},
		{
			Key: "$inc",
			Value: bson.M{
				FieldKeySequence: n,
			},
		},
	}

	return update
}
