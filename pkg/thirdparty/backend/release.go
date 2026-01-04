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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/pageexecutor"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandlerRelease defines the backend Handler for release.
// nolint: interfacebloat
type IHandlerRelease interface {
	IHandlerReleaseAgent
	IHandlerReleaseProxy
	IHandlerReleasePlugin
	IHandlerReleaseCert
	IHandlerReleaseBinTool
	IHandlerReleasePluginBinTool
}

// IHandlerReleaseAgent defines the backend Handler for release agent.
type IHandlerReleaseAgent interface {
	// ListReleaseAgent lists agent releases by page and conditions.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param gen the generation info.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return agent release list with page and the total count with filter and error.
	ListReleaseAgent(nCtx contextx.IContext, gen types.Generation, page types.Page, condition *types.ReleaseCondition) (
		[]*types.ReleaseAgent, int64, error)

	// CountReleaseAgent counts agent releases by conditions.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param gen the generation info.
	// @param condition the filter conditions.
	// @return the agent release count with filter and error.
	CountReleaseAgent(nCtx contextx.IContext, gen types.Generation, condition *types.ReleaseCondition) (int64, error)

	// DistinctReleaseAgent distincts agent releases by conditions.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param gen the generation info.
	// @param distinctField the distinct field.
	// @param condition the filter conditions.
	// @return the agent release distinct result and error.
	DistinctReleaseAgent(nCtx contextx.IContext, gen types.Generation, distinctField types.ReleaseDistinctField, condition *types.ReleaseCondition) (
		*types.ReleaseDistinctResult, error)

	// SetReleaseAgentLabelsMany sets many agent release labels.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param gen the generation info.
	// @param labels the labels to set.
	// @param condition the filter conditions.
	// @return the error.
	SetReleaseAgentLabelsMany(nCtx contextx.IContext, gen types.Generation, labels []string, condition *types.ReleaseCondition) error

	// EnableReleaseAgent enables agent release.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param key the release agent key.
	// @return the error.
	EnableReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error

	// DisableReleaseAgent disables agent release.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param key the release agent key.
	// @return the error.
	DisableReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error

	// SetAsDefaultReleaseAgent sets agent release as default.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param key the release agent key.
	// @return the error.
	SetAsDefaultReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error

	// CancelAsDefaultReleaseAgent cancels agent release as default.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param key the release agent key.
	// @return the error.
	CancelAsDefaultReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error

	// DeleteReleaseAgent deletes agent release.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param key the release agent key.
	// @return the error.
	DeleteReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error
}

// IHandlerReleaseProxy defines the backend Handler for release proxy.
type IHandlerReleaseProxy interface {
	// ListReleaseProxy lists proxy releases by page and conditions.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param gen the generation info.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return proxy release list with page and the total count with filter and error.
	ListReleaseProxy(nCtx contextx.IContext, gen types.Generation, page types.Page, condition *types.ReleaseCondition) (
		[]*types.ReleaseProxy, int64, error)

	// CountReleaseProxy counts proxy releases by conditions.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param gen the generation info.
	// @param condition the filter conditions.
	// @return the proxy release count with filter and error.
	CountReleaseProxy(nCtx contextx.IContext, gen types.Generation, condition *types.ReleaseCondition) (int64, error)

	// DistinctReleaseProxy distincts proxy releases by conditions.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param gen the generation info.
	// @param distinctField the distinct field.
	// @param condition the filter conditions.
	// @return the proxy release distinct result and error.
	DistinctReleaseProxy(nCtx contextx.IContext, gen types.Generation, distinctField types.ReleaseDistinctField, condition *types.ReleaseCondition) (
		*types.ReleaseDistinctResult, error)

	// SetReleaseProxyLabelsMany sets many proxy release labels.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param gen the generation info.
	// @param labels the labels to set.
	// @param condition the filter conditions.
	// @return the error.
	SetReleaseProxyLabelsMany(nCtx contextx.IContext, gen types.Generation, labels []string, condition *types.ReleaseCondition) error

	// EnableReleaseProxy enables proxy release.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param key the release proxy key.
	// @return the error.
	EnableReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error

	// DisableReleaseProxy disables proxy release.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param key the release proxy key.
	// @return the error.
	DisableReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error

	// SetAsDefaultReleaseProxy sets proxy release as default.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param key the release proxy key.
	// @return the error.
	SetAsDefaultReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error

	// CancelAsDefaultReleaseProxy cancels proxy release as default.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param key the release proxy key.
	// @return the error.
	CancelAsDefaultReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error

	// DeleteReleaseProxy deletes proxy release.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param key the release proxy key.
	// @return the error.
	DeleteReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error
}

