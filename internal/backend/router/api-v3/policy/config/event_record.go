/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package config

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/goasync"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// recordConfigPolicyEvent records a single config policy event.
func (h *handler) recordConfigPolicyEvent(rCtx restserver.IContext, event *types.ConfigPolicyEvent) {
	err := h.goAsyncPool.Run(
		rCtx,
		func(nCtx contextx.IContext) error {
			return h.storageConfigPolicy.CreateManyConfigPolicyEvent(nCtx, event)
		},
		goasync.WithName("record_config_policy_event"),
	)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to submit policy event recording task")
	}
}

// recordConfigPolicyEventsByPolicies builds events from policies and records them via the goasync pool.
func (h *handler) recordConfigPolicyEventsByPolicies(
	rCtx restserver.IContext,
	eventType types.ConfigPolicyEventType,
	policies ...*types.ConfigPolicy,
) {

	err := h.goAsyncPool.Run(
		rCtx,
		func(nCtx contextx.IContext) error {
			events := make([]*types.ConfigPolicyEvent, len(policies))
			for i, cp := range policies {
				events[i] = &types.ConfigPolicyEvent{
					TenantID:         nCtx.TenantID(),
					BizID:            cp.BizID,
					Type:             eventType,
					ConfigPolicyID:   cp.ID,
					ConfigPolicyName: cp.Name,
					ConfigPolicyType: cp.Type,
					Version:          int64(cp.Version),
					Operator:         cp.Operator,
					OperateTime:      time.Now(),
				}
			}

			return h.storageConfigPolicy.CreateManyConfigPolicyEvent(nCtx, events...)
		},
		goasync.WithName("record_config_policy_event_by_policies"),
	)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to submit policy event recording task")
	}
}

// recordConfigPolicyEventsByPolicyIDs queries policies by IDs, builds events, and records them via the goasync pool.
func (h *handler) recordConfigPolicyEventsByPolicyIDs(rCtx restserver.IContext, eventType types.ConfigPolicyEventType, policyIDs ...int64) {
	err := h.goAsyncPool.Run(
		rCtx,
		func(nCtx contextx.IContext) error {
			policies, err := h.getConfigPolicy(nCtx, policyIDs)
			if err != nil {
				return err
			}

			events := make([]*types.ConfigPolicyEvent, len(policies))
			for i, cp := range policies {
				events[i] = &types.ConfigPolicyEvent{
					TenantID:         nCtx.TenantID(),
					BizID:            cp.BizID,
					Type:             eventType,
					ConfigPolicyID:   cp.ID,
					ConfigPolicyName: cp.Name,
					ConfigPolicyType: cp.Type,
					Version:          int64(cp.Version),
					Operator:         cp.Operator,
					OperateTime:      time.Now(),
				}
			}

			return h.storageConfigPolicy.CreateManyConfigPolicyEvent(nCtx, events...)
		},
		goasync.WithName("record_config_policy_event_by_ids"),
	)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to submit policy event recording task")
	}
}
