/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package monitor

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
)

// IHandler defines monitor business-facing operations.
type IHandler interface {
	// GetOrCreateAgentEventDataID gets or creates an agent event data-id by business ID.
	GetOrCreateAgentEventDataID(nCtx contextx.IContext, bkBizID int64) (int64, bool, error)
}

// Handler is the monitor-backed implementation of IHandler.
type Handler struct {
	cli *cli
}

var _ IHandler = (*Handler)(nil)

// New creates a monitor handler.
func New(c *restclient.Capability, conf *Config) (IHandler, error) {
	cli, err := newClient(c, conf)
	if err != nil {
		return nil, err
	}

	return &Handler{cli: cli}, nil
}

// GetOrCreateAgentEventDataID gets or creates an agent event data-id by business ID.
func (h *Handler) GetOrCreateAgentEventDataID(nCtx contextx.IContext, bkBizID int64) (int64, bool, error) {
	if bkBizID <= 0 {
		return 0, false, fmt.Errorf("bk_biz_id must be positive, got %d", bkBizID)
	}

	data, err := h.cli.getOrCreateAgentEventDataID(nCtx, bkBizID)
	if err != nil {
		return 0, false, err
	}

	if data == nil || data.BkDataID <= 0 {
		return 0, false, fmt.Errorf("invalid monitor agent event data-id response for bk_biz_id %d", bkBizID)
	}

	return data.BkDataID, true, nil
}
