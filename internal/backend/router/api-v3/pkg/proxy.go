/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pkg

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	authProvider "github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth/provider"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ListReleaseProxy lists proxy releases with page and conditions.
func (h *handler) ListReleaseProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleaseProxyListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release proxy, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())
	exactIncludeCond := req.ConvertExactIncludeConditionsToTypes()
	exactIncludeCond.Generation = append(exactIncludeCond.Generation, gen)
	cond := &types.ReleaseCondition{
		ExactInclude: exactIncludeCond,
	}

	// only count.
	if req.GetOnlyCount() {
		num, err := h.daoReleaseProxy.CountReleaseProxy(rCtx, cond)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list release proxy. failed to count release proxy")
			return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.PackageReleaseProxyListResp)
		resp.ConvertReleasesFromTypes(num, nil)

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release proxy, invalid page info")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	hosts, num, err := h.daoReleaseProxy.ListReleaseProxy(rCtx, page, cond)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release proxy")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PackageReleaseProxyListResp)
	resp.ConvertReleasesFromTypes(num, hosts)

	return resp.GetData(), nil
}

// DistinctReleaseProxy distincts proxy releases.
func (h *handler) DistinctReleaseProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleaseProxyDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct proxy release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())
	exactIncludeCond := req.ConvertExactIncludeConditionsToTypes()
	exactIncludeCond.Generation = append(exactIncludeCond.Generation, gen)
	cond := &types.ReleaseCondition{
		ExactInclude: exactIncludeCond,
	}

	distinctField := types.ReleaseDistinctField{
		OSType:  req.GetDistinctField().GetOsType(),
		CPUArch: req.GetDistinctField().GetCpuArch(),
	}

	result, err := h.daoReleaseProxy.DistinctReleaseProxy(rCtx, distinctField, cond)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct proxy release. failed to distinct proxy release fields")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PackageReleaseProxyDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}

// SetReleaseProxyLabelsMany sets many proxy releases labels.
func (h *handler) SetReleaseProxyLabelsMany(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleaseProxySetLabelsManyReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set many proxy release labels, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// check permission.
	resources := authProvider.BuildPackageResources(string(types.ReleaseTypeProxy))
	if err := h.authorizer.Check(rCtx, auth.ActionPackageManage, resources); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set many proxy release labels, permission denied")
		return nil, err
	}

	gen := types.Generation(req.GetGeneration())
	exactIncludeCond := req.ConvertExactIncludeConditionsToTypes()
	exactIncludeCond.Generation = append(exactIncludeCond.Generation, gen)
	cond := &types.ReleaseCondition{
		ExactInclude: exactIncludeCond,
	}

	if err := h.daoReleaseProxy.SetReleaseProxyLabelsMany(rCtx, req.GetLabels(), cond); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("labels", req.GetLabels(), "condition", *exactIncludeCond).
			Error("failed to set many proxy release labels")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	logger.G.Biz(rCtx).
		With("labels", req.GetLabels(), "condition", *exactIncludeCond).
		Info("set many proxy release labels")

	resp := new(protoBackend.PackageReleaseProxySetLabelsManyResp)

	return resp.GetData(), nil
}

// EnableReleaseProxy enables proxy release.
func (h *handler) EnableReleaseProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleaseProxyEnableReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to enable proxy release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// check permission.
	resources := authProvider.BuildPackageResources(string(types.ReleaseTypeProxy))
	if err := h.authorizer.Check(rCtx, auth.ActionPackageManage, resources); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to enable proxy release, permission denied")
		return nil, err
	}

	gen := types.Generation(req.GetGeneration())
	plat := protoBackend.ConvertPlatformToTypes(req.GetPlatform())
	version := req.GetVersion()
	key := types.ReleaseProxyKey{
		Generation: gen,
		Platform:   plat,
		Version:    version,
	}

	if err := h.daoReleaseProxy.EnableReleaseProxy(rCtx, key); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "platform", plat, "version", version).
			Error("failed to enable proxy release")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// record package events.
	h.recordProxyEvent(rCtx, gen, version, plat, types.PackageEventTypeEnable)

	logger.G.Biz(rCtx).
		With("gen", gen, "platform", plat, "version", version).
		Info("enabled proxy release")

	resp := new(protoBackend.PackageReleaseProxyEnableResp)

	return resp.GetData(), nil
}

// DisableReleaseProxy disables proxy release.
func (h *handler) DisableReleaseProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleaseProxyDisableReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to disable proxy release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// check permission.
	resources := authProvider.BuildPackageResources(string(types.ReleaseTypeProxy))
	if err := h.authorizer.Check(rCtx, auth.ActionPackageManage, resources); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to disable proxy release, permission denied")
		return nil, err
	}

	gen := types.Generation(req.GetGeneration())
	plat := protoBackend.ConvertPlatformToTypes(req.GetPlatform())
	version := req.GetVersion()
	key := types.ReleaseProxyKey{
		Generation: gen,
		Platform:   plat,
		Version:    version,
	}

	if err := h.daoReleaseProxy.DisableReleaseProxy(rCtx, key); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "platform", plat, "version", version).
			Error("failed to disable proxy release")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// record package events.
	h.recordProxyEvent(rCtx, gen, version, plat, types.PackageEventTypeDisable)

	logger.G.Biz(rCtx).
		With("gen", gen, "platform", plat, "version", version).
		Info("disabled proxy release")

	resp := new(protoBackend.PackageReleaseProxyDisableResp)

	return resp.GetData(), nil
}

