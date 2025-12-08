/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cmdb

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/combine"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
)

const (
	// combinedMax defines the max items in one combined task.
	combinedMax = 100

	// combinedGap defines the max gap between two combined tasks.
	combinedGap = 1 * time.Second
)

func (h *Handler) getCombinedHandler(nCtx contextx.IContext) *combinedHandler {
	tenantID := nCtx.TenantID()

	h.combinedHandlerGroupMu.RLock()
	handler, ok := h.combinedHandlerGroup[tenantID]
	h.combinedHandlerGroupMu.RUnlock()

	if ok {
		return handler
	}

	h.combinedHandlerGroupMu.Lock()
	defer h.combinedHandlerGroupMu.Unlock()

	handler, ok = h.combinedHandlerGroup[tenantID]
	if ok {
		return handler
	}

	// create new handler.
	handler = &combinedHandler{
		tenantID: tenantID,
		username: h.cli.config.VirtualUser,
		cli:      h.cli,
	}
	handler.init()

	h.combinedHandlerGroup[tenantID] = handler

	return handler
}

type combinedHandler struct {
	cli      *cli
	tenantID string
	username string

	bindHostAgentCombinedHandler                combine.IHandler[*HostAgentIDInfo, interface{}]
	unbindHostAgentCombinedHandler              combine.IHandler[*HostAgentIDInfo, interface{}]
	pushHostIdentifierCombinedHandler           combine.IHandler[int64, *PushHostIdentifierResp]
	findHostIdentifierPushResultCombinedHandler combine.IHandler[struct{}, *FindHostIdentifierPushResultResp]
	addHostToBusinessIdleCombinedHandler        combine.IHandler[*CreateHostInfo, *AddHostToBusinessIdleResp]
	addHostToResourcePoolCombinedHandler        combine.IHandler[*CreateHostInfo, *AddHostToResourcePoolResp]
}

func (hdl *combinedHandler) init() {
	hdl.registerAddHostToBusinessIdleCombinedHandler()
	hdl.registerBindHostAgentCombinedHandler()
	hdl.registerUnbindHostAgentCombinedHandler()
	hdl.registerPushHostIdentifierCombinedHandler()
	hdl.registerFindHostIdentifierPushResultCombinedHandler()
	hdl.registerAddHostToResourcePoolCombinedHandler()
}

func (hdl *combinedHandler) registerBindHostAgentCombinedHandler() {
	hdl.bindHostAgentCombinedHandler = combine.New[*HostAgentIDInfo, interface{}](
		combinedMax,
		combinedGap,
		func(_ string, data []*HostAgentIDInfo) (interface{}, error) {
			reqList := make([]*HostAgentIDInfo, 0, len(data))
			for _, item := range data {
				found := false
				for _, req := range reqList {
					if req.BKHostID == item.BKHostID {
						req.BKAgentID = item.BKAgentID
						found = true

						break
					}
				}
				if found {
					continue
				}

				reqList = append(reqList, item)
			}

			req := &BindHostAgentReq{
				List: reqList,
			}

			nCtx := contextx.From(contextx.Background(), contextx.WithTenantID(hdl.tenantID), contextx.WithBKUsername(hdl.username))
			logger.G.Biz(nCtx).With("req", req).Info("use virtual user to bind host agent")

			return nil, hdl.cli.bindHostAgent(nCtx, req)
		},
	)
}

func (hdl *combinedHandler) registerUnbindHostAgentCombinedHandler() {
	hdl.unbindHostAgentCombinedHandler = combine.New[*HostAgentIDInfo, interface{}](
		combinedMax,
		combinedGap,
		func(_ string, data []*HostAgentIDInfo) (interface{}, error) {
			reqList := make([]*HostAgentIDInfo, 0, len(data))
			for _, item := range data {
				found := false
				for _, req := range reqList {
					if req.BKHostID == item.BKHostID {
						req.BKAgentID = item.BKAgentID
						found = true

						break
					}
				}
				if found {
					continue
				}

				reqList = append(reqList, item)
			}

			req := &UnbindHostAgentReq{
				List: reqList,
			}

			nCtx := contextx.From(contextx.Background(), contextx.WithTenantID(hdl.tenantID), contextx.WithBKUsername(hdl.username))
			logger.G.Biz(nCtx).With("req", req).Info("use virtual user to unbind host agent")

			return nil, hdl.cli.unbindHostAgent(nCtx, req)
		},
	)
}

