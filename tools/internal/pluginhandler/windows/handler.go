//go:build windows

/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package windows

import (
	"fmt"
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/pluginhandler"
)

// NewPluginHandler creates a new PluginHandler for plugin.
func NewPluginHandler(rootAbsDir, pluginGroup, pluginName string) (pluginhandler.IPluginHandler, error) {
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

var _ pluginhandler.IPluginHandler = &PluginHandler{}

// PluginHandler provides the unix plugin handler.
type PluginHandler struct {
	rootAbsDir string

	pluginName  string
	pluginGroup string

	setupDir string
	binDir   string
	etcDir   string
}

// Group implement pluginhandler.IPluginHandler.
func (handler *PluginHandler) Group() string {
	return handler.pluginGroup
}

// FS implement pluginhandler.IPluginHandler.
func (handler *PluginHandler) FS() pluginhandler.IPluginFSHandler {
	return handler
}

// Process implement pluginhandler.IPluginHandler.
func (handler *PluginHandler) Process() pluginhandler.IPluginProcessHandler {
	return handler
}

// initConfigs initializes the configurations via root-abs-dir.
func (handler *PluginHandler) initConfigs() error {
	handler.setupDir = fmt.Sprintf("plugin/%s/%s", handler.pluginGroup, handler.pluginName)
	handler.binDir = filepath.Join(handler.setupDir, "bin")
	handler.etcDir = filepath.Join(handler.setupDir, "etc")

	return nil
}
