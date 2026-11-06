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

// Package processv2 provides storage for the node manager v2 compatible plugin process.
package processv2

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(client *mongo.Database, tableName string) *dao {
	d := &dao{
		client:    client.Collection(tableName),
		tableName: tableName,
		IOrm:      nil,
	}
	d.IOrm = base.NewOrm[*Process, Process](d)

	return d
}

type dao struct {
	client    *mongo.Collection
	tableName string

	base.IOrm[*Process, Process]
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
//
// The process v2 table is always read by host id and plugin name, so it keeps one compound index for those lookups.
func (d *dao) GetIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: FieldKeyHostID, Value: 1},
				{Key: FieldKeyPluginName, Value: 1},
			},
			Options: mongoOptions.Index(),
		},
	}
}
