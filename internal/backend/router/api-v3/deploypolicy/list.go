/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package deploypolicy

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

const (
	// maxDeployPolicyLimit the max limit of deploy policies.
	maxDeployPolicyLimit = 1000
)

// List lists the deploy policies.
func (h *handler) List(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.DeployPolicyListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list deploy policies, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	conditions, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list deploy policies, failed to convert conditions")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	page, err := req.ConvertPageToTypes(maxDeployPolicyLimit)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list deploy policies, failed to convert page")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	deployPolicies, total, err := h.daoDeployPolicy.ListDeployPolicies(rCtx, page, conditions)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list deploy policies")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.DeployPolicyListResp)
	if err := resp.ConvertDeployPoliciesFromTypes(total, deployPolicies); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list deploy policies, failed to convert deploy policies")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	return resp.GetData(), nil
}
