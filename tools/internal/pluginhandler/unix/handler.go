/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package unix

import (
	"fmt"
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/pluginhandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/types"
)

// NewOfficialPluginHandler creates a new PluginHandler for official plugin.
func NewOfficialPluginHandler(rootAbsDir, pluginName string) (pluginhandler.IPluginHandler, error) {
	handler := &PluginHandler{
		rootAbsDir: rootAbsDir,
		pluginName: pluginName,
		pluginType: types.PluginTypeOfficial,
	}
	if err := handler.initConfigs(); err != nil {
		return nil, err
	}

	return handler, nil
}

// NewExternalPluginHandler creates a new PluginHandler for external plugin.
func NewExternalPluginHandler(rootAbsDir, pluginName string) (pluginhandler.IPluginHandler, error) {
	handler := &PluginHandler{
		rootAbsDir: rootAbsDir,
		pluginName: pluginName,
		pluginType: types.PluginTypeExternal,
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

	pluginName string
	pluginType types.PluginType

	setupDir string
	binDir   string
	etcDir   string
}

// PluginType implement pluginhandler.IPluginHandler.
func (handler *PluginHandler) PluginType() types.PluginType {
	return handler.pluginType
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
	switch handler.pluginType {
	case types.PluginTypeOfficial, types.PluginTypeExternal:
		handler.setupDir = fmt.Sprintf("plugin/%s/%s", handler.pluginType, handler.pluginName)
	default:
		return fmt.Errorf("unsupported plugin type: %s", handler.pluginType)
	}

	handler.binDir = filepath.Join(handler.setupDir, "bin")
	handler.etcDir = filepath.Join(handler.setupDir, "etc")

	return nil
}
