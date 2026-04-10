/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package base provides base orm for mongodb.
// nolint: nonamedreturns
package base

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	daomongo "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

// IDao this is a common dao to dao some common crud.
type IDao interface {
	// GetClient get the mongo client.
	GetClient() *mongo.Collection

	// GetTableName get the table name.
	GetTableName() string

	// GetIndexes get the indexes.
	GetIndexes() []mongo.IndexModel
}

// NewOrm this is a common orm to operate mongo db.
func NewOrm[P DataPoint[T], T any](dao IDao) *Orm[P, T] {
	return &Orm[P, T]{
		dao: dao,
	}
}

// IOrm this defines the orm interface.
// nolint: interfacebloat
type IOrm[P DataPoint[T], T any] interface {
	// EnsureIndexes ensure the indexes of given table.
	EnsureIndexes() error

	// Get single data by given filter.
	Get(nCtx contextx.IContext, filter bson.D, fields ...string) (P, error)

	// Exist check if the data by given filter exist.
	Exist(nCtx contextx.IContext, filter bson.D) (bool, error)

	// Create single data.
	Create(nCtx contextx.IContext, data P) error

	// CreateMany create multiple data.
	CreateMany(nCtx contextx.IContext, datas []P) error

	// UpdateField update single field of given data.
	UpdateField(nCtx contextx.IContext, filter bson.D, field string, value any) error

	// Count count the number of given data.
	Count(nCtx contextx.IContext, filter bson.D) (int64, error)

	// List list the data by given filter.
	List(nCtx contextx.IContext, filter bson.D, findOpt *mongoOptions.FindOptions, field ...string) ([]P, error)

	// DistinctString distinct the string value of given key.
	DistinctString(
		nCtx contextx.IContext, key string, filter bson.D, distinctOpt *mongoOptions.DistinctOptions) ([]string, error)

	// DistinctInt64 distinct the int64 value of given key.
	DistinctInt64(
		nCtx contextx.IContext, key string, filter bson.D, distinctOpt *mongoOptions.DistinctOptions) ([]int64, error)

	// DeleteMany delete multiple data.
	DeleteMany(nCtx contextx.IContext, filter bson.D) error

	// HardDelete delete single data permanently from database.
	// NOTE: this is a hard delete, record will be permanently removed.
	// WARNING: empty filter is not allowed to prevent accidental deletion of all data.
	HardDelete(nCtx contextx.IContext, filter bson.D) error

	// HardDeleteMany delete multiple data permanently from database.
	// NOTE: this is a hard delete, records will be permanently removed.
	// WARNING: empty filter is not allowed to prevent accidental deletion of all data.
	HardDeleteMany(nCtx contextx.IContext, filter bson.D) error

	// UpdateFieldsBulk updates multiple documents in bulk based on the provided updates.
	UpdateFieldsBulk(nCtx contextx.IContext, updates []*DocumentFieldUpdate) error
}

// Orm this is a common orm to operate mongo db.
type Orm[P DataPoint[T], T any] struct {
	dao IDao
}

// DataPoint is a pointer.
type DataPoint[T any] interface {
	*T
	IData
}

// Get this is a common operation for mongo db.
func (orm *Orm[P, T]) Get(nCtx contextx.IContext, filter bson.D, fields ...string) (dataPoint P, err error) {
	// record metric.
	metric := orm.metric().start(daomongo.MetricOperationFindOne, len(filter))
	defer func() {
		metric.end(err, func() int {
			if dataPoint == nil {
				return 0
			}

			return 1
		}())
	}()

	if nCtx == nil {
		return nil, errors.New("context is nil")
	}

	// set projection.
	projection := make(bson.D, 0)
	for _, field := range fields {
		projection = append(projection, bson.E{Key: field, Value: 1})
	}
	findOptions := mongoOptions.FindOne().SetProjection(projection)

	// find one as get.
	table := &TableBroker[P]{}
	result := orm.dao.GetClient().FindOne(nCtx, filter, findOptions)
	if err = result.Err(); err != nil {
		logger.G.Sys().Ctx(nCtx).WithErr(err).With("table", orm.dao.GetTableName()).Warn("failed to find one")

		return nil, err
	}

	if err = result.Decode(table); err != nil {
		logger.G.Sys().Ctx(nCtx).WithErr(err).With("table", orm.dao.GetTableName()).Warn("failed to decode found document")

		return nil, err
	}

	// invalid data.
	if table.Data == nil {
		return nil, errors.New("record found but data is nil")
	}

	return table.Data, nil
}

