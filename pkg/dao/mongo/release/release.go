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

// Package release provides the release models.
package release

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func newDao(tableName string, client *mongo.Database) *dao {
	d := &dao{
		client:    client.Collection(tableName),
		tableName: tableName,
	}

	d.IOrm = base.NewOrm[*Release, Release](d)

	return d
}

type dao struct {
	client    *mongo.Collection
	tableName string

	base.IOrm[*Release, Release]
}

// GetClient get the dao's client.
func (d *dao) GetClient() *mongo.Collection {
	return d.client
}

// GetTableName get the dao's table name.
func (d *dao) GetTableName() string {
	return d.tableName
}

// GetIndexes get the dao's indexes.
func (d *dao) GetIndexes() []mongo.IndexModel {
	var indexes []mongo.IndexModel

	return indexes
}

func (d *dao) upsertMany(nCtx contextx.IContext, releases []*Release) error {
	models := buildUpsertManyParams(releases)

	result, err := d.client.BulkWrite(nCtx, models)
	if err != nil {
		return err
	}

	if result.UpsertedCount > 0 {
		logger.G.Sys().With("inserted-count", result.UpsertedCount).Info("upserted releases")
	}

	if result.MatchedCount > 0 {
		logger.G.Sys().With("matched-count", result.MatchedCount).Info("upserted releases")
	}

	return nil
}

func buildUpsertManyParams(releases []*Release) []mongo.WriteModel {
	models := make([]mongo.WriteModel, 0, len(releases))
	for _, release := range releases {
		filter := bson.D{
			bson.E{Key: FieldKeyName, Value: release.Name},
			bson.E{Key: FieldKeyGeneration, Value: release.Generation},
			bson.E{Key: FieldKeyType, Value: release.Type},
			bson.E{Key: FieldKeyCPUArch, Value: release.CPUArch},
			bson.E{Key: FieldKeyOSType, Value: release.OSType},
			bson.E{Key: FieldKeyVersion, Value: release.Version},
		}

		update := base.BuildUpsertParam(release)

		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	}

	return models
}
