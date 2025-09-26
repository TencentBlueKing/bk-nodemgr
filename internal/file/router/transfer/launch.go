/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package transfer

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// TransferLaunchNode launch transfer node package.
func (h *handler) TransferLaunchNode(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.TransferLaunchNodeReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to launch transfer node, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	host, err := h.storageTopo.GetHostByID(rCtx, req.GetTargetHostId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to launch transfer node, failed to get host")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	tf, err := h.manager.LaunchTransferNode(rCtx, gen, rt, plat, version, req.GetTargetDir(), host)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to launch transfer node")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("task-id", tf.GetTaskID()).Info("launched node transfer")

	resp := new(protoFile.TransferLaunchNodeResp)
	resp.ConvertResult(tf)

	return resp.GetData(), nil
}

// TransferLaunchInstaller launch transfer installer package.
func (h *handler) TransferLaunchInstaller(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.TransferLaunchInstallerReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to launch transfer installer, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	host, err := h.storageTopo.GetHostByID(rCtx, req.GetTargetHostId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to launch transfer installer, failed to get host")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	tf, err := h.manager.LaunchTransferInstaller(rCtx,
		protoFile.ConvertPlatformToTypes(req.GetPlatform()),
		req.GetTargetDir(),
		host)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to launch transfer installer")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("task-id", tf.GetTaskID()).Info("launched installer transfer")

	resp := new(protoFile.TransferLaunchInstallerResp)
	resp.ConvertResult(tf)

	return resp.GetData(), nil
}

// TransferLaunchPlugin launch transfer plugin package.
func (h *handler) TransferLaunchPlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.TransferLaunchPluginReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to launch transfer plugin, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	host, err := h.storageTopo.GetHostByID(rCtx, req.GetTargetHostId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to launch transfer plugin, failed to get host")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	name, gen, rt, plat, version := req.GetIdentifier()
	tf, err := h.manager.LaunchTransferPlugin(rCtx, name, gen, rt, plat, version, req.GetTargetDir(), host)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to launch transfer plugin")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("task-id", tf.GetTaskID()).Info("launched plugin transfer")

	resp := new(protoFile.TransferLaunchPluginResp)
	resp.ConvertResult(tf)

	return resp.GetData(), nil
}