// IHandlerReleasePlugin defines the backend Handler for release plugin.
type IHandlerReleasePlugin interface {
	// ListReleasePlugin lists plugin releases by page and conditions.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param gen the generation info.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return plugin release list with page and the total count with filter and error.
	ListReleasePlugin(nCtx contextx.IContext, gen types.Generation, page types.Page, condition *types.ReleaseCondition) (
		[]*types.ReleasePlugin, int64, error)

	// CountReleasePlugin counts plugin releases by conditions.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param gen the generation info.
	// @param condition the filter conditions.
	// @return the plugin release count with filter and error.
	CountReleasePlugin(nCtx contextx.IContext, gen types.Generation, condition *types.ReleaseCondition) (int64, error)

	// EnableReleasePlugin enables plugin release.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param key the release plugin key.
	// @return the error.
	EnableReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error

	// DisableReleasePlugin disables plugin release.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param key the release plugin key.
	// @return the error.
	DisableReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error

	// SetAsDefaultReleasePlugin sets plugin release as default.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param key the release plugin key.
	// @return the error.
	SetAsDefaultReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error

	// CancelAsDefaultReleasePlugin cancels plugin release as default.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param key the release plugin key.
	// @return the error.
	CancelAsDefaultReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error

	// DeleteReleasePlugin deletes plugin release.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param key the release plugin key.
	// @return the error.
	DeleteReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error
}

// IHandlerReleaseCert defines the backend Handler for release cert.
type IHandlerReleaseCert interface {
	// ListReleaseCert lists cert releases by page and conditions.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param gen the generation info.
	// @return cert release list with page and the total count with filter and error.
	ListReleaseCert(nCtx contextx.IContext, gen types.Generation) ([]*types.ReleaseCert, int64, error)

	// DeleteReleaseCert deletes cert release.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param key the release cert key.
	// @return the error.
	DeleteReleaseCert(nCtx contextx.IContext, key types.ReleaseCertKey) error
}

// IHandlerReleaseBinTool defines the backend Handler for release bintool.
type IHandlerReleaseBinTool interface {
	// ListReleaseBinTool lists bintool releases by page and conditions.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param gen the generation info.
	// @return bintool release list with page and the total count with filter and error.
	ListReleaseBinTool(nCtx contextx.IContext, gen types.Generation) ([]*types.ReleaseBinTool, int64, error)

	// DeleteReleaseBinTool deletes bintool release.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param key the release bintool key.
	// @return the error.
	DeleteReleaseBinTool(nCtx contextx.IContext, key types.ReleaseBinToolKey) error
}

// IHandlerReleasePluginBinTool defines the backend Handler for release plugin bintool.
type IHandlerReleasePluginBinTool interface {
	// ListReleasePluginBinTool lists plugin-bintool releases by page and conditions.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param gen the generation info.
	// @return plugin-bintool release list with page and the total count with filter and error.
	ListReleasePluginBinTool(nCtx contextx.IContext, gen types.Generation) ([]*types.ReleasePluginBinTool, int64, error)

	// DeleteReleasePluginBinTool deletes plugin-bintool release.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param key the release plugin-bintool key.
	// @return the error.
	DeleteReleasePluginBinTool(nCtx contextx.IContext, key types.ReleasePluginBinToolKey) error
}

// ===============================================================================
// Release Agent Related Interfaces
// ===============================================================================

