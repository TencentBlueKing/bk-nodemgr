/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package options provides the various capabilities the service supports.
package options

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/frontsetting"
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/storage/cptemplate"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backend"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tracing"
	"go.mongodb.org/mongo-driver/mongo"
)

// Capability encapsulates the various capabilities the service supports.
type Capability struct {
	// MongoClient mongo client.
	MongoClient *mongo.Client

	// BackendHandler the backend api hanler.
	BackendHandler backend.IHandler

	// FileHandler the file handler.
	FileHandler file.IHandler

	// StorageConfigPolicyTemplate the storage of config policy template.
	StorageConfigPolicyTemplate cptemplate.IStorage

	// Discover provides discover handler.
	DiscoverProvider discover.Provider

	// FrontSetting front setting
	FrontSetting frontsetting.IFrontSetting
}

// Start starts all services in capability.
func (c *Capability) Start(nCtx contextx.IContext) error {
	if err := c.DiscoverProvider.Start(nCtx); err != nil {
		return err
	}

	if err := c.StorageConfigPolicyTemplate.Start(nCtx); err != nil {
		return err
	}

	return nil
}

// GracefulShutdown graceful shutdown all services in capability.
func (c *Capability) GracefulShutdown() error {
	if err := tracing.G().ShutdownAll(contextx.Background()); err != nil {
		return err
	}

	return nil
}
