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

// Package configpolicy provides the config policy data models.
package configpolicy

import (
	"errors"
	"fmt"
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

func (d *dao) create(nCtx contextx.IContext, configPolicy *ConfigPolicy) (*ConfigPolicy, error) {
	newSequence, err := d.counter.Generate(nCtx, tableNamePrefix)
	if err != nil {
		return nil, err
	}

	prioritySeq, err := d.counter.Generate(nCtx, counterKeyPriority)
	if err != nil {
		return nil, err
	}

	configPolicy.Raw.ConfigPolicyID = newSequence
	configPolicy.Raw.Priority = prioritySeq + 1
	if err := d.Create(nCtx, configPolicy); err != nil {
		return nil, err
	}

	return configPolicy, nil
}

func (d *dao) updateMany(nCtx contextx.IContext, updates []*base.DocumentFieldUpdate) ([]*ConfigPolicy, error) {
	policies := make([]*ConfigPolicy, 0, len(updates))
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
		policy, err := d.updateWithSnapshot(nCtx, update.Filter, updateDoc, options.After)
		if errors.Is(err, mongo.ErrNoDocuments) {
			continue
		}
		if err != nil {
			return policies, err
		}
		policies = append(policies, policy)
	}

	if len(policies) > 0 {
		logger.G.Sys().With("matched-count", len(policies)).Info("updated config policies")
	}

	return policies, nil
}

func (d *dao) deleteMany(nCtx contextx.IContext, tenantID string, configPolicyIDs ...int64) ([]*ConfigPolicy, error) {
	policies := make([]*ConfigPolicy, 0, len(configPolicyIDs))
	for _, id := range configPolicyIDs {
		filter := configPolicyMutationFilter(tenantID, id)
		// Preserve soft deletion and return the policy immediately before it is deleted.
		policy, err := d.updateWithSnapshot(nCtx, filter, base.BuildDeleteParam(), options.Before)
		if errors.Is(err, mongo.ErrNoDocuments) {
			continue
		}
		if err != nil {
			return policies, err
		}
		policies = append(policies, policy)
	}

	if len(policies) > 0 {
		logger.G.Sys().With("deleted-count", len(policies)).Info("deleted config policies")
	}

	return policies, nil
}

func (d *dao) enableMany(nCtx contextx.IContext, tenantID string, configPolicyIDs ...int64) ([]*ConfigPolicy, error) {
	n := int64(len(configPolicyIDs))
	seqBase, err := d.counter.GenerateN(nCtx, counterKeyPriority, n)
	if err != nil {
		return nil, err
	}

	policies := make([]*ConfigPolicy, 0, len(configPolicyIDs))
	for i, id := range configPolicyIDs {
		update := bson.D{{Key: "$set", Value: bson.M{
			base.FieldKeyIsDeleted: false,
			base.FieldKeyUpdatedAt: time.Now(),
			FieldKeyEnabled:        true,
			FieldKeyPriority:       seqBase + int64(i) + 1,
		}}}
		policy, err := d.updateWithSnapshot(nCtx, configPolicyMutationFilter(tenantID, id), update, options.After)
		if errors.Is(err, mongo.ErrNoDocuments) {
			continue
		}
		if err != nil {
			return policies, err
		}
		policies = append(policies, policy)
	}

	return policies, nil
}

func (d *dao) disableMany(nCtx contextx.IContext, tenantID string, priorityDisabled int64, configPolicyIDs ...int64) ([]*ConfigPolicy, error) {
	policies := make([]*ConfigPolicy, 0, len(configPolicyIDs))
	for _, id := range configPolicyIDs {
		update := bson.D{{Key: "$set", Value: bson.M{
			base.FieldKeyIsDeleted: false,
			base.FieldKeyUpdatedAt: time.Now(),
			FieldKeyEnabled:        false,
			FieldKeyPriority:       priorityDisabled,
		}}}
		policy, err := d.updateWithSnapshot(nCtx, configPolicyMutationFilter(tenantID, id), update, options.After)
		if errors.Is(err, mongo.ErrNoDocuments) {
			continue
		}
		if err != nil {
			return policies, err
		}
		policies = append(policies, policy)
	}

	return policies, nil
}

func (d *dao) updatePriorityMany(nCtx contextx.IContext, tenantID string, priorities map[int64]int64) ([]*ConfigPolicy, error) {
	policies := make([]*ConfigPolicy, 0, len(priorities))
	for id, priority := range priorities {
		update := bson.D{{Key: "$set", Value: bson.M{
			base.FieldKeyIsDeleted: false,
			base.FieldKeyUpdatedAt: time.Now(),
			FieldKeyPriority:       priority,
		}}}
		policy, err := d.updateWithSnapshot(nCtx, configPolicyMutationFilter(tenantID, id), update, options.After)
		if errors.Is(err, mongo.ErrNoDocuments) {
			continue
		}
		if err != nil {
			return policies, err
		}
		policies = append(policies, policy)
	}

	return policies, nil
}

func configPolicyMutationFilter(tenantID string, id int64) bson.D {
	return append(base.AliveFilter(),
		bson.E{Key: FieldKeyConfigPolicyID, Value: id},
		bson.E{Key: FieldKeyTenantID, Value: tenantID})
}

// updateWithSnapshot preserves ErrNoDocuments so batch callers can skip unmatched policies.
func (d *dao) updateWithSnapshot(nCtx contextx.IContext, filter, update bson.D, returnDocument options.ReturnDocument) (*ConfigPolicy, error) {
	result := d.client.FindOneAndUpdate(nCtx, filter, update,
		options.FindOneAndUpdate().SetReturnDocument(returnDocument).SetUpsert(false))
	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("failed to mutate config policy: %w", err)
	}

	table := new(TableConfigPolicy)
	if err := result.Decode(table); err != nil {
		return nil, fmt.Errorf("failed to decode config policy snapshot: %w", err)
	}
	if table.Data == nil {
		return nil, errors.New("config policy snapshot is missing data")
	}

	return table.Data, nil
}
