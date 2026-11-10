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

// nolint: revive // the package name follows the plugin callback router layout.
package plugin

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoCallback "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/callback"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// pluginMainConfigFile is one main config file returned to the plugin installer.
type pluginMainConfigFile struct {
	// FileName is the config file name generated under the plugin config dir.
	FileName string `json:"file_name"`
	// Content is the rendered config file content.
	Content string `json:"content"`
}

// pluginMainConfigsResp is the response of the get main config callback api.
type pluginMainConfigsResp struct {
	// MainConfigs is all main config files of the plugin deployment, it can be empty for plugin package without main config.
	MainConfigs []*pluginMainConfigFile `json:"main_configs"`
}

// GetMainConfig returns all main config files of the plugin deployment, a plugin package with multiple main config
// templates is supported.
func (h *handler) GetMainConfig(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoCallback.PluginGetMainConfigReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get main config, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	token := req.GetToken()
	configDetails, err := h.daoPluginDeployment.GetPluginDeploymentPluginConfConfigFilesDetail(rCtx, token)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get main config, failed to get plugin deployment main config")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := &pluginMainConfigsResp{MainConfigs: make([]*pluginMainConfigFile, 0, len(configDetails))}
	for _, config := range configDetails {
		if config == nil || !config.IsMainConfig {
			continue
		}

		resp.MainConfigs = append(resp.MainConfigs, &pluginMainConfigFile{
			FileName: config.Name,
			Content:  config.Content,
		})
	}

	logger.G.Biz(rCtx).With("file-count", len(resp.MainConfigs)).Info("got main configs")

	return resp, nil
}
