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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/access"
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

func (h *Handler) registerBindHostAgentCombinedHandler() {
	h.bindHostAgentCombinedHandler = combine.New[*HostAgentIDInfo, interface{}](
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

			virtualUser := access.GetVirtualUser()
			newCtx := contextx.From(contextx.Background(), contextx.WithBKUsername(virtualUser))

			req := &BindHostAgentReq{
				List: reqList,
			}

			logger.G.Biz(newCtx).With("req", req).Info("use virtual user to bind host agent")

			return nil, h.cli.bindHostAgent(newCtx, req)
		},
	)
}

func (h *Handler) registerUnbindHostAgentCombinedHandler() {
	h.unbindHostAgentCombinedHandler = combine.New[*HostAgentIDInfo, interface{}](
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

			virtualUser := access.GetVirtualUser()
			newCtx := contextx.From(contextx.Background(), contextx.WithBKUsername(virtualUser))

			req := &UnbindHostAgentReq{
				List: reqList,
			}

			logger.G.Biz(newCtx).With("req", req).Info("use virtual user to unbind host agent")

			return nil, h.cli.unbindHostAgent(newCtx, req)
		},
	)
}

func (h *Handler) registerPushHostIdentifierCombinedHandler() {
	h.pushHostIdentifierCombinedHandler = combine.New[int64, *PushHostIdentifierResp](
		combinedMax,
		combinedGap,
		func(_ string, data []int64) (*PushHostIdentifierResp, error) {
			reqMap := make(map[int64]struct{})
			for _, hostID := range data {
				reqMap[hostID] = struct{}{}
			}

			virtualUser := access.GetVirtualUser()
			newCtx := contextx.From(contextx.Background(), contextx.WithBKUsername(virtualUser))

			req := &PushHostIdentifierReq{
				BKHostIDs: conv.MapKeyToSlice(reqMap),
			}

			logger.G.Biz(newCtx).With("req", req).Info("use virtual user to push host identifier")

			return h.cli.pushHostIdentifier(newCtx, req)
		},
	)
}

func (h *Handler) registerFindHostIdentifierPushResultCombinedHandler() {
	h.findHostIdentifierPushResultCombinedHandler = combine.New[struct{}, *FindHostIdentifierPushResultResp](
		combinedMax,
		combinedGap,
		func(taskID string, _ []struct{}) (*FindHostIdentifierPushResultResp, error) {
			req := &FindHostIdentifierPushResultReq{
				TaskID: taskID,
			}

			virtualUser := access.GetVirtualUser()
			newCtx := contextx.From(contextx.Background(), contextx.WithBKUsername(virtualUser))

			logger.G.Biz(newCtx).With("req", req).Info("use virtual user to find host identifier push result")

			resp, err := h.cli.findHostIdentifierPushResult(newCtx, req)
			if err != nil {
				return nil, err
			}

			return resp, nil
		},
	)
}
