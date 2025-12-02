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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandlerRelease defines the backend Handler for release.
// nolint: interfacebloat
type IHandlerRelease interface {
	// ListRelease lists release by page and conditions.
	ListRelease(
		ctx contextx.IContext,
		releaseType types.ReleaseType,
		gen types.Generation,
		page types.Page,
		condition *types.ReleaseCondition,
	) ([]*types.Release, int64, error)

	// CountRelease counts release by conditions.
	CountRelease(
		ctx contextx.IContext,
		releaseType types.ReleaseType,
		gen types.Generation,
		condition *types.ReleaseCondition,
	) (int64, error)

	// ListReleaseAgent lists release by page and conditions.
	ListReleaseAgent(
		ctx contextx.IContext,
		gen types.Generation,
		page types.Page,
		condition *types.ReleaseCondition,
	) ([]*types.ReleaseAgent, int64, error)

	// CountReleaseAgent counts release by conditions.
	CountReleaseAgent(
		ctx contextx.IContext,
		gen types.Generation,
		condition *types.ReleaseCondition,
	) (int64, error)

	// ListReleaseProxy lists release proxy by page and conditions.
	ListReleaseProxy(
		ctx contextx.IContext,
		gen types.Generation,
		page types.Page,
		condition *types.ReleaseCondition,
	) ([]*types.ReleaseProxy, int64, error)

	// CountReleaseProxy counts release by conditions.
	CountReleaseProxy(
		ctx contextx.IContext,
		gen types.Generation,
		condition *types.ReleaseCondition,
	) (int64, error)

	// DistinctRelease distincts release by conditions.
	DistinctRelease(
		ctx contextx.IContext,
		releaseType types.ReleaseType,
		gen types.Generation,
		distinctField types.ReleaseDistinctField,
		condition *types.ReleaseCondition,
	) (*types.ReleaseDistinctResult, error)

	// SetReleaseLabels sets release labels.
	SetReleaseLabels(
		ctx contextx.IContext,
		gen types.Generation,
		releaseType types.ReleaseType,
		plat platfmt.Platform,
		version string,
		labels []string,
	) error

	// SetReleaseLabelsMany sets many release labels.
	SetReleaseLabelsMany(
		ctx contextx.IContext,
		releaseType types.ReleaseType,
		gen []types.Generation,
		plat []platfmt.Platform,
		version []string,
		labels []string,
	) error

	// EnableRelease enables release active by generation, release type, platform and version.
	EnableRelease(
		ctx contextx.IContext,
		gen types.Generation,
		releaseType types.ReleaseType,
		plat platfmt.Platform,
		version string,
	) error

	// DisableRelease disables release disactive by generation, release type, platform and version.
	DisableRelease(
		ctx contextx.IContext,
		gen types.Generation,
		releaseType types.ReleaseType,
		plat platfmt.Platform,
		version string,
	) error

	// SetAsDefaultRelease sets the release as default.
	SetAsDefaultRelease(
		ctx contextx.IContext,
		gen types.Generation,
		releaseType types.ReleaseType,
		plat platfmt.Platform,
		version string,
	) error

	// CancelAsDefaultRelease cancels the release as default.
	CancelAsDefaultRelease(
		ctx contextx.IContext,
		gen types.Generation,
		releaseType types.ReleaseType,
		plat platfmt.Platform,
		version string,
	) error

	// 	DeleteRelease deletes release by generation, release type, platform and version.
	DeleteRelease(
		ctx contextx.IContext,
		gen types.Generation,
		releaseType types.ReleaseType,
		plat platfmt.Platform,
		version string,
	) error
}

// IHandlerReleasePlugin defines the backend Handler for release plugin.
type IHandlerReleasePlugin interface {
	// ListReleasePlugin lists release plugin by page and conditions.
	ListReleasePlugin(
		ctx contextx.IContext,
		gen types.Generation,
		page types.Page,
		condition *types.ReleaseCondition,
	) ([]*types.ReleasePlugin, int64, error)

	// CountReleasePlugin counts release plugin by conditions.
	CountReleasePlugin(
		ctx contextx.IContext,
		gen types.Generation,
		condition *types.ReleaseCondition,
	) (int64, error)

	// EnableReleasePlugin enables release plugin active by name, generation, platform and version.
	EnableReleasePlugin(
		ctx contextx.IContext,
		name string,
		gen types.Generation,
		plat platfmt.Platform,
		version string,
	) error

	// DisableReleasePlugin disables release plugin disactive by name, generation, platform and version.
	DisableReleasePlugin(
		ctx contextx.IContext,
		name string,
		gen types.Generation,
		plat platfmt.Platform,
		version string,
	) error

	// SetAsDefaultReleasePlugin sets the release plugin as default.
	SetAsDefaultReleasePlugin(
		ctx contextx.IContext,
		name string,
		gen types.Generation,
		plat platfmt.Platform,
		version string,
	) error

	// CancelAsDefaultReleasePlugin cancels the release plugin as default.
	CancelAsDefaultReleasePlugin(
		ctx contextx.IContext,
		name string,
		gen types.Generation,
		plat platfmt.Platform,
		version string,
	) error

	// DeleteReleasePlugin deletes release plugin by name, generation, platform and version.
	DeleteReleasePlugin(
		ctx contextx.IContext,
		name string,
		gen types.Generation,
		plat platfmt.Platform,
		version string,
	) error
}

