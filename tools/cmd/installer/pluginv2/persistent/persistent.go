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

// Package persistent provides the persistent vars.
package persistent

import (
	"fmt"
	"os"
	"path/filepath"

	pluginV2Flag "github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/pluginv2/flag"
	"github.com/spf13/cobra"
)

const (
	dirMode = 0600
)

// Variables represents the persistent vars.
type Variables struct {
	// required flags.
	DeployEnv string

	// optional flags.
	PluginName    string
	PluginGroup   string
	PluginPkgName string

	// generates.
	DeployDir string
	WorkDir   string
	DataDir   string
	ConfigDir string
}

// EnsureDirs ensures the dirs.
func (vars Variables) EnsureDirs() error {
	// nolint: mnd
	if err := os.MkdirAll(vars.WorkDir, dirMode); err != nil {
		return err
	}

	// nolint: mnd
	if err := os.MkdirAll(vars.DataDir, dirMode); err != nil {
		return err
	}

	// nolint: mnd
	if err := os.MkdirAll(vars.ConfigDir, dirMode); err != nil {
		return err
	}

	return nil
}

// GetVariables get the persistent vars.
func GetVariables(cmd *cobra.Command) (*Variables, error) {
	deployEnv, err := cmd.Flags().GetString(pluginV2Flag.DeployEnv)
	if err != nil {
		return nil, err
	}

	pluginName, err := cmd.Flags().GetString(pluginV2Flag.PluginName)
	if err != nil {
		return nil, err
	}

	pluginGroup, err := cmd.Flags().GetString(pluginV2Flag.PluginGroup)
	if err != nil {
		return nil, err
	}

	pluginPkgName, err := cmd.Flags().GetString(pluginV2Flag.PluginPkgName)
	if err != nil {
		return nil, err
	}

	baseDeployDir, err := cmd.Flags().GetString(pluginV2Flag.BaseDeployDir)
	if err != nil {
		return nil, err
	}
	if baseDeployDir == "" {
		return nil, fmt.Errorf("%s is empty", pluginV2Flag.BaseDeployDir)
	}

	baseWorkDir, err := cmd.Flags().GetString(pluginV2Flag.BaseWorkDir)
	if err != nil {
		return nil, err
	}
	if baseWorkDir == "" {
		return nil, fmt.Errorf("%s is empty", pluginV2Flag.BaseWorkDir)
	}

	return &Variables{
		PluginName:    pluginName,
		PluginGroup:   pluginGroup,
		PluginPkgName: pluginPkgName,
		DeployEnv:     deployEnv,
		DeployDir:     filepath.Join(baseDeployDir, deployEnv),
		WorkDir:       filepath.Join(baseWorkDir, deployEnv),
		DataDir:       filepath.Join(baseWorkDir, deployEnv, "data", "plugin", pluginName),
		ConfigDir:     filepath.Join(baseWorkDir, deployEnv, "data", "plugin", pluginName, "config"),
	}, nil
}
