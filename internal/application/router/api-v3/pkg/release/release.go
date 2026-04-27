/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package release provides the release router.
package release

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/options"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backend"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg             *gin.RouterGroup
	backendHandler backend.IHandler
	fileHandler    file.IHandler
}

type packageItem interface {
	GetGeneration() int64
	GetPlatform() *protoApplication.Platform
	GetVersion() string
}

func buildReleaseAgentKey(item packageItem) types.ReleaseAgentKey {
	return types.ReleaseAgentKey{
		Generation: types.Generation(item.GetGeneration()),
		Platform:   protoApplication.ConvertPlatformToTypes(item.GetPlatform()),
		Version:    item.GetVersion(),
	}
}

// countDeployedBatch represents a batch of items that share the same host query condition.
// Used exclusively by CountDeployed handlers to optimize backend queries.
type countDeployedBatch struct {
	sharedCondition *types.HostCondition
	itemIndices     []int
	versions        []string
}

// countDeployedQueryKey represents the dimensions used to query host distribution.
// Items with the same key can be queried together in a single backend call.
type countDeployedQueryKey struct {
	role       types.NodeRole
	generation int64
	osType     string
	cpuArch    string
}

// calCountDeployedBatch batches items by host query condition to reduce backend calls.
// This function is specifically designed for CountDeployed handlers.
// Do NOT reuse for other scenarios without understanding the assumptions.
func calCountDeployedBatch[T packageItem](items []T, role types.NodeRole) []countDeployedBatch {
	if len(items) == 0 {
		return nil
	}

	// Step 1: Extract query keys from all items.
	keys := make([]countDeployedQueryKey, len(items))
	for i, item := range items {
		keys[i] = extractCountDeployedQueryKey(item, role)
	}

	// Step 2: Batch items by query key.
	batchIndices := make(map[countDeployedQueryKey]int)
	batches := make([]countDeployedBatch, 0)

	for i, key := range keys {
		batchIndex, exists := batchIndices[key]
		if !exists {
			batchIndex = len(batches)
			batchIndices[key] = batchIndex
			batches = append(batches, countDeployedBatch{
				sharedCondition: buildHostConditionForCountDeployed(key),
				itemIndices:     []int{},
				versions:        []string{},
			})
		}

		batches[batchIndex].itemIndices = append(batches[batchIndex].itemIndices, i)
		batches[batchIndex].versions = append(batches[batchIndex].versions, items[i].GetVersion())
	}

	return batches
}

// extractCountDeployedQueryKey extracts the query key from a packageItem.
func extractCountDeployedQueryKey(item packageItem, role types.NodeRole) countDeployedQueryKey {
	platform := item.GetPlatform()
	osType, cpuArch := "", ""
	if platform != nil {
		osType = platform.GetOsType()
		cpuArch = platform.GetCpuArch()
	}

	return countDeployedQueryKey{
		role:       role,
		generation: item.GetGeneration(),
		osType:     osType,
		cpuArch:    cpuArch,
	}
}

// buildHostConditionForCountDeployed constructs HostCondition from query key.
func buildHostConditionForCountDeployed(key countDeployedQueryKey) *types.HostCondition {
	return &types.HostCondition{
		DynamicExactInclude: &types.HostDynamicExactFields{
			NodeRole:       []types.NodeRole{key.role},
			NodeGeneration: []int64{key.generation},
			OSType:         []string{key.osType},
			Arch:           []string{key.cpuArch},
		},
	}
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:             rg.Group("/release"),
		backendHandler: capability.BackendHandler,
		fileHandler:    capability.FileHandler,
	}
}

