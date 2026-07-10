/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package backendadmin

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

type syncUnassignedNetworkUnitReq struct {
	BKBizID []int64 `json:"bk_biz_id"`
}

type syncUnassignedNetworkUnitResp struct {
	Code      int32                          `json:"code"`
	Message   string                         `json:"message"`
	RequestID string                         `json:"request_id"`
	Data      *syncUnassignedNetworkUnitData `json:"data"`
}

type syncUnassignedNetworkUnitData struct {
	SuccessCount  int64    `json:"success_count"`
	FailedCount   int64    `json:"failed_count"`
	FailedReasons []string `json:"failed_reasons"`
}

func (c *cli) syncUnassignedAgentNetworkUnit(
	nCtx contextx.IContext,
	bizIDs []int64,
) (*types.NodeAgentAssignUnitResult, error) {

	if c.syncUnassignedAgentNetworkUnitFn != nil {
		return c.syncUnassignedAgentNetworkUnitFn(nCtx, bizIDs)
	}

	header, err := c.getCommonHeader(nCtx)
	if err != nil {
		return nil, err
	}

	resp := new(syncUnassignedNetworkUnitResp)
	err = c.client.Post().
		SubResourcef("/node/agent/sync_unassigned_network_unit").
		WithContext(nCtx).
		WithHeaders(header).
		Body(&syncUnassignedNetworkUnitReq{BKBizID: bizIDs}).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("sync unassigned agent network unit failed, code(%d), message(%s), request-id(%s)",
			resp.Code, resp.Message, resp.RequestID)
	}

	if resp.Data == nil {
		return nil, fmt.Errorf("sync unassigned agent network unit failed, empty response data, request-id(%s)", resp.RequestID)
	}

	return &types.NodeAgentAssignUnitResult{
		SuccessCount:  resp.Data.SuccessCount,
		FailedCount:   resp.Data.FailedCount,
		FailedReasons: resp.Data.FailedReasons,
	}, nil
}
