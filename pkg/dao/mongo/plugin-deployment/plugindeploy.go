/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package plugin-deployment this package is used to store the need data for plugin deployment.
package plugindeployment

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(client *mongo.Database, logger logger.ILogger) *dao {
	d := &dao{client: client.Collection(TableName), logger: logger}
	d.IOrm = base.NewOrm[*Data, Data](d)

	return d
}

type dao struct {
	client *mongo.Collection
	logger logger.ILogger
	base.IOrm[*Data, Data]
}

// GetClient get the dao's client.
func (d *dao) GetClient() *mongo.Collection {
	return d.client
}

// GetLogger get the dao's logger.
func (d *dao) GetLogger() logger.ILogger {
	return d.logger
}

// GetTableName get the dao's table name.
func (d *dao) GetTableName() string {
	return TableName
}

// ExpireTimeSec is the expire time for the data in the database.
const ExpireTimeSec = 7 * 24 * 60

// GetIndexes get the dao's indexes.
func (d *dao) GetIndexes() []mongo.IndexModel {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: FieldKeyExpireAt, Value: 1}},
			Options: mongoOptions.Index().
				SetExpireAfterSeconds(ExpireTimeSec),
		},
	}

	return indexes
}
