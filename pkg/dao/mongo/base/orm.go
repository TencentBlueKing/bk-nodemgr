/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package base ...
package base

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

// Dao this is a common dao to dao some common crud.
type Dao interface {
	// GetClient get the mongo client.
	GetClient() *mongo.Collection

	// GetLogger get the logger.
	GetLogger() logger.Logger

	// GetTableName get the table name.
	GetTableName() string

	// GetIndexes get the indexes.
	GetIndexes() []mongo.IndexModel
}

// NewOrm this is a common orm to operate mongo db.
func NewOrm[P Pointer[T], T any](dao Dao) *Orm[P, T] {
	return &Orm[P, T]{
		dao: dao,
	}
}

// IOrm this defines the orm interface.
type IOrm[P Pointer[T], T any] interface {
	// EnsureIndexes ensure the indexes of given table.
	EnsureIndexes() error

	// Get single data by given filter.
	Get(ctx context.Context, filter bson.D, fields ...string) (P, error)

	// Create single data.
	Create(ctx context.Context, data P) error

	// CreateMany create multiple data.
	CreateMany(ctx context.Context, datas []P) error

	// UpdateField update single field of given data.
	UpdateField(ctx context.Context, filter bson.D, field string, value any) error

	// Count count the number of given data.
	Count(ctx context.Context, filter bson.D) (int64, error)

	// List list the data by given filter.
	List(ctx context.Context, filter bson.D, findOpt *mongoOptions.FindOptions) ([]P, error)

	// DistinctString distinct the string value of given key.
	DistinctString(
		ctx context.Context, key string, filter bson.D, distinctOpt *mongoOptions.DistinctOptions) ([]string, error)

	// DistinctInt64 distinct the int64 value of given key.
	DistinctInt64(
		ctx context.Context, key string, filter bson.D, distinctOpt *mongoOptions.DistinctOptions) ([]int64, error)

	// DeleteMany delete multiple data.
	DeleteMany(ctx context.Context, filter bson.D) error
}

// Orm this is a common orm to operate mongo db.
type Orm[P Pointer[T], T any] struct {
	dao Dao
}

// Pointer is a pointer.
type Pointer[T any] interface {
	*T
	Data
}

// Get this is a common operation for mongo db.
func (orm *Orm[P, T]) Get(ctx context.Context, filter bson.D, fields ...string) (P, error) {
	projection := bson.D{}
	for _, field := range fields {
		projection = append(projection, bson.E{Key: field, Value: 1})
	}
	findOptions := mongoOptions.FindOne().SetProjection(projection)

	table := &TableBroker[P]{}
	err := orm.dao.GetClient().FindOne(ctx, filter, findOptions).Decode(table)
	if err != nil {
		orm.dao.GetLogger().Warnf("failed to decode %s, err %v", orm.dao.GetTableName(), err)
		return nil, err
	}

	if table.Data == nil {
		return nil, errors.New("record found but data is nil")
	}

	return table.Data, nil
}

// CreateMany this is a common operation for mongo db.
func (orm *Orm[P, T]) CreateMany(ctx context.Context, datas []P) error {
	if len(datas) == 0 {
		return nil
	}

	timeNow := time.Now()
	tables := make([]interface{}, 0, len(datas))
	for idx := range datas {
		tables = append(tables, &TableBroker[P]{
			BasicInfo: BasicInfo{
				IsDeleted: false,
				CreatedAt: timeNow,
			},
			Data: datas[idx],
		})
	}

	result, err := orm.dao.GetClient().InsertMany(ctx, tables, nil)
	if err != nil {
		return err
	}

	orm.dao.GetLogger().Infof("created multi documents. table(%s), count(%d)",
		orm.dao.GetTableName(), len(result.InsertedIDs))

	return nil
}

// Create this is a common operation for mongo db.
func (orm *Orm[P, T]) Create(ctx context.Context, data P) error {
	table := &TableBroker[P]{
		BasicInfo: BasicInfo{
			IsDeleted: false,
			CreatedAt: time.Now(),
		},
		Data: data,
	}

	if _, err := orm.dao.GetClient().InsertOne(ctx, table); err != nil {
		return err
	}

	orm.dao.GetLogger().Infof("created document. table(%s), unique-key(%s)", orm.dao.GetTableName(), data.UniqueKey())

	return nil
}

