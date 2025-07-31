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
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// TransferLaunchRelease launch transfer release package.
func (h *handler) TransferLaunchRelease(ctx *restserver.Context) (interface{}, error) {
	req := new(protoFile.TransferLaunchReleaseReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to launch transfer release, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	host, err := h.storageTopo.GetHostByID(ctx, req.GetTargetHostId())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to launch transfer release, failed to get host. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	tf, err := h.manager.LaunchTransferRelease(ctx, gen, rt, plat, version, req.GetTargetDir(), host)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to launch transfer release. err: %v", err)

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	h.logger.InfoCtxf(ctx, "launched release transfer. task-id(%s)", tf.GetTaskID())

	resp := new(protoFile.TransferLaunchReleaseResp)
	resp.ConvertResult(tf)

	return resp.GetData(), nil
}

// TransferLaunchInstaller launch transfer installer package.
func (h *handler) TransferLaunchInstaller(ctx *restserver.Context) (interface{}, error) {
	req := new(protoFile.TransferLaunchInstallerReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to launch transfer installer, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	host, err := h.storageTopo.GetHostByID(ctx, req.GetTargetHostId())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to launch transfer installer, failed to get host. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	tf, err := h.manager.LaunchTransferInstaller(ctx,
		protoFile.ConvertPlatformToTypes(req.GetPlatform()),
		req.GetTargetDir(),
		host)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to launch transfer installer. err: %v", err)

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	h.logger.InfoCtxf(ctx, "launched installer transfer. task-id(%s)", tf.GetTaskID())

	resp := new(protoFile.TransferLaunchInstallerResp)
	resp.ConvertResult(tf)

	return resp.GetData(), nil
}