// Load loads node handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	// release agent.
	h.rg.POST("/agent/list", restserver.Handler(h.ListReleaseAgent))
	h.rg.POST("/agent/list/brief", restserver.Handler(h.ListReleaseAgentBrief))
	h.rg.POST("/agent/distinct", restserver.Handler(h.DistinctReleaseAgent))
	h.rg.POST("/agent/set_labels_many", restserver.Handler(h.SetReleaseAgentLabelsMany))
	h.rg.POST("/agent/enable", restserver.Handler(h.EnableReleaseAgent))
	h.rg.POST("/agent/disable", restserver.Handler(h.DisableReleaseAgent))
	h.rg.POST("/agent/set_as_default", restserver.Handler(h.SetAsDefaultReleaseAgent))
	h.rg.POST("/agent/cancel_as_default", restserver.Handler(h.CancelAsDefaultReleaseAgent))
	h.rg.POST("/agent/delete", restserver.Handler(h.DeleteReleaseAgent))
	h.rg.POST("/agent/count_deployed", restserver.Handler(h.CountDeployedReleaseAgent))
	h.rg.POST("/agent/download", restserver.StreamHandler(h.DownloadReleaseAgent))

	// release proxy.
	h.rg.POST("/proxy/list", restserver.Handler(h.ListReleaseProxy))
	h.rg.POST("/proxy/list/brief", restserver.Handler(h.ListReleaseProxyBrief))
	h.rg.POST("/proxy/distinct", restserver.Handler(h.DistinctReleaseProxy))
	h.rg.POST("/proxy/set_labels_many", restserver.Handler(h.SetReleaseProxyLabelsMany))
	h.rg.POST("/proxy/enable", restserver.Handler(h.EnableReleaseProxy))
	h.rg.POST("/proxy/disable", restserver.Handler(h.DisableReleaseProxy))
	h.rg.POST("/proxy/set_as_default", restserver.Handler(h.SetAsDefaultReleaseProxy))
	h.rg.POST("/proxy/cancel_as_default", restserver.Handler(h.CancelAsDefaultReleaseProxy))
	h.rg.POST("/proxy/delete", restserver.Handler(h.DeleteReleaseProxy))
	h.rg.POST("/proxy/count_deployed", restserver.Handler(h.CountDeployedReleaseProxy))
	h.rg.POST("/proxy/download", restserver.StreamHandler(h.DownloadReleaseProxy))

	// release plugin.
	h.rg.POST("/plugin/list", restserver.Handler(h.ListReleasePlugin))
	h.rg.POST("/plugin/list/brief", restserver.Handler(h.ListReleasePluginBrief))
	h.rg.POST("/plugin/enable", restserver.Handler(h.EnableReleasePlugin))
	h.rg.POST("/plugin/disable", restserver.Handler(h.DisableReleasePlugin))
	h.rg.POST("/plugin/set_as_default", restserver.Handler(h.SetAsDefaultReleasePlugin))
	h.rg.POST("/plugin/cancel_as_default", restserver.Handler(h.CancelAsDefaultReleasePlugin))
	h.rg.POST("/plugin/delete", restserver.Handler(h.DeleteReleasePlugin))
	h.rg.POST("/plugin/download", restserver.StreamHandler(h.DownloadReleasePlugin))
	h.rg.POST("/plugin/get_config_variables", restserver.Handler(h.GetConfigVariablesReleasePlugin))

	// release cert.
	h.rg.POST("/cert/list", restserver.Handler(h.ListReleaseCert))
	h.rg.POST("/cert/delete", restserver.Handler(h.DeleteReleaseCert))
	h.rg.POST("/cert/download", restserver.StreamHandler(h.DownloadReleaseCert))

	// release bintool.
	h.rg.POST("/bintool/list", restserver.Handler(h.ListReleaseBinTool))
	h.rg.POST("/bintool/delete", restserver.Handler(h.DeleteReleaseBinTool))
	h.rg.POST("/bintool/download", restserver.StreamHandler(h.DownloadReleaseBinTool))

	// release plugin bintool.
	h.rg.POST("/plugin_bintool/list", restserver.Handler(h.ListReleasePluginBinTool))
	h.rg.POST("/plugin_bintool/delete", restserver.Handler(h.DeleteReleasePluginBinTool))
	h.rg.POST("/plugin_bintool/download", restserver.StreamHandler(h.DownloadReleasePluginBinTool))
}

// ==================== Release Agent ====================

// ListReleaseAgent lists release agent with page and conditions.
func (h *handler) ListReleaseAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseAgentListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release agent, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())

	// only count.
	if req.GetOnlyCount() {
		num, err := h.backendHandler.CountReleaseAgent(rCtx, gen, req.ConvertConditionsToTypes())
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list release agent. failed to count release agent")
			return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
		}

		resp := new(protoApplication.PackageReleaseAgentListResp)
		resp.ConvertReleasesFromTypes(num, nil)

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release agent, invalid page info")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// special logic:
	// release agent request with limit 0 means unlimited
	if page.Limit == 0 {
		page = types.UnlimitedPage()
	}

	releases, num, err := h.backendHandler.ListReleaseAgent(rCtx, gen, page, req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release agent")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.PackageReleaseAgentListResp)
	resp.ConvertReleasesFromTypes(num, releases)

	return resp.GetData(), nil
}

