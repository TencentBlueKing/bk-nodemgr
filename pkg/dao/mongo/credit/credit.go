/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package credit

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(tenantID string, client *mongo.Database) *dao {
	tableName := TableName(tenantID)
	d := &dao{
		client:    client.Collection(tableName),
		tableName: tableName,
	}

	d.IOrm = base.NewOrm[*Credit, Credit](d)

	return d
}

type dao struct {
	client    *mongo.Collection
	tableName string
	base.IOrm[*Credit, Credit]
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
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: FieldKeyExpireAt, Value: 1}},
			Options: mongoOptions.Index().
				SetExpireAfterSeconds(0),
		},
	}

	return indexes
}

// upsert updates or inserts a credit.
func (d *dao) upsert(nCtx contextx.IContext, credit *Credit) error {
	filter, upsert, opts := buildUpsertParams(credit)
	result, err := d.client.UpdateOne(nCtx, filter, upsert, opts)
	if err != nil {
		return err
	}

	switch {
	case result.UpsertedCount > 0:
		logger.G.Biz(nCtx).With("unique-key", credit.UniqueKey(), "table", d.tableName).Info("upserted credit")

	case result.MatchedCount > 0:
		logger.G.Biz(nCtx).With("unique-key", credit.UniqueKey(), "table", d.tableName).Info("updated credit")

	default:
		logger.G.Biz(nCtx).With("unique-key", credit.UniqueKey(), "table", d.tableName).Warn("try to upsert credit but no changes made")
	}

	return nil
}

// buildUpsertParams build update params.
func buildUpsertParams(credit *Credit) (bson.D, bson.D, *mongoOptions.UpdateOptions) {
	// update business by credit_id.
	filter := bson.D{{
		Key: FieldKeyCreditID, Value: credit.CreditID},
		{
			Key:   FieldKeyTenantID,
			Value: credit.TenantID,
		},
	}

	// insert as creation or update data only.
	update := base.BuildUpsertParam(credit)

	// do upsert.
	opts := mongoOptions.Update().SetUpsert(true)

	return filter, update, opts
}
