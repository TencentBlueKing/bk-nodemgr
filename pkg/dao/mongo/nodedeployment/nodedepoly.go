/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package nodedeployment this package is used to store the need data for node deployment.
package nodedeployment

import (
	"context"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/counter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(client *mongo.Database, logger logger.Logger) *dao {
	d := &dao{client: client.Collection(TableName), logger: logger}
	d.baseOrm = base.NewOrm[*NodeDeployment, NodeDeployment](d)

	return d
}

type dao struct {
	client  *mongo.Collection
	logger  logger.Logger
	counter counter.Handler
	baseOrm base.IOrm[*NodeDeployment, NodeDeployment]
}

// GetClient get the dao's client.
func (d *dao) GetClient() *mongo.Collection {
	return d.client
}

// GetLogger get the dao's logger.
func (d *dao) GetLogger() logger.Logger {
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
			Keys:    bson.D{{Key: FieldKeyToken, Value: 1}},
			Options: new(mongoOptions.IndexOptions).SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "data.expire_at", Value: 1}},
			Options: mongoOptions.Index().
				SetExpireAfterSeconds(ExpireTimeSec),
		},
	}

	return indexes
}

func (d *dao) create(ctx context.Context, nodeDeployment *NodeDeployment) error {
	return d.baseOrm.Create(ctx, nodeDeployment)
}

func (d *dao) get(ctx context.Context, filter bson.D, fields ...string) (*NodeDeployment, error) {
	return d.baseOrm.Get(ctx, filter, fields...)
}

func (d *dao) updateField(ctx context.Context, filter bson.D, field string, value any) error {
	return d.baseOrm.UpdateField(ctx, filter, field, value)
}

func (d *dao) ensureIndexes() error {
	return d.baseOrm.EnsureIndexes()
}
