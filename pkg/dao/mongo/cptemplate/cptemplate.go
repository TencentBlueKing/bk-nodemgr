/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cptemplate

import (
	"context"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func newDao(tenantID string, client *mongo.Database, logger logger.Logger) *dao {
	tableName := TableName(tenantID)
	d := &dao{
		tenantID:  tenantID,
		tableName: tableName,
		client:    client.Collection(tableName),
		logger:    logger,
	}

	d.IOrm = base.NewOrm[*ConfigPolicyTemplate, ConfigPolicyTemplate](d)

	return d
}

type dao struct {
	tenantID  string
	tableName string
	client    *mongo.Collection
	logger    logger.Logger

	base.IOrm[*ConfigPolicyTemplate, ConfigPolicyTemplate]
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
			Keys: bson.D{bson.E{Key: FieldKeyConfigPolicyID, Value: 1}},
		},
	}

	return indexes
}

func (d *dao) upsertMany(ctx context.Context, templates []*ConfigPolicyTemplate) error {
	models := buildUpsertManyParams(templates)

	result, err := d.client.BulkWrite(ctx, models)
	if err != nil {
		return err
	}

	if result.UpsertedCount > 0 {
		d.logger.Infof("inserted config policy template, inserted-count(%v)", result.UpsertedCount)
	}

	if result.MatchedCount > 0 {
		d.logger.Infof("updated config policy template, update-count(%v)", result.MatchedCount)
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
