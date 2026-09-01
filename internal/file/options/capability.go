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

// Package options provides the various capabilities the service supports.
package options

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/manager"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/storage/packageevent"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/storage/packageexport"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/storage/release"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/storage/upload"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/bkrepo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tracing"
	"go.mongodb.org/mongo-driver/mongo"
)

// Capability encapsulates the various capabilities the service supports.
type Capability struct {
	// MongoClient mongo client.
	MongoClient *mongo.Client

	// Discover provides discover handler.
	DiscoverProvider discover.IProvider

	// BKRepo provides bkrepo handler.
	BKRepo bkrepo.IHandler

	// StorageUpload provides storage upload handler.
	StorageUpload upload.IStorage

	// StorageRelease provides storage release handler.
	StorageRelease release.IStorage

	// StorageTopo provides storage topo handler.
	StorageTopo topo.IStorage

	// StorageEvent provides storage event handler.
	StorageEvent packageevent.IStorage

	// StoragePackageExport provides package export storage handler.
	StoragePackageExport packageexport.IStorage

	// GSEHandler provides gse handler.
	GSEHandler gse.IHandler

	// Manager provides manager handler.
	Manager manager.IManager
}

// Start start the capability.
// nolint: varnamelen
func (c *Capability) Start(nCtx contextx.IContext) error {
	if err := c.DiscoverProvider.Start(nCtx); err != nil {
		return err
	}

	if err := c.StorageUpload.Start(nCtx); err != nil {
		return err
	}

	if err := c.StorageRelease.Start(nCtx); err != nil {
		return err
	}

	if err := c.StorageTopo.Start(nCtx); err != nil {
		return err
	}

	if err := c.StorageEvent.Start(nCtx); err != nil {
		return err
	}

	if err := c.StoragePackageExport.Start(nCtx); err != nil {
		return err
	}

	if err := c.Manager.Start(nCtx); err != nil {
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
