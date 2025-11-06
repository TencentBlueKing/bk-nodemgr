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

// ApplySubConfig apply plugin sub config.
func (h *handler) ApplySubConfig(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PluginApplySubConfigReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to apply plugin sub config, failed to decode request body.")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	pluginDeployments, hostIDs, err := h.generateApplyPluginSubConfigDeployments(rCtx, req)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to install plugin, failed to generate plugin deployments.")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	workflowID, err := h.manager.LaunchApplyPluginSubConfig(rCtx, manager.ApplyPluginSubConfigParam{
		Type:              types.PluginWorkflowTypeApplyPluginSubConfig,
		HostIDs:           hostIDs,
		Operator:          rCtx.BKUsername(),
		PluginDeployments: pluginDeployments,
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to apply plugin sub config.")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	respData := &protoBackend.PluginApplySubConfigResp_Data{
		WorkflowId: workflowID,
	}

	logger.G.Biz(rCtx).With("workflow-id", workflowID).Info("launched apply plugin subconfig workflow")

	return respData, nil
}

func (h *handler) generateApplyPluginSubConfigDeployments(nCtx contextx.IContext, req *protoBackend.PluginApplySubConfigReq) (
	[]*types.PluginDeployment, []int64, error) {

	gp := gopool.NewPool()
	pluginDeployments := make([]*types.PluginDeployment, len(req.GetPlugin()))
	for i := range req.GetPlugin() {
		idx := i
		reqProcess := req.GetPlugin()[idx]

		gp.Go(func() error {
			conf := &types.PluginDeploymentPluginConf{
				ConfigFilesDetail:   make([]*types.PluginConfigDetail, 0, len(reqProcess.GetConfigName())),
				CustomConfigContext: reqProcess.GetCustomConfigContext().AsMap(),
			}
			for _, item := range reqProcess.GetConfigName() {
				conf.ConfigFilesDetail = append(conf.ConfigFilesDetail, &types.PluginConfigDetail{Name: item})
			}

			deploymentInfo := &types.PluginDeploymentInfo{
				Process: types.Process{
					TenantID:   nCtx.TenantID(),
					HostID:     reqProcess.GetBkHostId(),
					PluginName: reqProcess.GetPluginName(),
				},
				InstallOptions: types.PluginDeploymentInstallOptions{
					Version: reqProcess.GetVersion(),
				},
				TransferOptions: types.PluginDeploymentTransferOptions{
					SelectDownloads:      true,
					EnableReleasePackage: false,
					EnableInstaller:      true,
				},
			}

			pluginDeployments[idx] = types.NewPluginDeployment(deploymentInfo, conf)

			return nil
		})
	}

	if err := gp.Wait(); err != nil {
		return nil, nil, err
	}

	hostIDMap := make(map[int64]struct{})
	for _, host := range req.GetPlugin() {
		hostIDMap[host.GetBkHostId()] = struct{}{}
	}
	hostIDs := conv.MapKeyToSlice(hostIDMap)

	return pluginDeployments, hostIDs, nil
}
