/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package gse

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

func (c *cli) getProcOperateResultV2(nCtx contextx.IContext, req *getProcOperateResultV2Req) (getProcOperateResultV2Resp, error) {
	resp := new(BaseBroker[getProcOperateResultV2Resp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/proc/get_proc_operate_result_v2").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get proc operate result v2: %w", err)
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("failed to get proc operate result v2: %w", err)
	}

	return resp.Data, nil
}

func (c *cli) operateProcV2(nCtx contextx.IContext, req *operateProcV2Req) (*operateProcV2Resp, error) {
	resp := new(BaseBroker[*operateProcV2Resp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/proc/operate_proc_v2").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to operate proc v2: %w", err)
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("failed to operate proc v2: %w", err)
	}

	return resp.Data, nil
}

// operateProcMulti operates multiple processes.
func (c *cli) operateProcMulti(nCtx contextx.IContext, req *operateProcMultiReq) (*operateProcMultiResp, error) {
	resp := new(BaseBroker[*operateProcMultiResp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/proc/operate_proc_multi").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to operate proc multi: %w", err)
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("failed to operate proc multi: %w", err)
	}

	return resp.Data, nil
}