// ListReleaseAgentBrief lists release agent brief with page and conditions.
func (h *handler) ListReleaseAgentBrief(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseAgentListBriefReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release agent brief, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())

	if req.GetOnlyCount() {
		num, err := h.backendHandler.CountReleaseAgent(rCtx, gen, req.ConvertConditionsToTypes())
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list release agent brief. failed to count release agent")
			return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
		}

		resp := new(protoApplication.PackageReleaseAgentListBriefResp)
		resp.ConvertReleasesFromTypes(num, nil)

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release agent brief, invalid page info")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if page.Limit == 0 {
		page = types.UnlimitedPage()
	}

	releases, num, err := h.backendHandler.ListReleaseAgentBrief(rCtx, gen, page, req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release agent brief")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.PackageReleaseAgentListBriefResp)
	resp.ConvertReleasesFromTypes(num, releases)

	return resp.GetData(), nil
}

// DistinctReleaseAgent distincts release agent by conditions.
func (h *handler) DistinctReleaseAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseAgentDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct release agent, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())
	distinctField := req.ConvertDistinctFieldToTypes()
	condition := req.ConvertConditionsToTypes()

	result, err := h.backendHandler.DistinctReleaseAgent(rCtx, gen, distinctField, condition)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct release agent")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.PackageReleaseAgentDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}

// SetReleaseAgentLabelsMany sets many agent release labels.
func (h *handler) SetReleaseAgentLabelsMany(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseAgentSetLabelsManyReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set many agent release labels, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())
	labels := req.GetLabels()
	condition := req.ConvertConditionsToTypes()

	if err := h.backendHandler.SetReleaseAgentLabelsMany(rCtx, gen, labels, condition); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("gen", gen, "labels", labels).Error("failed to set many agent release labels")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("gen", gen, "labels", labels).Info("set many agent release labels")

	resp := new(protoApplication.PackageReleaseAgentSetLabelsManyResp)

	return resp.GetData(), nil
}

// EnableReleaseAgent enables agent release.
func (h *handler) EnableReleaseAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseAgentEnableReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to enable release agent, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	key := buildReleaseAgentKey(req)

	if err := h.backendHandler.EnableReleaseAgent(rCtx, key); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("gen", key.Generation, "plat", key.Platform, "version", key.Version).Error("failed to enable release agent")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("gen", key.Generation, "plat", key.Platform, "version", key.Version).Info("enabled release agent")

	resp := new(protoApplication.PackageReleaseAgentEnableResp)

	return resp.GetData(), nil
}

// DisableReleaseAgent disables agent release.
func (h *handler) DisableReleaseAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseAgentDisableReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to disable release agent, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	key := buildReleaseAgentKey(req)

	if err := h.backendHandler.DisableReleaseAgent(rCtx, key); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("gen", key.Generation, "plat", key.Platform, "version", key.Version).Error("failed to disable release agent")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("gen", key.Generation, "plat", key.Platform, "version", key.Version).Info("disabled release agent")

	resp := new(protoApplication.PackageReleaseAgentDisableResp)

	return resp.GetData(), nil
}

// SetAsDefaultReleaseAgent sets agent release as default.
func (h *handler) SetAsDefaultReleaseAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseAgentSetAsDefaultReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set default release agent, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	key := buildReleaseAgentKey(req)

	if err := h.backendHandler.SetAsDefaultReleaseAgent(rCtx, key); err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			With("gen", key.Generation, "plat", key.Platform, "version", key.Version).Error("failed to set default release agent")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("gen", key.Generation, "plat", key.Platform, "version", key.Version).Info("set default release agent")

	resp := new(protoApplication.PackageReleaseAgentSetAsDefaultResp)

	return resp.GetData(), nil
}

// CancelAsDefaultReleaseAgent cancels agent release as default.
func (h *handler) CancelAsDefaultReleaseAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseAgentCancelAsDefaultReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to cancel default release agent, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	key := buildReleaseAgentKey(req)

	if err := h.backendHandler.CancelAsDefaultReleaseAgent(rCtx, key); err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			With("gen", key.Generation, "plat", key.Platform, "version", key.Version).Error("failed to cancel default release agent")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("gen", key.Generation, "plat", key.Platform, "version", key.Version).Info("canceled default release agent")

	resp := new(protoApplication.PackageReleaseAgentCancelAsDefaultResp)

	return resp.GetData(), nil
}

