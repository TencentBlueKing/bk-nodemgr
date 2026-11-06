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

package pkg

import (
	"errors"
	"math"
	"sort"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ListPackageEvent lists events with page and conditions.
func (h *handler) ListPackageEvent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageEventListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list event, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	conditions, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list event, failed to convert conditions")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	fixedTypes, pluginNames, scopeIsAny, authErr := h.narrowAuthorizedPackageNamesForHistory(rCtx)
	if authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to list event, permission denied")

		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}
	fixedCondition, pluginCondition, hasAuthorizedPackages := narrowPackageEventConditions(
		conditions, fixedTypes, pluginNames, scopeIsAny,
	)

	// only count.
	if req.GetOnlyCount() {
		if !hasAuthorizedPackages {
			resp := new(protoBackend.PackageEventListResp)
			resp.ConvertPackageEventsFromTypes(0, nil)

			return resp.GetData(), nil
		}

		num, err := h.countAuthorizedPackageEvents(rCtx, fixedCondition, pluginCondition)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list event, failed to count event")

			return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.PackageEventListResp)
		resp.ConvertPackageEventsFromTypes(num, nil)

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list event, invalid page info")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if !hasAuthorizedPackages {
		resp := new(protoBackend.PackageEventListResp)
		resp.ConvertPackageEventsFromTypes(0, nil)

		return resp.GetData(), nil
	}

	events, num, err := h.listAuthorizedPackageEvents(rCtx, page, fixedCondition, pluginCondition)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list event")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PackageEventListResp)
	resp.ConvertPackageEventsFromTypes(num, events)

	return resp.GetData(), nil
}

func (h *handler) countAuthorizedPackageEvents(
	rCtx restserver.IContext, fixedCondition, pluginCondition *types.PackageEventCondition,
) (
	int64,
	error,
) {

	var total int64
	for _, condition := range []*types.PackageEventCondition{fixedCondition, pluginCondition} {
		if condition == nil {
			continue
		}
		count, err := h.fileHandler.CountPackageEvent(rCtx, condition)
		if err != nil {
			return 0, err
		}
		total += count
	}

	return total, nil
}

func (h *handler) listAuthorizedPackageEvents(
	rCtx restserver.IContext, page types.Page, fixedCondition, pluginCondition *types.PackageEventCondition,
) (
	[]*types.PackageEvent,
	int64,
	error,
) {

	if fixedCondition == nil {
		return h.fileHandler.ListPackageEvent(rCtx, page, pluginCondition)
	}
	if pluginCondition == nil {
		return h.fileHandler.ListPackageEvent(rCtx, page, fixedCondition)
	}
	if page.Offset > math.MaxInt-page.Limit {
		return nil, 0, errors.New("package event page offset plus limit exceeds maximum value")
	}
	queryPage := page
	queryPage.Offset = 0
	queryPage.Limit += page.Offset

	var events []*types.PackageEvent
	var total int64
	for _, condition := range []*types.PackageEventCondition{fixedCondition, pluginCondition} {
		if condition == nil {
			continue
		}
		items, count, err := h.fileHandler.ListPackageEvent(rCtx, queryPage, condition)
		if err != nil {
			return nil, 0, err
		}
		events = append(events, items...)
		total += count
	}

	// Preserve source order for equal timestamps; the File API does not expose event IDs.
	sort.SliceStable(events, func(i, j int) bool {
		return events[i].OperateTime.After(events[j].OperateTime)
	})
	if page.Offset >= len(events) {
		return nil, total, nil
	}
	end := page.Offset + page.Limit
	if end > len(events) {
		end = len(events)
	}

	return events[page.Offset:end], total, nil
}

// DistinctPackageEvent distincts events with conditions.
// nolint: dupl
func (h *handler) DistinctPackageEvent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageEventDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct topoevent, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	conditions, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct topoevent, failed to convert conditions")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.fileHandler.DistinctPackageEvent(
		rCtx,
		types.NewPackageEventDistinctRequestAllSet(),
		conditions)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct topoevent, failed to distinct host fields")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PackageEventDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}