// ListRelease lists release by page and conditions.
func (h *Handler) ListRelease(
	ctx contextx.IContext,
	releaseType types.ReleaseType,
	gen types.Generation,
	page types.Page,
	condition *types.ReleaseCondition,
) ([]*types.Release, int64, error) {

	req := &protoBackend.PackageReleaseListReq{
		Page:        convertPage(page),
		ReleaseType: string(releaseType),
		Generation:  int64(gen),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listRelease(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	total, releases := resp.ConvertReleasesToTypes()

	return releases, total, nil
}

// CountRelease counts release by conditions.
func (h *Handler) CountRelease(
	ctx contextx.IContext,
	releaseType types.ReleaseType,
	gen types.Generation,
	condition *types.ReleaseCondition,
) (int64, error) {

	req := &protoBackend.PackageReleaseListReq{
		ReleaseType: string(releaseType),
		Generation:  int64(gen),
		OnlyCount:   true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}

	resp, err := h.cli.listRelease(ctx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// ListReleaseAgent lists release by page and conditions.
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

	total, releases := resp.ConvertReleasesAgentToTypes()

	return releases, total, nil
}

// CountReleaseAgent counts release by conditions.
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

	return resp.GetData().GetTotal(), nil
}

// ListReleaseProxy lists release by page and conditions.
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

	total, releases := resp.ConvertReleasesProxyToTypes()

	return releases, total, nil
}

// CountReleaseProxy counts release by conditions.
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

	return resp.GetData().GetTotal(), nil
}

// DistinctRelease distincts release by conditions.
func (h *Handler) DistinctRelease(
	ctx contextx.IContext,
	releaseType types.ReleaseType,
	gen types.Generation,
	distinctField types.ReleaseDistinctField,
	condition *types.ReleaseCondition,
) (*types.ReleaseDistinctResult, error) {

	req := &protoBackend.PackageReleaseDistinctReq{
		ReleaseType: string(releaseType),
		Generation:  int64(gen),
		DistinctField: &protoBackend.PackageReleaseDistinctReq_DistinctField{
			OsType:  distinctField.OSType,
			CpuArch: distinctField.CPUArch,
		},
	}

	if condition != nil && condition.ExactExclude != nil {
		exactIncludeConditions := &protoBackend.PackageReleaseExactConditions{
			Version:   condition.ExactExclude.Version,
			AsDefault: condition.ExactExclude.AsDefault,
			Enabled:   condition.ExactExclude.Enabled,
		}

		for _, item := range condition.ExactExclude.Platform {
			exactIncludeConditions.Platform = append(exactIncludeConditions.Platform, protoBackend.ConvertPlatformFromTypes(item))
		}

		req.ExactIncludeConditions = exactIncludeConditions
	}

	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, err
	}

	resp, err := h.cli.distinctRelease(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertResultToTypes(), nil
}

// SetReleaseLabels sets release labels.
func (h *Handler) SetReleaseLabels(
	ctx contextx.IContext,
	gen types.Generation,
	releaseType types.ReleaseType,
	plat platfmt.Platform,
	version string,
	labels []string,
) error {

	req := &protoBackend.PackageReleaseSetLabelsReq{Labels: labels}
	req.SetIdentifer(gen, releaseType, plat, version)

	return h.cli.setReleaseLabels(ctx, req)
}

// SetReleaseLabelsMany sets many release labels.
func (h *Handler) SetReleaseLabelsMany(
	ctx contextx.IContext,
	releaseType types.ReleaseType,
	gen []types.Generation,
	plat []platfmt.Platform,
	version []string,
	labels []string,
) error {

	req := &protoBackend.PackageReleaseSetLabelsManyReq{Labels: labels}
	req.SetIdentifers(releaseType, gen, plat, version)

	return h.cli.setReleaseLabelsMany(ctx, req)
}