func (hdl *combinedHandler) registerPushHostIdentifierCombinedHandler() {
	hdl.pushHostIdentifierCombinedHandler = combine.New[int64, *PushHostIdentifierResp](
		combinedMax,
		combinedGap,
		func(_ string, data []int64) (*PushHostIdentifierResp, error) {
			reqMap := make(map[int64]struct{})
			for _, hostID := range data {
				reqMap[hostID] = struct{}{}
			}

			req := &PushHostIdentifierReq{
				BKHostIDs: conv.MapKeyToSlice(reqMap),
			}

			nCtx := contextx.From(contextx.Background(), contextx.WithTenantID(hdl.tenantID), contextx.WithBKUsername(hdl.username))
			logger.G.Biz(nCtx).With("req", req).Info("use virtual user to push host identifier")

			return hdl.cli.pushHostIdentifier(nCtx, req)
		},
	)
}

func (hdl *combinedHandler) registerFindHostIdentifierPushResultCombinedHandler() {
	hdl.findHostIdentifierPushResultCombinedHandler = combine.New[struct{}, *FindHostIdentifierPushResultResp](
		combinedMax,
		combinedGap,
		func(taskID string, _ []struct{}) (*FindHostIdentifierPushResultResp, error) {
			req := &FindHostIdentifierPushResultReq{
				TaskID: taskID,
			}

			nCtx := contextx.From(contextx.Background(), contextx.WithTenantID(hdl.tenantID), contextx.WithBKUsername(hdl.username))
			logger.G.Biz(nCtx).With("req", req).Info("use virtual user to find host identifier push result")

			return hdl.cli.findHostIdentifierPushResult(nCtx, req)
		},
	)
}

func (hdl *combinedHandler) registerAddHostToBusinessIdleCombinedHandler() {
	hdl.addHostToBusinessIdleCombinedHandler = combine.New[*CreateHostInfo, *AddHostToBusinessIdleResp](
		combinedMax,
		combinedGap,
		func(bizIDKey string, data []*CreateHostInfo) (*AddHostToBusinessIdleResp, error) {
			// convert bizIDKey to int64.
			bizID, err := conv.ToInt64(bizIDKey)
			if err != nil {
				return nil, fmt.Errorf("invalid biz-id. key(%s)", bizIDKey)
			}

			req := &AddHostToBusinessIdleReq{
				BKBizID:    bizID,
				BKHostList: data,
			}

			nCtx := contextx.From(contextx.Background(), contextx.WithTenantID(hdl.tenantID), contextx.WithBKUsername(hdl.username))
			logger.G.Biz(nCtx).With("req", req).Info("use virtual user to add host to business idle")

			return hdl.cli.addHostToBusinessIdle(nCtx, req)
		},
	)
}

func (hdl *combinedHandler) registerAddHostToResourcePoolCombinedHandler() {
	hdl.addHostToResourcePoolCombinedHandler = combine.New[*CreateHostInfo, *AddHostToResourcePoolResp](
		combinedMax,
		combinedGap,
		func(_ string, data []*CreateHostInfo) (*AddHostToResourcePoolResp, error) {
			req := &AddHostToResourcePoolReq{
				HostInfo: data,
			}

			nCtx := contextx.From(contextx.Background(), contextx.WithTenantID(hdl.tenantID), contextx.WithBKUsername(hdl.username))
			logger.G.Biz(nCtx).With("req", req).Info("use virtual user to add host to resource pool")

			return hdl.cli.addHostToResource(nCtx, req)
		},
	)
}
