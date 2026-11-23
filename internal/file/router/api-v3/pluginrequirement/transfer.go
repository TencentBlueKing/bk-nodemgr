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

package pluginrequirement

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// Transfer launches an asynchronous requirement file transfer to the target host.
func (h *handler) Transfer(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.PluginRequirementTransferReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to transfer plugin requirement, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	host, err := h.storageTopo.GetHostByID(rCtx, req.GetTargetHostId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to transfer plugin requirement, failed to get host")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	tf, err := h.manager.LaunchTransferPluginRequirement(rCtx, req.GetRequirementFileId(), req.GetTargetDir(), host)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to transfer plugin requirement")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("task-id", tf.GetTaskID()).Info("launched plugin requirement transfer")

	resp := new(protoFile.PluginRequirementTransferResp)
	resp.ConvertResult(tf)

	return resp.GetData(), nil
}
