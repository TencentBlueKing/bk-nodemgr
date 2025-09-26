/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package mongotest provides a test suite for MongoDB using Testcontainers.
package mongotest

import (
	"fmt"
	"os"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/docker/go-connections/nat"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestSuite is a test suite for MongoDB using Testcontainers.
type TestSuite struct {
	suite.Suite
	mongoContainer tc.Container
	client         *mongo.Client
	db             *mongo.Database
	ctx            contextx.IContext
	dbName         string
	mongoURI       string
}

// LoadEnv loads environment variables from a .env file.
func (suit *TestSuite) LoadEnv() error {
	return godotenv.Load(".env")
}

// InitMongo initializes the MongoDB test suite.
func (suit *TestSuite) InitMongo(nCtx contextx.IContext, testConfig *Config) {
	suit.ctx = nCtx
	if testConfig == nil {
		testConfig = DefaultConfig()
	}
	suit.dbName = testConfig.DatabaseName

	if _, exist := os.LookupEnv("SKIP_CONTAINERS"); exist {
		suit.T().Log("SKIP_CONTAINERS is set, skipping container setup")

		mongoURI := os.Getenv("TEST_MONGO_URI")
		if mongoURI == "" {
			suit.T().Fatal("TEST_MONGO_URI environment variable must be set when SKIP_CONTAINERS=true")
		}
		suit.setupMongoClient(mongoURI)
	} else {
		suit.setupMongoContainer(testConfig)
	}
}

// Config holds the configuration for the MongoDB test container.
type Config struct {
	MongoImage    string
	DatabaseName  string
	ContainerName string
}

// DefaultConfig returns the default configuration for the MongoDB test container.
func DefaultConfig() *Config {
	return &Config{
		MongoImage:    "mongo:6.0.10",
		DatabaseName:  fmt.Sprintf("testdb_%d", time.Now().UnixNano()),
		ContainerName: fmt.Sprintf("mongo_test_%d", time.Now().UnixNano()),
	}
}

type loggerAdapter struct {
}

// Printf implements the Logging interface.
func (l *loggerAdapter) Printf(format string, v ...interface{}) {
	logger.G.Sys().Info(format, v...)
}

// MongoPort is the default port for MongoDB.
const MongoPort = "27017/tcp"

func (suit *TestSuite) setupMongoContainer(config *Config) {
	natPort := nat.Port(MongoPort)
	req := tc.ContainerRequest{
		Image:        config.MongoImage,
		ExposedPorts: []string{MongoPort},
		Env:          map[string]string{"MONGO_INITDB_DATABASE": config.DatabaseName},
		WaitingFor:   wait.ForListeningPort(MongoPort),
		Name:         config.ContainerName,
	}

	mongoC, err := tc.GenericContainer(suit.ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Logger:           &loggerAdapter{},
		Reuse:            true,
	})
	suit.Require().NoError(err)
	suit.mongoContainer = mongoC

	mappedPort, err := mongoC.MappedPort(suit.ctx, natPort)
	suit.Require().NoError(err)

	host, err := mongoC.Host(suit.ctx)
	suit.Require().NoError(err)

	mongoURI := fmt.Sprintf("mongodb://%s:%s", host, mappedPort.Port()) // nolint: nosprintfhostport
	suit.setupMongoClient(mongoURI)
}

func (suit *TestSuite) setupMongoClient(mongoURI string) {
	clientOpts := options.Client().ApplyURI(mongoURI)
	var err error
	suit.client, err = mongo.Connect(suit.ctx, clientOpts)
	suit.Require().NoError(err)
	err = suit.client.Ping(suit.ctx, nil)
	suit.Require().NoError(err, "Failed to ping MongoDB")

	suit.db = suit.client.Database(suit.dbName)
	suit.mongoURI = mongoURI
}

// TearDownMongo tears down the MongoDB test suite.
func (suit *TestSuite) TearDownMongo() {
	if suit.client != nil {
		err := suit.db.Drop(suit.ctx)
		suit.Require().NoError(err, "Failed to drop database")

		err = suit.client.Disconnect(suit.ctx)
		suit.Require().NoError(err, "Failed to disconnect client")
	}
	if suit.mongoContainer != nil {
		err := suit.mongoContainer.Terminate(suit.ctx)
		suit.Require().NoError(err, "Failed to terminate container")
	}
}

// GetDatabase returns the MongoDB database.
func (suit *TestSuite) GetDatabase() *mongo.Database { return suit.db }

// GetClient returns the MongoDB client.
func (suit *TestSuite) GetClient() *mongo.Client { return suit.client }

// GetDatabaseName returns the name of the MongoDB database.
func (suit *TestSuite) GetDatabaseName() string { return suit.dbName }

// GetMongoURI returns the MongoDB URI.
func (suit *TestSuite) GetMongoURI() string { return suit.mongoURI }

// GetContext returns the context for the test suite.
func (suit *TestSuite) GetContext() contextx.IContext { return suit.ctx }

// CreateCollection creates a new collection in the MongoDB database.
func (suit *TestSuite) CreateCollection(collectionName string) error {
	return suit.db.CreateCollection(suit.ctx, collectionName)
}

// ClearCollection clears the specified collection in the MongoDB database.
func (suit *TestSuite) ClearCollection(collectionName string) error {
	_, err := suit.db.Collection(collectionName).DeleteMany(suit.ctx, map[string]interface{}{})
	return err
}
