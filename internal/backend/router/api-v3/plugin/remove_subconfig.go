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

package plugin

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

var errInvalidRemoveSubConfigTarget = errors.New("invalid remove sub config target")

// RemoveSubConfig removes plugin sub config.
func (h *handler) RemoveSubConfig(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PluginRemoveSubConfigReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to remove plugin subconfig, failed to decode request body.")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	if err := req.Validate(); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to remove plugin subconfig, request validation failed.")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	pluginName := conv.SliceToSlice(req.GetPlugin(), func(item *protoBackend.PluginRemoveSubConfigReq_RemoveSubConfigInfo) string {
		return item.GetPluginName()
	})
	if authErr := h.authorizedPluginOperate(rCtx, pluginName...); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).
			With("plugin-name", pluginName).
			Error("failed to remove plugin subconfig, permission denied.")

		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	pluginDeploymentParam, err := h.generateDeployParams(rCtx, req)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			With("plugin-name", pluginName).
			Error("failed to remove plugin subconfig, generate deploy params failed.")

		if errors.Is(err, errInvalidRemoveSubConfigTarget) {
			return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
		}

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	applyPluginCompatibilityModePolicy(h.resolvePluginCompatibilityModePolicy(rCtx), rCtx.TenantID(), pluginDeploymentParam...)
	pluginDeployments, hostIDs, bizIDs, err := types.NewPluginDeploymentsByParams(
		rCtx.TenantID(), types.PluginDeploymentTransferOptionsOnlyTransferInstaller(), pluginDeploymentParam...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to remove plugin subconfig, failed to generate plugin deployments.")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	workflowID, err := h.pluginMgrIface.LaunchRemovePluginSubConfig(rCtx, types.RemovePluginSubConfigParam{
		Type:              types.PluginWorkflowTypeRemovePluginSubConfig,
		HostIDs:           hostIDs,
		BizIDs:            bizIDs,
		Operator:          rCtx.BKUsername(),
		PluginDeployments: pluginDeployments,
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to remove plugin subconfig.")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	respData := &protoBackend.PluginRemoveSubConfigResp_Data{
		WorkflowId: workflowID,
	}

	logger.G.Biz(rCtx).With("workflow-id", workflowID).Info("launched remove plugin subconfig workflow")

	return respData, nil
}

func (h *handler) generateDeployParams(rCtx restserver.IContext, req *protoBackend.PluginRemoveSubConfigReq) ([]*types.PluginDeploymentParam, error) {
	hostTopoMapping, err := h.daoHost.GetHostTopoRelationMapping(rCtx, req.GetHostIDs()...)
	if err != nil {
		return nil, err
	}

	params := make([]*types.PluginDeploymentParam, 0, len(req.GetPlugin()))
	for _, item := range req.GetPlugin() {
		targetConfigNames, err := h.getTargetConfigNames(rCtx, item.GetBkHostId(), item.GetPluginName(), item.GetConfigTemplateName())
		if err != nil {
			return nil, err
		}

		params = append(params, &types.PluginDeploymentParam{
			HostID:                  item.GetBkHostId(),
			BizID:                   hostTopoMapping[item.GetBkHostId()].BizID,
			PluginName:              item.GetPluginName(),
			RemoveSubConfigFileName: append(item.GetConfigFileName(), targetConfigNames...),
		})
	}

	return params, nil
}

func (h *handler) getTargetConfigNames(rCtx restserver.IContext, hostID int64, pluginName string, targetTemplateNames []string) ([]string, error) {
	if len(targetTemplateNames) == 0 {
		return []string{}, nil
	}

	configs, _, err := h.daoProcessConfig.ListProcessConfigs(rCtx, types.UnlimitedPage(), &types.ProcessConfigCondition{
		ExactInclude: &types.ProcessConfigExactFields{
			HostID:       []int64{hostID},
			ProcessName:  []string{pluginName},
			TemplateName: targetTemplateNames,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("list process configs failed, err: %w", err)
	}

	matchedTemplateNames := make(map[string]struct{}, len(configs))
	for _, config := range configs {
		matchedTemplateNames[config.TemplateName] = struct{}{}
	}
	for _, targetTemplateName := range targetTemplateNames {
		if _, ok := matchedTemplateNames[targetTemplateName]; ok {
			continue
		}

		return nil, fmt.Errorf("%w: config_template_name not found, template-name(%s)", errInvalidRemoveSubConfigTarget, targetTemplateName)
	}

	return conv.SliceToSlice(configs, func(config *types.ProcessConfig) string {
		return config.Name
	}), nil
}
