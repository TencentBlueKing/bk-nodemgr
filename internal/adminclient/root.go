/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package adminclient provides the root command for the admin client CLI.
package adminclient

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/adminclient/backend"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backendadmin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/version"
	"github.com/spf13/cobra"
)

// NewRootCMD creates the root command for the admin client.
func NewRootCMD(handlerFactory func(configPath string) (backendadmin.IHandler, error)) *cobra.Command {
	rootCMD := &cobra.Command{
		Use:     "bk-nodeman-adminclient",
		Short:   "bk-nodeman admin client",
		Version: version.FormatVersion(),
		PreRun: func(_ *cobra.Command, _ []string) {
			fmt.Println(version.GetStartInfo())
		},
	}

	rootCMD.AddCommand(backend.NewBackendCMD(handlerFactory))

	return rootCMD
}
