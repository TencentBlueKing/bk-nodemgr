/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package counter provides global counter management for dao.
package counter

import (
	"context"
	"errors"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(client *mongo.Database, logger logger.Logger) *dao {
	return &dao{client: client.Collection(TableName()), logger: logger}
}

type dao struct {
	client *mongo.Collection
	logger logger.Logger
}

// ensureIndexes ensures the required indexes for the collection.
func (d *dao) ensureIndexes() error {
	var indexes []mongo.IndexModel

	indexes = append(indexes, mongo.IndexModel{
		Keys: bson.D{{Key: "data.key", Value: 1}},
	})

	_, err := d.client.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return err
	}

	d.logger.Infof("successfully created required indexes")

	return nil
}

func (d *dao) generate(ctx context.Context, key string) (int64, error) {
	filter := append(base.AliveFilter(), bson.E{Key: "data.key", Value: key})
	opts := new(options.FindOneAndUpdateOptions)
	opts.SetUpsert(true)

	result := d.client.FindOneAndUpdate(ctx, filter, buildGenerateParam(), opts)
	if errors.Is(result.Err(), mongo.ErrNoDocuments) {
		return 0, nil
	}
	if err := result.Err(); err != nil {
		d.logger.Errorf("failed to generate counter. key(%s), err: %v", key, err)

		return -1, err
	}

	data := &TableCounter{}
	if err := result.Decode(data); err != nil {
		d.logger.Errorf("failed to decode counter. key(%s), err: %v", key, err)

		return -1, err
	}

	return data.Data.Seqeunce, nil
}

func buildGenerateParam() bson.D {
	nowTime := time.Now()
	update := bson.D{
		{
			Key: "$set",
			Value: bson.M{
				"basic.is_deleted": false,
				"basic.updated_at": nowTime,
			},
		},
		{
			Key: "$setOnInsert",
			Value: bson.M{
				"basic.created_at": nowTime,
			},
		},
		{
			Key: "$inc",
			Value: bson.M{
				"data.sequence": 1,
			},
		},
	}

	return update
}
