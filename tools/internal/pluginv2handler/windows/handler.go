//go:build windows

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

package windows

import (
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/pluginv2handler"
)

// NewPluginHandler creates a new plugin handler for plugin.
func NewPluginHandler(rootAbsDir, pluginGroup, pluginName string) (pluginv2handler.IPluginHandler, error) {
	handler := &PluginHandler{
		rootAbsDir:  rootAbsDir,
		pluginName:  pluginName,
		pluginGroup: pluginGroup,
	}
	if err := handler.initConfigs(); err != nil {
		return nil, err
	}

	return handler, nil
}

var _ pluginv2handler.IPluginHandler = &PluginHandler{}

// PluginHandler provides the windows plugin handler.
type PluginHandler struct {
	rootAbsDir string

	pluginName  string
	pluginGroup string

	setupDir string
	binDir   string
	etcDir   string
}

// Group implement pluginv2handler.IPluginHandler.
func (handler *PluginHandler) Group() string {
	return handler.pluginGroup
}

// FS implement pluginv2handler.IPluginHandler.
func (handler *PluginHandler) FS() pluginv2handler.IPluginFSHandler {
	return handler
}

// Process implement pluginv2handler.IPluginHandler.
func (handler *PluginHandler) Process() pluginv2handler.IPluginProcessHandler {
	return handler
}

// initConfigs initializes the configurations via root-abs-dir.
// All plugins share a single 'plugins' directory with merged bin/ and etc/ subdirectories.
func (handler *PluginHandler) initConfigs() error {
	handler.setupDir = "plugins"
	handler.binDir = filepath.Join(handler.setupDir, "bin")
	handler.etcDir = filepath.Join(handler.setupDir, "etc")

	return nil
}
