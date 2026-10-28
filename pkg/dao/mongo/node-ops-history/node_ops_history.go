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

// Package nodeopshistory provides persistence for archived out-of-band operations.
package nodeopshistory

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func newDao(tenantID string, client *mongo.Database) *dao {
	tableName := TableName(tenantID)
	d := &dao{
		client:    client.Collection(tableName),
		tableName: tableName,
	}
	d.IOrm = base.NewOrm[*Data, Data](d)

	return d
}

type dao struct {
	client    *mongo.Collection
	tableName string
	base.IOrm[*Data, Data]
}

// GetClient returns the tenant collection.
func (d *dao) GetClient() *mongo.Collection {
	return d.client
}

// GetTableName returns the tenant collection name.
func (d *dao) GetTableName() string {
	return d.tableName
}

// GetIndexes returns query indexes; the ORM adds the operation identity index.
func (d *dao) GetIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys: bson.D{
				{
					Key:   FieldKeyNodeRole,
					Value: 1,
				},
				{
					Key:   FieldKeyBizIDs,
					Value: 1,
				},
				{
					Key:   FieldKeyOperateTime,
					Value: -1,
				},
			},
		},
		{
			Keys: bson.D{
				{
					Key:   FieldKeyOperateTime,
					Value: -1,
				},
				{
					Key:   FieldKeyOperateID,
					Value: 1,
				},
			},
		},
	}
}
