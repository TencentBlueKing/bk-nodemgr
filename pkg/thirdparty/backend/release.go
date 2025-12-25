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
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandlerRelease defines the backend Handler for release.
// nolint: interfacebloat
type IHandlerRelease interface {

	// ==================== Agent Methods ====================

	// ListReleaseAgent lists agent releases by page and conditions.
	ListReleaseAgent(ctx contextx.IContext, gen types.Generation, page types.Page, condition *types.ReleaseCondition) (
		[]*types.ReleaseAgent, int64, error)

	// CountReleaseAgent counts agent releases by conditions.
	CountReleaseAgent(ctx contextx.IContext, gen types.Generation, condition *types.ReleaseCondition) (int64, error)

	// DistinctReleaseAgent distincts agent releases by conditions.
	DistinctReleaseAgent(ctx contextx.IContext, gen types.Generation, distinctField types.ReleaseDistinctField, condition *types.ReleaseCondition) (
		*types.ReleaseDistinctResult, error)

	// SetReleaseAgentLabelsMany sets many agent release labels.
	SetReleaseAgentLabelsMany(ctx contextx.IContext, gen types.Generation, labels []string, condition *types.ReleaseCondition) error

	// EnableReleaseAgent enables agent release.
	EnableReleaseAgent(ctx contextx.IContext, key types.ReleaseAgentKey) error

	// DisableReleaseAgent disables agent release.
	DisableReleaseAgent(ctx contextx.IContext, key types.ReleaseAgentKey) error

	// SetAsDefaultReleaseAgent sets agent release as default.
	SetAsDefaultReleaseAgent(ctx contextx.IContext, key types.ReleaseAgentKey) error

	// CancelAsDefaultReleaseAgent cancels agent release as default.
	CancelAsDefaultReleaseAgent(ctx contextx.IContext, key types.ReleaseAgentKey) error

	// DeleteReleaseAgent deletes agent release.
	DeleteReleaseAgent(ctx contextx.IContext, key types.ReleaseAgentKey) error

	// ==================== Proxy Methods ====================

	// ListReleaseProxy lists proxy releases by page and conditions.
	ListReleaseProxy(ctx contextx.IContext, gen types.Generation, page types.Page, condition *types.ReleaseCondition) (
		[]*types.ReleaseProxy, int64, error)

	// CountReleaseProxy counts proxy releases by conditions.
	CountReleaseProxy(ctx contextx.IContext, gen types.Generation, condition *types.ReleaseCondition) (int64, error)

	// DistinctReleaseProxy distincts proxy releases by conditions.
	DistinctReleaseProxy(ctx contextx.IContext, gen types.Generation, distinctField types.ReleaseDistinctField, condition *types.ReleaseCondition) (
		*types.ReleaseDistinctResult, error)

	// SetReleaseProxyLabelsMany sets many proxy release labels.
	SetReleaseProxyLabelsMany(ctx contextx.IContext, gen types.Generation, labels []string, condition *types.ReleaseCondition) error

	// EnableReleaseProxy enables proxy release.
	EnableReleaseProxy(ctx contextx.IContext, key types.ReleaseProxyKey) error

	// DisableReleaseProxy disables proxy release.
	DisableReleaseProxy(ctx contextx.IContext, key types.ReleaseProxyKey) error

	// SetAsDefaultReleaseProxy sets proxy release as default.
	SetAsDefaultReleaseProxy(ctx contextx.IContext, key types.ReleaseProxyKey) error

	// CancelAsDefaultReleaseProxy cancels proxy release as default.
	CancelAsDefaultReleaseProxy(ctx contextx.IContext, key types.ReleaseProxyKey) error

	// DeleteReleaseProxy deletes proxy release.
	DeleteReleaseProxy(ctx contextx.IContext, key types.ReleaseProxyKey) error

	// ==================== Plugin Methods ====================

	// ListReleasePlugin lists plugin releases by page and conditions.
	ListReleasePlugin(ctx contextx.IContext, gen types.Generation, page types.Page, condition *types.ReleaseCondition) (
		[]*types.ReleasePlugin, int64, error)

	// CountReleasePlugin counts plugin releases by conditions.
	CountReleasePlugin(ctx contextx.IContext, gen types.Generation, condition *types.ReleaseCondition) (int64, error)

	// EnableReleasePlugin enables plugin release.
	EnableReleasePlugin(ctx contextx.IContext, key types.ReleasePluginKey) error

	// DisableReleasePlugin disables plugin release.
	DisableReleasePlugin(ctx contextx.IContext, key types.ReleasePluginKey) error

	// SetAsDefaultReleasePlugin sets plugin release as default.
	SetAsDefaultReleasePlugin(ctx contextx.IContext, key types.ReleasePluginKey) error

	// CancelAsDefaultReleasePlugin cancels plugin release as default.
	CancelAsDefaultReleasePlugin(ctx contextx.IContext, key types.ReleasePluginKey) error

	// DeleteReleasePlugin deletes plugin release.
	DeleteReleasePlugin(ctx contextx.IContext, key types.ReleasePluginKey) error

	// ==================== Cert Methods ====================

	// ListReleaseCert lists cert releases by page and conditions.
	ListReleaseCert(ctx contextx.IContext, gen types.Generation) ([]*types.ReleaseCert, int64, error)

	// DeleteReleaseCert deletes cert release.
	DeleteReleaseCert(ctx contextx.IContext, key types.ReleaseCertKey) error

	// ==================== BinTool Methods ====================

	// ListReleaseBinTool lists bintool releases by page and conditions.
	ListReleaseBinTool(ctx contextx.IContext, gen types.Generation) ([]*types.ReleaseBinTool, int64, error)

	// DeleteReleaseBinTool deletes bintool release.
	DeleteReleaseBinTool(ctx contextx.IContext, key types.ReleaseBinToolKey) error

	// ==================== PluginBinTool Methods ====================

	// ListReleasePluginBinTool lists plugin-bintool releases by page and conditions.
	ListReleasePluginBinTool(ctx contextx.IContext, gen types.Generation) ([]*types.ReleasePluginBinTool, int64, error)

	// DeleteReleasePluginBinTool deletes plugin-bintool release.
	DeleteReleasePluginBinTool(ctx contextx.IContext, key types.ReleasePluginBinToolKey) error
}

