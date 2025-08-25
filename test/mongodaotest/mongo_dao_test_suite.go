/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package mongodaotest provides a test suite for MongoDB using Testcontainers.
package mongodaotest

import (
	"context"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/test/mongotest"
	"go.mongodb.org/mongo-driver/mongo"
)

// TestSuite provides a basic suite of mongodb dao based testing.
// can be embedded and extended by other test suites.
type TestSuite[P base.Pointer[T], T any] struct {
	mongotest.TestSuite
	baseOrm   base.IOrm[P, T]
	TestDatas []P
	Dao       base.Dao
	logger    logger.ILogger
	initFn    func(client *mongo.Database, logger logger.ILogger)
}

// SetupSuite Initialize resources before all tests begin.
func (suit *TestSuite[P, T]) SetupSuite() {
	err := suit.LoadEnv()
	if err != nil {
		suit.logger.Warnf("load env failed: %v", err)
	}

	suit.InitMongo(context.Background(), &mongotest.Config{
		MongoImage:    "mongo:latest",
		DatabaseName:  "bknodemgr",
		ContainerName: "mongo_test_bknodemgr",
	})

	suit.initFn(suit.GetDatabase(), suit.logger)
	suit.baseOrm = base.NewOrm[P, T](suit.Dao)
}

// TearDownSuite clean up resources after all tests are done.
func (suit *TestSuite[P, T]) TearDownSuite() {
	suit.TearDownMongo()
}

// SetupTest Initialize resources before each test.
func (suit *TestSuite[P, T]) SetupTest() {
	err := suit.ClearCollection(suit.Dao.GetTableName())
	suit.Require().NoError(err, "Failed to clear collection")

	err = suit.baseOrm.CreateMany(suit.GetContext(), suit.TestDatas)
	suit.Require().NoError(err, "Failed to prepare test data")
}

// NewMongoDaoTestSuite creates a new MongoDaoTestSuite instance.
func NewMongoDaoTestSuite[P base.Pointer[T], T any](
	logger logger.ILogger,
	initFn func(client *mongo.Database, logger logger.ILogger),
) TestSuite[P, T] {

	return TestSuite[P, T]{
		TestSuite: mongotest.TestSuite{},
		logger:    logger,
		initFn:    initFn,
	}
}
