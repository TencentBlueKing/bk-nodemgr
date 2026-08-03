/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package backend

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backendadmin"
	"github.com/spf13/cobra"
)

const targetTenantIDFlag = "target-tenant-id"

// NewTenantCMD creates a command for backend tenant operations.
func NewTenantCMD(
	getHandler func() backendadmin.IInitTenantHandler,
	getAuthInfo func() (string, string),
) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "tenant",
		Short: "manage backend tenant operations",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(NewInitTenantCMD(getHandler, getAuthInfo))

	return cmd
}

// NewInitTenantCMD creates a command for initializing a backend tenant.
func NewInitTenantCMD(
	getHandler func() backendadmin.IInitTenantHandler,
	getAuthInfo func() (string, string),
) *cobra.Command {

	var targetTenantID string
	cmd := &cobra.Command{
		Use:          "init",
		Short:        "initialize backend tenant data",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			handler := getHandler()
			if handler == nil {
				return fmt.Errorf("backend handler does not support initializing tenants")
			}

			tenantID, loginName := getAuthInfo()
			nCtx := contextx.New(cmd.Context(), contextx.WithTenantID(tenantID), contextx.WithLoginName(loginName))
			if err := handler.InitTenant(nCtx, targetTenantID); err != nil {
				return err
			}

			_, err := fmt.Fprintln(cmd.OutOrStdout(), "Successfully initialized tenant")

			return err
		},
	}

	cmd.Flags().StringVar(&targetTenantID, targetTenantIDFlag, "", "tenant id to initialize")
	_ = cmd.MarkFlagRequired(targetTenantIDFlag)

	return cmd
}

func validateInitTenantFlags(cmd *cobra.Command) error {
	targetTenantID := cmd.Flags().Lookup(targetTenantIDFlag)
	if targetTenantID == nil || targetTenantID.Changed {
		return nil
	}

	return fmt.Errorf("required flag(s) \"%s\" not set", targetTenantIDFlag)
}