// ==================== Agent Methods ====================

// ListReleaseAgent lists agent releases by page and conditions.
func (h *Handler) ListReleaseAgent(
	ctx contextx.IContext,
	gen types.Generation,
	page types.Page,
	condition *types.ReleaseCondition,
) ([]*types.ReleaseAgent, int64, error) {

	req := &protoBackend.PackageReleaseAgentListReq{
		Page:       convertPage(page),
		Generation: int64(gen),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listReleaseAgent(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	total, releases := resp.ConvertReleasesToTypes()

	return releases, total, nil
}

// CountReleaseAgent counts agent releases by conditions.
func (h *Handler) CountReleaseAgent(
	ctx contextx.IContext,
	gen types.Generation,
	condition *types.ReleaseCondition,
) (int64, error) {

	req := &protoBackend.PackageReleaseAgentListReq{
		Generation: int64(gen),
		OnlyCount:  true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}

	resp, err := h.cli.listReleaseAgent(ctx, req)
	if err != nil {
		return 0, err
	}

	total, _ := resp.ConvertReleasesToTypes()

	return total, nil
}

// DistinctReleaseAgent distincts agent releases by conditions.
func (h *Handler) DistinctReleaseAgent(
	ctx contextx.IContext,
	gen types.Generation,
	distinctField types.ReleaseDistinctField,
	condition *types.ReleaseCondition,
) (*types.ReleaseDistinctResult, error) {

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

	resp, err := h.cli.distinctReleaseAgent(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertResultToTypes(), nil
}

// SetReleaseAgentLabelsMany sets many agent release labels.
func (h *Handler) SetReleaseAgentLabelsMany(
	ctx contextx.IContext,
	gen types.Generation,
	labels []string,
	condition *types.ReleaseCondition,
) error {

	req := &protoBackend.PackageReleaseAgentSetLabelsManyReq{
		Generation: int64(gen),
		Labels:     labels,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return err
	}

	return h.cli.setReleaseAgentLabelsMany(ctx, req)
}

// EnableReleaseAgent enables agent release.
func (h *Handler) EnableReleaseAgent(
	ctx contextx.IContext,
	key types.ReleaseAgentKey,
) error {

	req := &protoBackend.PackageReleaseAgentEnableReq{
		Generation: int64(key.Generation),
		Platform:   protoBackend.ConvertPlatformFromTypes(key.Platform),
		Version:    key.Version,
	}

	return h.cli.enableReleaseAgent(ctx, req)
}

// DisableReleaseAgent disables agent release.
func (h *Handler) DisableReleaseAgent(
	ctx contextx.IContext,
	key types.ReleaseAgentKey,
) error {

	req := &protoBackend.PackageReleaseAgentDisableReq{
		Generation: int64(key.Generation),
		Platform:   protoBackend.ConvertPlatformFromTypes(key.Platform),
		Version:    key.Version,
	}

	return h.cli.disableReleaseAgent(ctx, req)
}

// SetAsDefaultReleaseAgent sets agent release as default.
func (h *Handler) SetAsDefaultReleaseAgent(
	ctx contextx.IContext,
	key types.ReleaseAgentKey,
) error {

	req := &protoBackend.PackageReleaseAgentSetAsDefaultReq{
		Generation: int64(key.Generation),
		Platform:   protoBackend.ConvertPlatformFromTypes(key.Platform),
		Version:    key.Version,
	}

	return h.cli.setAsDefaultReleaseAgent(ctx, req)
}

// CancelAsDefaultReleaseAgent cancels agent release as default.
func (h *Handler) CancelAsDefaultReleaseAgent(
	ctx contextx.IContext,
	key types.ReleaseAgentKey,
) error {

	req := &protoBackend.PackageReleaseAgentCancelAsDefaultReq{
		Generation: int64(key.Generation),
		Platform:   protoBackend.ConvertPlatformFromTypes(key.Platform),
		Version:    key.Version,
	}

	return h.cli.cancelAsDefaultReleaseAgent(ctx, req)
}

// DeleteReleaseAgent deletes agent release.
func (h *Handler) DeleteReleaseAgent(
	ctx contextx.IContext,
	key types.ReleaseAgentKey,
) error {

	req := &protoBackend.PackageReleaseAgentDeleteReq{
		Generation: int64(key.Generation),
		Platform:   protoBackend.ConvertPlatformFromTypes(key.Platform),
		Version:    key.Version,
	}

	return h.cli.deleteReleaseAgent(ctx, req)
}

// ==================== Proxy Methods ====================

// ListReleaseProxy lists proxy releases by page and conditions.
func (h *Handler) ListReleaseProxy(
	ctx contextx.IContext,
	gen types.Generation,
	page types.Page,
	condition *types.ReleaseCondition,
) ([]*types.ReleaseProxy, int64, error) {

	req := &protoBackend.PackageReleaseProxyListReq{
		Page:       convertPage(page),
		Generation: int64(gen),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listReleaseProxy(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	total, releases := resp.ConvertReleasesToTypes()

	return releases, total, nil
}

// CountReleaseProxy counts proxy releases by conditions.
func (h *Handler) CountReleaseProxy(
	ctx contextx.IContext,
	gen types.Generation,
	condition *types.ReleaseCondition,
) (int64, error) {

	req := &protoBackend.PackageReleaseProxyListReq{
		Generation: int64(gen),
		OnlyCount:  true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}

	resp, err := h.cli.listReleaseProxy(ctx, req)
	if err != nil {
		return 0, err
	}

	total, _ := resp.ConvertReleasesToTypes()

	return total, nil
}

// DistinctReleaseProxy distincts proxy releases by conditions.
func (h *Handler) DistinctReleaseProxy(
	ctx contextx.IContext,
	gen types.Generation,
	distinctField types.ReleaseDistinctField,
	condition *types.ReleaseCondition,
) (*types.ReleaseDistinctResult, error) {

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

	resp, err := h.cli.distinctReleaseProxy(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertResultToTypes(), nil
}

// SetReleaseProxyLabelsMany sets many proxy release labels.
func (h *Handler) SetReleaseProxyLabelsMany(
	ctx contextx.IContext,
	gen types.Generation,
	labels []string,
	condition *types.ReleaseCondition,
) error {

	req := &protoBackend.PackageReleaseProxySetLabelsManyReq{
		Generation: int64(gen),
		Labels:     labels,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return err
	}

	return h.cli.setReleaseProxyLabelsMany(ctx, req)
}

// EnableReleaseProxy enables proxy release.
func (h *Handler) EnableReleaseProxy(
	ctx contextx.IContext,
	key types.ReleaseProxyKey,
) error {

	req := &protoBackend.PackageReleaseProxyEnableReq{
		Generation: int64(key.Generation),
		Platform:   protoBackend.ConvertPlatformFromTypes(key.Platform),
		Version:    key.Version,
	}

	return h.cli.enableReleaseProxy(ctx, req)
}

// DisableReleaseProxy disables proxy release.
func (h *Handler) DisableReleaseProxy(
	ctx contextx.IContext,
	key types.ReleaseProxyKey,
) error {

	req := &protoBackend.PackageReleaseProxyDisableReq{
		Generation: int64(key.Generation),
		Platform:   protoBackend.ConvertPlatformFromTypes(key.Platform),
		Version:    key.Version,
	}

	return h.cli.disableReleaseProxy(ctx, req)
}

// SetAsDefaultReleaseProxy sets proxy release as default.
func (h *Handler) SetAsDefaultReleaseProxy(
	ctx contextx.IContext,
	key types.ReleaseProxyKey,
) error {

	req := &protoBackend.PackageReleaseProxySetAsDefaultReq{
		Generation: int64(key.Generation),
		Platform:   protoBackend.ConvertPlatformFromTypes(key.Platform),
		Version:    key.Version,
	}

	return h.cli.setAsDefaultReleaseProxy(ctx, req)
}

// CancelAsDefaultReleaseProxy cancels proxy release as default.
func (h *Handler) CancelAsDefaultReleaseProxy(
	ctx contextx.IContext,
	key types.ReleaseProxyKey,
) error {

	req := &protoBackend.PackageReleaseProxyCancelAsDefaultReq{
		Generation: int64(key.Generation),
		Platform:   protoBackend.ConvertPlatformFromTypes(key.Platform),
		Version:    key.Version,
	}

	return h.cli.cancelAsDefaultReleaseProxy(ctx, req)
}

// DeleteReleaseProxy deletes proxy release.
func (h *Handler) DeleteReleaseProxy(
	ctx contextx.IContext,
	key types.ReleaseProxyKey,
) error {

	req := &protoBackend.PackageReleaseProxyDeleteReq{
		Generation: int64(key.Generation),
		Platform:   protoBackend.ConvertPlatformFromTypes(key.Platform),
		Version:    key.Version,
	}

	return h.cli.deleteReleaseProxy(ctx, req)
}

// ==================== Plugin Methods ====================

// ListReleasePlugin lists plugin releases by page and conditions.
func (h *Handler) ListReleasePlugin(
	ctx contextx.IContext,
	gen types.Generation,
	page types.Page,
	condition *types.ReleaseCondition,
) ([]*types.ReleasePlugin, int64, error) {

	req := &protoBackend.PackageReleasePluginListReq{
		Page:       convertPage(page),
		Generation: int64(gen),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listReleasePlugin(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	total, releases := resp.ConvertReleasePluginsToTypes()

	return releases, total, nil
}

// CountReleasePlugin counts plugin releases by conditions.
func (h *Handler) CountReleasePlugin(ctx contextx.IContext,
	gen types.Generation,
	condition *types.ReleaseCondition,
) (int64, error) {

	req := &protoBackend.PackageReleasePluginListReq{
		Generation: int64(gen),
		OnlyCount:  true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}

	resp, err := h.cli.listReleasePlugin(ctx, req)
	if err != nil {
		return 0, err
	}

	total, _ := resp.ConvertReleasePluginsToTypes()

	return total, nil
}

// EnableReleasePlugin enables plugin release.
func (h *Handler) EnableReleasePlugin(
	ctx contextx.IContext,
	key types.ReleasePluginKey,
) error {

	req := &protoBackend.PackageReleasePluginEnableReq{}
	req.SetIdentifer(key.Name, key.Generation, key.Platform, key.Version)

	return h.cli.enableReleasePlugin(ctx, req)
}

// DisableReleasePlugin disables plugin release.
func (h *Handler) DisableReleasePlugin(
	ctx contextx.IContext,
	key types.ReleasePluginKey,
) error {

	req := &protoBackend.PackageReleasePluginDisableReq{}
	req.SetIdentifer(key.Name, key.Generation, key.Platform, key.Version)

	return h.cli.disableReleasePlugin(ctx, req)
}

// SetAsDefaultReleasePlugin sets plugin release as default.
func (h *Handler) SetAsDefaultReleasePlugin(
	ctx contextx.IContext,
	key types.ReleasePluginKey,
) error {

	req := &protoBackend.PackageReleasePluginSetAsDefaultReq{}
	req.SetIdentifer(key.Name, key.Generation, key.Platform, key.Version)

	return h.cli.setAsDefaultReleasePlugin(ctx, req)
}

// CancelAsDefaultReleasePlugin cancels plugin release as default.
func (h *Handler) CancelAsDefaultReleasePlugin(
	ctx contextx.IContext,
	key types.ReleasePluginKey,
) error {

	req := &protoBackend.PackageReleasePluginCancelAsDefaultReq{}
	req.SetIdentifer(key.Name, key.Generation, key.Platform, key.Version)

	return h.cli.cancelAsDefaultReleasePlugin(ctx, req)
}

// DeleteReleasePlugin deletes plugin release.
func (h *Handler) DeleteReleasePlugin(
	ctx contextx.IContext,
	key types.ReleasePluginKey,
) error {

	req := &protoBackend.PackageReleasePluginDeleteReq{}
	req.SetIdentifer(key.Name, key.Generation, key.Platform, key.Version)

	return h.cli.deleteReleasePlugin(ctx, req)
}

// ==================== Cert Methods ====================

// ListReleaseCert lists cert releases by page and conditions.
func (h *Handler) ListReleaseCert(
	ctx contextx.IContext,
	gen types.Generation,
) ([]*types.ReleaseCert, int64, error) {

	req := &protoBackend.PackageReleaseCertListReq{
		Generation: int64(gen),
	}

	resp, err := h.cli.listReleaseCert(ctx, req)
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
func (h *Handler) DeleteReleaseCert(
	ctx contextx.IContext,
	key types.ReleaseCertKey,
) error {

	req := &protoBackend.PackageReleaseCertDeleteReq{
		Generation: int64(key.Generation),
	}

	return h.cli.deleteReleaseCert(ctx, req)
}

// ==================== BinTool Methods ====================

// ListReleaseBinTool lists bintool releases by page and conditions.
func (h *Handler) ListReleaseBinTool(
	ctx contextx.IContext,
	gen types.Generation,
) ([]*types.ReleaseBinTool, int64, error) {

	req := &protoBackend.PackageReleaseBinToolListReq{
		Generation: int64(gen),
	}

	resp, err := h.cli.listReleaseBinTool(ctx, req)
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
func (h *Handler) DeleteReleaseBinTool(
	ctx contextx.IContext,
	key types.ReleaseBinToolKey,
) error {

	req := &protoBackend.PackageReleaseBinToolDeleteReq{
		Generation: int64(key.Generation),
	}

	return h.cli.deleteReleaseBinTool(ctx, req)
}

// ==================== PluginBinTool Methods ====================

// ListReleasePluginBinTool lists plugin-bintool releases by page and conditions.
func (h *Handler) ListReleasePluginBinTool(
	ctx contextx.IContext,
	gen types.Generation,
) ([]*types.ReleasePluginBinTool, int64, error) {

	req := &protoBackend.PackageReleasePluginBinToolListReq{
		Generation: int64(gen),
	}

	resp, err := h.cli.listReleasePluginBinTool(ctx, req)
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
func (h *Handler) DeleteReleasePluginBinTool(
	ctx contextx.IContext,
	key types.ReleasePluginBinToolKey,
) error {

	req := &protoBackend.PackageReleasePluginBinToolDeleteReq{
		Generation: int64(key.Generation),
		Name:       key.Name,
	}

	return h.cli.deleteReleasePluginBinTool(ctx, req)
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
