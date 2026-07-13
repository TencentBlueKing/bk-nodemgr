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
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ReorderPrioritiesConfigPolicy reorders config policy priorities within a (biz, type) scope.
func (h *handler) ReorderPrioritiesConfigPolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ConfigPolicyPriorityReorderReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to reorder priorities for config policy, failed to decode request body")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	bizID := req.GetBkBizId()
	policyType := types.ConfigPolicyType(req.GetConfigpolicyType())
	orderedPolicyIDs := req.GetOrderedConfigpolicyId()

	// Check permission with biz resource.
	resources := authRouter.BuildBizResources(bizID)
	if authErr := h.authorizer.Check(rCtx, auth.ActionConfigPolicyManage, resources); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to reorder priorities for config policy, permission denied")
		return nil, errf.ErrWrap(errf.PermissionDenied, authErr)
	}

	// list all enabled config policies for this (biz, type) scope, ordered by priority ascending.
	all, err := h.storageConfigPolicy.ListConfigPolicyWithoutCount(rCtx, types.UnlimitedPage(), &types.ConfigPolicyCondition{
		ExactInclude: &types.ConfigPolicyExactFields{
			BizID:   []int64{bizID},
			Type:    []types.ConfigPolicyType{policyType},
			Enabled: []bool{true},
		},
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to reorder priorities for config policy, failed to list config policies")

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	// build sorted policy IDs from all and ordered IDs.
	sortedIDs, err := buildReorderedPolicyIDs(all, orderedPolicyIDs)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to reorder priorities for config policy, failed to build reordered policy IDs")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// assign priorities to the reordered policy IDs.
	// this will assign priorities 1..N to the ordered IDs, and N+1.. to the remaining IDs.
	priorities := assignPolicyPriorities(sortedIDs)

	// update priorities.
	if err := h.storageConfigPolicy.UpdatePriorityManyConfigPolicy(rCtx, priorities); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to reorder priorities for config policy")

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	// record event.
	// reorder is a biz-level operation, so only biz_id is recorded without specific policy info.
	reorderEvent := &types.ConfigPolicyEvent{
		TenantID:         rCtx.TenantID(),
		BizID:            bizID,
		Type:             types.ConfigPolicyEventTypeReorderPriorities,
		ConfigPolicyType: policyType,
		Operator:         rCtx.BKUsername(),
		OperateTime:      time.Now(),
	}
	h.recordConfigPolicyEvent(rCtx, reorderEvent)

	resp := new(protoBackend.ConfigPolicyPriorityReorderResp)

	return resp.GetData(), nil
}

// buildReorderedPolicyIDs returns orderedIDs followed by the remaining IDs in all in their original order.
func buildReorderedPolicyIDs(all []*types.ConfigPolicy, orderedIDs []int64) ([]int64, error) {
	if err := checkPolicyIDsExist(all, orderedIDs); err != nil {
		return nil, err
	}

	orderedSet := make(map[int64]struct{}, len(orderedIDs))
	for _, id := range orderedIDs {
		orderedSet[id] = struct{}{}
	}

	result := make([]int64, 0, len(all))
	result = append(result, orderedIDs...)
	for _, cp := range all {
		if _, ok := orderedSet[cp.ID]; ok {
			continue
		}

		result = append(result, cp.ID)
	}

	return result, nil
}

func checkPolicyIDsExist(all []*types.ConfigPolicy, givenIDs []int64) error {
	allSet := make(map[int64]struct{}, len(all))
	for _, cp := range all {
		allSet[cp.ID] = struct{}{}
	}

	for _, id := range givenIDs {
		if _, ok := allSet[id]; !ok {
			return fmt.Errorf("ordered config policy id not found. config-policy-id(%d)", id)
		}
	}

	return nil
}

func assignPolicyPriorities(ids []int64) map[int64]int64 {
	priorities := make(map[int64]int64, len(ids))
	for i, id := range ids {
		priorities[id] = minConfigPolicyPriority + int64(i)
	}

	return priorities
}
