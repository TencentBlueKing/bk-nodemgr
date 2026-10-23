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

package config

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/goasync"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// recordConfigPolicyEventsByPolicies snapshots successful mutations before asynchronous persistence.
func (h *handler) recordConfigPolicyEventsByPolicies(
	rCtx restserver.IContext,
	eventType types.ConfigPolicyEventType,
	policies ...*types.ConfigPolicy,
) {

	if len(policies) == 0 {
		return
	}

	operator := rCtx.BKUsername()
	operateTime := time.Now()
	events := make([]*types.ConfigPolicyEvent, len(policies))
	for i, cp := range policies {
		events[i] = &types.ConfigPolicyEvent{
			TenantID:         cp.TenantID,
			BizID:            cp.BizID,
			Type:             eventType,
			ConfigPolicyID:   cp.ID,
			ConfigPolicyName: cp.Name,
			ConfigPolicyType: cp.Type,
			Version:          int64(cp.Version),
			Operator:         operator,
			OperateTime:      operateTime,
		}
	}

	err := h.goAsyncPool.Run(
		rCtx,
		func(nCtx contextx.IContext) error {
			return h.storageConfigPolicy.CreateManyConfigPolicyEvent(nCtx, events...)
		},
		goasync.WithName("record_config_policy_event_by_policies"),
	)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to submit policy event recording task")
	}
}
