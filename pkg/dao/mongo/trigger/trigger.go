/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package trigger ...
package trigger

import (
	"context"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(client *mongo.Database, logger logger.ILogger) *dao {
	d := &dao{
		client:    client.Collection(TableName),
		logger:    logger,
		tableName: TableName,
	}

	d.IOrm = base.NewOrm[*Trigger, Trigger](d)

	return d
}

type dao struct {
	client    *mongo.Collection
	tableName string
	logger    logger.ILogger
	base.IOrm[*Trigger, Trigger]
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
	return d.tableName
}

// GetIndexes get the dao's indexes.
func (d *dao) GetIndexes() []mongo.IndexModel {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: FieldKeyTriggerID, Value: 1}},
			Options: mongoOptions.Index(),
		},
	}

	return indexes
}

func (d *dao) update(ctx context.Context, trig *Trigger) error {
	filter := WithTriggerID(trig.TriggerID)(base.AliveFilter())

	result, err := d.client.UpdateOne(ctx, filter, base.BuildUpsertParam(trig), mongoOptions.Update())
	if err != nil {
		return err
	}

	switch {
	case result.MatchedCount > 0:
		{
			d.logger.Infof("successfully updated trigger, unique-key(%s)", trig.UniqueKey())
		}
	default:
		d.logger.Warnf("try to update trigger but no changes made, unique-key(%s", trig.UniqueKey())
	}

	return nil
}
