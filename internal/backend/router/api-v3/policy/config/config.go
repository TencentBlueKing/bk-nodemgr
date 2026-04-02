/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package config provides the config policy router.
package config

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/configpolicy"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/goasync"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
)

const (
	initVersion = 1

	minConfigPolicyPriority int64 = 1

	asyncPoolNum  = 10
	asyncPoolSize = 100
)

type handler struct {
	rg                  *gin.RouterGroup
	storageConfigPolicy configpolicy.IStorage
	goAsyncPool         goasync.IHandler
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	goAsyncPool, _ := goasync.NewHandler(goasync.HandlerOption{
		PoolNum:               asyncPoolNum,
		PerPoolSize:           asyncPoolSize,
		LoadBalancingStrategy: goasync.LoadBalancingStrategyLeastFirst,
	})

	return &handler{
		rg:                  rg.Group("/config"),
		storageConfigPolicy: capability.StorageConfigPolicy,
		goAsyncPool:         goAsyncPool,
	}
}

// Load loads node handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/list", restserver.Handler(h.ListConfigPolicy))
	h.rg.POST("/get", restserver.Handler(h.GetConfigPolicy))
	h.rg.POST("/create", restserver.Handler(h.CreateConfigPolicy))
	h.rg.POST("/update", restserver.Handler(h.UpdateConfigPolicy))
	h.rg.POST("/enable", restserver.Handler(h.EnableConfigPolicy))
	h.rg.POST("/disable", restserver.Handler(h.DisableConfigPolicy))
	h.rg.POST("/delete", restserver.Handler(h.DeleteConfigPolicy))
	h.rg.POST("/reorder_priorities", restserver.Handler(h.ReorderPrioritiesConfigPolicy))
	h.rg.POST("/preview", restserver.Handler(h.PreviewConfigPolicy))

	// package event apis.
	h.rg.POST("/event/list", restserver.Handler(h.ListConfigPolicyEvent))
	h.rg.POST("/event/distinct", restserver.Handler(h.DistinctConfigPolicyEvent))
}

// ListConfigPolicy lists config policy with page and conditions.
func (h *handler) ListConfigPolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ConfigPolicyListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list config policy, failed to decode request body")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	conditions, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list config policy, failed to convert conditions")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}
	// only count.
	if req.GetOnlyCount() {
		num, err := h.storageConfigPolicy.CountConfigPolicy(
			rCtx,
			conditions)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list config policy. failed to count host")

			return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.ConfigPolicyListResp)
		resp.ConvertConfigPoliciesFromTypes(num, nil)

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list config policy, invalid page info")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	hosts, num, err := h.storageConfigPolicy.ListConfigPolicy(rCtx, page, conditions)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list config policy")

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.ConfigPolicyListResp)
	resp.ConvertConfigPoliciesFromTypes(num, hosts)

	return resp.GetData(), nil
}

// GetConfigPolicy gets config policy.
func (h *handler) GetConfigPolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ConfigPolicyGetReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get config policy, failed to decode request body")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// get config policy.
	configPolicy, err := h.storageConfigPolicy.GetConfigPolicy(rCtx, req.GetConfigpolicyId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get config policy")

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.ConfigPolicyGetResp)
	resp.ConvertConfigPolicyFromTypes(configPolicy)

	return resp.GetData(), nil
}

// CreateConfigPolicy creates config policy.
func (h *handler) CreateConfigPolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ConfigPolicyCreateReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create config policy, failed to decode request body")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	configPolicy := req.ConvertConfigPolicyToTypes()
	configPolicy.TenantID = rCtx.TenantID()
	configPolicy.Version = initVersion
	configPolicyID, err := h.storageConfigPolicy.CreateConfigPolicy(rCtx, configPolicy)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create config policy")

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	// record event.
	h.recordConfigPolicyEventsByPolicyIDs(rCtx, types.ConfigPolicyEventTypeCreate, configPolicyID)

	resp := new(protoBackend.ConfigPolicyCreateResp)
	resp.ConvertConfigPolicyID(configPolicyID)

	return resp.GetData(), nil
}

// UpdateConfigPolicy updates config policy.
func (h *handler) UpdateConfigPolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ConfigPolicyUpdateReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update config policy, failed to decode request body")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	configPolicy := req.ConvertConfigPolicyToTypes()
	configPolicy.TenantID = rCtx.TenantID()
	if err := h.storageConfigPolicy.UpdateConfigPolicy(rCtx, configPolicy); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update config policy")

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	// record event.
	h.recordConfigPolicyEventsByPolicyIDs(rCtx, types.ConfigPolicyEventTypeUpdate, configPolicy.ID)

	resp := new(protoBackend.ConfigPolicyUpdateResp)
	resp.ConvertConfigPolicyID(req.GetConfigpolicyId())

	return resp.GetData(), nil
}

