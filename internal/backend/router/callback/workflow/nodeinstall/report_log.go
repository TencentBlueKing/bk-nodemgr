/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package nodeinstall ...
package nodeinstall

import (
	"fmt"
	"net/http"
	"time"

	proto "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/callback"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/gin-gonic/gin"
)

// ReportLog report agent install shell script log.
func (h *handler) ReportLog(gCtx *gin.Context) {
	req := new(proto.ReportLogReq)
	if err := gCtx.BindJSON(req); err != nil {
		h.logger.Errorf("report log failed, err: %s", err)
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Errorf("report log failed, err: %s", err)
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	info, err := h.GetInfo(gCtx, req.Token)
	if err != nil {
		h.logger.Errorf("token is invalid, err: %s", err)
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	logs := make([]action.Message, len(req.Logs))
	for idx, log := range req.Logs {
		logs[idx] = action.Message{
			Time: time.Now(),
			Text: fmt.Sprintf("%s\t:%s\t:%s\t",
				time.Unix(log.GetTimestamp(), 0).Format("2006-01-02 15:04:05"),
				log.GetStep(),
				log.GetLog()),
			Level: log.GetLevel(),
		}
	}

	if err = h.PushActionInstanceMessage(gCtx, req.GetOperInstId(), info.BlockingActionName, logs...); err != nil {
		h.logger.Errorf("report log failed, err: %s", err)
		gCtx.JSON(http.StatusInternalServerError, err)

		return
	}

	gCtx.JSON(http.StatusOK, nil)
}
