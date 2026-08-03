/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package backend provides CLI commands for backend admin operations.
package backend

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backendadmin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/spf13/cobra"
)

// NewBackendCMD creates a new backend command with the given handler factory.
func NewBackendCMD(handlerFactory func(configPath string) (backendadmin.IHandler, error)) *cobra.Command {
	var configPath string
	var tenantID string
	var loginName string
	var handler backendadmin.IHandler

	cmd := &cobra.Command{
		Use:   "backend",
		Short: "backend admin operations",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateInitTenantFlags(cmd); err != nil {
				return err
			}

			if err := validateSyncUnassignedNetworkUnitFlags(cmd); err != nil {
				return err
			}

			h, err := handlerFactory(configPath)
			if err != nil {
				return fmt.Errorf("failed to create backend handler: %w", err)
			}
			handler = h

			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&configPath, "file", "f", "/bk-nodemgr/etc/backend_conf.yaml", "path of backend config file")
	cmd.PersistentFlags().StringVar(&tenantID, "tenant-id", "default", "tenant id for authentication")
	cmd.PersistentFlags().StringVar(&loginName, "login-name", "admin", "login name for authentication")
	// Closure to access handler and auth info after PersistentPreRunE
	getHandler := func() backendadmin.IHandler { return handler }
	getSyncHandler := func() backendadmin.ISyncUnassignedAgentNetworkUnitHandler {
		syncHandler, _ := handler.(backendadmin.ISyncUnassignedAgentNetworkUnitHandler)
		return syncHandler
	}
	getTenantHandler := func() backendadmin.IInitTenantHandler {
		tenantHandler, _ := handler.(backendadmin.IInitTenantHandler)
		return tenantHandler
	}
	getAuthInfo := func() (string, string) { return tenantID, loginName }
	cmd.AddCommand(NewNetworkUnitSegmentRulesCMD(getHandler, getAuthInfo))
	cmd.AddCommand(NewNodeCMD(getSyncHandler, getAuthInfo))
	cmd.AddCommand(NewTenantCMD(getTenantHandler, getAuthInfo))

	return cmd
}

// NewNetworkUnitSegmentRulesCMD creates a new command for managing network unit segment rules.
func NewNetworkUnitSegmentRulesCMD(getHandler func() backendadmin.IHandler, getAuthInfo func() (string, string)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "networkunit_segment_rules",
		Short: "manage backend networkunit segment rules",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "get backend networkunit segment rules",
		RunE: func(cmd *cobra.Command, _ []string) error {
			tenantID, loginName := getAuthInfo()
			nCtx := contextx.New(cmd.Context(), contextx.WithTenantID(tenantID), contextx.WithLoginName(loginName))
			rules, err := getHandler().GetNetworkUnitSegmentRules(nCtx)
			if err != nil {
				return err
			}

			data, err := json.MarshalIndent(rules, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal networkunit segment rules: %w", err)
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), string(data))

			return err
		},
	})

	var rulesFile string
	upsertCMD := &cobra.Command{
		Use:   "upsert",
		Short: "upsert backend networkunit segment rules",
		RunE: func(cmd *cobra.Command, _ []string) error {
			tenantID, loginName := getAuthInfo()

			// nolint: gosec
			content, err := os.ReadFile(rulesFile)
			if err != nil {
				return fmt.Errorf("read rules file: %w", err)
			}

			var cfg types.NetworkUnitSegmentRuleConfig
			if err := json.Unmarshal(content, &cfg); err != nil {
				return fmt.Errorf("unmarshal rules file: %w", err)
			}

			nCtx := contextx.New(cmd.Context(), contextx.WithTenantID(tenantID), contextx.WithLoginName(loginName))
			if err := getHandler().UpsertNetworkUnitSegmentRules(nCtx, cfg); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Successfully upserted networkunit segment rules")

			return err
		},
	}
	upsertCMD.Flags().StringVar(&rulesFile, "rules-file", "", "path to rules json file")
	_ = upsertCMD.MarkFlagRequired("rules-file")
	cmd.AddCommand(upsertCMD)

	return cmd
}
