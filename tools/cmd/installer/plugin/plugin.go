/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package plugin provides plugin related commands.
package plugin

import (
	"runtime"

	pluginflag "github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/plugin/flag"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/plugin/step"
	"github.com/spf13/cobra"
)

// NewPluginCommand creates a new plugin sub command.
func NewPluginCommand() *cobra.Command {
	pluginCommand := &cobra.Command{
		Use:   "plugin",
		Short: "Plugin command",
		Long:  "Plugin command",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	pluginCommand.AddCommand(step.NewStepCommand())
	pluginCommand.AddCommand(NewFullInstall())

	/*
	 * persistent required flags.
	 */
	pluginCommand.PersistentFlags().StringP(pluginflag.DeployEnv, pluginflag.DeployEnvS, "", "the deploy environment which to operate at")
	_ = pluginCommand.MarkFlagRequired(pluginflag.DeployEnv)

	/*
	 * persistent optional flags.
	 */
	pluginCommand.PersistentFlags().String(pluginflag.PluginName, "", "plugin name")
	pluginCommand.PersistentFlags().String(pluginflag.PluginGroup, "", "plugin group, this will be used to build the plugin directory")
	pluginCommand.PersistentFlags().String(pluginflag.PluginPkgName, "", "plugin package name, this will be used to download the plugin package")
	pluginCommand.PersistentFlags().String(pluginflag.BaseDeployDir, defaultBaseDeployDir(), "base deployed directory of this node, the deploy dir will be created under this directory with deploy-env")
	pluginCommand.PersistentFlags().String(pluginflag.BaseWorkDir, defaultBaseWorkDir(), "base work directory of this node, the work dir will be created under this directory with deploy-env")

	return pluginCommand
}

func defaultBaseDeployDir() string {
	if runtime.GOOS == "windows" {
		return `C:\`
	}

	return "/usr/local/"
}

func defaultBaseWorkDir() string {
	if runtime.GOOS == "windows" {
		return `C:\tmp\bknm\`
	}

	return "/tmp/bknm/"
}
