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
	"context"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/configpolicy"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
)

const (
	initVersion = 1
)

type handler struct {
	rg      *gin.RouterGroup
	storage configpolicy.IStorage
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:      rg.Group("/config"),
		storage: capability.StorageConfigPolicy,
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
		num, err := h.storage.CountConfigPolicy(
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

	hosts, num, err := h.storage.ListConfigPolicy(rCtx, page, conditions)
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
	configPolicy, err := h.storage.GetConfigPolicy(rCtx, req.GetConfigpolicyId())
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
	configPolicyID, err := h.storage.CreateConfigPolicy(rCtx, configPolicy)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create config policy")

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	// record event
	configPolicy.ID = configPolicyID
	h.recordCreateEvent(rCtx, types.ConfigPolicyEventTypeCreate, configPolicy)

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
	if err := h.storage.UpdateConfigPolicy(rCtx, configPolicy); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update config policy")

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	// record event
	h.recordChangesEvent(rCtx, types.ConfigPolicyEventTypeUpdate, configPolicy.ID)

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

	if err := h.storage.EnableManyConfigPolicy(rCtx, req.GetConfigpolicyId()...); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to enable config policy")

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	// record event
	h.recordChangesEvent(rCtx, types.ConfigPolicyEventTypeEnable, req.GetConfigpolicyId()...)

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

	if err := h.storage.DisableManyConfigPolicy(rCtx, req.GetConfigpolicyId()...); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to disable config policy")

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	// record event
	h.recordChangesEvent(rCtx, types.ConfigPolicyEventTypeDisable, req.GetConfigpolicyId()...)

	resp := new(protoBackend.ConfigPolicyDisableResp)

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
	if err := h.storage.DeleteManyConfigPolicy(rCtx, req.GetConfigpolicyId()...); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete config policy")

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	// record event
	h.recordDeleteEvent(rCtx, types.ConfigPolicyEventTypeDelete, configpolicy)

	resp := new(protoBackend.ConfigPolicyDeleteResp)

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
		num, err := h.storage.CountConfigPolicyEvent(
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

	events, num, err := h.storage.ListConfigPolicyEvent(rCtx, page, conditions)
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

	result, err := h.storage.DistinctConfigPolicyEvent(
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

func (h *handler) recordCreateEvent(rCtx restserver.IContext, eventType types.ConfigPolicyEventType, configPolicy *types.ConfigPolicy) {
	tenantID := rCtx.TenantID()

	go func() {
		event := &types.ConfigPolicyEvent{
			TenantID:         tenantID,
			Type:             eventType,
			ConfigPolicyID:   configPolicy.ID,
			ConfigPolicyName: configPolicy.Name,
			ConfigPolicyType: configPolicy.Type,
			Version:          int64(configPolicy.Version),
			OperateTime:      time.Now(),
			Operator:         configPolicy.Operator,
		}

		if err := h.storage.CreateManyConfigPolicyEvent(contextx.New(context.Background(), contextx.WithTenantID(tenantID)), event); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to record policy event, failed to create event")
		}
	}()
}

func (h *handler) recordDeleteEvent(rCtx restserver.IContext, eventType types.ConfigPolicyEventType, configPolicy []*types.ConfigPolicy) {
	tenantID := rCtx.TenantID()

	go func() {
		events := make([]*types.ConfigPolicyEvent, len(configPolicy))
		for idx, cp := range configPolicy {
			event := &types.ConfigPolicyEvent{
				TenantID:         tenantID,
				Type:             eventType,
				ConfigPolicyID:   cp.ID,
				ConfigPolicyName: cp.Name,
				ConfigPolicyType: cp.Type,
				Version:          int64(cp.Version),
				OperateTime:      time.Now(),
				Operator:         cp.Operator,
			}
			events[idx] = event
		}
		if err := h.storage.CreateManyConfigPolicyEvent(contextx.New(context.Background(), contextx.WithTenantID(tenantID)), events...); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to record policy event, failed to create event")
		}
	}()
}
func (h *handler) recordChangesEvent(rCtx restserver.IContext, eventType types.ConfigPolicyEventType, configpolicyID ...int64) {
	tenantID := rCtx.TenantID()

	go func() {
		newCtx := contextx.New(contextx.Background(), contextx.WithTenantID(tenantID))

		// get the config policy info.
		configpolicy, err := h.getConfigPolicy(newCtx, configpolicyID)
		if err != nil {
			logger.G.Sys().WithErr(err).Error("failed to record policy event, failed to get policy info")

			return
		}

		events := make([]*types.ConfigPolicyEvent, len(configpolicy))
		for idx, cp := range configpolicy {
			event := &types.ConfigPolicyEvent{
				TenantID:         tenantID,
				Type:             eventType,
				ConfigPolicyID:   cp.ID,
				ConfigPolicyName: cp.Name,
				ConfigPolicyType: cp.Type,
				Version:          int64(cp.Version),
				Operator:         cp.Operator,
				OperateTime:      time.Now(),
			}

			events[idx] = event
		}

		if err := h.storage.CreateManyConfigPolicyEvent(newCtx, events...); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to record policy event, failed to create event")
		}
	}()
}

func (h *handler) getConfigPolicy(nCtx contextx.IContext, configpolicyID []int64) ([]*types.ConfigPolicy, error) {
	// get the config policy info.
	configpolicies, _, err := h.storage.ListConfigPolicy(nCtx, types.UnlimitedPage(),
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
