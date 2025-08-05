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
	"encoding/json"

	"github.com/TencentBlueKing/bk-nodemgr/internal/application/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/storage/cptemplate"
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backend"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
)

const (
	maxConfigPolicyLimit = 1000
)

type handler struct {
	rg                          *gin.RouterGroup
	backendHandler              backend.Handler
	storageConfigPolicyTemplate cptemplate.IStorage
	logger                      logger.Logger
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:                          rg.Group("/config"),
		backendHandler:              capability.BackendHandler,
		storageConfigPolicyTemplate: capability.StorageConfigPolicyTemplate,
		logger:                      capability.Logger,
	}
}

// Load loads node handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/list", restserver.Handler(h.ListConfigPolicy))
	h.rg.POST("/get", restserver.Handler(h.GetConfigPolicy))
	h.rg.POST("/template", restserver.Handler(h.TemplateConfigPolicy))
	h.rg.POST("/create", restserver.Handler(h.CreateConfigPolicy))
	h.rg.POST("/update", restserver.Handler(h.UpdateConfigPolicy))
	h.rg.POST("/enable", restserver.Handler(h.EnableConfigPolicy))
	h.rg.POST("/disable", restserver.Handler(h.DisableConfigPolicy))
	h.rg.POST("/delete", restserver.Handler(h.DeleteConfigPolicy))
}

// ListConfigPolicy lists config policy with page and conditions.
func (h *handler) ListConfigPolicy(ctx *restserver.Context) (interface{}, error) {
	req := new(protoApplication.ConfigPolicyListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list config policy, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// only count.
	if req.GetOnlyCount() {
		num, err := h.backendHandler.CountConfigPolicy(
			ctx,
			req.ConvertConditionsToTypes())
		if err != nil {
			h.logger.ErrorCtxf(ctx, "failed to list config policy. failed to count host. err: %v", err)
			return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
		}

		resp := new(protoApplication.ConfigPolicyListResp)
		resp.ConvertConfigPoliciesFromTypes(num, nil)

		return resp.GetData(), nil
	}

	hosts, num, err := h.backendHandler.ListConfigPolicy(
		ctx,
		req.ConvertPageToTypes(maxConfigPolicyLimit),
		req.ConvertConditionsToTypes())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list config policy. err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.ConfigPolicyListResp)
	resp.ConvertConfigPoliciesFromTypes(num, hosts)

	return resp.GetData(), nil
}

// GetConfigPolicy gets config policy.
func (h *handler) GetConfigPolicy(ctx *restserver.Context) (interface{}, error) {
	req := new(protoApplication.ConfigPolicyGetReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to get config policy, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// get config policy.
	configPolicy, err := h.backendHandler.GetConfigPolicy(ctx, req.GetConfigpolicyId())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to get config policy. err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	// get template.
	template, err := h.storageConfigPolicyTemplate.GetConfigPolicyTemplate(ctx, configPolicy.ID)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to get config policy. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	var blocks []types.ConfigPolicyTemplateBlock
	if err := json.Unmarshal([]byte(template.Template), &blocks); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to get config policy template. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// refresh template with new items.
	templates := getConfigTemplate(configPolicy.NodeRole)
	for _, block := range templates {
		blocks = insertTemplateBlock(blocks, block)
	}

	resp := new(protoApplication.ConfigPolicyGetResp)
	resp.ConvertConfigPolicyFromTypes(configPolicy, blocks)

	return resp.GetData(), nil
}

// TemplateConfigPolicy gets config policy template.
func (h *handler) TemplateConfigPolicy(ctx *restserver.Context) (interface{}, error) {
	req := new(protoApplication.ConfigPolicyTemplateReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to get config policy template, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	resp := new(protoApplication.ConfigPolicyTemplateResp)
	resp.ConvertTemplateFromTypes(getConfigTemplate(types.NodeRole(req.GetNodeRole())))

	return resp.GetData(), nil
}

// CreateConfigPolicy creates config policy.
func (h *handler) CreateConfigPolicy(ctx *restserver.Context) (interface{}, error) {
	req := new(protoApplication.ConfigPolicyCreateReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to create config policy, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	configPolicy, cpTemplate := req.ConvertConfigPolicyToTypes()

	// create config policy.
	addCustomConfig(configPolicy, cpTemplate)
	configPolicy.TenantID = ctx.TenantID
	configPolicyID, err := h.backendHandler.CreateConfigPolicy(ctx, configPolicy)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to create config policy. err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	// create template.
	template, err := json.Marshal(cpTemplate)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to create config policy, failed to marshal template. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}
	if err := h.storageConfigPolicyTemplate.UpsertManyConfigPolicyTemplate(ctx, &types.ConfigPolicyTemplate{
		TenantID: configPolicy.TenantID,
		ID:       configPolicyID,
		Template: string(template),
	}); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to create config policy, failed to create template. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoApplication.ConfigPolicyCreateResp)
	resp.ConvertConfigPolicyID(configPolicyID)

	return resp.GetData(), nil
}

// UpdateConfigPolicy updates config policy.
func (h *handler) UpdateConfigPolicy(ctx *restserver.Context) (interface{}, error) {
	req := new(protoApplication.ConfigPolicyUpdateReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to update config policy, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	configPolicy, cpTemplate := req.ConvertConfigPolicyToTypes()

	// update config policy.
	addCustomConfig(configPolicy, cpTemplate)
	configPolicy.TenantID = ctx.TenantID
	if _, err := h.backendHandler.UpdateConfigPolicy(ctx, configPolicy); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to update config policy. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	// upsert template.
	template, err := json.Marshal(cpTemplate)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to update config policy, failed to marshal template. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}
	if err := h.storageConfigPolicyTemplate.UpsertManyConfigPolicyTemplate(ctx, &types.ConfigPolicyTemplate{
		TenantID: configPolicy.TenantID,
		ID:       configPolicy.ID,
		Template: string(template),
	}); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to update config policy, failed to update template. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoApplication.ConfigPolicyUpdateResp)
	resp.ConvertConfigPolicyID(req.GetConfigpolicyId())

	return resp.GetData(), nil
}