// Exist check if the data by given filter exist.
func (orm *Orm[P, T]) Exist(nCtx contextx.IContext, filter bson.D) (exist bool, err error) {
	// record metric.
	metric := orm.metric().start(daomongo.MetricOperationFindOne, len(filter))
	defer func() {
		metric.end(err, func() int {
			if err != nil {
				return 0
			}

			return 1
		}())
	}()

	if nCtx == nil {
		return false, errors.New("context is nil")
	}

	// use FindOne for presence checks and query only _id fields to improve performance.
	opts := mongoOptions.FindOne().SetProjection(bson.D{{"_id", 1}})

	var result struct {
		ID interface{} `bson:"_id"`
	}

	err = orm.dao.GetClient().FindOne(nCtx, filter, opts).Decode(&result)
	if err != nil {
		// not found, not an error
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// CreateMany this is a common operation for mongo db.
func (orm *Orm[P, T]) CreateMany(nCtx contextx.IContext, datas []P) (err error) {
	if len(datas) == 0 {
		return nil
	}

	var result *mongo.InsertManyResult

	// record metric.
	metric := orm.metric().start(daomongo.MetricOperationInsertMany, len(datas))
	defer func() {
		metric.end(err, func() int {
			if result == nil {
				return 0
			}

			return len(result.InsertedIDs)
		}())
	}()

	if nCtx == nil {
		return errors.New("context is nil")
	}

	// generates documents.
	timeNow := time.Now()
	documents := make([]interface{}, 0, len(datas))
	for idx := range datas {
		documents = append(documents, &TableBroker[P]{
			BasicInfo: BasicInfo{
				IsDeleted: false,
				CreatedAt: timeNow,
			},
			Data: datas[idx],
		})
	}

	// insert many as create many.
	if result, err = orm.dao.GetClient().InsertMany(nCtx, documents, nil); err != nil {
		return err
	}

	logger.G.Sys().Ctx(nCtx).With("table", orm.dao.GetTableName(), "count", len(result.InsertedIDs)).Info("created multi documents")

	return nil
}

// Create this is a common operation for mongo db.
func (orm *Orm[P, T]) Create(nCtx contextx.IContext, data P) (err error) {
	// record metric.
	metric := orm.metric().start(daomongo.MetricOperationInsertOne, 1)
	defer func() {
		metric.end(err, func() int {
			if err != nil {
				return 0
			}

			return 1
		}())
	}()

	if nCtx == nil {
		return errors.New("context is nil")
	}

	// generates document.
	document := &TableBroker[P]{
		BasicInfo: BasicInfo{
			IsDeleted: false,
			CreatedAt: time.Now(),
		},
		Data: data,
	}

	// insert one as create.
	if _, err = orm.dao.GetClient().InsertOne(nCtx, document); err != nil {
		return err
	}

	logger.G.Sys().Ctx(nCtx).With("table", orm.dao.GetTableName(), "unique-key", data.UniqueKey()).Info("created document")

	return nil
}

// EnsureIndexes this is a common operation for mongo db.
func (orm *Orm[P, T]) EnsureIndexes() (err error) {
	indexes := orm.dao.GetIndexes()
	var result []string

	// record metric.
	metric := orm.metric().start(daomongo.MetricOperationCreateIndexes, len(indexes))
	defer func() {
		metric.end(err, len(result))
	}()

	uniqueFieldKey := make(bson.D, 0)
	uniqueFields := TableBroker[P]{}.Data.UniqueFields()
	for _, fieldKey := range uniqueFields {
		uniqueFieldKey = append(uniqueFieldKey, bson.E{Key: fieldKey, Value: 1})
	}

	indexes = append(indexes, mongo.IndexModel{
		Keys:    uniqueFieldKey,
		Options: mongoOptions.Index().SetUnique(true),
	})

	indexes = append(indexes, mongo.IndexModel{
		Keys:    bson.D{{Key: FieldKeyIsDeleted, Value: -1}},
		Options: mongoOptions.Index(),
	})

	// this is a common operation for mongo db, so we use background context.
	if result, err = orm.dao.GetClient().Indexes().CreateMany(context.Background(), indexes); err != nil {
		return err
	}

	logger.G.Sys().Ctx(contextx.Background()).With("table", orm.dao.GetTableName(), "indexes", result).Info("created indexes")

	return nil
}

// UpdateField this is a common operation for mongo db.
func (orm *Orm[P, T]) UpdateField(nCtx contextx.IContext, filter bson.D, field string, value any) (err error) {
	var result *mongo.UpdateResult

	// record metric.
	metric := orm.metric().start(daomongo.MetricOperationUpdateMany, len(filter))
	defer func() {
		metric.end(err, func() int {
			if result == nil {
				return 0
			}

			return int(result.ModifiedCount)
		}())
	}()

	// update field.
	if result, err = orm.dao.GetClient().UpdateMany(nCtx, filter, buildUpdateField(field, value)); err != nil {
		return err
	}

	logger.G.Sys().Ctx(nCtx).
		With("table", orm.dao.GetTableName(), "field", field, "matched-count", result.MatchedCount, "modified-count", result.ModifiedCount).
		Debug("updated field")

	return nil
}

// Count this is a common operation for mongo db.
func (orm *Orm[P, T]) Count(nCtx contextx.IContext, filter bson.D) (num int64, err error) {
	// record metric.
	metric := orm.metric().start(daomongo.MetricOperationCountDucuments, len(filter))
	defer func() {
		metric.end(err, func() int {
			if err != nil {
				return 0
			}

			return 1
		}())
	}()

	if nCtx == nil {
		return 0, errors.New("context is nil")
	}

	if num, err = orm.dao.GetClient().CountDocuments(nCtx, filter); err != nil {
		return 0, err
	}

	if num < 0 {
		return 0, fmt.Errorf("count documents get unexpected result: %d", num)
	}

	return num, nil
}

// List this is a common operation for mongo db.
func (orm *Orm[P, T]) List(nCtx contextx.IContext, filter bson.D, findOpt *mongoOptions.FindOptions, field ...string) (dataPoints []P, err error) {
	// record metric.
	metric := orm.metric().start(daomongo.MetricOperationFind, len(filter))
	defer func() {
		metric.end(err, len(dataPoints))
	}()

	if nCtx == nil {
		return nil, errors.New("context is nil")
	}

	// set projection if fields are specified.
	if len(field) > 0 {
		projection := make(bson.D, 0)
		for _, f := range field {
			projection = append(projection, bson.E{Key: f, Value: 1})
		}
		if findOpt == nil {
			findOpt = mongoOptions.Find()
		}
		findOpt.SetProjection(projection)
	}

	var cursor *mongo.Cursor
	if cursor, err = orm.dao.GetClient().Find(nCtx, filter, findOpt); err != nil {
		return nil, err
	}

	dataPoints = make([]P, 0)
	for cursor.Next(nCtx) {
		document := &TableBroker[P]{}
		if err := cursor.Decode(document); err != nil {
			logger.G.Sys().Ctx(nCtx).WithErr(err).With("table", orm.dao.GetTableName()).Info("failed to list, failed to decode document")

			continue
		}
		dataPoints = append(dataPoints, document.Data)
	}

	return dataPoints, nil
}

// DistinctString this is a common operation for mongo db.
func (orm *Orm[P, T]) DistinctString(
	nCtx contextx.IContext, key string, filter bson.D, distinctOpt *mongoOptions.DistinctOptions) (result []string, err error) {

	var values []interface{}

	// record metric.
	metric := orm.metric().start(daomongo.MetricOperationDistinct, len(filter))
	defer func() {
		metric.end(err, len(values))
	}()

	if values, err = orm.dao.GetClient().Distinct(nCtx, key, filter, distinctOpt); err != nil {
		return nil, err
	}

	result = make([]string, 0)
	for _, value := range values {
		if v, ok := value.(string); ok {
			result = append(result, v)
		}
	}

	return result, nil
}

// DistinctInt64 this is a common operation for mongo db.
func (orm *Orm[P, T]) DistinctInt64(
	nCtx contextx.IContext, key string, filter bson.D, distinctOpt *mongoOptions.DistinctOptions) (result []int64, err error) {

	var values []interface{}

	// record metric.
	metric := orm.metric().start(daomongo.MetricOperationDistinct, len(filter))
	defer func() {
		metric.end(err, len(values))
	}()

	if values, err = orm.dao.GetClient().Distinct(nCtx, key, filter, distinctOpt); err != nil {
		return nil, err
	}

	result = make([]int64, 0)
	for _, value := range values {
		if v, ok := value.(int64); ok {
			result = append(result, v)
		}
	}

	return result, nil
}

// DeleteMany this is a common operation for mongo db.
// NOTE: this is a soft delete, not real delete.
func (orm *Orm[P, T]) DeleteMany(nCtx contextx.IContext, filter bson.D) (err error) {
	models := []mongo.WriteModel{mongo.NewUpdateManyModel().SetFilter(filter).SetUpdate(BuildDeleteParam()).SetUpsert(false)}
	var result *mongo.BulkWriteResult

	// record metric.
	metric := orm.metric().start(daomongo.MetricOperationBulkWrite, len(models))
	defer func() {
		metric.end(err, func() int {
			if result == nil {
				return 0
			}

			return int(result.MatchedCount)
		}())
	}()

	if nCtx == nil {
		return errors.New("context is nil")
	}

	if result, err = orm.dao.GetClient().BulkWrite(nCtx, models); err != nil {
		return err
	}

	if result.MatchedCount > 0 {
		logger.G.Sys().Ctx(nCtx).With("table", orm.dao.GetTableName(), "deleted-count", result.MatchedCount).Info("deleted many")
	}

	return nil
}

// HardDelete this is a hard delete operation for mongo db.
// NOTE: this is a hard delete, record will be permanently removed.
// WARNING: empty filter is not allowed to prevent accidental deletion of all data.
func (orm *Orm[P, T]) HardDelete(nCtx contextx.IContext, filter bson.D) (err error) {
	var result *mongo.DeleteResult

	// record metric.
	metric := orm.metric().start(daomongo.MetricOperationBulkWrite, len(filter))
	defer func() {
		metric.end(err, func() int {
			if result == nil {
				return 0
			}

			return int(result.DeletedCount)
		}())
	}()

	if nCtx == nil {
		return errors.New("context is nil")
	}

	if len(filter) == 0 {
		return errors.New("empty filter is not allowed for hard delete to prevent accidental deletion of all data")
	}

	if result, err = orm.dao.GetClient().DeleteOne(nCtx, filter); err != nil {
		return err
	}

	if result.DeletedCount > 0 {
		logger.G.Sys().Ctx(nCtx).With("table", orm.dao.GetTableName(), "deleted-count", result.DeletedCount).Info("hard deleted document")
	}

	return nil
}

// HardDeleteMany this is a hard delete operation for mongo db.
// NOTE: this is a hard delete, records will be permanently removed.
// WARNING: empty filter is not allowed to prevent accidental deletion of all data.
func (orm *Orm[P, T]) HardDeleteMany(nCtx contextx.IContext, filter bson.D) (err error) {
	var result *mongo.DeleteResult

	// record metric.
	metric := orm.metric().start(daomongo.MetricOperationBulkWrite, len(filter))
	defer func() {
		metric.end(err, func() int {
			if result == nil {
				return 0
			}

			return int(result.DeletedCount)
		}())
	}()

	if nCtx == nil {
		return errors.New("context is nil")
	}

	if len(filter) == 0 {
		return errors.New("empty filter is not allowed for hard delete to prevent accidental deletion of all data")
	}

	if result, err = orm.dao.GetClient().DeleteMany(nCtx, filter); err != nil {
		return err
	}

	if result.DeletedCount > 0 {
		logger.G.Sys().Ctx(nCtx).With("table", orm.dao.GetTableName(), "deleted-count", result.DeletedCount).Info("hard deleted many documents")
	}

	return nil
}

// DocumentFieldUpdate represents a document update operation with specific fields.
type DocumentFieldUpdate struct {
	Filter bson.D
	Fields map[string]any
}

// UpdateFieldsBulk updates multiple documents in bulk based on the provided updates.
func (orm *Orm[P, T]) UpdateFieldsBulk(nCtx contextx.IContext, updates []*DocumentFieldUpdate) (err error) {
	if len(updates) == 0 {
		return nil
	}

	var result *mongo.BulkWriteResult

	// record metric.
	metric := orm.metric().start(daomongo.MetricOperationBulkWrite, len(updates))
	defer func() {
		metric.end(err, func() int {
			if result == nil {
				return 0
			}

			return int(result.MatchedCount)
		}())
	}()

	models := make([]mongo.WriteModel, 0, len(updates))
	for _, update := range updates {
		if len(update.Fields) == 0 {
			continue
		}

		updateDoc := buildUpdateFields(update.Fields)
		model := mongo.NewUpdateManyModel().
			SetFilter(update.Filter).
			SetUpdate(updateDoc).
			SetUpsert(false)

		models = append(models, model)
	}

	if len(models) == 0 {
		return nil
	}

	if result, err = orm.dao.GetClient().BulkWrite(nCtx, models); err != nil {
		return err
	}

	logger.G.Sys().Ctx(nCtx).
		With("table", orm.dao.GetTableName(), "matched-count", result.MatchedCount, "modified-count", result.ModifiedCount).
		Info("bulk updated fields")

	return nil
}

func (orm *Orm[P, T]) metric() *metricData {
	return newMetricData(orm.dao.GetTableName())
}

// buildUpdateField build update field param.
func buildUpdateField(key string, value any) bson.D {
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

// buildUpdateFields build update fields param for multiple fields.
func buildUpdateFields(fields map[string]any) bson.D {
	nowTime := time.Now()
	updateFields := bson.M{
		"basic.is_deleted": false,
		"basic.updated_at": nowTime,
	}

	for key, value := range fields {
		updateFields[key] = value
	}

	return bson.D{
		{
			Key:   "$set",
			Value: updateFields,
		},
	}
}
