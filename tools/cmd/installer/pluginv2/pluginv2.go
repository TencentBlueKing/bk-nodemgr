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

// Package pluginv2 provides plugin v2 related commands.
package pluginv2

import (
	"fmt"
	"runtime"

	pluginV2Flag "github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/pluginv2/flag"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/pluginv2/step"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
	"github.com/spf13/cobra"
)

// NewPluginCommand creates a new plugin v2 sub command.
// nolint: lll
func NewPluginCommand() *cobra.Command {
	pluginCommand := &cobra.Command{
		Use:   "pluginv2",
		Short: "Plugin v2 command",
		Long:  "Plugin v2 command",
		// PersistentPreRunE is inherited by every pluginv2 leaf command (none of them defines
		// its own PersistentPreRunE) and runs before any HTTP request, so overriding the
		// proxy env vars here makes every http.Client apply the override.
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// distinguish "flag not given" from "given with an empty value": only an
			// explicitly given flag overrides the proxy, and an empty value skips it.
			if !cmd.Flags().Changed(pluginV2Flag.HTTPProxy) {
				return nil
			}

			httpProxy, err := cmd.Flags().GetString(pluginV2Flag.HTTPProxy)
			if err != nil {
				return err
			}

			fmt.Printf("flag --%s is set, http proxy is overridden for all http requests\n", pluginV2Flag.HTTPProxy)

			return utils.OverrideProxyEnv(httpProxy)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	pluginCommand.AddCommand(step.NewStepCommand())
	pluginCommand.AddCommand(NewFullInstall())
	pluginCommand.AddCommand(NewFullUninstall())

	/*
	 * persistent required flags.
	 */
	pluginCommand.PersistentFlags().StringP(pluginV2Flag.DeployEnv, pluginV2Flag.DeployEnvS, "", "the deploy environment which to operate at")
	_ = pluginCommand.MarkFlagRequired(pluginV2Flag.DeployEnv)

	/*
	 * persistent optional flags.
	 */
	pluginCommand.PersistentFlags().String(pluginV2Flag.PluginName, "", "plugin name")
	pluginCommand.PersistentFlags().String(pluginV2Flag.PluginGroup, "", "plugin group, this will be used to build the plugin directory")
	pluginCommand.PersistentFlags().String(pluginV2Flag.PluginPkgName, "", "plugin package name, this will be used to download the plugin package")
	pluginCommand.PersistentFlags().String(pluginV2Flag.BaseDeployDir, defaultBaseDeployDir(), "base deployed directory of this node, the deploy dir will be created under this directory with deploy-env")
	pluginCommand.PersistentFlags().String(pluginV2Flag.WorkDir, "", "work directory of this plugin operation, carrying the per-user isolation segment")
	_ = pluginCommand.MarkPersistentFlagRequired(pluginV2Flag.WorkDir)
	pluginCommand.PersistentFlags().String(pluginV2Flag.HTTPProxy, "",
		"override the http proxy for all http requests in the installer, empty value means skipping the proxy")

	return pluginCommand
}

func defaultBaseDeployDir() string {
	if runtime.GOOS == "windows" {
		return `C:\`
	}

	return "/usr/local/"
}
