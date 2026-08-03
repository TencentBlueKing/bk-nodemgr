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
)

type initTenantReq struct {
	TenantID string `json:"tenant_id"`
}

type initTenantResp struct {
	Code      int32  `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func (c *cli) initTenant(nCtx contextx.IContext, tenantID string) error {
	if c.initTenantFn != nil {
		return c.initTenantFn(nCtx, tenantID)
	}

	header, err := c.getCommonHeader(nCtx)
	if err != nil {
		return err
	}

	resp := new(initTenantResp)
	err = c.client.Post().
		SubResourcef("/tenant/init").
		WithContext(nCtx).
		WithHeaders(header).
		Body(&initTenantReq{TenantID: tenantID}).
		Do().Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return fmt.Errorf("init tenant failed, code(%d), message(%s), request-id(%s)",
			resp.Code, resp.Message, resp.RequestID)
	}

	return nil
}
