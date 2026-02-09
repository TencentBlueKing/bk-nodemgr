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
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoCallback "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/callback"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/common"
)

// ReportLog report plugin installer log.
func (h *handler) ReportLog(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoCallback.PluginReportLogReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to report log, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	token := req.GetToken()
	info, err := h.daoPluginDeployment.GetPluginDeploymentInfo(rCtx, token)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to report log, failed to get plugin deployment info")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	logs := make([]common.Message, len(req.GetLogs()))
	for idx, log := range req.GetLogs() {
		text := fmt.Sprintf("%s\t:%s\t:%s\t",
			time.Unix(log.GetTimestamp(), 0).Format("2006-01-02 15:04:05"),
			log.GetStep(),
			log.GetLog())
		logs[idx] = common.Message{
			Time:   time.Now(),
			Text:   text,   // 过渡期保留
			TextZh: text,   // 同时写入中文字段
			TextEn: text,   // 同时写入英文字段
			Level:  log.GetLevel(),
		}
	}

	operInstID := req.GetOperInstId()
	if err = h.stgWorkflow.PushActionInstanceMessage(rCtx, operInstID, info.BlockingActionName, logs...); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to report log, failed to push message")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoCallback.PluginReportLogResp)

	return resp.GetData(), nil
}
