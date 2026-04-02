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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/goasync"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ListReleaseAgent lists agent releases with page and conditions.
func (h *handler) ListReleaseAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleaseAgentListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release agent, failed to decode request body")
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
		num, err := h.daoReleaseAgent.CountReleaseAgent(rCtx, cond)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list release agent. failed to count release agent")
			return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.PackageReleaseAgentListResp)
		resp.ConvertReleasesFromTypes(num, nil)

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release agent, invalid page info")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	hosts, num, err := h.daoReleaseAgent.ListReleaseAgent(rCtx, page, cond)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release agent")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PackageReleaseAgentListResp)
	resp.ConvertReleasesFromTypes(num, hosts)

	return resp.GetData(), nil
}

// DistinctReleaseAgent distinct agent releases.
func (h *handler) DistinctReleaseAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleaseAgentDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct agent release, failed to decode request body")
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

	result, err := h.daoReleaseAgent.DistinctReleaseAgent(rCtx, distinctField, cond)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct agent release. failed to distinct agent release fields")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PackageReleaseAgentDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}

// SetReleaseAgentLabelsMany sets many agent releases labels.
func (h *handler) SetReleaseAgentLabelsMany(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleaseAgentSetLabelsManyReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set many agent release labels, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())
	exactIncludeCond := req.ConvertExactIncludeConditionsToTypes()
	exactIncludeCond.Generation = append(exactIncludeCond.Generation, gen)
	cond := &types.ReleaseCondition{
		ExactInclude: exactIncludeCond,
	}

	if err := h.daoReleaseAgent.SetReleaseAgentLabelsMany(rCtx, req.GetLabels(), cond); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("labels", req.GetLabels(), "condition", *exactIncludeCond).
			Error("failed to set many agent release labels")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	logger.G.Biz(rCtx).
		With("labels", req.GetLabels(), "condition", *exactIncludeCond).
		Info("set many agent release labels")

	resp := new(protoBackend.PackageReleaseAgentSetLabelsManyResp)

	return resp.GetData(), nil
}

// EnableReleaseAgent enables agent release.
func (h *handler) EnableReleaseAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleaseAgentEnableReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to enable agent release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())
	plat := protoBackend.ConvertPlatformToTypes(req.GetPlatform())
	version := req.GetVersion()
	key := types.ReleaseAgentKey{
		Generation: gen,
		Platform:   plat,
		Version:    version,
	}

	if err := h.daoReleaseAgent.EnableReleaseAgent(rCtx, key); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "platform", plat, "version", version).
			Error("failed to enable agent release")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// record package events.
	h.recordAgentEvent(rCtx, gen, version, plat, types.PackageEventTypeEnable)

	logger.G.Biz(rCtx).
		With("gen", gen, "platform", plat, "version", version).
		Info("enabled agent release")

	resp := new(protoBackend.PackageReleaseAgentEnableResp)

	return resp.GetData(), nil
}

// DisableReleaseAgent disables agent release.
func (h *handler) DisableReleaseAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleaseAgentDisableReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to disable agent release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())
	plat := protoBackend.ConvertPlatformToTypes(req.GetPlatform())
	version := req.GetVersion()
	key := types.ReleaseAgentKey{
		Generation: gen,
		Platform:   plat,
		Version:    version,
	}

	if err := h.daoReleaseAgent.DisableReleaseAgent(rCtx, key); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "platform", plat, "version", version).
			Error("failed to disable agent release")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// record package events.
	h.recordAgentEvent(rCtx, gen, version, plat, types.PackageEventTypeDisable)

	logger.G.Biz(rCtx).
		With("gen", gen, "platform", plat, "version", version).
		Info("disabled agent release")

	resp := new(protoBackend.PackageReleaseAgentDisableResp)

	return resp.GetData(), nil
}

// SetAsDefaultReleaseAgent sets agent release as default.
func (h *handler) SetAsDefaultReleaseAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleaseAgentSetAsDefaultReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set default agent release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())
	plat := protoBackend.ConvertPlatformToTypes(req.GetPlatform())
	version := req.GetVersion()
	key := types.ReleaseAgentKey{
		Generation: gen,
		Platform:   plat,
		Version:    version,
	}

	if err := h.daoReleaseAgent.SetAsDefaultReleaseAgent(rCtx, key); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "platform", plat, "version", version).
			Error("failed to set as default agent release")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// record package events.
	h.recordAgentEvent(rCtx, gen, version, plat, types.PackageEventTypeSetAsDefault)

	logger.G.Biz(rCtx).
		With("gen", gen, "platform", plat, "version", version).
		Info("set as default agent release")

	resp := new(protoBackend.PackageReleaseAgentSetAsDefaultResp)

	return resp.GetData(), nil
}

// CancelAsDefaultReleaseAgent cancels agent release as default.
func (h *handler) CancelAsDefaultReleaseAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleaseAgentCancelAsDefaultReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to cancel default agent release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())
	plat := protoBackend.ConvertPlatformToTypes(req.GetPlatform())
	version := req.GetVersion()
	key := types.ReleaseAgentKey{
		Generation: gen,
		Platform:   plat,
		Version:    version,
	}

	if err := h.daoReleaseAgent.CancelAsDefaultReleaseAgent(rCtx, key); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "platform", plat, "version", version).
			Error("failed to cancel as default agent release")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// record package events.
	h.recordAgentEvent(rCtx, gen, version, plat, types.PackageEventTypeCancelAsDefault)

	logger.G.Biz(rCtx).
		With("gen", gen, "platform", plat, "version", version).
		Info("canceled as default agent release")

	resp := new(protoBackend.PackageReleaseAgentCancelAsDefaultResp)

	return resp.GetData(), nil
}

// DeleteReleaseAgent deletes agent release.
func (h *handler) DeleteReleaseAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleaseAgentDeleteReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete agent release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())
	plat := protoBackend.ConvertPlatformToTypes(req.GetPlatform())
	version := req.GetVersion()
	key := types.ReleaseAgentKey{
		Generation: gen,
		Platform:   plat,
		Version:    version,
	}

	if err := h.daoReleaseAgent.DeleteReleaseAgent(rCtx, key); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "platform", plat, "version", version).
			Error("failed to delete agent release")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// record package events.
	h.recordAgentEvent(rCtx, gen, version, plat, types.PackageEventTypeDelete)

	logger.G.Biz(rCtx).
		With("gen", gen, "platform", plat, "version", version).
		Info("deleted agent release")

	resp := new(protoBackend.PackageReleaseAgentDeleteResp)

	return resp.GetData(), nil
}

func (h *handler) recordAgentEvent(rCtx restserver.IContext, gen types.Generation, version string, plat platfmt.Platform,
	eventType types.PackageEventType) {

	event := &types.PackageEvent{
		Name:        types.ReleaseNameAgent,
		ReleaseType: types.ReleaseTypeAgent,
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

func (h *handler) recordPackageEvents(rCtx restserver.IContext, events ...*types.PackageEvent) {
	err := h.goAsyncPool.Run(
		rCtx,
		func(nCtx contextx.IContext) error {
			return h.daoPackageEvent.CreateManyPackageEvent(nCtx, events...)
		},
		goasync.WithName("record_package_event"),
	)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to submit package event recording task")
	}
}
