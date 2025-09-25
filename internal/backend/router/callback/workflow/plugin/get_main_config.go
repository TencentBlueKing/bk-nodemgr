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
	"bytes"
	"fmt"
	"io"

	protoCallback "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/callback"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// GetMainConfig get plugin main config.
func (h *handler) GetMainConfig(rCtx restserver.IContext) (*restserver.FileResponse, error) {
	req := new(protoCallback.PluginGetMainConfigReq)
	if err := rCtx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to get main config, failed to decode request body: %v", err)

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	token := req.GetToken()

	info, err := h.daoPluginDeployment.GetPluginDeploymentInfo(rCtx, token)
	if err != nil {
		h.logger.Errorf("failed to get main config, failed to get plugin deployment info: %v", err)

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	mainConfigBytes, err := h.daoPluginDeployment.GetPluginDeploymentMainConfig(rCtx, token)
	if err != nil {
		h.logger.Errorf("failed to get main config, failed to get plugin deployment main config: %v", err)

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	fileName := fmt.Sprintf("%s.conf", info.Plugin.Name)
	data := io.NopCloser(bytes.NewReader(mainConfigBytes))

	resp := &restserver.FileResponse{
		Data:        data,
		Size:        0,
		FilePath:    "",
		FileName:    fileName,
		ContentType: restserver.MIMETypeText,
		Headers:     nil,
	}

	return resp, nil
}
