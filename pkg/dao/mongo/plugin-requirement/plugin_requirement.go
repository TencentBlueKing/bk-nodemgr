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

package pluginrequirement

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(tableName string, client *mongo.Database) *dao {
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

// GetClient returns the collection client.
func (d *dao) GetClient() *mongo.Collection {
	return d.client
}

// GetTableName returns the collection name.
func (d *dao) GetTableName() string {
	return d.tableName
}

// GetIndexes returns additional indexes; the base ORM creates the unique ID index.
func (d *dao) GetIndexes() []mongo.IndexModel {
	return nil
}

func (d *dao) update(nCtx contextx.IContext, data *Data) error {
	filter := append(base.AliveFilter(), bson.E{Key: FieldKeyID, Value: data.ID})
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "data", Value: data},
		{Key: base.FieldKeyUpdatedAt, Value: time.Now()},
	}}}
	// ORM field updates do not expose matched counts; a missing ID must fail atomically.
	result, err := d.client.UpdateOne(nCtx, filter, update, options.Update().SetUpsert(false))
	if err != nil {
		return fmt.Errorf("failed to update requirement document: %w", err)
	}
	if result.MatchedCount == 0 {
		return base.ErrRecordNoFound()
	}

	return nil
}