// DeleteReleaseAgent deletes agent release.
func (h *handler) DeleteReleaseAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseAgentDeleteReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete release agent, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	key := buildReleaseAgentKey(req)

	if err := h.backendHandler.DeleteReleaseAgent(rCtx, key); err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			With("gen", key.Generation, "plat", key.Platform, "version", key.Version).Error("failed to delete release agent")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("gen", key.Generation, "plat", key.Platform, "version", key.Version).Info("deleted release agent")

	resp := new(protoApplication.PackageReleaseAgentDeleteResp)

	return resp.GetData(), nil
}

// CountDeployedReleaseAgent count deployed release agent.
func (h *handler) CountDeployedReleaseAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseAgentCountDeployedReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to count deployed agent release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if _, err := req.ConvertConditionsToTypes(); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to count deployed agent release, failed to convert conditions")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	groups := calCountDeployedBatch(req.GetItems(), types.NodeRoleAgent)
	results := make([]int64, len(req.GetItems()))

	gp := gopool.NewPool()
	for _, group := range groups {
		currentGroup := group

		gp.Go(func() error {
			hostDistributionByNodeVersion, err := h.backendHandler.GetHostDistributionByNodeVersion(
				rCtx, currentGroup.sharedCondition,
			)
			if err != nil {
				logger.G.Biz(rCtx).WithErr(err).Error("failed to get host distribution by node version")

				return err
			}

			for i, version := range currentGroup.versions {
				results[currentGroup.itemIndices[i]] = hostDistributionByNodeVersion[version]
			}

			return nil
		})
	}
	if err := gp.Wait(); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to count deployed agent release")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).Info("count deployed agent release")

	resp := new(protoApplication.PackageReleaseAgentCountDeployedResp)
	resp.ConvertResultFromTypes(results)

	return resp.GetData(), nil
}

// DownloadReleaseAgent download release agent.
func (h *handler) DownloadReleaseAgent(rCtx restserver.IContext) (*restserver.StreamResponse, error) {
	req := new(protoApplication.PackageReleaseAgentDownloadReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download release agent, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, plat, version := req.GetIdentifier()
	resp, err := h.fileHandler.DownloadReleaseAgent(rCtx, gen, plat, version)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download release agent")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("gen", gen, "plat", plat, "version", version).Info("downloaded release agent")

	return resp, nil
}

// ==================== Release Proxy ====================

// ListReleaseProxy lists release proxy with page and conditions.
func (h *handler) ListReleaseProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseProxyListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release proxy, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())

	// only count.
	if req.GetOnlyCount() {
		num, err := h.backendHandler.CountReleaseProxy(rCtx, gen, req.ConvertConditionsToTypes())
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list release proxy. failed to count release proxy")
			return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
		}

		resp := new(protoApplication.PackageReleaseProxyListResp)
		resp.ConvertReleasesFromTypes(num, nil)

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release proxy, invalid page info")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// special logic:
	// release proxy request with limit 0 means unlimited
	if page.Limit == 0 {
		page = types.UnlimitedPage()
	}

	releases, num, err := h.backendHandler.ListReleaseProxy(rCtx, gen, page, req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release proxy")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.PackageReleaseProxyListResp)
	resp.ConvertReleasesFromTypes(num, releases)

	return resp.GetData(), nil
}

// ListReleaseProxyBrief lists release proxy brief with page and conditions.
func (h *handler) ListReleaseProxyBrief(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseProxyListBriefReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release proxy brief, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())

	if req.GetOnlyCount() {
		num, err := h.backendHandler.CountReleaseProxy(rCtx, gen, req.ConvertConditionsToTypes())
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list release proxy brief. failed to count release proxy")
			return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
		}

		resp := new(protoApplication.PackageReleaseProxyListBriefResp)
		resp.ConvertReleasesFromTypes(num, nil)

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release proxy brief, invalid page info")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if page.Limit == 0 {
		page = types.UnlimitedPage()
	}

	releases, num, err := h.backendHandler.ListReleaseProxyBrief(rCtx, gen, page, req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release proxy brief")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.PackageReleaseProxyListBriefResp)
	resp.ConvertReleasesFromTypes(num, releases)

	return resp.GetData(), nil
}