// SetAsDefaultReleaseProxy sets proxy release as default.
func (h *handler) SetAsDefaultReleaseProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleaseProxySetAsDefaultReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set default proxy release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// check permission.
	resources := authProvider.BuildPackageResources(string(types.ReleaseTypeProxy))
	if err := h.authorizer.Check(rCtx, auth.ActionPackageManage, resources); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set default proxy release, permission denied")
		return nil, err
	}

	gen := types.Generation(req.GetGeneration())
	plat := protoBackend.ConvertPlatformToTypes(req.GetPlatform())
	version := req.GetVersion()
	key := types.ReleaseProxyKey{
		Generation: gen,
		Platform:   plat,
		Version:    version,
	}

	if err := h.daoReleaseProxy.SetAsDefaultReleaseProxy(rCtx, key); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "platform", plat, "version", version).
			Error("failed to set as default proxy release")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// record package events.
	h.recordProxyEvent(rCtx, gen, version, plat, types.PackageEventTypeSetAsDefault)

	logger.G.Biz(rCtx).
		With("gen", gen, "platform", plat, "version", version).
		Info("set as default proxy release")

	resp := new(protoBackend.PackageReleaseProxySetAsDefaultResp)

	return resp.GetData(), nil
}

// CancelAsDefaultReleaseProxy cancels proxy release as default.
func (h *handler) CancelAsDefaultReleaseProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleaseProxyCancelAsDefaultReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to cancel default proxy release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// check permission.
	resources := authProvider.BuildPackageResources(string(types.ReleaseTypeProxy))
	if err := h.authorizer.Check(rCtx, auth.ActionPackageManage, resources); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to cancel default proxy release, permission denied")
		return nil, err
	}

	gen := types.Generation(req.GetGeneration())
	plat := protoBackend.ConvertPlatformToTypes(req.GetPlatform())
	version := req.GetVersion()
	key := types.ReleaseProxyKey{
		Generation: gen,
		Platform:   plat,
		Version:    version,
	}

	if err := h.daoReleaseProxy.CancelAsDefaultReleaseProxy(rCtx, key); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "platform", plat, "version", version).
			Error("failed to cancel as default proxy release")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// record package events.
	h.recordProxyEvent(rCtx, gen, version, plat, types.PackageEventTypeCancelAsDefault)

	logger.G.Biz(rCtx).
		With("gen", gen, "platform", plat, "version", version).
		Info("canceled as default proxy release")

	resp := new(protoBackend.PackageReleaseProxyCancelAsDefaultResp)

	return resp.GetData(), nil
}

// DeleteReleaseProxy deletes proxy release.
func (h *handler) DeleteReleaseProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleaseProxyDeleteReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete proxy release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// check permission.
	resources := authProvider.BuildPackageResources(string(types.ReleaseTypeProxy))
	if err := h.authorizer.Check(rCtx, auth.ActionPackageManage, resources); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete proxy release, permission denied")
		return nil, err
	}

	gen := types.Generation(req.GetGeneration())
	plat := protoBackend.ConvertPlatformToTypes(req.GetPlatform())
	version := req.GetVersion()
	key := types.ReleaseProxyKey{
		Generation: gen,
		Platform:   plat,
		Version:    version,
	}

	if err := h.daoReleaseProxy.DeleteReleaseProxy(rCtx, key); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "platform", plat, "version", version).
			Error("failed to delete proxy release")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// record package events.
	h.recordProxyEvent(rCtx, gen, version, plat, types.PackageEventTypeDelete)

	logger.G.Biz(rCtx).
		With("gen", gen, "platform", plat, "version", version).
		Info("deleted proxy release")

	resp := new(protoBackend.PackageReleaseProxyDeleteResp)

	return resp.GetData(), nil
}

func (h *handler) recordProxyEvent(rCtx restserver.IContext, gen types.Generation, version string, plat platfmt.Platform,
	eventType types.PackageEventType) {

	event := &types.PackageEvent{
		Name:        types.ReleaseNameProxy,
		ReleaseType: types.ReleaseTypeProxy,
		Generation:  gen,
		OSType:      plat.OS,
		CPUArch:     plat.Arch,
		Version:     version,
		EventType:   eventType,
		Operator:    rCtx.Data().GetLoginName(),
		OperateTime: time.Now(),
	}

	h.recordPackageEvents(rCtx, event)
}
