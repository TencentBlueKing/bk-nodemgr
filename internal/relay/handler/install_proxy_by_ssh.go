/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package handler

import (
	"encoding/json"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
)

const installProxyBySSHUnsupported = "relay install proxy by ssh executor is not implemented"

// InstallProxyBySSH installs proxy by SSH.
func (h *handler) InstallProxyBySSH(nCtx contextx.IContext, payload []byte) {
	logger.G.Biz(nCtx).Info("handler install proxy by ssh event")

	var event protoRelay.InstallProxyBySSHReq
	if err := json.Unmarshal(payload, &event); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to unmarshal install proxy by ssh event")

		return
	}

	if err := h.reportInstallResult(nCtx, event.ActionName, event.OperInstID, "", installProxyBySSHUnsupported); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to report unsupported install proxy by ssh result")

		return
	}

	logger.G.Biz(nCtx).With("ip", event.IP, "port", event.Port, "user", event.User).
		Info("done report unsupported install proxy by ssh result")
}
