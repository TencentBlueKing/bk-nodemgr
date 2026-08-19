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

// Package config provides the config policy router.
package config

import (
	"encoding/json"

	"github.com/TencentBlueKing/bk-nodemgr/internal/application/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/storage/cptemplate"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backend"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg                          *gin.RouterGroup
	backendHandler              backend.IHandler
	storageConfigPolicyTemplate cptemplate.IStorage
	configPolicyOptionSet       types.ConfigPolicyOptionSet
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:                          rg.Group("/config"),
		backendHandler:              capability.BackendHandler,
		storageConfigPolicyTemplate: capability.StorageConfigPolicyTemplate,
		configPolicyOptionSet:       capability.ConfigPolicyOptionSet,
	}
}

// Load loads node handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/list", restserver.Handler(h.ListConfigPolicy))
	h.rg.POST("/get", restserver.Handler(h.GetConfigPolicy))
	h.rg.POST("/list_platform", restserver.Handler(h.ListConfigPolicyPlatform))
	h.rg.POST("/get_template", restserver.Handler(h.GetConfigPolicyTemplate))
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
	req := new(protoApplication.ConfigPolicyListReq)
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
		num, err := h.backendHandler.CountConfigPolicy(
			rCtx,
			conditions)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list config policy. failed to count host")
			return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
		}

		resp := new(protoApplication.ConfigPolicyListResp)
		resp.ConvertConfigPoliciesFromTypes(num, nil)

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list config policy, invalid page info")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	hosts, num, err := h.backendHandler.ListConfigPolicy(rCtx, page, conditions)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list config policy")
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.ConfigPolicyListResp)
	resp.ConvertConfigPoliciesFromTypes(num, hosts)

	return resp.GetData(), nil
}

// GetConfigPolicy gets config policy.
func (h *handler) GetConfigPolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.ConfigPolicyGetReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get config policy, failed to decode request body")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// get config policy.
	configPolicy, err := h.backendHandler.GetConfigPolicy(rCtx, req.GetConfigpolicyId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get config policy")
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	// get template.
	template, err := h.storageConfigPolicyTemplate.GetConfigPolicyTemplate(rCtx, configPolicy.ID)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get config policy")
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	var blocks []types.ConfigPolicyTemplateBlock
	if err := json.Unmarshal([]byte(template.Template), &blocks); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get config policy template")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// refresh template with new items.
	templates, err := h.configPolicyOptionSet.GetOptionsByPolicyType(configPolicy.Type)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get config policy template")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	for _, block := range templates {
		blocks = insertTemplateBlock(blocks, block)
	}

	resp := new(protoApplication.ConfigPolicyGetResp)
	resp.ConvertConfigPolicyFromTypes(configPolicy, blocks)

	return resp.GetData(), nil
}

// ListConfigPolicyPlatform lists config policy platform.
func (h *handler) ListConfigPolicyPlatform(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.ConfigPolicyListPlatformReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list config policy platform, failed to decode request body")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// get release type.
	releaseType, err := types.ConvertConfigPolicyTypeToReleaseType(types.ConfigPolicyType(req.GetConfigpolicyType()))
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list config policy platform, failed to convert release type")
	}

	gen := types.Generation(req.GetGeneration())
	distinctField := types.ReleaseDistinctField{
		OSType:  true,
		CPUArch: true,
	}

	var result *types.ReleaseDistinctResult
	switch releaseType {
	case types.ReleaseTypeAgent:
		result, err = h.backendHandler.DistinctReleaseAgent(rCtx, gen, distinctField, req.ConvertConditionsToTypes())
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list config policy platform, failed to distinct release")
			return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
		}

	case types.ReleaseTypeProxy:
		result, err = h.backendHandler.DistinctReleaseProxy(rCtx, gen, distinctField, req.ConvertConditionsToTypes())
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list config policy platform, failed to distinct release")
			return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
		}

	default:
		logger.G.Biz(rCtx).WithErr(err).With("type", releaseType).Error("failed to list config policy platform, failed to convert release type")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	resp := new(protoApplication.ConfigPolicyListPlatformResp)
	resp.ConvertPlatformFromTypes(result)

	return resp.GetData(), nil
}

// GetConfigPolicyTemplate gets config policy template.
func (h *handler) GetConfigPolicyTemplate(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.ConfigPolicyGetTemplateReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get config policy template, failed to decode request body")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	configTemplate, err := h.configPolicyOptionSet.GetOptionsByPolicyType(types.ConfigPolicyType(req.GetConfigpolicyType()))
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get config policy template")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	resp := new(protoApplication.ConfigPolicyGetTemplateResp)
	resp.ConvertTemplateFromTypes(configTemplate)

	return resp.GetData(), nil
}

// CreateConfigPolicy creates config policy.
func (h *handler) CreateConfigPolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.ConfigPolicyCreateReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create config policy, failed to decode request body")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	configPolicy, cpTemplate := req.ConvertConfigPolicyToTypes()

	configTemplate, err := h.configPolicyOptionSet.GetOptionsByPolicyType(types.ConfigPolicyType(req.GetConfigpolicyType()))
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get config policy template")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// create config policy.
	addCustomConfig(configPolicy, cpTemplate, configTemplate)
	configPolicy.TenantID = rCtx.TenantID()
	configPolicyID, err := h.backendHandler.CreateConfigPolicy(rCtx, configPolicy)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create config policy")
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	// create template.
	template, err := json.Marshal(cpTemplate)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create config policy, failed to marshal template")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}
	if err := h.storageConfigPolicyTemplate.UpsertManyConfigPolicyTemplate(rCtx, &types.ConfigPolicyTemplate{
		TenantID: configPolicy.TenantID,
		ID:       configPolicyID,
		Template: string(template),
	}); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create config policy, failed to create template")
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoApplication.ConfigPolicyCreateResp)
	resp.ConvertConfigPolicyID(configPolicyID)

	return resp.GetData(), nil
}

