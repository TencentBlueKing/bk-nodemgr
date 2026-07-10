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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backendadmin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/spf13/cobra"
)

type syncUnassignedNetworkUnitOptions struct {
	bkBizID string
	all     bool
}

// NewNodeCMD creates a new command for backend node operations.
func NewNodeCMD(
	getHandler func() backendadmin.ISyncUnassignedAgentNetworkUnitHandler,
	getAuthInfo func() (string, string),
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "manage backend node operations",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	agentCMD := &cobra.Command{
		Use:   "agent",
		Short: "manage backend node agent operations",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	agentCMD.AddCommand(NewSyncUnassignedNetworkUnitCMD(getHandler, getAuthInfo))
	cmd.AddCommand(agentCMD)

	return cmd
}

// NewSyncUnassignedNetworkUnitCMD creates a command for syncing unassigned agent network units.
func NewSyncUnassignedNetworkUnitCMD(
	getHandler func() backendadmin.ISyncUnassignedAgentNetworkUnitHandler,
	getAuthInfo func() (string, string),
) *cobra.Command {

	opts := syncUnassignedNetworkUnitOptions{}
	cmd := &cobra.Command{
		Use:          "sync-unassigned-network-unit",
		Short:        "sync unassigned agent hosts to recommended network units",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			bizIDs, err := opts.bizIDs()
			if err != nil {
				return err
			}

			handler := getHandler()
			if handler == nil {
				return fmt.Errorf("backend handler does not support syncing unassigned agent network units")
			}

			tenantID, loginName := getAuthInfo()
			nCtx := contextx.New(cmd.Context(), contextx.WithTenantID(tenantID), contextx.WithLoginName(loginName))
			result, err := handler.SyncUnassignedAgentNetworkUnit(nCtx, bizIDs)
			if err != nil {
				return err
			}

			if err := writeSyncUnassignedNetworkUnitResult(cmd, result); err != nil {
				return err
			}

			if result.FailedCount > 0 {
				return fmt.Errorf("sync unassigned network unit completed with %d failed hosts", result.FailedCount)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&opts.bkBizID, "bk-biz-id", "", "comma-separated bk biz ids to sync, for example 2,3")
	cmd.Flags().BoolVar(&opts.all, "all", false, "sync all businesses explicitly")
	cmd.MarkFlagsOneRequired("bk-biz-id", "all")
	cmd.MarkFlagsMutuallyExclusive("bk-biz-id", "all")

	return cmd
}

func validateSyncUnassignedNetworkUnitFlags(cmd *cobra.Command) error {
	bkBizIDFlag := cmd.Flags().Lookup("bk-biz-id")
	allFlag := cmd.Flags().Lookup("all")
	if bkBizIDFlag == nil || allFlag == nil {
		return nil
	}

	bkBizIDChanged := bkBizIDFlag.Changed
	allChanged := allFlag.Changed
	if bkBizIDChanged == allChanged {
		return fmt.Errorf("exactly one of --bk-biz-id or --all is required")
	}

	return nil
}

func (o syncUnassignedNetworkUnitOptions) bizIDs() ([]int64, error) {
	if o.all {
		return nil, nil
	}

	parts := strings.Split(o.bkBizID, ",")
	bizIDs := make([]int64, 0, len(parts))
	for _, part := range parts {
		bizID, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid bk biz id(%s): %w", part, err)
		}

		if bizID <= 0 {
			return nil, fmt.Errorf("invalid bk biz id(%d): must be positive", bizID)
		}

		bizIDs = append(bizIDs, bizID)
	}

	return bizIDs, nil
}

func writeSyncUnassignedNetworkUnitResult(cmd *cobra.Command, result *types.NodeAgentAssignUnitResult) error {
	data, err := json.MarshalIndent(struct {
		SuccessCount  int64    `json:"success_count"`
		FailedCount   int64    `json:"failed_count"`
		FailedReasons []string `json:"failed_reasons"`
	}{
		SuccessCount:  result.SuccessCount,
		FailedCount:   result.FailedCount,
		FailedReasons: result.FailedReasons,
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal sync unassigned network unit result: %w", err)
	}

	_, err = fmt.Fprintln(cmd.OutOrStdout(), string(data))

	return err
}
