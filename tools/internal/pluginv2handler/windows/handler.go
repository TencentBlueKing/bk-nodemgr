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
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/pluginv2handler"
)

// NewPluginV2Handler creates a new pluginv2Handler for plugin.
func NewPluginV2Handler(rootAbsDir, pluginGroup, pluginName string) (pluginv2handler.IPluginV2Handler, error) {
	handler := &PluginV2Handler{
		rootAbsDir:  rootAbsDir,
		pluginName:  pluginName,
		pluginGroup: pluginGroup,
	}
	if err := handler.initConfigs(); err != nil {
		return nil, err
	}

	return handler, nil
}

var _ pluginv2handler.IPluginV2Handler = &PluginV2Handler{}

// PluginV2Handler provides the windows plugin handler.
type PluginV2Handler struct {
	rootAbsDir string

	pluginName  string
	pluginGroup string

	setupDir string
	binDir   string
	etcDir   string
}

// Group implement pluginv2handler.IPluginV2Handler.
func (handler *PluginV2Handler) Group() string {
	return handler.pluginGroup
}

// FS implement pluginv2handler.IPluginV2Handler.
func (handler *PluginV2Handler) FS() pluginv2handler.IPluginV2FSHandler {
	return handler
}

// Process implement pluginv2handler.IPluginV2Handler.
func (handler *PluginV2Handler) Process() pluginv2handler.IPluginV2ProcessHandler {
	return handler
}

// initConfigs initializes the configurations via root-abs-dir.
// All plugins share a single 'plugins' directory with merged bin/ and etc/ subdirectories.
func (handler *PluginV2Handler) initConfigs() error {
	handler.setupDir = "plugins"
	handler.binDir = filepath.Join(handler.setupDir, "bin")
	handler.etcDir = filepath.Join(handler.setupDir, "etc")

	return nil
}
