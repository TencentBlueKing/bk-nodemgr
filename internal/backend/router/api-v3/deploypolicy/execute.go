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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// Execute executes the deploy policy.
func (h *handler) Execute(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.DeployPolicyExecuteReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to execute deploy policy, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	deployPolicyID := req.GetDeployPolicyId()

	// 1. get deploy policy by id
	deployPolicy, err := h.daoDeployPolicy.GetDeployPolicyByID(rCtx, deployPolicyID)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to execute deploy policy, failed to get deploy policy by id")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// 2. check deploy policy is enabled
	if !deployPolicy.Enabled {
		logger.G.Biz(rCtx).Error("failed to execute deploy policy, deploy policy is not enabled")
		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("deploy policy is not enabled"))
	}

	// 3. async execute deploy policy.
	err = h.goAsyncPool.Run(rCtx, func(nCtx contextx.IContext) error {
		err := h.deployPolicyMgr.Do(nCtx, deployPolicy)
		if err != nil {
			logger.G.Biz(nCtx).WithErr(err).Error("failed to execute deploy policy, failed to do deploy policy")
			return fmt.Errorf("failed to execute deploy policy, failed to do deploy policy: %w", err)
		}

		return nil
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to execute deploy policy, failed to run async task")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	resp := new(protoBackend.DeployPolicyExecuteResp)

	return resp.GetData(), nil
}
