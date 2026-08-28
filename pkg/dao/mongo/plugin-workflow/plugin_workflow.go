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

// Package pluginworkflow is the workflow definition for plugin workflow manager.
package pluginworkflow

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(client *mongo.Database, tenantID string) *dao {
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
			Keys: bson.D{{Key: FieldKeyDeployPolicyIDs, Value: 1}},
		},
	}

	return indexes
}

// distinctString distinct string field.
func (d *dao) distinctString(
	nCtx contextx.IContext, key string, filter bson.D, distinctOpt *mongoOptions.DistinctOptions) ([]string, error) {

	return d.DistinctString(nCtx, key, filter, distinctOpt)
}

// distinctInt64 distinct int64 field.
func (d *dao) distinctInt64(
	nCtx contextx.IContext, key string, filter bson.D, distinctOpt *mongoOptions.DistinctOptions) ([]int64, error) {

	return d.DistinctInt64(nCtx, key, filter, distinctOpt)
}
