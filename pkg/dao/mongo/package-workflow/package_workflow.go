/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package packageworkflow is the DAO for package workflow lifecycle records.
package packageworkflow

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

// GetClient gets the dao client.
func (d *dao) GetClient() *mongo.Collection {
	return d.client
}

// GetTableName gets the dao table name.
func (d *dao) GetTableName() string {
	return d.tableName
}

// GetIndexes gets the dao indexes.
func (d *dao) GetIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{Keys: bson.D{{Key: FieldKeyTriggerID, Value: 1}}},
	}
}