// EnableRelease enables release active by generation, release type, platform and version.
func (h *Handler) EnableRelease(
	ctx contextx.IContext,
	gen types.Generation,
	releaseType types.ReleaseType,
	plat platfmt.Platform,
	version string,
) error {

	req := &protoBackend.PackageReleaseEnableReq{}
	req.SetIdentifer(gen, releaseType, plat, version)

	return h.cli.enableRelease(ctx, req)
}

// DisableRelease disables release disactive by generation, release type, platform and version.
func (h *Handler) DisableRelease(
	ctx contextx.IContext,
	gen types.Generation,
	releaseType types.ReleaseType,
	plat platfmt.Platform,
	version string,
) error {

	req := &protoBackend.PackageReleaseDisableReq{}
	req.SetIdentifer(gen, releaseType, plat, version)

	return h.cli.disableRelease(ctx, req)
}

// SetAsDefaultRelease sets the release as default.
func (h *Handler) SetAsDefaultRelease(
	ctx contextx.IContext,
	gen types.Generation,
	releaseType types.ReleaseType,
	plat platfmt.Platform,
	version string,
) error {

	req := &protoBackend.PackageReleaseSetAsDefaultReq{}
	req.SetIdentifer(gen, releaseType, plat, version)

	return h.cli.setAsDefaultRelease(ctx, req)
}

// CancelAsDefaultRelease cancels the release as default.
func (h *Handler) CancelAsDefaultRelease(
	ctx contextx.IContext,
	gen types.Generation,
	releaseType types.ReleaseType,
	plat platfmt.Platform,
	version string,
) error {

	req := &protoBackend.PackageReleaseCancelAsDefaultReq{}
	req.SetIdentifer(gen, releaseType, plat, version)

	return h.cli.cancelAsDefaultRelease(ctx, req)
}

// DeleteRelease deletes release by generation, release type, platform and version.
func (h *Handler) DeleteRelease(
	ctx contextx.IContext,
	gen types.Generation,
	releaseType types.ReleaseType,
	plat platfmt.Platform,
	version string,
) error {

	req := &protoBackend.PackageReleaseDeleteReq{}
	req.SetIdentifer(gen, releaseType, plat, version)

	return h.cli.deleteRelease(ctx, req)
}

// ListReleasePlugin lists release plugin by page and conditions.
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

// CountReleasePlugin counts release plugin by conditions.
func (h *Handler) CountReleasePlugin(
	ctx contextx.IContext,
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

	return resp.GetData().GetTotal(), nil
}

// EnableReleasePlugin enables release plugin active by name, generation, platform and version.
func (h *Handler) EnableReleasePlugin(
	ctx contextx.IContext,
	name string,
	gen types.Generation,
	plat platfmt.Platform,
	version string,
) error {

	req := &protoBackend.PackageReleasePluginEnableReq{}
	req.SetIdentifer(name, gen, plat, version)

	return h.cli.enableReleasePlugin(ctx, req)
}

// DisableReleasePlugin disables release plugin disactive by name, generation, platform and version.
func (h *Handler) DisableReleasePlugin(
	ctx contextx.IContext,
	name string,
	gen types.Generation,
	plat platfmt.Platform,
	version string,
) error {

	req := &protoBackend.PackageReleasePluginDisableReq{}
	req.SetIdentifer(name, gen, plat, version)

	return h.cli.disableReleasePlugin(ctx, req)
}

// SetAsDefaultReleasePlugin sets the release plugin as default.
func (h *Handler) SetAsDefaultReleasePlugin(
	ctx contextx.IContext,
	name string,
	gen types.Generation,
	plat platfmt.Platform,
	version string,
) error {

	req := &protoBackend.PackageReleasePluginSetAsDefaultReq{}
	req.SetIdentifer(name, gen, plat, version)

	return h.cli.setAsDefaultReleasePlugin(ctx, req)
}

// CancelAsDefaultReleasePlugin cancels the release plugin as default.
func (h *Handler) CancelAsDefaultReleasePlugin(
	ctx contextx.IContext,
	name string,
	gen types.Generation,
	plat platfmt.Platform,
	version string,
) error {

	req := &protoBackend.PackageReleasePluginCancelAsDefaultReq{}
	req.SetIdentifer(name, gen, plat, version)

	return h.cli.cancelAsDefaultReleasePlugin(ctx, req)
}

// DeleteReleasePlugin deletes release plugin by name, generation, platform and version.
func (h *Handler) DeleteReleasePlugin(
	ctx contextx.IContext,
	name string,
	gen types.Generation,
	plat platfmt.Platform,
	version string,
) error {

	req := &protoBackend.PackageReleasePluginDeleteReq{}
	req.SetIdentifer(name, gen, plat, version)

	return h.cli.deleteReleasePlugin(ctx, req)
}