// DistinctReleaseProxy distincts release proxy by conditions.
func (h *handler) DistinctReleaseProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseProxyDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct release proxy, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())
	distinctField := req.ConvertDistinctFieldToTypes()
	condition := req.ConvertConditionsToTypes()

	result, err := h.backendHandler.DistinctReleaseProxy(rCtx, gen, distinctField, condition)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct release proxy")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.PackageReleaseProxyDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}

// SetReleaseProxyLabelsMany sets many proxy release labels.
func (h *handler) SetReleaseProxyLabelsMany(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseProxySetLabelsManyReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set many proxy release labels, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())
	labels := req.GetLabels()
	condition := req.ConvertConditionsToTypes()

	if err := h.backendHandler.SetReleaseProxyLabelsMany(rCtx, gen, labels, condition); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("gen", gen, "labels", labels).Error("failed to set many proxy release labels")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("gen", gen, "labels", labels).Info("set many proxy release labels")

	resp := new(protoApplication.PackageReleaseProxySetLabelsManyResp)

	return resp.GetData(), nil
}

// EnableReleaseProxy enables proxy release.
func (h *handler) EnableReleaseProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseProxyEnableReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to enable release proxy, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	key := buildReleaseProxyKey(req)

	if err := h.backendHandler.EnableReleaseProxy(rCtx, key); err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			With("gen", key.Generation, "plat", key.Platform, "version", key.Version).Error("failed to enable release proxy")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("gen", key.Generation, "plat", key.Platform, "version", key.Version).Info("enabled release proxy")

	resp := new(protoApplication.PackageReleaseProxyEnableResp)

	return resp.GetData(), nil
}

func buildReleaseProxyKey(req packageItem) types.ReleaseProxyKey {
	return types.ReleaseProxyKey{
		Generation: types.Generation(req.GetGeneration()),
		Platform:   protoApplication.ConvertPlatformToTypes(req.GetPlatform()),
		Version:    req.GetVersion(),
	}
}

// DisableReleaseProxy disables proxy release.
func (h *handler) DisableReleaseProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseProxyDisableReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to disable release proxy, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	key := buildReleaseProxyKey(req)

	if err := h.backendHandler.DisableReleaseProxy(rCtx, key); err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			With("gen", key.Generation, "plat", key.Platform, "version", key.Version).Error("failed to disable release proxy")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("gen", key.Generation, "plat", key.Platform, "version", key.Version).Info("disabled release proxy")

	resp := new(protoApplication.PackageReleaseProxyDisableResp)

	return resp.GetData(), nil
}

// SetAsDefaultReleaseProxy sets proxy release as default.
func (h *handler) SetAsDefaultReleaseProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseProxySetAsDefaultReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set default release proxy, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	key := buildReleaseProxyKey(req)

	if err := h.backendHandler.SetAsDefaultReleaseProxy(rCtx, key); err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			With("gen", key.Generation, "plat", key.Platform, "version", key.Version).Error("failed to set default release proxy")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("gen", key.Generation, "plat", key.Platform, "version", key.Version).Info("set default release proxy")

	resp := new(protoApplication.PackageReleaseProxySetAsDefaultResp)

	return resp.GetData(), nil
}

// CancelAsDefaultReleaseProxy cancels proxy release as default.
func (h *handler) CancelAsDefaultReleaseProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseProxyCancelAsDefaultReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to cancel default release proxy, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	key := buildReleaseProxyKey(req)

	if err := h.backendHandler.CancelAsDefaultReleaseProxy(rCtx, key); err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			With("gen", key.Generation, "plat", key.Platform, "version", key.Version).Error("failed to cancel default release proxy")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("gen", key.Generation, "plat", key.Platform, "version", key.Version).Info("canceled default release proxy")

	resp := new(protoApplication.PackageReleaseProxyCancelAsDefaultResp)

	return resp.GetData(), nil
}

// DeleteReleaseProxy deletes proxy release.
func (h *handler) DeleteReleaseProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseProxyDeleteReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete release proxy, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	key := buildReleaseProxyKey(req)

	if err := h.backendHandler.DeleteReleaseProxy(rCtx, key); err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			With("gen", key.Generation, "plat", key.Platform, "version", key.Version).Error("failed to delete release proxy")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("gen", key.Generation, "plat", key.Platform, "version", key.Version).Info("deleted release proxy")

	resp := new(protoApplication.PackageReleaseProxyDeleteResp)

	return resp.GetData(), nil
}

