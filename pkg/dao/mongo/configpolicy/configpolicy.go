/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package configpolicy provides the config policy data models.
package configpolicy

import (
	"maps"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/counter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(tenantID string, client *mongo.Database) *dao {
	tableName := TableName(tenantID)
	d := &dao{
		tenantID:  tenantID,
		tableName: tableName,
		client:    client.Collection(tableName),
		counter:   counter.New(client),
	}

	d.IOrm = base.NewOrm[*ConfigPolicy, ConfigPolicy](d)

	return d
}

type dao struct {
	tenantID  string
	tableName string
	client    *mongo.Collection

	counter counter.Handler

	base.IOrm[*ConfigPolicy, ConfigPolicy]
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
	return []mongo.IndexModel{
		{
			Keys: bson.D{{Key: FieldKeyPriority, Value: 1}},
			Options: options.Index().SetPartialFilterExpression(bson.D{
				{Key: base.FieldKeyIsDeleted, Value: false},
			}),
		},
	}
}

func (d *dao) create(nCtx contextx.IContext, configPolicy *ConfigPolicy) (int64, error) {
	newSequence, err := d.counter.Generate(nCtx, tableNamePrefix)
	if err != nil {
		return 0, err
	}

	prioritySeq, err := d.counter.Generate(nCtx, counterKeyPriority)
	if err != nil {
		return 0, err
	}

	configPolicy.Raw.ConfigPolicyID = newSequence
	configPolicy.Raw.Priority = prioritySeq + 1
	if err := d.Create(nCtx, configPolicy); err != nil {
		return -1, err
	}

	return newSequence, nil
}

func (d *dao) updateMany(nCtx contextx.IContext, updates []*base.DocumentFieldUpdate) error {
	models := make([]mongo.WriteModel, 0, len(updates))
	for _, update := range updates {
		nowTime := time.Now()
		setFields := bson.M{
			base.FieldKeyIsDeleted: false,
			base.FieldKeyUpdatedAt: nowTime,

			// configpolicy need to show update time
			FieldKeyUpdatedAt: nowTime,
		}
		maps.Copy(setFields, update.Fields)

		updateDoc := bson.D{
			{Key: "$set", Value: setFields},
			{Key: "$inc", Value: bson.M{FieldKeyVersion: 1}},
		}
		models = append(models, mongo.NewUpdateOneModel().
			SetFilter(update.Filter).
			SetUpdate(updateDoc).
			SetUpsert(false))
	}

	result, err := d.client.BulkWrite(nCtx, models)
	if err != nil {
		return err
	}

	if result.MatchedCount > 0 {
		logger.G.Sys().With("matched-count", result.MatchedCount).Info("updated config policies")
	}

	return nil
}

func (d *dao) deleteMany(nCtx contextx.IContext, tenantID string, configPolicyIDs ...int64) error {
	models := buildDeleteManyParams(tenantID, configPolicyIDs...)

	result, err := d.client.BulkWrite(nCtx, models)
	if err != nil {
		return err
	}

	if result.MatchedCount > 0 {
		logger.G.Sys().With("deleted-count", result.MatchedCount).Info("deleted config policies")
	}

	return nil
}

func (d *dao) enableMany(nCtx contextx.IContext, tenantID string, configPolicyIDs ...int64) error {
	n := int64(len(configPolicyIDs))
	seqBase, err := d.counter.GenerateN(nCtx, counterKeyPriority, n)
	if err != nil {
		return err
	}

	models := buildEnableManyParams(tenantID, seqBase, configPolicyIDs...)

	_, err = d.client.BulkWrite(nCtx, models)

	return err
}

func (d *dao) disableMany(nCtx contextx.IContext, tenantID string, priorityDisabled int64, configPolicyIDs ...int64) error {
	filter := append(base.AliveFilter(),
		bson.E{Key: FieldKeyConfigPolicyID, Value: bson.D{{Key: "$in", Value: configPolicyIDs}}},
		bson.E{Key: FieldKeyTenantID, Value: tenantID},
	)

	nowTime := time.Now()
	update := bson.D{
		{Key: "$set", Value: bson.M{
			base.FieldKeyIsDeleted: false,
			base.FieldKeyUpdatedAt: nowTime,
			FieldKeyEnabled:        false,
			FieldKeyPriority:       priorityDisabled,
		}},
	}

	_, err := d.client.UpdateMany(nCtx, filter, update)

	return err
}

func (d *dao) updatePriorityMany(nCtx contextx.IContext, tenantID string, priorities map[int64]int64) error {
	if len(priorities) == 0 {
		return nil
	}

	models := buildUpdatePriorityManyParams(tenantID, priorities)

	_, err := d.client.BulkWrite(nCtx, models)

	return err
}

func buildEnableManyParams(tenantID string, seqBase int64, configPolicyIDs ...int64) []mongo.WriteModel {
	models := make([]mongo.WriteModel, 0, len(configPolicyIDs))

	for i, id := range configPolicyIDs {
		filter := append(base.AliveFilter(),
			bson.E{Key: FieldKeyConfigPolicyID, Value: id},
			bson.E{Key: FieldKeyTenantID, Value: tenantID})

		nowTime := time.Now()
		update := bson.D{
			{Key: "$set", Value: bson.M{
				base.FieldKeyIsDeleted: false,
				base.FieldKeyUpdatedAt: nowTime,
				FieldKeyEnabled:        true,
				FieldKeyPriority:       seqBase + int64(i) + 1,
			}},
		}

		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(false))
	}

	return models
}
func buildUpdatePriorityManyParams(tenantID string, priorities map[int64]int64) []mongo.WriteModel {
	models := make([]mongo.WriteModel, 0, len(priorities))

	for id, priority := range priorities {
		filter := append(base.AliveFilter(),
			bson.E{Key: FieldKeyConfigPolicyID, Value: id},
			bson.E{Key: FieldKeyTenantID, Value: tenantID})

		nowTime := time.Now()
		update := bson.D{
			{Key: "$set", Value: bson.M{
				base.FieldKeyIsDeleted: false,
				base.FieldKeyUpdatedAt: nowTime,
				FieldKeyPriority:       priority,
			}},
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
