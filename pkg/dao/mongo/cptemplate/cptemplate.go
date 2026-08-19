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

// Package cptemplate provides the config policy template data models.
package cptemplate

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func newDao(tenantID string, client *mongo.Database) *dao {
	tableName := TableName(tenantID)
	d := &dao{
		tenantID:  tenantID,
		tableName: tableName,
		client:    client.Collection(tableName),
	}

	d.IOrm = base.NewOrm[*ConfigPolicyTemplate, ConfigPolicyTemplate](d)

	return d
}

type dao struct {
	tenantID  string
	tableName string
	client    *mongo.Collection

	base.IOrm[*ConfigPolicyTemplate, ConfigPolicyTemplate]
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

func (d *dao) upsertMany(nCtx contextx.IContext, templates []*ConfigPolicyTemplate) error {
	models := buildUpsertManyParams(templates)

	result, err := d.client.BulkWrite(nCtx, models)
	if err != nil {
		return err
	}

	if result.UpsertedCount > 0 {
		logger.G.Sys().With("inserted-count", result.UpsertedCount).Info("upserted config policy template")
	}

	if result.MatchedCount > 0 {
		logger.G.Sys().With("matched-count", result.MatchedCount).Info("upserted config policy template")
	}

	return nil
}

func buildUpsertManyParams(templates []*ConfigPolicyTemplate) []mongo.WriteModel {
	models := make([]mongo.WriteModel, 0, len(templates))
	for _, template := range templates {
		filter := bson.D{
			bson.E{Key: FieldKeyConfigPolicyID, Value: template.ConfigPolicyID},
		}

		update := base.BuildUpsertParam(template)

		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	}

	return models
}