// CountDeployedReleaseProxy count deployed release proxy.
func (h *handler) CountDeployedReleaseProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseProxyCountDeployedReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to count deployed proxy release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if _, err := req.ConvertConditionsToTypes(); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to count deployed proxy release, failed to convert conditions")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	groups := calCountDeployedBatch(req.GetItems(), types.NodeRoleProxy)
	results := make([]int64, len(req.GetItems()))

	gp := gopool.NewPool()
	for _, group := range groups {
		currentGroup := group

		gp.Go(func() error {
			hostDistributionByNodeVersion, err := h.backendHandler.GetHostDistributionByNodeVersion(
				rCtx, currentGroup.sharedCondition,
			)
			if err != nil {
				logger.G.Biz(rCtx).WithErr(err).Error("failed to get host distribution by node version")

				return err
			}

			for i, version := range currentGroup.versions {
				results[currentGroup.itemIndices[i]] = hostDistributionByNodeVersion[version]
			}

			return nil
		})
	}
	if err := gp.Wait(); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to count deployed proxy release")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).Info("count deployed proxy release")

	resp := new(protoApplication.PackageReleaseProxyCountDeployedResp)
	resp.ConvertResultFromTypes(results)

	return resp.GetData(), nil
}

// DownloadReleaseProxy download release proxy.
func (h *handler) DownloadReleaseProxy(rCtx restserver.IContext) (*restserver.StreamResponse, error) {
	req := new(protoApplication.PackageReleaseProxyDownloadReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download release proxy, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, plat, version := req.GetIdentifier()
	resp, err := h.fileHandler.DownloadReleaseProxy(rCtx, gen, plat, version)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download release proxy")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("gen", gen, "plat", plat, "version", version).Info("downloaded release proxy")

	return resp, nil
}

// ==================== Release Plugin ====================

// ListReleasePlugin lists release plugin with page and conditions.
func (h *handler) ListReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleasePluginListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release plugin, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())

	// only count.
	if req.GetOnlyCount() {
		num, err := h.backendHandler.CountReleasePlugin(rCtx, gen, req.ConvertConditionsToTypes())
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list release plugin. failed to count release plugin")
			return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
		}

		resp := new(protoApplication.PackageReleasePluginListResp)
		resp.ConvertReleasePluginsFromTypes(num, nil)

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release plugin, invalid page info")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// special logic:
	// release plugin request with limit 0 means unlimited
	if page.Limit == 0 {
		page = types.UnlimitedPage()
	}

	releases, num, err := h.backendHandler.ListReleasePlugin(rCtx, gen, page, req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release plugin")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.PackageReleasePluginListResp)
	resp.ConvertReleasePluginsFromTypes(num, releases)

	return resp.GetData(), nil
}

// ListReleasePluginBrief lists release plugin brief with page and conditions.
func (h *handler) ListReleasePluginBrief(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleasePluginListBriefReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release plugin brief, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())

	if req.GetOnlyCount() {
		num, err := h.backendHandler.CountReleasePlugin(rCtx, gen, req.ConvertConditionsToTypes())
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list release plugin brief. failed to count release plugin")
			return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
		}

		resp := new(protoApplication.PackageReleasePluginListBriefResp)
		resp.ConvertReleasePluginsFromTypes(num, nil)

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release plugin brief, invalid page info")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if page.Limit == 0 {
		page = types.UnlimitedPage()
	}

	releases, num, err := h.backendHandler.ListReleasePluginBrief(rCtx, gen, page, req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release plugin brief")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.PackageReleasePluginListBriefResp)
	resp.ConvertReleasePluginsFromTypes(num, releases)

	return resp.GetData(), nil
}

// EnableReleasePlugin enable release plugin.
func (h *handler) EnableReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleasePluginEnableReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to enable release plugin, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	name, gen, plat, version := req.GetIdentifier()
	if err := h.backendHandler.EnableReleasePlugin(rCtx, types.ReleasePluginKey{
		Name:       name,
		Generation: gen,
		Platform:   plat,
		Version:    version,
	}); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("name", name, "gen", gen, "plat", plat, "version", version).Error("failed to enable release plugin")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("name", name, "gen", gen, "plat", plat, "version", version).Info("enabled release plugin")

	resp := new(protoApplication.PackageReleasePluginEnableResp)

	return resp.GetData(), nil
}

