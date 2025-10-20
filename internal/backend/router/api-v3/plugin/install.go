/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package plugin

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Install defines the handler to install plugin.
func (h *handler) Install(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PluginInstallReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to install plugin, failed to decode request body.")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	pluginDeployments, hostIDs, err := h.generateInstallPluginDeployments(rCtx, req)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to install plugin, failed to generate plugin deployments.")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	workflowID, err := h.manager.LaunchInstallPlugin(rCtx, manager.InstallPluginParam{
		Type:              types.PluginWorkflowTypeInstall,
		HostIDs:           hostIDs,
		Operator:          rCtx.BKUsername(),
		PluginDeployments: pluginDeployments,
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to install plugin.")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	respData := &protoBackend.PluginInstallResp_Data{
		WorkflowId: workflowID,
	}

	logger.G.Biz(rCtx).With("workflow-id", workflowID).Info("launched install plugin workflow")

	return respData, nil
}

func (h *handler) generateInstallPluginDeployments(nCtx contextx.IContext, req *protoBackend.PluginInstallReq) (
	[]*types.PluginDeployment, []int64, error) {

	gp := gopool.NewPool()
	pluginDeployments := make([]*types.PluginDeployment, len(req.GetProcess()))
	for i := range req.GetProcess() {
		idx := i
		reqProcess := req.GetProcess()[idx]

		gp.Go(func() error {
			pluginDeployment := types.NewPluginDeployment(&types.PluginDeploymentInfo{
				Process: types.Process{
					TenantID: nCtx.TenantID(),
					HostID:   reqProcess.GetBkHostId(),
					PluginID: reqProcess.GetPluginId(),
					Info: types.ProcessInfo{
						Version: reqProcess.GetVersion(),
					},
				},
				InstallOptions: types.PluginDeploymentInstallOptions{},
				TransferOptions: types.PluginDeploymentTransferOptions{
					SelectDownloads:      true,
					EnableReleasePackage: false,
					EnableInstaller:      true,
				},
			})

			pluginDeployments[idx] = pluginDeployment

			return nil
		})
	}

	if err := gp.Wait(); err != nil {
		return nil, nil, err
	}

	hostIDMap := make(map[int64]struct{})
	for _, host := range req.GetProcess() {
		hostIDMap[host.GetBkHostId()] = struct{}{}
	}
	hostIDs := conv.MapKeyToSlice(hostIDMap)

	return pluginDeployments, hostIDs, nil
}
