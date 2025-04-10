/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package main ...
package main

import (
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/filedownloader"
	"github.com/spf13/cobra"
)

// NewStepDownloadFiles ...
func NewStepDownloadFiles() *cobra.Command {
	var (
		downloadEndpoint string
		version          int
		tag              string
		token            string
		callBackEndPoint string
	)
	stepCmd := &cobra.Command{
		Use:   "step_download_files",
		Short: "download files",
		Long:  "download files",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			if err := SetNodeGeneration(version); err != nil {
				return err
			}

			if err := SetNodeVersion(tag); err != nil {
				return err
			}

			if err := SetDownloadEndPoint(downloadEndpoint); err != nil {
				return err
			}

			if err := SetCallbackEndPoint(callBackEndPoint); err != nil {
				return err
			}

			if err := SetToken(token); err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			step := filedownloader.NewStep(filedownloader.StepArgs{
				DownloadPoint:        GetDownloadEndPoint(),
				CallbackEndpoint:     GetCallBackEndpoint(),
				PkgGeneration:        GetNodePkgGeneration(),
				PkgPath:              GetGsePkgPath(),
				PkgVersion:           GetNodePkgVersion(),
				NodeRole:             GetNodeRole(),
				Token:                GetToken(),
				TmpAgentConfPath:     GetTmpAgentConfPath(),
				TmpFileProxyConfPath: GetTmpFileProxyConfPath(),
				TmpDataProxyConfPath: GetTmpDataProxyConfPath(),
				CheckListPath:        GetPreCheckFilePath(),
			})
			err := step.Run(cmd.Context())
			if err != nil {
				return err
			}

			return nil
		},
	}

	stepCmd.Flags().
		StringVar(&downloadEndpoint, CmdFlagDownloadEndpoint, "", "download endpoint")
	stepCmd.Flags().
		IntVar(&version, CmdFlagPkgGeneration, CmdDefaultPkgGeneration, "this is the gse version which will be installed")
	stepCmd.Flags().
		StringVar(&tag, CmdFlagPkgVersion, "", "this gse node version tag which will be installed")
	stepCmd.Flags().
		StringVar(&callBackEndPoint, CmdFlagCallbackEndpoint, "", "callback endpoint")
	stepCmd.Flags().
		StringVar(&token, CmdFlagToken, "", "token")

	_ = stepCmd.MarkPersistentFlagRequired(CmdFlagPkgVersion)
	_ = stepCmd.MarkPersistentFlagRequired(CmdFlagCallbackEndpoint)
	_ = stepCmd.MarkPersistentFlagRequired(CmdFlagDownloadEndpoint)

	return stepCmd
}
