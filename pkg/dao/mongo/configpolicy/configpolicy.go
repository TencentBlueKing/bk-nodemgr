/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package configpolicy

import (
	"context"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/counter"
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
		counter:   counter.New(client, logger),
	}

	d.IOrm = base.NewOrm[*ConfigPolicy, ConfigPolicy](d)

	return d
}

type dao struct {
	tenantID  string
	tableName string
	client    *mongo.Collection
	logger    logger.Logger
	counter   counter.Handler

	base.IOrm[*ConfigPolicy, ConfigPolicy]
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

func (d *dao) create(ctx context.Context, configPolicy *ConfigPolicy) (int64, error) {
	newSequence, err := d.counter.Generate(ctx, "configpolicy")
	if err != nil {
		return 0, err
	}

	configPolicy.Raw.ConfigPolicyID = newSequence
	if err := d.Create(ctx, configPolicy); err != nil {
		return -1, err
	}

	return newSequence, nil
}

func (d *dao) updateMany(ctx context.Context, tenantID string, configPolicies []*ConfigPolicy) error {
	models := buildUpdateManyParams(tenantID, configPolicies)

	result, err := d.client.BulkWrite(ctx, models)
	if err != nil {
		return err
	}

	if result.MatchedCount > 0 {
		d.logger.Infof("successfully updated config policies, update-count(%v)", result.MatchedCount)
	}

	return nil
}

func (d *dao) deleteMany(ctx context.Context, tenantID string, configPolicyIDs ...int64) error {
	models := buildDeleteManyParams(tenantID, configPolicyIDs...)

	result, err := d.client.BulkWrite(ctx, models)
	if err != nil {
		return err
	}

	if result.MatchedCount > 0 {
		d.logger.Infof("successfully deleted config policies, deleted-count(%v)", result.MatchedCount)
	}

	return nil
}

func (d *dao) setEnabledMany(ctx context.Context, tenantID string, enabled bool, configPolicyIDs ...int64) error {
	filter := bson.D{
		bson.E{Key: FieldKeyConfigPolicyID, Value: bson.D{bson.E{Key: "$in", Value: configPolicyIDs}}},
		bson.E{Key: FieldKeyTenantID, Value: tenantID}}

	return d.UpdateField(ctx, filter, FieldKeyEnabled, enabled)
}

// buildUpdateManyParams build update many params.
func buildUpdateManyParams(tenantID string, configPolicies []*ConfigPolicy) []mongo.WriteModel {
	models := make([]mongo.WriteModel, 0)

	for _, configPolicy := range configPolicies {
		filter := append(base.AliveFilter(),
			bson.E{Key: FieldKeyConfigPolicyID, Value: configPolicy.Raw.ConfigPolicyID},
			bson.E{Key: FieldKeyTenantID, Value: tenantID})

		nowTime := time.Now()
		update := bson.D{
			bson.E{
				Key: "$set",
				Value: bson.M{
					"basic.is_deleted": false,
					"basic.updated_at": nowTime,
					"data.raw":         configPolicy.Raw,
				},
			},
			bson.E{
				Key: "$inc",
				Value: bson.M{
					FieldKeyVersion: 1,
				},
			},
		}

		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(false))
	}

	return models
}

// buildDeleteManyParams build delete many params.
func buildDeleteManyParams(tenantID string, configPolicyIDs ...int64) []mongo.WriteModel {
	filter := bson.D{
		bson.E{Key: FieldKeyConfigPolicyID, Value: bson.D{bson.E{Key: "$in", Value: configPolicyIDs}}},
		bson.E{Key: FieldKeyTenantID, Value: tenantID}}

	update := base.BuildDeleteParam()

	return []mongo.WriteModel{mongo.NewUpdateManyModel().SetFilter(filter).SetUpdate(update).SetUpsert(false)}
}