// EnableConfigPolicy enables config policy.
func (h *handler) EnableConfigPolicy(ctx *restserver.Context) (interface{}, error) {
	req := new(protoApplication.ConfigPolicyEnableReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to enable config policy, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := h.backendHandler.EnableConfigPolicy(ctx, req.GetConfigpolicyId()...); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to enable config policy. err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.ConfigPolicyEnableResp)

	return resp.GetData(), nil
}

// DisableConfigPolicy disables config policy.
func (h *handler) DisableConfigPolicy(ctx *restserver.Context) (interface{}, error) {
	req := new(protoApplication.ConfigPolicyDisableReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to disable config policy, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := h.backendHandler.DisableConfigPolicy(ctx, req.GetConfigpolicyId()...); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to disable config policy. err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.ConfigPolicyDisableResp)

	return resp.GetData(), nil
}

// DeleteConfigPolicy deletes config policy.
func (h *handler) DeleteConfigPolicy(ctx *restserver.Context) (interface{}, error) {
	req := new(protoApplication.ConfigPolicyDeleteReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to delete config policy, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := h.backendHandler.DeleteConfigPolicy(ctx, req.GetConfigpolicyId()...); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to delete config policy. err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.ConfigPolicyDeleteResp)

	return resp.GetData(), nil
}

func insertTemplateBlock(
	src []types.ConfigPolicyTemplateBlock, extra types.ConfigPolicyTemplateBlock) []types.ConfigPolicyTemplateBlock {

	result := make([]types.ConfigPolicyTemplateBlock, 0)

	blockFound := false
	for _, block := range src {
		if block.ID != extra.ID {
			result = append(result, block)

			continue
		}
		blockFound = true

		for _, newItem := range extra.Items {
			itenFound := false
			for idx, item := range block.Items {
				if newItem.ID != item.ID {
					continue
				}
				itenFound = true

				block.Items[idx].NameEN = newItem.NameEN
				block.Items[idx].NameZH = newItem.NameZH
				block.Items[idx].RemarkEN = newItem.RemarkEN
				block.Items[idx].RemarkZH = newItem.RemarkZH
				block.Items[idx].Type = newItem.Type
				block.Items[idx].Key = newItem.Key

				break
			}

			if !itenFound {
				block.Items = append(block.Items, newItem)
			}
		}

		block.TitleEN = extra.TitleEN
		block.TitleZH = extra.TitleZH

		result = append(result, block)
	}

	if !blockFound {
		result = append(result, extra)
	}

	return result
}

// nolint:funlen,fnsize,gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func addCustomConfig(configPolicy *types.ConfigPolicy, blocks []types.ConfigPolicyTemplateBlock) {
	if configPolicy == nil {
		return
	}

	if configPolicy.Configs == nil {
		configPolicy.Configs = make(map[string]any)
	}

	templates := getConfigTemplate(configPolicy.NodeRole)
	for _, block := range blocks {
		for _, item := range block.Items {
			if !item.Enabled {
				continue
			}

			// common configs, not value group. set key value directly.
			if !isValueGroupKey(item.Key) {
				switch item.Type {
				case types.ConfigPolicyTemplateTypeString, types.ConfigPolicyTemplateTypeStringSelect:
					configPolicy.Configs[item.Key] = item.ValueString

				case types.ConfigPolicyTemplateTypeInt, types.ConfigPolicyTemplateTypeIntSelect:
					configPolicy.Configs[item.Key] = item.ValueInt

				case types.ConfigPolicyTemplateTypeBool:
					configPolicy.Configs[item.Key] = item.ValueBool

				default:
					continue
				}

				continue
			}

			templateItem, ok := findTemplateItem(templates, block.ID, item.ID)
			if !ok {
				continue
			}

			// value group configs, fixed values.
			if len(templateItem.ValueGroupFixed) > 0 {
				if item.Type == types.ConfigPolicyTemplateTypeBool && item.ValueBool {
					for k, v := range templateItem.ValueGroupFixed {
						configPolicy.Configs[k] = v
					}
				}

				continue
			}

			// value group configs, assign values.
			if len(templateItem.ValueGroupFixed) > 0 {
				switch item.Type {
				case types.ConfigPolicyTemplateTypeString, types.ConfigPolicyTemplateTypeStringSelect:
					for k := range templateItem.ValueGroupAssigned {
						configPolicy.Configs[k] = item.ValueString
					}

				case types.ConfigPolicyTemplateTypeInt, types.ConfigPolicyTemplateTypeIntSelect:
					for k := range templateItem.ValueGroupAssigned {
						configPolicy.Configs[k] = item.ValueInt
					}

				case types.ConfigPolicyTemplateTypeBool:
					for k := range templateItem.ValueGroupAssigned {
						configPolicy.Configs[k] = item.ValueBool
					}

				default:
					continue
				}

				continue
			}
		}
	}
}
