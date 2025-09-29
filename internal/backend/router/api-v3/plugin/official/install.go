/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package official

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Install defines the handler to install external plugin.
func (h *handler) Install(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PluginOfficialInstallReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to install official plugin, failed to decode request body.")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	pluginDeployments, hostIDs, err := h.generateInstallPluginDeployments(req)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to install official plugin, failed to generate plugin deployments.")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	workflowID, err := h.manager.LaunchInstallPlugin(rCtx, manager.InstallPluginParam{
		Type:              types.PluginWorkflowTypeInstallOfficial,
		HostIDs:           hostIDs,
		Operator:          rCtx.BKUsername(),
		PluginDeployments: pluginDeployments,
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to install plugin.")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	respData := &protoBackend.PluginOfficialInstallResp_Data{
		WorkflowId: workflowID,
	}

	logger.G.Biz(rCtx).With("workflow-id", workflowID).Info("launched install plugin workflow")

	return respData, nil
}

func (h *handler) generateInstallPluginDeployments(req *protoBackend.PluginOfficialInstallReq) ([]*types.PluginDeployment, []int64, error) {
	targetVersions := make([]types.TargetPluginVersion, len(req.GetTargetVersion()))
	for idx, version := range req.GetTargetVersion() {
		targetVersions[idx] = types.TargetPluginVersion{
			Platform: platform.Platform{
				OS:   criteria.OSType(version.GetOsType()),
				Arch: criteria.CPUArch(version.GetCpuArch()),
			},
			Version: version.GetVersion(),
		}
	}

	hostIDMap := make(map[int64]struct{})
	for _, host := range req.GetPlugin() {
		hostIDMap[host.GetBkHostId()] = struct{}{}
	}
	hostIDs := conv.MapKeyToSlice(hostIDMap)

	gp := gopool.NewPool()
	pluginDeployments := make([]*types.PluginDeployment, len(req.GetPlugin()))
	for i := range req.GetPlugin() {
		idx := i
		reqPlugin := req.GetPlugin()[idx]

		gp.Go(func() error {
			pluginDeployment := types.NewPluginDeployment(&types.PluginDeploymentInfo{
				Plugin: types.Plugin{
					Name:       reqPlugin.GetName(),
					HostID:     reqPlugin.GetBkHostId(),
					Type:       types.PluginTypeOfficial,
					Generation: types.Generation2,
					Platform:   platform.Platform{},
					Version:    reqPlugin.GetVersion(),
				},
				InstallOptions: types.PluginDeploymentInstallOptions{},
				TransferOptions: types.PluginDeploymentTransferOptions{
					SelectDownloads:      true,
					EnableReleasePackage: true,
					EnableInstaller:      true,
				},
				TargetVersion: targetVersions,
			})

			pluginDeployments[idx] = pluginDeployment

			return nil
		})
	}

	if err := gp.Wait(); err != nil {
		return nil, nil, err
	}

	return pluginDeployments, hostIDs, nil
}
