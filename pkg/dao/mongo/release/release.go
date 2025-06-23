/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package release provides the release models.
package release

import (
	"context"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(client *mongo.Database, logger logger.Logger) *dao {
	d := &dao{
		client:    client.Collection(TableName),
		logger:    logger,
		tableName: TableName,
	}

	d.IOrm = base.NewOrm[*Release, Release](d)

	return d
}

type dao struct {
	client    *mongo.Collection
	tableName string
	logger    logger.Logger

	base.IOrm[*Release, Release]
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
	return d.tableName
}

// GetIndexes get the dao's indexes.
func (d *dao) GetIndexes() []mongo.IndexModel {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: FieldKeyGeneration, Value: 1},
				{Key: FieldKeyType, Value: 1},
				{Key: FieldKeyCPUArch, Value: 1},
				{Key: FieldKeyOSType, Value: 1},
				{Key: FieldKeyVersion, Value: 1}},
			Options: mongoOptions.Index().SetUnique(true),
		},
	}

	return indexes
}

func (d *dao) upsertMany(ctx context.Context, releases []*Release) error {
	models := buildUpsertManyParams(releases)

	result, err := d.client.BulkWrite(ctx, models)
	if err != nil {
		return err
	}

	if result.UpsertedCount > 0 {
		d.logger.Infof("inserted releases, inserted-count(%v)", result.UpsertedCount)
	}

	if result.MatchedCount > 0 {
		d.logger.Infof("updated releases, update-count(%v)", result.MatchedCount)
	}

	return nil
}

func buildUpsertManyParams(releases []*Release) []mongo.WriteModel {
	models := make([]mongo.WriteModel, 0, len(releases))
	for _, release := range releases {
		filter := bson.D{
			{Key: FieldKeyGeneration, Value: release.Generation},
			{Key: FieldKeyType, Value: release.Type},
			{Key: FieldKeyCPUArch, Value: release.CPUArch},
			{Key: FieldKeyOSType, Value: release.OSType},
			{Key: FieldKeyVersion, Value: release.Version},
		}

		update := base.BuildUpsertParam(release)

		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	}

	return models
}
