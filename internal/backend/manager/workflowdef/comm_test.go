/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package workflowdef ...
package workflowdef

import (
	"context"
	"os"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/nodedeployment"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/operation"
	operinstdataStorage "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/operinstdata"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/crypter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/locker"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigengine"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

// Capability encapsulates the various capabilities the service supports.
type Capability struct {
	// TopoStorage bk nodeman topo storage.
	TopoStorage topoStg.IStorage

	// TrigEngineStorage bk nodeman trigengine storage.
	TrigEngineStorage trigengine.Storage

	// OperInstStorage bk nodeman operation_inst storage.
	OperInstStorage operinstdataStorage.IStorage

	// OperStorage bk nodeman operation storage.
	OperStorage operation.Storage

	// NodeDeploymentStorage bk nodeman node deployment storage.
	NodeDeploymentStorage nodedeployment.Storage

	// CmdbHandler cmdb handler.
	CmdbHandler cmdb.IHandler

	// GseHandler gse handler.
	GseHandler gse.IHandler

	// Logger logger
	Logger logger.Logger

	// LockerFactory locker factory
	LockerFactory locker.MutexFactory

	// Crypter ...
	Crypter crypter.Crypter
}

type testHeaderSetter struct{}

// GetAuthHeader ...
func (testHeaderSetter) GetAuthHeader() (string, error) {
	return os.Getenv("BK_APIGW_AUTHHEADER"), nil
}

// testCapability ...
func testCapability(t *testing.T) *Capability {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	loggerDefault := logger.LoggerDefault{}

	ctx, _ := tenant.SetID(context.Background(), "bk_nodeman")
	mongoClient, err := mongo.Connect(
		ctx,
		&mongoOptions.ClientOptions{
			Hosts: []string{
				os.Getenv("MONGO_ADDRESS"),
			},
			Auth: &mongoOptions.Credential{
				Username:      os.Getenv("MONGO_USER"),
				Password:      os.Getenv("MONGO_PASSWORD"),
				AuthSource:    os.Getenv("MONGO_AUTH_SOURCE"),
				AuthMechanism: os.Getenv("MONGO_AUTH_MECHANISM"),
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	httpClient, err := client.NewClient(&ssl.TLSConfig{
		InsecureSkipVerify: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	topoStorage, err := topoStg.NewStorage(mongoClient, "bk_nodeman", loggerDefault)
	if err != nil {
		t.Fatal(err)
	}

	if err = topoStorage.Start(ctx); err != nil {
		t.Fatal(err)
	}

	operInstStorage, err := operinstdataStorage.NewStorage(mongoClient, "bk_nodeman", loggerDefault)
	if err != nil {
		t.Fatal(err)
	}

	if err = operInstStorage.Start(ctx); err != nil {
		t.Fatal(err)
	}

	cmdbHandler, err := cmdb.New(
		&client.Capability{
			Client:               httpClient,
			Discover:             discovery.NewDiscovery("apigateway", []string{os.Getenv("BK_APIGW_CMDB_ENDPOINT")}),
			ToleranceLatencyTime: client.ToleranceLatencyTimeDefault,
			MetricOpts:           client.MetricOption{},
			Logger:               loggerDefault,
		},
		&cmdb.Config{
			HeaderSetter: testHeaderSetter{},
		})

	gseHandler, err := gse.New(
		&client.Capability{
			Client:               httpClient,
			Discover:             discovery.NewDiscovery("apigateway", []string{os.Getenv("BK_APIGW_GSE_ENDPOINT")}),
			ToleranceLatencyTime: client.ToleranceLatencyTimeDefault,
			MetricOpts:           client.MetricOption{},
			Logger:               loggerDefault,
		},
		&gse.Config{
			HeaderSetter: testHeaderSetter{},
		})

	crypt, err := crypter.NewAESCrypter([]byte(os.Getenv("SSH_ENCRYPT_KEY")))
	if err != nil {
		t.Fatal(err)
	}

	capability := &Capability{
		TopoStorage:       topoStorage,
		TrigEngineStorage: nil,
		OperInstStorage:   operInstStorage,
		OperStorage:       nil,
		GseHandler:        gseHandler,
		CmdbHandler:       cmdbHandler,
		Logger:            loggerDefault,
		LockerFactory:     nil,
		Crypter:           crypt,
	}

	return capability
}
