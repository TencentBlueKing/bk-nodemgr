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
	"bytes"
	"runtime/debug"
	"time"

	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	maxReleaseLimit = 1000
)

// ListRelease lists releases with page and conditions.
func (h *handler) ListRelease(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleaseListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	releaseType := types.ReleaseType(req.GetReleaseType())
	gen := types.Generation(req.GetGeneration())
	exactIncludeCond := req.ConvertExactIncludeConditionsToTypes()
	exactIncludeCond.Generation = append(exactIncludeCond.Generation, gen)
	cond := &types.ReleaseCondition{
		ExactInclude: exactIncludeCond,
	}

	// only count.
	if req.GetOnlyCount() {
		num, err := h.storage.CountRelease(rCtx, releaseType, cond)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list release. failed to count host")
			return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.PackageReleaseListResp)
		resp.ConvertReleasesFromTypes(num, nil)

		return resp.GetData(), nil
	}

	page := req.ConvertPageToTypes(maxReleaseLimit)
	hosts, num, err := h.storage.ListRelease(rCtx, releaseType, page, cond)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PackageReleaseListResp)
	resp.ConvertReleasesFromTypes(num, hosts)

	return resp.GetData(), nil
}

// DistinctRelease distinct releases.
func (h *handler) DistinctRelease(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleaseDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	releaseType := types.ReleaseType(req.GetReleaseType())
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

	result, err := h.storage.DistinctRelease(rCtx, releaseType, distinctField, cond)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct host. failed to distinct host fields")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PackageReleaseDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}

// SetReleaseLabels set release labels.
func (h *handler) SetReleaseLabels(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleaseSetLabelsReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set release labels, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.storage.SetReleaseLabels(rCtx, gen, rt, plat, version, req.GetLabels()); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "release-type", rt, "platform", plat, "version", version).
			Error("failed to set release labels")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	logger.G.Biz(rCtx).
		With("gen", gen, "release-type", rt, "platform", plat, "version", version).
		Info("set release labels")

	resp := new(protoBackend.PackageReleaseSetLabelsResp)

	return resp.GetData(), nil
}

// SetReleaseLabelsMany set release labels.
func (h *handler) SetReleaseLabelsMany(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleaseSetLabelsManyReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set many release labels, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	rt, gen, plat, version := req.GetIdentifiers()
	if err := h.storage.SetReleaseLabelsMany(rCtx, rt, gen, plat, version, req.GetLabels()); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "release-type", rt, "platform", plat, "version", version, "labels", req.GetLabels()).
			Error("failed to set many release labels")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	logger.G.Biz(rCtx).
		With("gen", gen, "release-type", rt, "platform", plat, "version", version, "labels", req.GetLabels()).
		Info("set many release labels")

	resp := new(protoBackend.PackageReleaseSetLabelsManyResp)

	return resp.GetData(), nil
}

// EnableRelease enable release.
func (h *handler) EnableRelease(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleaseEnableReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to enable release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.storage.EnableRelease(rCtx, gen, rt, plat, version); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "release-type", rt, "platform", plat, "version", version).
			Error("failed to enable release")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// record package events.
	go h.recordPackageEvent(rCtx, gen, version, plat, rt, types.PackageEventTypeEnable)

	logger.G.Biz(rCtx).
		With("gen", gen, "release-type", rt, "platform", plat, "version", version).
		Info("enabled release")

	resp := new(protoBackend.PackageReleaseEnableResp)

	return resp.GetData(), nil
}

// DisableRelease disable release.
func (h *handler) DisableRelease(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleaseDisableReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to disable release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.storage.DisableRelease(rCtx, gen, rt, plat, version); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "release-type", rt, "platform", plat, "version", version).
			Error("failed to disable release")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// record package events.
	go h.recordPackageEvent(rCtx, gen, version, plat, rt, types.PackageEventTypeDisable)

	logger.G.Biz(rCtx).
		With("gen", gen, "release-type", rt, "platform", plat, "version", version).
		Info("disable release")

	resp := new(protoBackend.PackageReleaseDisableResp)

	return resp.GetData(), nil
}

// SetAsDefaultRelease set default release.
func (h *handler) SetAsDefaultRelease(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleaseSetAsDefaultReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set default release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.storage.SetAsDefaultRelease(rCtx, gen, rt, plat, version); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "release-type", rt, "platform", plat, "version", version).
			Error("failed to set as default release")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// record package events.
	go h.recordPackageEvent(rCtx, gen, version, plat, rt, types.PackageEventTypeSetAsDefault)

	logger.G.Biz(rCtx).
		With("gen", gen, "release-type", rt, "platform", plat, "version", version).
		Info("set as default release")

	resp := new(protoBackend.PackageReleaseSetAsDefaultResp)

	return resp.GetData(), nil
}

// CancelAsDefaultRelease cancel default release.
func (h *handler) CancelAsDefaultRelease(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleaseCancelAsDefaultReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to cancel default release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.storage.CancelAsDefaultRelease(rCtx, gen, rt, plat, version); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "release-type", rt, "platform", plat, "version", version).
			Error("failed to cancel as default release")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// record package events.
	go h.recordPackageEvent(rCtx, gen, version, plat, rt, types.PackageEventTypeCancelAsDefault)

	logger.G.Biz(rCtx).
		With("gen", gen, "release-type", rt, "platform", plat, "version", version).
		Info("canceled as default release")

	resp := new(protoBackend.PackageReleaseCancelAsDefaultResp)

	return resp.GetData(), nil
}

func (h *handler) DeleteRelease(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleaseDeleteReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.storage.DeleteRelease(rCtx, gen, rt, plat, version); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "release-type", rt, "platform", plat, "version", version).
			Error("failed to delete release")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// record package events.
	go h.recordPackageEvent(rCtx, gen, version, plat, rt, types.PackageEventTypeDelete)

	logger.G.Biz(rCtx).
		With("gen", gen, "release-type", rt, "platform", plat, "version", version).
		Info("deleted release")

	resp := new(protoBackend.PackageReleaseDeleteResp)

	return resp.GetData(), nil
}

func (h *handler) recordPackageEvent(rCtx restserver.IContext,
	gen types.Generation, version string, plat platfmt.Platform, rt types.ReleaseType,
	eventType types.PackageEventType) {
	// recover panic.
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()

			// The first line of the stack trace is of the form "goroutine N [status]:",
			// but by the time the panic reaches Do the goroutine may no longer exist,
			// and its status will have changed. Trim out the misleading line.
			if line := bytes.IndexByte(stack[:], '\n'); line >= 0 { //nolint: gocritic
				stack = stack[line+1:]
			}

			logger.G.Sys().With("event-type", eventType, "recover", r, "stack", stack).Error("failed to record package event")
		}
	}()

	event := &types.PackageEvent{
		EventType:   eventType,
		ReleaseType: rt,
		Generation:  gen,
		Version:     version,
		OSType:      plat.OS,
		CPUArch:     plat.Arch,
		OperateTime: time.Now(),
		Operator:    rCtx.Data().GetLoginName(),
	}

	if err := h.storage.CreateManyPackageEvent(rCtx, event); err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			With("release-type", rt,
				"gen", gen,
				"event-type", eventType,
				"os-type", plat.OS,
				"cpu-arch", plat.Arch,
				"version", version).
			Warn("failed to record package event event, failed to create package event")
	}
}
