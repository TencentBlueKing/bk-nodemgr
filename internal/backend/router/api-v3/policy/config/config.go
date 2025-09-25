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
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/configpolicy"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/gin-gonic/gin"
)

const (
	maxConfigPolicyLimit = 1000
)

type handler struct {
	rg      *gin.RouterGroup
	storage configpolicy.IStorage
	logger  logger.ILogger
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:      rg.Group("/config"),
		storage: capability.StorageConfigPolicy,
		logger:  capability.Logger,
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
}

// ListConfigPolicy lists config policy with page and conditions.
func (h *handler) ListConfigPolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ConfigPolicyListReq)
	if err := rCtx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to list config policy, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// only count.
	if req.GetOnlyCount() {
		num, err := h.storage.CountConfigPolicy(
			rCtx,
			req.ConvertConditionsToTypes())
		if err != nil {
			h.logger.ErrorCtxf(rCtx, "failed to list config policy. failed to count host. err: %v", err)
			return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.ConfigPolicyListResp)
		resp.ConvertConfigPoliciesFromTypes(num, nil)

		return resp.GetData(), nil
	}

	hosts, num, err := h.storage.ListConfigPolicy(
		rCtx,
		req.ConvertPageToTypes(maxConfigPolicyLimit),
		req.ConvertConditionsToTypes())
	if err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to list config policy. err: %v", err)
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
		h.logger.ErrorCtxf(rCtx, "failed to get config policy, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// get config policy.
	configPolicy, err := h.storage.GetConfigPolicy(rCtx, req.GetConfigpolicyId())
	if err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to get config policy. err: %v", err)
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
		h.logger.ErrorCtxf(rCtx, "failed to create config policy, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	configPolicy := req.ConvertConfigPolicyToTypes()
	configPolicy.TenantID = rCtx.TenantID()
	configPolicyID, err := h.storage.CreateConfigPolicy(rCtx, configPolicy)
	if err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to create config policy. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.ConfigPolicyCreateResp)
	resp.ConvertConfigPolicyID(configPolicyID)

	return resp.GetData(), nil
}

// UpdateConfigPolicy updates config policy.
func (h *handler) UpdateConfigPolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ConfigPolicyUpdateReq)
	if err := rCtx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to update config policy, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	configPolicy := req.ConvertConfigPolicyToTypes()
	configPolicy.TenantID = rCtx.TenantID()
	if err := h.storage.UpdateConfigPolicy(rCtx, configPolicy); err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to update config policy. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.ConfigPolicyUpdateResp)
	resp.ConvertConfigPolicyID(req.GetConfigpolicyId())

	return resp.GetData(), nil
}

// EnableConfigPolicy enables config policy.
func (h *handler) EnableConfigPolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ConfigPolicyEnableReq)
	if err := rCtx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to enable config policy, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := h.storage.EnableManyConfigPolicy(rCtx, req.GetConfigpolicyId()...); err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to enable config policy. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.ConfigPolicyEnableResp)

	return resp.GetData(), nil
}

// DisableConfigPolicy disables config policy.
func (h *handler) DisableConfigPolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ConfigPolicyDisableReq)
	if err := rCtx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to disable config policy, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := h.storage.DisableManyConfigPolicy(rCtx, req.GetConfigpolicyId()...); err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to disable config policy. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.ConfigPolicyDisableResp)

	return resp.GetData(), nil
}

// DeleteConfigPolicy deletes config policy.
func (h *handler) DeleteConfigPolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ConfigPolicyDeleteReq)
	if err := rCtx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to delete config policy, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := h.storage.DeleteManyConfigPolicy(rCtx, req.GetConfigpolicyId()...); err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to delete config policy. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.ConfigPolicyDeleteResp)

	return resp.GetData(), nil
}
