/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package watcher provides the internal watcher.
package watcher

import (
	"context"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
)

// Config defines the configuration of watcher.
type Config struct {
	CmdbHandler cmdb.IHandler
	TopoStorage topo.IStorage
	Manager     manager.Manager
}

// Watcher defines a watcher manager.
type Watcher struct {
	// config.
	conf Config

	// logger.
	logger logger.Logger
}

// NewWatcher creates a new watcher manager.
func NewWatcher(conf Config, logger logger.Logger) (*Watcher, error) {
	return &Watcher{
		conf:   conf,
		logger: logger,
	}, nil
}

// Start starts the watcher manager.
func (w *Watcher) Start(_ context.Context) error {
	w.logger.Info("successfully started watcher manager")

	return nil
}
