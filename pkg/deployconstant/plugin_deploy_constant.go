/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package deployconstant provides constants and configuration for deployment.
package deployconstant

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/winpath"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// PluginDeployConf defines the deployment configuration for agent.
type PluginDeployConf struct {
	DeployConf
}

// Validate checks if the deployment configuration is valid.
func (conf PluginDeployConf) Validate() error {
	if err := conf.Generation.Validate(); err != nil {
		return fmt.Errorf("invalid generation: %w", err)
	}

	if err := conf.OsType.Validate(); err != nil {
		return err
	}

	if conf.BaseDeployDir == "" {
		return errors.New("baseDeployDir is empty")
	}

	return nil
}

// nolint: gochecknoglobals
var pluginDeployConfMap = make(map[types.Generation]map[criteria.OSType]PluginDeployConf)

// GetPluginDeployConf returns the deployment configuration for the specified OS type.
func GetPluginDeployConf(generation types.Generation, osType criteria.OSType) (PluginDeployConf, error) {
	confMap, ok := pluginDeployConfMap[generation]
	if !ok {
		return PluginDeployConf{}, fmt.Errorf("deploy conf not found for generation: %d", generation)
	}

	conf, ok := confMap[osType]
	if !ok {
		return PluginDeployConf{}, fmt.Errorf("deploy conf not found for os type: %s", osType)
	}

	return conf, nil
}

// SetPluginDeployConf sets the deployment configuration for the specified OS type.
// this map only set once, if the osType already exists, it will not be set again.
func SetPluginDeployConf(conf PluginDeployConf) error {
	// Populate default values if not set
	populatePluginDefaultValues(&conf)

	if err := conf.Validate(); err != nil {
		return fmt.Errorf("set deploy conf failed: %w", err)
	}

	if _, ok := pluginDeployConfMap[conf.Generation]; !ok {
		pluginDeployConfMap[conf.Generation] = make(map[criteria.OSType]PluginDeployConf)
	}

	if _, ok := pluginDeployConfMap[conf.Generation][conf.OsType]; !ok {
		pluginDeployConfMap[conf.Generation][conf.OsType] = conf
	}

	return nil
}

func populatePluginDefaultValues(conf *PluginDeployConf) {
	if conf.OsType == criteria.OSWindows {
		populatePluginDefaultValuesWindows(conf)

		return
	}

	populatePluginDefaultValuesUnix(conf)
}

func populatePluginDefaultValuesUnix(conf *PluginDeployConf) {
	env := system.GetEnv()

	conf.DeployDir = filepath.Join(conf.BaseDeployDir, env)
}

func populatePluginDefaultValuesWindows(conf *PluginDeployConf) {
	env := system.GetEnv()

	conf.DeployDir = winpath.Join(conf.BaseDeployDir, env)
}
