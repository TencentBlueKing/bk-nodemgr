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

// Package deploypolicy this package is used to store the need data for deploy policy.
package deploypolicy

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/counter"
	"go.mongodb.org/mongo-driver/mongo"
)

func newDao(client *mongo.Database, tableName string) *dao {
	d := &dao{
		client:    client.Collection(tableName),
		tableName: tableName,
		counter:   counter.New(client),
		IOrm:      nil,
	}
	d.IOrm = base.NewOrm[*DeployPolicy, DeployPolicy](d)

	return d
}

type dao struct {
	client    *mongo.Collection
	tableName string

	counter counter.Handler

	base.IOrm[*DeployPolicy, DeployPolicy]
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