// EnableConfigPolicy enables config policy.
func (h *handler) EnableConfigPolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ConfigPolicyEnableReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to enable config policy, failed to decode request body")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := h.storageConfigPolicy.EnableManyConfigPolicy(rCtx, req.GetConfigpolicyId()...); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to enable config policy")

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	// record event.
	h.recordConfigPolicyEventsByPolicyIDs(rCtx, types.ConfigPolicyEventTypeEnable, req.GetConfigpolicyId()...)

	resp := new(protoBackend.ConfigPolicyEnableResp)

	return resp.GetData(), nil
}

// DisableConfigPolicy disables config policy.
func (h *handler) DisableConfigPolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ConfigPolicyDisableReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to disable config policy, failed to decode request body")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := h.storageConfigPolicy.DisableManyConfigPolicy(rCtx, req.GetConfigpolicyId()...); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to disable config policy")

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	// record event.
	h.recordConfigPolicyEventsByPolicyIDs(rCtx, types.ConfigPolicyEventTypeDisable, req.GetConfigpolicyId()...)

	resp := new(protoBackend.ConfigPolicyDisableResp)

	return resp.GetData(), nil
}

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

	// list all enabled config policies for this (biz, type) scope, ordered by priority ascending.
	all, _, err := h.storageConfigPolicy.ListConfigPolicy(rCtx, types.UnlimitedPage(), &types.ConfigPolicyCondition{
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

// DeleteConfigPolicy deletes config policy.
func (h *handler) DeleteConfigPolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ConfigPolicyDeleteReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete config policy, failed to decode request body")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// get the config policy info.
	configpolicy, err := h.getConfigPolicy(rCtx, req.GetConfigpolicyId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete config policy")

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	// delete the config policy.
	if err := h.storageConfigPolicy.DeleteManyConfigPolicy(rCtx, req.GetConfigpolicyId()...); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete config policy")

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	// record event.
	h.recordConfigPolicyEventsByPolicies(rCtx, types.ConfigPolicyEventTypeDelete, configpolicy...)

	resp := new(protoBackend.ConfigPolicyDeleteResp)

	return resp.GetData(), nil
}

// PreviewConfigPolicy previews the merged config for each host.
func (h *handler) PreviewConfigPolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ConfigPolicyPreviewReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to preview config policy, failed to decode request body")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	policyType := types.ConfigPolicyType(req.GetPolicyType())
	previewHosts := req.ConvertPreviewHostsToTypes()

	results, err := h.storageConfigPolicy.PreviewConfigPolicy(rCtx, req.GetBkBizId(), policyType, previewHosts)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to preview config policy")

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.ConfigPolicyPreviewResp)
	resp.ConvertMatchResultsFromTypes(results)

	return resp.GetData(), nil
}

// ListConfigPolicyEvent lists events with page and conditions.
func (h *handler) ListConfigPolicyEvent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ConfigPolicyEventListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list policy event, failed to decode request body")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	conditions, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list policy event, failed to convert conditions")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// only count.
	if req.GetOnlyCount() {
		num, err := h.storageConfigPolicy.CountConfigPolicyEvent(
			rCtx,
		)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list policy event, failed to count event")

			return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.ConfigPolicyEventListResp)
		resp.ConvertConfigPolicyEventsFromTypes(num, nil)

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list policy event, invalid page info")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	events, num, err := h.storageConfigPolicy.ListConfigPolicyEvent(rCtx, page, conditions)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list policy event")

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.ConfigPolicyEventListResp)
	resp.ConvertConfigPolicyEventsFromTypes(num, events)

	return resp.GetData(), nil
}

// DistinctConfigPolicyEvent distincts events with conditions.
// nolint: dupl
func (h *handler) DistinctConfigPolicyEvent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ConfigPolicyEventDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct policy event, failed to decode request body")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	conditions, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct policy event, failed to convert conditions")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	result, err := h.storageConfigPolicy.DistinctConfigPolicyEvent(
		rCtx,
		types.NewConfigPolicyEventDistinctRequestAllSet(),
		conditions)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct policy event, failed to distinct host fields")

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.ConfigPolicyEventDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}

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
				return fmt.Errorf("failed to get policy info: %w", err)
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

func (h *handler) getConfigPolicy(nCtx contextx.IContext, configpolicyID []int64) ([]*types.ConfigPolicy, error) {
	// get the config policy info.
	configpolicies, _, err := h.storageConfigPolicy.ListConfigPolicy(nCtx, types.UnlimitedPage(),
		&types.ConfigPolicyCondition{
			ExactInclude: &types.ConfigPolicyExactFields{
				ConfigPolicyID: configpolicyID,
			},
		})
	if err != nil {
		return nil, fmt.Errorf("failed to get config policy: %w", err)
	}

	if len(configpolicies) == 0 {
		return nil, fmt.Errorf("config policy not found")
	}

	return configpolicies, nil
}