// UpdateConfigPolicy updates config policy.
func (h *handler) UpdateConfigPolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.ConfigPolicyUpdateReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update config policy, failed to decode request body")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	configPolicy, cpTemplate := req.ConvertConfigPolicyToTypes()

	configTemplate, err := h.configPolicyOptionSet.GetOptionsByPolicyType(types.ConfigPolicyType(req.GetConfigpolicyType()))
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get config policy template")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// update config policy.
	addCustomConfig(configPolicy, cpTemplate, configTemplate)
	configPolicy.TenantID = rCtx.TenantID()
	if _, err := h.backendHandler.UpdateConfigPolicy(rCtx, configPolicy); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update config policy")
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	// upsert template.
	template, err := json.Marshal(cpTemplate)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update config policy, failed to marshal template")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}
	if err := h.storageConfigPolicyTemplate.UpsertManyConfigPolicyTemplate(rCtx, &types.ConfigPolicyTemplate{
		TenantID: configPolicy.TenantID,
		ID:       configPolicy.ID,
		Template: string(template),
	}); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update config policy, failed to update template")
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoApplication.ConfigPolicyUpdateResp)
	resp.ConvertConfigPolicyID(req.GetConfigpolicyId())

	return resp.GetData(), nil
}

// EnableConfigPolicy enables config policy.
func (h *handler) EnableConfigPolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.ConfigPolicyEnableReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to enable config policy, failed to decode request body")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := h.backendHandler.EnableConfigPolicy(rCtx, req.GetConfigpolicyId()...); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to enable config policy")
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.ConfigPolicyEnableResp)

	return resp.GetData(), nil
}

// DisableConfigPolicy disables config policy.
func (h *handler) DisableConfigPolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.ConfigPolicyDisableReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to disable config policy, failed to decode request body")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := h.backendHandler.DisableConfigPolicy(rCtx, req.GetConfigpolicyId()...); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to disable config policy")
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.ConfigPolicyDisableResp)

	return resp.GetData(), nil
}

// DeleteConfigPolicy deletes config policy.
func (h *handler) DeleteConfigPolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.ConfigPolicyDeleteReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete config policy, failed to decode request body")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := h.backendHandler.DeleteConfigPolicy(rCtx, req.GetConfigpolicyId()...); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete config policy")
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.ConfigPolicyDeleteResp)

	return resp.GetData(), nil
}

// ReorderPrioritiesConfigPolicy reorders config policy priorities within a (biz, type) scope.
func (h *handler) ReorderPrioritiesConfigPolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.ConfigPolicyPriorityReorderReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to reorder priorities for config policy, failed to decode request body")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	policyType := types.ConfigPolicyType(req.GetConfigpolicyType())
	if err := h.backendHandler.ReorderPrioritiesConfigPolicy(rCtx, req.GetBkBizId(), policyType, req.GetOrderedConfigpolicyId()); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to reorder priorities for config policy")
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.ConfigPolicyPriorityReorderResp)

	return resp.GetData(), nil
}

// ListConfigPolicyEvent lists events with page and conditions.
func (h *handler) ListConfigPolicyEvent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.ConfigPolicyEventListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list config policy event, failed to decode request body")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	conditions, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list config policy event, failed to convert conditions")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}
	// only count.
	if req.GetOnlyCount() {
		num, err := h.backendHandler.CountConfigPolicyEvent(
			rCtx,
			conditions)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list config policy event. failed to count event")
			return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
		}

		resp := new(protoApplication.ConfigPolicyEventListResp)
		resp.ConvertConfigPolicyEventsFromTypes(num, nil)

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list config policy event, invalid page info")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	events, num, err := h.backendHandler.ListConfigPolicyEvent(rCtx, page, conditions)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list config policy event")
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.ConfigPolicyEventListResp)
	resp.ConvertConfigPolicyEventsFromTypes(num, events)

	return resp.GetData(), nil
}

// DistinctEvent distincts events with conditions.
// nolint: dupl
func (h *handler) DistinctConfigPolicyEvent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.ConfigPolicyEventDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct config policy event, failed to decode request body")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	conditions, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct config policy event, failed to convert conditions")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}
	result, err := h.backendHandler.DistinctConfigPolicyEvent(
		rCtx,
		conditions)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct config policy event. failed to distinct config policy fields: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.ConfigPolicyEventDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}

// PreviewConfigPolicy previews the merged config for each host.
func (h *handler) PreviewConfigPolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.ConfigPolicyPreviewReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to preview config policy, failed to decode request body")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	policyType := types.ConfigPolicyType(req.GetPolicyType())
	previewHosts := req.ConvertPreviewHostsToTypes()

	result, err := h.backendHandler.PreviewConfigPolicy(rCtx, req.GetBkBizId(), policyType, req.GetPluginName(), previewHosts)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to preview config policy")

		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.ConfigPolicyPreviewResp)
	resp.ConvertMatchResultsFromTypes(result)

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

// nolint:funlen,gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func addCustomConfig(configPolicy *types.ConfigPolicy, blocks, preDefinedBlocks []types.ConfigPolicyTemplateBlock) {
	if configPolicy == nil {
		return
	}

	if configPolicy.Configs == nil {
		configPolicy.Configs = make(map[string]any)
	}

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

			templateItem, ok := findTemplateItem(preDefinedBlocks, block.ID, item.ID)
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
			if len(templateItem.ValueGroupAssigned) > 0 {
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
