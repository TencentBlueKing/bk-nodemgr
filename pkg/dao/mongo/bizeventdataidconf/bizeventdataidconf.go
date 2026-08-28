/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package bizeventdataidconf

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(client *mongo.Database, tenantID string) *dao {
	tableName := TableName(tenantID)
	d := &dao{
		client:    client.Collection(tableName),
		tableName: tableName,
	}

	d.IOrm = base.NewOrm[*BizEventDataIDConf](d)

	return d
}

type dao struct {
	client    *mongo.Collection
	tableName string

	base.IOrm[*BizEventDataIDConf, BizEventDataIDConf]
}

// GetClient gets the dao's client.
func (d *dao) GetClient() *mongo.Collection {
	return d.client
}

// GetTableName gets the dao's table name.
func (d *dao) GetTableName() string {
	return d.tableName
}

// GetIndexes gets the dao's indexes.
func (d *dao) GetIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{}
}

func (d *dao) upsertAgentBaseAlarmEventDataID(nCtx contextx.IContext, bkBizID, eventDataID int64) error {
	nowTime := time.Now()
	filter := bson.D{{Key: FieldKeyBizID, Value: bkBizID}}
	update := bson.D{
		{
			Key: "$set",
			Value: bson.M{
				base.FieldKeyIsDeleted:         false,
				base.FieldKeyUpdatedAt:         nowTime,
				FieldKeyBizID:                  bkBizID,
				fieldAgentBaseAlarmEventDataID: eventDataID,
			},
		},
		{
			Key: "$setOnInsert",
			Value: bson.M{
				base.FieldKeyCreatedAt: nowTime,
			},
		},
	}

	_, err := d.client.UpdateOne(nCtx, filter, update, options.Update().SetUpsert(true))

	return err
}