// DisableReleasePlugin disable release plugin.
func (h *handler) DisableReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleasePluginDisableReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to disable release plugin, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	name, gen, plat, version := req.GetIdentifier()
	if err := h.backendHandler.DisableReleasePlugin(rCtx, types.ReleasePluginKey{
		Name:       name,
		Generation: gen,
		Platform:   plat,
		Version:    version,
	}); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("name", name, "gen", gen, "plat", plat, "version", version).Error("failed to disable release plugin.")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("name", name, "gen", gen, "plat", plat, "version", version).Info("disabled release plugin")

	resp := new(protoApplication.PackageReleasePluginDisableResp)

	return resp.GetData(), nil
}

// SetAsDefaultReleasePlugin set default release plugin.
func (h *handler) SetAsDefaultReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleasePluginSetAsDefaultReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set default release plugin, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	name, gen, plat, version := req.GetIdentifier()
	if err := h.backendHandler.SetAsDefaultReleasePlugin(rCtx, types.ReleasePluginKey{
		Name:       name,
		Generation: gen,
		Platform:   plat,
		Version:    version,
	}); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("name", name, "gen", gen, "plat", plat, "version", version).Error("failed to set default release plugin")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("name", name, "gen", gen, "plat", plat, "version", version).Info("set default release plugin")

	resp := new(protoApplication.PackageReleasePluginSetAsDefaultResp)

	return resp.GetData(), nil
}

// CancelAsDefaultReleasePlugin cancel default release plugin.
func (h *handler) CancelAsDefaultReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleasePluginCancelAsDefaultReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to cancel default release plugin, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	name, gen, plat, version := req.GetIdentifier()
	if err := h.backendHandler.CancelAsDefaultReleasePlugin(rCtx, types.ReleasePluginKey{
		Name:       name,
		Generation: gen,
		Platform:   plat,
		Version:    version,
	}); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("name", name, "gen", gen, "plat", plat, "version", version).
			Error("failed to cancel default release plugin")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("name", name, "gen", gen, "plat", plat, "version", version).Info("canceled default release plugin")

	resp := new(protoApplication.PackageReleasePluginCancelAsDefaultResp)

	return resp.GetData(), nil
}

// DeleteReleasePlugin delete release plugin.
func (h *handler) DeleteReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleasePluginDeleteReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete release plugin, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	name, gen, plat, version := req.GetIdentifier()
	if err := h.backendHandler.DeleteReleasePlugin(rCtx, types.ReleasePluginKey{
		Name:       name,
		Generation: gen,
		Platform:   plat,
		Version:    version,
	}); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("name", name, "gen", gen, "plat", plat, "version", version).
			Error("failed to delete default release plugin")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("name", name, "gen", gen, "plat", plat, "version", version).Info("deleted default release plugin")

	resp := new(protoApplication.PackageReleasePluginDeleteResp)

	return resp.GetData(), nil
}

// DownloadReleasePlugin download release plugin.
func (h *handler) DownloadReleasePlugin(rCtx restserver.IContext) (*restserver.StreamResponse, error) {
	req := new(protoApplication.PackageReleasePluginDownloadReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download release plugin, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	name, plat, version := req.GetIdentifier()
	resp, err := h.fileHandler.DownloadReleasePlugin(rCtx, name, plat, version)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download release plugin")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("plugin_name", name, "plat", plat, "version", version).Info("downloaded release plugin")

	return resp, nil
}

// GetConfigVariablesReleasePlugin gets release plugin config variables.
func (h *handler) GetConfigVariablesReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleasePluginGetConfigVariablesReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			Error("failed to get release plugin config variables, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	name, version, gen, plat := req.GetIdentifier()
	configVariables, err := h.backendHandler.GetConfigVariablesReleasePlugin(rCtx, name, version, gen, plat)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			With("name", name, "gen", gen, "plat", plat, "version", version).
			Error("failed to get release plugin config variables")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.PackageReleasePluginGetConfigVariablesResp)
	resp.ConvertConfigVariablesFromTypes(configVariables)

	return resp.GetData(), nil
}

// ==================== Release Cert ====================

