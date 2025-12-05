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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/access"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/combine"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
)

const (
	combinedMax = 100
	combinedGap = 1 * time.Second
)

func (h *Handler) registerBindHostAgentCombinedHandler() {
	h.bindHostAgentCombinedHandler = combine.New(combinedMax, combinedGap, func(_ string, data []interface{}) (interface{}, error) {
		reqList := make([]*HostAgentIDInfo, len(data))
		for idx := range data {
			item, ok := data[idx].(*HostAgentIDInfo)
			if !ok {
				return nil, fmt.Errorf("invalid data: %v", data[idx])
			}

			found := false
			for preIdx := 0; preIdx < idx; preIdx++ {
				if reqList[preIdx].BKHostID == item.BKHostID {
					reqList[preIdx].BKAgentID = item.BKAgentID
					found = true

					break
				}
			}
			if found {
				continue
			}

			reqList[idx] = item
		}

		virtualUser := access.GetVirtualUser()
		newCtx := contextx.From(contextx.Background(), contextx.WithBKUsername(virtualUser))

		req := &BindHostAgentReq{
			List: reqList,
		}

		logger.G.Biz(newCtx).With("req", req).Info("use virtual user to bind host agent")

		return nil, h.cli.bindHostAgent(newCtx, req)
	})
}

func (h *Handler) registerUnbindHostAgentCombinedHandler() {
	h.unbindHostAgentCombinedHandler = combine.New(combinedMax, combinedGap, func(_ string, data []interface{}) (interface{}, error) {
		reqList := make([]*HostAgentIDInfo, len(data))
		for idx := range data {
			item, ok := data[idx].(*HostAgentIDInfo)
			if !ok {
				return nil, fmt.Errorf("invalid data: %v", data[idx])
			}

			found := false
			for preIdx := 0; preIdx < idx; preIdx++ {
				if reqList[preIdx].BKHostID == item.BKHostID {
					reqList[preIdx].BKAgentID = item.BKAgentID
					found = true

					break
				}
			}
			if found {
				continue
			}

			reqList[idx] = item
		}

		virtualUser := access.GetVirtualUser()
		newCtx := contextx.From(contextx.Background(), contextx.WithBKUsername(virtualUser))

		req := &UnbindHostAgentReq{
			List: reqList,
		}

		logger.G.Biz(newCtx).With("req", req).Info("use virtual user to unbind host agent")

		return nil, h.cli.unbindHostAgent(newCtx, req)
	})
}

func (h *Handler) registerPushHostIdentifierCombinedHandler() {
	h.pushHostIdentifierCombinedHandler = combine.New(combinedMax, combinedGap, func(_ string, data []interface{}) (interface{}, error) {
		reqMap := make(map[int64]struct{})
		for idx := range data {
			hostID, ok := data[idx].(int64)
			if !ok {
				return nil, fmt.Errorf("invalid data: %v", data[idx])
			}

			reqMap[hostID] = struct{}{}
		}

		virtualUser := access.GetVirtualUser()
		newCtx := contextx.From(contextx.Background(), contextx.WithBKUsername(virtualUser))

		req := &PushHostIdentifierReq{
			BKHostIDs: conv.MapKeyToSlice(reqMap),
		}

		logger.G.Biz(newCtx).With("req", req).Info("use virtual user to push host identifier")

		return h.cli.pushHostIdentifier(newCtx, req)
	})
}

func (h *Handler) registerFindHostIdentifierPushResultCombinedHandler() {
	h.findHostIdentifierPushResultCombinedHandler = combine.New(combinedMax, combinedGap, func(taskID string, _ []interface{}) (interface{}, error) {
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
	})
}
