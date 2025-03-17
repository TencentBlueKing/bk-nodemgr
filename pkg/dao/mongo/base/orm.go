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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

// Dao this is a common dao to dao some common crud.
type Dao interface {
	GetClient() *mongo.Collection
	GetLogger() logger.Logger
	GetTableName() string
}

// NewOrm this is a common orm to operate mongo db.
func NewOrm[P Pointer[T], T any](dao Dao) *Orm[P, T] {
	return &Orm[P, T]{
		dao: dao,
	}
}

// IOrm this defines the orm interface.
type IOrm[P Pointer[T], T any] interface {
	Get(ctx context.Context, filter bson.D, fields ...string) (P, error)
	Create(ctx context.Context, data P) error
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

// Create this is a common operation for mongo db.
func (orm *Orm[P, T]) Create(ctx context.Context, data P) error {
	table := &TableBroker[P]{
		Data: data,
	}

	if _, err := orm.dao.GetClient().InsertOne(ctx, table); err != nil {
		return err
	}

	orm.dao.GetLogger().Infof("successfully created %s, unique-key(%s)", orm.dao.GetTableName(), data.UniqueKey())

	return nil
}