// ListReleaseAgent lists agent releases by page and conditions.
func (h *Handler) ListReleaseAgent(nCtx contextx.IContext,gen types.Generation,page types.Page,condition *types.ReleaseCondition) (
	[]*types.ReleaseAgent, int64, error) {

	req := &protoBackend.PackageReleaseAgentListReq{
		Generation: int64(gen),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	var (
		total    int64
		releases []*types.ReleaseAgent
	)
	executor := pageexecutor.NewPageExecutor[*types.ReleaseAgent](req.PageLimit(), req.PageTimeout())
	fn := func(nCtx contextx.IContext, page types.Page) ([]*types.ReleaseAgent, error) {
		req.Page = convertPage(page)
		resp, err := h.cli.listReleaseAgent(nCtx, req)
		if err != nil {
			return nil, err
		}

		total, releases = resp.ConvertReleasesToTypes()

		return releases, nil
	}

	result, err := executor.Execute(nCtx, page, fn)
	if err != nil {
		return nil, 0, err
	}

	return result.Items, total, nil
}

// CountReleaseAgent counts agent releases by conditions.
func (h *Handler) CountReleaseAgent(nCtx contextx.IContext,gen types.Generation,condition *types.ReleaseCondition) (int64, error) {
	req := &protoBackend.PackageReleaseAgentListReq{
		Generation: int64(gen),
		OnlyCount:  true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}

	resp, err := h.cli.listReleaseAgent(nCtx, req)
	if err != nil {
		return 0, err
	}

	total, _ := resp.ConvertReleasesToTypes()

	return total, nil
}

// DistinctReleaseAgent distincts agent releases by conditions.
func (h *Handler) DistinctReleaseAgent(
	nCtx contextx.IContext,gen types.Generation,distinctField types.ReleaseDistinctField,condition *types.ReleaseCondition) (
	*types.ReleaseDistinctResult, error) {

	req := &protoBackend.PackageReleaseAgentDistinctReq{
		Generation: int64(gen),
		DistinctField: &protoBackend.PackageReleaseDistinctField{
			OsType:  distinctField.OSType,
			CpuArch: distinctField.CPUArch,
		},
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, err
	}

	resp, err := h.cli.distinctReleaseAgent(nCtx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertResultToTypes(), nil
}

// SetReleaseAgentLabelsMany sets many agent release labels.
func (h *Handler) SetReleaseAgentLabelsMany(nCtx contextx.IContext,gen types.Generation,labels []string,condition *types.ReleaseCondition) error {
	req := &protoBackend.PackageReleaseAgentSetLabelsManyReq{
		Generation: int64(gen),
		Labels:     labels,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return err
	}

	return h.cli.setReleaseAgentLabelsMany(nCtx, req)
}

// EnableReleaseAgent enables agent release.
func (h *Handler) EnableReleaseAgent(nCtx contextx.IContext,key types.ReleaseAgentKey) error {
	req := &protoBackend.PackageReleaseAgentEnableReq{
		Generation: int64(key.Generation),
		Platform:   protoBackend.ConvertPlatformFromTypes(key.Platform),
		Version:    key.Version,
	}

	return h.cli.enableReleaseAgent(nCtx, req)
}

// DisableReleaseAgent disables agent release.
func (h *Handler) DisableReleaseAgent(nCtx contextx.IContext,key types.ReleaseAgentKey) error {
	req := &protoBackend.PackageReleaseAgentDisableReq{
		Generation: int64(key.Generation),
		Platform:   protoBackend.ConvertPlatformFromTypes(key.Platform),
		Version:    key.Version,
	}

	return h.cli.disableReleaseAgent(nCtx, req)
}

// SetAsDefaultReleaseAgent sets agent release as default.
func (h *Handler) SetAsDefaultReleaseAgent(nCtx contextx.IContext,key types.ReleaseAgentKey) error {
	req := &protoBackend.PackageReleaseAgentSetAsDefaultReq{
		Generation: int64(key.Generation),
		Platform:   protoBackend.ConvertPlatformFromTypes(key.Platform),
		Version:    key.Version,
	}

	return h.cli.setAsDefaultReleaseAgent(nCtx, req)
}

// CancelAsDefaultReleaseAgent cancels agent release as default.
func (h *Handler) CancelAsDefaultReleaseAgent(nCtx contextx.IContext,key types.ReleaseAgentKey) error {
	req := &protoBackend.PackageReleaseAgentCancelAsDefaultReq{
		Generation: int64(key.Generation),
		Platform:   protoBackend.ConvertPlatformFromTypes(key.Platform),
		Version:    key.Version,
	}

	return h.cli.cancelAsDefaultReleaseAgent(nCtx, req)
}

// DeleteReleaseAgent deletes agent release.
func (h *Handler) DeleteReleaseAgent(nCtx contextx.IContext,key types.ReleaseAgentKey) error {
	req := &protoBackend.PackageReleaseAgentDeleteReq{
		Generation: int64(key.Generation),
		Platform:   protoBackend.ConvertPlatformFromTypes(key.Platform),
		Version:    key.Version,
	}

	return h.cli.deleteReleaseAgent(nCtx, req)
}

// ===============================================================================
// Release Proxy Related Interfaces
// ===============================================================================

// ListReleaseProxy lists proxy releases by page and conditions.
func (h *Handler) ListReleaseProxy(nCtx contextx.IContext,gen types.Generation,page types.Page,condition *types.ReleaseCondition) (
	[]*types.ReleaseProxy, int64, error) {

	req := &protoBackend.PackageReleaseProxyListReq{
		Generation: int64(gen),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	var (
		total    int64
		releases []*types.ReleaseProxy
	)
	executor := pageexecutor.NewPageExecutor[*types.ReleaseProxy](req.PageLimit(), req.PageTimeout())
	fn := func(nCtx contextx.IContext, page types.Page) ([]*types.ReleaseProxy, error) {
		req.Page = convertPage(page)
		resp, err := h.cli.listReleaseProxy(nCtx, req)
		if err != nil {
			return nil, err
		}

		total, releases = resp.ConvertReleasesToTypes()

		return releases, nil
	}

	result, err := executor.Execute(nCtx, page, fn)
	if err != nil {
		return nil, 0, err
	}

	return result.Items, total, nil
}

// CountReleaseProxy counts proxy releases by conditions.
func (h *Handler) CountReleaseProxy(nCtx contextx.IContext,gen types.Generation,condition *types.ReleaseCondition) (int64, error) {
	req := &protoBackend.PackageReleaseProxyListReq{
		Generation: int64(gen),
		OnlyCount:  true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}

	resp, err := h.cli.listReleaseProxy(nCtx, req)
	if err != nil {
		return 0, err
	}

	total, _ := resp.ConvertReleasesToTypes()

	return total, nil
}

// DistinctReleaseProxy distincts proxy releases by conditions.
func (h *Handler) DistinctReleaseProxy(
	nCtx contextx.IContext,gen types.Generation,distinctField types.ReleaseDistinctField,condition *types.ReleaseCondition) (
	*types.ReleaseDistinctResult, error) {

	req := &protoBackend.PackageReleaseProxyDistinctReq{
		Generation: int64(gen),
		DistinctField: &protoBackend.PackageReleaseDistinctField{
			OsType:  distinctField.OSType,
			CpuArch: distinctField.CPUArch,
		},
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, err
	}

	resp, err := h.cli.distinctReleaseProxy(nCtx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertResultToTypes(), nil
}

// SetReleaseProxyLabelsMany sets many proxy release labels.
func (h *Handler) SetReleaseProxyLabelsMany(nCtx contextx.IContext,gen types.Generation,labels []string,condition *types.ReleaseCondition) error {
	req := &protoBackend.PackageReleaseProxySetLabelsManyReq{
		Generation: int64(gen),
		Labels:     labels,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return err
	}

	return h.cli.setReleaseProxyLabelsMany(nCtx, req)
}

// EnableReleaseProxy enables proxy release.
func (h *Handler) EnableReleaseProxy(nCtx contextx.IContext,key types.ReleaseProxyKey) error {
	req := &protoBackend.PackageReleaseProxyEnableReq{
		Generation: int64(key.Generation),
		Platform:   protoBackend.ConvertPlatformFromTypes(key.Platform),
		Version:    key.Version,
	}

	return h.cli.enableReleaseProxy(nCtx, req)
}

// DisableReleaseProxy disables proxy release.
func (h *Handler) DisableReleaseProxy(nCtx contextx.IContext,key types.ReleaseProxyKey) error {
	req := &protoBackend.PackageReleaseProxyDisableReq{
		Generation: int64(key.Generation),
		Platform:   protoBackend.ConvertPlatformFromTypes(key.Platform),
		Version:    key.Version,
	}

	return h.cli.disableReleaseProxy(nCtx, req)
}

// SetAsDefaultReleaseProxy sets proxy release as default.
func (h *Handler) SetAsDefaultReleaseProxy(nCtx contextx.IContext,key types.ReleaseProxyKey) error {
	req := &protoBackend.PackageReleaseProxySetAsDefaultReq{
		Generation: int64(key.Generation),
		Platform:   protoBackend.ConvertPlatformFromTypes(key.Platform),
		Version:    key.Version,
	}

	return h.cli.setAsDefaultReleaseProxy(nCtx, req)
}

// CancelAsDefaultReleaseProxy cancels proxy release as default.
func (h *Handler) CancelAsDefaultReleaseProxy(nCtx contextx.IContext,key types.ReleaseProxyKey) error {
	req := &protoBackend.PackageReleaseProxyCancelAsDefaultReq{
		Generation: int64(key.Generation),
		Platform:   protoBackend.ConvertPlatformFromTypes(key.Platform),
		Version:    key.Version,
	}

	return h.cli.cancelAsDefaultReleaseProxy(nCtx, req)
}

// DeleteReleaseProxy deletes proxy release.
func (h *Handler) DeleteReleaseProxy(nCtx contextx.IContext,key types.ReleaseProxyKey) error {
	req := &protoBackend.PackageReleaseProxyDeleteReq{
		Generation: int64(key.Generation),
		Platform:   protoBackend.ConvertPlatformFromTypes(key.Platform),
		Version:    key.Version,
	}

	return h.cli.deleteReleaseProxy(nCtx, req)
}

// ===============================================================================
// Release Plugin Related Interfaces
// ===============================================================================

// ListReleasePlugin lists plugin releases by page and conditions.
func (h *Handler) ListReleasePlugin(nCtx contextx.IContext,gen types.Generation,page types.Page,condition *types.ReleaseCondition) (
	[]*types.ReleasePlugin, int64, error) {

	req := &protoBackend.PackageReleasePluginListReq{
		Generation: int64(gen),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	var (
		total    int64
		releases []*types.ReleasePlugin
	)
	executor := pageexecutor.NewPageExecutor[*types.ReleasePlugin](req.PageLimit(), req.PageTimeout())
	fn := func(nCtx contextx.IContext, page types.Page) ([]*types.ReleasePlugin, error) {
		req.Page = convertPage(page)
		resp, err := h.cli.listReleasePlugin(nCtx, req)
		if err != nil {
			return nil, err
		}

		total, releases = resp.ConvertReleasePluginsToTypes()

		return releases, nil
	}

	result, err := executor.Execute(nCtx, page, fn)
	if err != nil {
		return nil, 0, err
	}

	return result.Items, total, nil
}

// CountReleasePlugin counts plugin releases by conditions.
func (h *Handler) CountReleasePlugin(nCtx contextx.IContext,gen types.Generation,condition *types.ReleaseCondition) (int64, error) {
	req := &protoBackend.PackageReleasePluginListReq{
		Generation: int64(gen),
		OnlyCount:  true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}

	resp, err := h.cli.listReleasePlugin(nCtx, req)
	if err != nil {
		return 0, err
	}

	total, _ := resp.ConvertReleasePluginsToTypes()

	return total, nil
}

// EnableReleasePlugin enables plugin release.
func (h *Handler) EnableReleasePlugin(nCtx contextx.IContext,key types.ReleasePluginKey) error {
	req := &protoBackend.PackageReleasePluginEnableReq{}
	req.SetIdentifer(key.Name, key.Generation, key.Platform, key.Version)

	return h.cli.enableReleasePlugin(nCtx, req)
}

// DisableReleasePlugin disables plugin release.
func (h *Handler) DisableReleasePlugin(nCtx contextx.IContext,key types.ReleasePluginKey) error {
	req := &protoBackend.PackageReleasePluginDisableReq{}
	req.SetIdentifer(key.Name, key.Generation, key.Platform, key.Version)

	return h.cli.disableReleasePlugin(nCtx, req)
}

// SetAsDefaultReleasePlugin sets plugin release as default.
func (h *Handler) SetAsDefaultReleasePlugin(nCtx contextx.IContext,key types.ReleasePluginKey) error {
	req := &protoBackend.PackageReleasePluginSetAsDefaultReq{}
	req.SetIdentifer(key.Name, key.Generation, key.Platform, key.Version)

	return h.cli.setAsDefaultReleasePlugin(nCtx, req)
}

// CancelAsDefaultReleasePlugin cancels plugin release as default.
func (h *Handler) CancelAsDefaultReleasePlugin(nCtx contextx.IContext,key types.ReleasePluginKey) error {
	req := &protoBackend.PackageReleasePluginCancelAsDefaultReq{}
	req.SetIdentifer(key.Name, key.Generation, key.Platform, key.Version)

	return h.cli.cancelAsDefaultReleasePlugin(nCtx, req)
}

// DeleteReleasePlugin deletes plugin release.
func (h *Handler) DeleteReleasePlugin(nCtx contextx.IContext,key types.ReleasePluginKey) error {
	req := &protoBackend.PackageReleasePluginDeleteReq{}
	req.SetIdentifer(key.Name, key.Generation, key.Platform, key.Version)

	return h.cli.deleteReleasePlugin(nCtx, req)
}

// ===============================================================================
// Release Cert Related Interfaces
// ===============================================================================

// ListReleaseCert lists cert releases by page and conditions.
func (h *Handler) ListReleaseCert(nCtx contextx.IContext,gen types.Generation) ([]*types.ReleaseCert, int64, error) {
	req := &protoBackend.PackageReleaseCertListReq{
		Generation: int64(gen),
	}

	resp, err := h.cli.listReleaseCert(nCtx, req)
	if err != nil {
		return nil, 0, err
	}

	data := resp.GetData()
	if data == nil {
		return nil, 0, nil
	}

	items := data.GetItems()
	releases := make([]*types.ReleaseCert, len(items))
	for idx, item := range items {
		releases[idx] = convertReleaseCertFromProto(item)
	}

	return releases, data.GetTotal(), nil
}

// DeleteReleaseCert deletes cert release.
func (h *Handler) DeleteReleaseCert(nCtx contextx.IContext,key types.ReleaseCertKey) error {
	req := &protoBackend.PackageReleaseCertDeleteReq{
		Generation: int64(key.Generation),
	}

	return h.cli.deleteReleaseCert(nCtx, req)
}

// convertReleaseCertFromProto converts proto ReleaseCert to types.ReleaseCert.
func convertReleaseCertFromProto(item *protoBackend.ReleaseCert) *types.ReleaseCert {
	release := item.GetRelease()
	return &types.ReleaseCert{
		Release: types.Release{
			Name:       release.GetName(),
			Generation: types.Generation(release.GetGeneration()),
			Type:       types.ReleaseType(release.GetReleaseType()),
			Version:    release.GetVersion(),
			Platform: platfmt.Platform{
				OS:   criteria.OSType(release.GetOsType()),
				Arch: criteria.CPUArch(release.GetCpuArch()),
			},
			Labels:    release.GetLabels(),
			FileName:  release.GetFileName(),
			MD5:       release.GetMd5(),
			Enabled:   release.GetEnabled(),
			AsDefault: release.GetAsDefault(),
			UpdatedAt: time.UnixMilli(int64(release.GetUpdatedAt())).Local(),
			Operator:  release.GetOperator(),
		},
	}
}

// ===============================================================================
// Release Bintool Related Interfaces
// ===============================================================================

// ListReleaseBinTool lists bintool releases by page and conditions.
func (h *Handler) ListReleaseBinTool(nCtx contextx.IContext,gen types.Generation) ([]*types.ReleaseBinTool, int64, error) {
	req := &protoBackend.PackageReleaseBinToolListReq{
		Generation: int64(gen),
	}

	resp, err := h.cli.listReleaseBinTool(nCtx, req)
	if err != nil {
		return nil, 0, err
	}

	data := resp.GetData()
	if data == nil {
		return nil, 0, nil
	}

	items := data.GetItems()
	releases := make([]*types.ReleaseBinTool, len(items))
	for idx, item := range items {
		releases[idx] = convertReleaseBinToolFromProto(item)
	}

	return releases, data.GetTotal(), nil
}

// DeleteReleaseBinTool deletes bintool release.
func (h *Handler) DeleteReleaseBinTool(nCtx contextx.IContext,key types.ReleaseBinToolKey) error {
	req := &protoBackend.PackageReleaseBinToolDeleteReq{
		Generation: int64(key.Generation),
	}

	return h.cli.deleteReleaseBinTool(nCtx, req)
}

// ===============================================================================
// Release Plugin Bintool Related Interfaces
// ===============================================================================

// ListReleasePluginBinTool lists plugin-bintool releases by page and conditions.
func (h *Handler) ListReleasePluginBinTool(nCtx contextx.IContext,gen types.Generation) ([]*types.ReleasePluginBinTool, int64, error) {
	req := &protoBackend.PackageReleasePluginBinToolListReq{
		Generation: int64(gen),
	}

	resp, err := h.cli.listReleasePluginBinTool(nCtx, req)
	if err != nil {
		return nil, 0, err
	}

	data := resp.GetData()
	if data == nil {
		return nil, 0, nil
	}

	items := data.GetItems()
	releases := make([]*types.ReleasePluginBinTool, len(items))
	for idx, item := range items {
		releases[idx] = convertReleasePluginBinToolFromProto(item)
	}

	return releases, data.GetTotal(), nil
}

// DeleteReleasePluginBinTool deletes plugin-bintool release.
func (h *Handler) DeleteReleasePluginBinTool(nCtx contextx.IContext,key types.ReleasePluginBinToolKey) error {
	req := &protoBackend.PackageReleasePluginBinToolDeleteReq{
		Generation: int64(key.Generation),
		Name:       key.Name,
	}

	return h.cli.deleteReleasePluginBinTool(nCtx, req)
}

// convertReleaseBinToolFromProto converts proto ReleaseBinTool to types.ReleaseBinTool.
func convertReleaseBinToolFromProto(item *protoBackend.ReleaseBinTool) *types.ReleaseBinTool {
	release := item.GetRelease()
	return &types.ReleaseBinTool{
		Release: types.Release{
			Name:       release.GetName(),
			Generation: types.Generation(release.GetGeneration()),
			Type:       types.ReleaseType(release.GetReleaseType()),
			Version:    release.GetVersion(),
			Platform: platfmt.Platform{
				OS:   criteria.OSType(release.GetOsType()),
				Arch: criteria.CPUArch(release.GetCpuArch()),
			},
			Labels:    release.GetLabels(),
			FileName:  release.GetFileName(),
			MD5:       release.GetMd5(),
			Enabled:   release.GetEnabled(),
			AsDefault: release.GetAsDefault(),
			UpdatedAt: time.UnixMilli(int64(release.GetUpdatedAt())).Local(),
			Operator:  release.GetOperator(),
		},
	}
}

// convertReleasePluginBinToolFromProto converts proto ReleasePluginBinTool to types.ReleasePluginBinTool.
func convertReleasePluginBinToolFromProto(item *protoBackend.ReleasePluginBinTool) *types.ReleasePluginBinTool {
	release := item.GetRelease()
	return &types.ReleasePluginBinTool{
		Release: types.Release{
			Name:       release.GetName(),
			Generation: types.Generation(release.GetGeneration()),
			Type:       types.ReleaseType(release.GetReleaseType()),
			Version:    release.GetVersion(),
			Platform: platfmt.Platform{
				OS:   criteria.OSType(release.GetOsType()),
				Arch: criteria.CPUArch(release.GetCpuArch()),
			},
			Labels:    release.GetLabels(),
			FileName:  release.GetFileName(),
			MD5:       release.GetMd5(),
			Enabled:   release.GetEnabled(),
			AsDefault: release.GetAsDefault(),
			UpdatedAt: time.UnixMilli(int64(release.GetUpdatedAt())).Local(),
			Operator:  release.GetOperator(),
		},
	}
}
