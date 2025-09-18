/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package step

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/flag"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/persistent"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/filedownloader"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/types"
	"github.com/spf13/cobra"
)

const (
	gseReleasePkgExt = "tgz"
)

// NewDownloadFiles creates a new download files step command.
// nolint: lll
func NewDownloadFiles() *cobra.Command {
	var (
		// required flags.
		downloadSvrAddr string
		callbackSvrAddr string
		deployToken     string
		nodeVersion     string

		// pre-run.
		persistentVars *persistent.Variables
	)

	stepCmd := &cobra.Command{
		Use:   "download-files",
		Short: "Download files",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			vars, err := persistent.GetVariables(cmd)
			if err != nil {
				return err
			}
			persistentVars = vars

			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			step := filedownloader.NewStep(filedownloader.StepArgs{
				DownloadSvrAddr:    downloadSvrAddr,
				CallbackSvrAddr:    callbackSvrAddr,
				NodeRole:           persistentVars.NodeRole,
				Generation:         persistentVars.Generation,
				DeployToken:        deployToken,
				PkgVersion:         nodeVersion,
				PkgSavedPath:       filepath.Join(persistentVars.DataDir, GenReleasePkgName(persistentVars.NodeRole, persistentVars.Generation, nodeVersion)),
				ConfigSavedDir:     filepath.Join(persistentVars.DataDir, "configs"),
				CheckListSavedPath: filepath.Join(persistentVars.DataDir, "precheck.json"),
			})

			if err := step.Run(cmd.Context()); err != nil {
				return err
			}

			fmt.Println("successfully downloaded files")

			return nil
		},
	}

	/*
	 * required flags.
	 */
	stepCmd.Flags().StringVar(&downloadSvrAddr, flag.DownloadSvrAddr, "", "download server address, for downloading release files")
	_ = stepCmd.MarkFlagRequired(flag.DownloadSvrAddr)

	stepCmd.Flags().StringVar(&callbackSvrAddr, flag.CallbackSvrAddr, "", "callback server address, for downloading config files")
	_ = stepCmd.MarkFlagRequired(flag.CallbackSvrAddr)

	stepCmd.Flags().StringVar(&deployToken, flag.DeployToken, "", "deploy token, contains the details of files")
	_ = stepCmd.MarkFlagRequired(flag.DeployToken)

	stepCmd.Flags().StringVar(&nodeVersion, flag.NodeVersion, "", "node version, for downloading package version")
	_ = stepCmd.MarkFlagRequired(flag.NodeVersion)

	return stepCmd
}

// GenReleasePkgName generates release package name.
func GenReleasePkgName(nodeRole types.NodeRole, generation types.Generation, version string) string {
	return fmt.Sprintf(
		"gse_%s-%d-%s-%s_%s.%s",
		string(nodeRole),
		int(generation),
		version,
		runtime.GOOS,
		runtime.GOARCH,
		gseReleasePkgExt)
}