// ListReleaseCert lists cert releases.
func (h *handler) ListReleaseCert(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseCertListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release cert, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())

	releases, num, err := h.backendHandler.ListReleaseCert(rCtx, gen)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release cert")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.PackageReleaseCertListResp)
	resp.ConvertReleasesFromTypes(num, releases)

	return resp.GetData(), nil
}

// DeleteReleaseCert deletes cert release.
func (h *handler) DeleteReleaseCert(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseCertDeleteReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete release cert, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	key := types.ReleaseCertKey{
		Generation: types.Generation(req.GetGeneration()),
	}

	if err := h.backendHandler.DeleteReleaseCert(rCtx, key); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("gen", key.Generation).Error("failed to delete release cert")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("gen", key.Generation).Info("deleted release cert")

	resp := new(protoApplication.PackageReleaseCertDeleteResp)

	return resp.GetData(), nil
}

// DownloadReleaseCert download release cert.
func (h *handler) DownloadReleaseCert(rCtx restserver.IContext) (*restserver.StreamResponse, error) {
	req := new(protoApplication.PackageReleaseCertDownloadReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download release cert, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())
	resp, err := h.fileHandler.DownloadReleaseCert(rCtx, gen)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download release cert")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("gen", gen).Info("downloaded release cert")

	return resp, nil
}

// ==================== Release BinTool ====================

// ListReleaseBinTool lists bintool releases.
func (h *handler) ListReleaseBinTool(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseBinToolListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release bintool, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())

	releases, num, err := h.backendHandler.ListReleaseBinTool(rCtx, gen)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release bintool")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.PackageReleaseBinToolListResp)
	resp.ConvertReleasesFromTypes(num, releases)

	return resp.GetData(), nil
}

// DeleteReleaseBinTool deletes bintool release.
func (h *handler) DeleteReleaseBinTool(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseBinToolDeleteReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete release bintool, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	key := types.ReleaseBinToolKey{
		Generation: types.Generation(req.GetGeneration()),
	}

	if err := h.backendHandler.DeleteReleaseBinTool(rCtx, key); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("gen", key.Generation).Error("failed to delete release bintool")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("gen", key.Generation).Info("deleted release bintool")

	resp := new(protoApplication.PackageReleaseBinToolDeleteResp)

	return resp.GetData(), nil
}

// DownloadReleaseBinTool download release bintool.
func (h *handler) DownloadReleaseBinTool(rCtx restserver.IContext) (*restserver.StreamResponse, error) {
	req := new(protoApplication.PackageReleaseBinToolDownloadReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download release bintool, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())
	resp, err := h.fileHandler.DownloadReleaseBinTool(rCtx, gen)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download release bintool")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("gen", gen).Info("downloaded release bintool")

	return resp, nil
}

// ==================== Release Plugin BinTool ====================

// ListReleasePluginBinTool lists plugin bintool releases.
func (h *handler) ListReleasePluginBinTool(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleasePluginBinToolListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release plugin bintool, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())

	releases, num, err := h.backendHandler.ListReleasePluginBinTool(rCtx, gen)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release plugin bintool")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.PackageReleasePluginBinToolListResp)
	resp.ConvertReleasesFromTypes(num, releases)

	return resp.GetData(), nil
}

// DeleteReleasePluginBinTool deletes plugin bintool release.
func (h *handler) DeleteReleasePluginBinTool(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleasePluginBinToolDeleteReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete release plugin bintool, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	key := types.ReleasePluginBinToolKey{
		Generation: types.Generation(req.GetGeneration()),
		Name:       req.GetName(),
	}

	if err := h.backendHandler.DeleteReleasePluginBinTool(rCtx, key); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("gen", key.Generation, "name", key.Name).Error("failed to delete release plugin bintool")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("gen", key.Generation, "name", key.Name).Info("deleted release plugin bintool")

	resp := new(protoApplication.PackageReleasePluginBinToolDeleteResp)

	return resp.GetData(), nil
}

// DownloadReleasePluginBinTool download release plugin bintool.
func (h *handler) DownloadReleasePluginBinTool(rCtx restserver.IContext) (*restserver.StreamResponse, error) {
	req := new(protoApplication.PackageReleasePluginBinToolDownloadReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download release plugin bintool, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())
	name := req.GetName()
	resp, err := h.fileHandler.DownloadReleasePluginBinTool(rCtx, gen, name)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download release plugin bintool")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("gen", gen, "name", name).Info("downloaded release plugin bintool")

	return resp, nil
}