// EnsureIndexes this is a common operation for mongo db.
func (orm *Orm[P, T]) EnsureIndexes() error {
	indexes := orm.dao.GetIndexes()

	if len(indexes) == 0 {
		return nil
	}

	// this is a common operation for mongo db, so we use background context.
	_, err := orm.dao.GetClient().Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return err
	}

	orm.dao.GetLogger().Infof("created required indexes, table(%s), indexes(%v)",
		orm.dao.GetTableName(), indexes)

	return nil
}

// UpdateField this is a common operation for mongo db.
func (orm *Orm[P, T]) UpdateField(ctx context.Context, filter bson.D, field string, value any) error {
	update := orm.buildUpdateField(field, value)

	result, err := orm.dao.GetClient().UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	orm.dao.GetLogger().Infof("updated field(%v), table(%s), updated-count(%d)",
		field, orm.dao.GetTableName(), result.MatchedCount)

	return nil
}

// buildUpdateField build update field param.
func (orm *Orm[P, T]) buildUpdateField(key string, value any) bson.D {
	nowTime := time.Now()
	update := bson.D{
		{
			Key: "$set",
			Value: bson.M{
				"basic.is_deleted": false,
				"basic.updated_at": nowTime,
				key:                value,
			},
		},
	}

	return update
}

// Count this is a common operation for mongo db.
func (orm *Orm[P, T]) Count(ctx context.Context, filter bson.D) (int64, error) {
	num, err := orm.dao.GetClient().CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	if num < 0 {
		return 0, fmt.Errorf("count documents get unexpected result: %d", num)
	}

	return num, nil
}

// List this is a common operation for mongo db.
func (orm *Orm[P, T]) List(ctx context.Context, filter bson.D, findOpt *mongoOptions.FindOptions) ([]P, error) {
	result, err := orm.dao.GetClient().Find(ctx, filter, findOpt)
	if err != nil {
		return nil, err
	}

	hosts := make([]P, 0)
	for result.Next(ctx) {
		table := &TableBroker[P]{}
		if err := result.Decode(table); err != nil {
			orm.dao.GetLogger().Warnf("failed to decode %s, err %v", orm.dao.GetTableName(), err)

			continue
		}
		hosts = append(hosts, table.Data)
	}

	return hosts, nil
}

// DistinctString this is a common operation for mongo db.
func (orm *Orm[P, T]) DistinctString(
	ctx context.Context, key string, filter bson.D, distinctOpt *mongoOptions.DistinctOptions) ([]string, error) {

	values, err := orm.dao.GetClient().Distinct(ctx, key, filter, distinctOpt)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0)
	for _, value := range values {
		if v, ok := value.(string); ok {
			result = append(result, v)
		}
	}

	return result, nil
}

// DistinctInt64 this is a common operation for mongo db.
func (orm *Orm[P, T]) DistinctInt64(
	ctx context.Context, key string, filter bson.D, distinctOpt *mongoOptions.DistinctOptions) ([]int64, error) {

	values, err := orm.dao.GetClient().Distinct(ctx, key, filter, distinctOpt)
	if err != nil {
		return nil, err
	}

	result := make([]int64, 0)
	for _, value := range values {
		if v, ok := value.(int64); ok {
			result = append(result, v)
		}
	}

	return result, nil
}

// DeleteMany this is a common operation for mongo db.
// NOTE: this is a soft delete, not real delete.
func (orm *Orm[P, T]) DeleteMany(ctx context.Context, filter bson.D) error {
	update := BuildDeleteParam()
	models := []mongo.WriteModel{mongo.NewUpdateManyModel().SetFilter(filter).SetUpdate(update).SetUpsert(false)}

	result, err := orm.dao.GetClient().BulkWrite(ctx, models)
	if err != nil {
		return err
	}

	if result.MatchedCount > 0 {
		orm.dao.GetLogger().Infof("deleted networkunits, deleted-count(%v)", result.MatchedCount)
	}

	return nil
}
