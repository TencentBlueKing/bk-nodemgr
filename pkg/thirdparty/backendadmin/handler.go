/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package backendadmin

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandler defines the interface for backend admin operations.
type IHandler interface {
	GetNetworkUnitSegmentRules(nCtx contextx.IContext) (types.NetworkUnitSegmentRuleConfig, error)
	UpsertNetworkUnitSegmentRules(nCtx contextx.IContext, cfg types.NetworkUnitSegmentRuleConfig) error
}

var _ IHandler = &Handler{}

// Handler implements IHandler for backend admin operations.
type Handler struct {
	cli *cli
}

// New creates a new Handler instance.
func New(c *restclient.Capability, conf *Config) (*Handler, error) {
	cli, err := newClient(c, conf)
	if err != nil {
		return nil, err
	}

	return &Handler{cli: cli}, nil
}

// GetNetworkUnitSegmentRules retrieves the network unit segment rules configuration.
func (h *Handler) GetNetworkUnitSegmentRules(nCtx contextx.IContext) (types.NetworkUnitSegmentRuleConfig, error) {
	return h.cli.getNetworkUnitSegmentRules(nCtx)
}

// UpsertNetworkUnitSegmentRules updates or inserts the network unit segment rules configuration.
func (h *Handler) UpsertNetworkUnitSegmentRules(
	nCtx contextx.IContext,
	cfg types.NetworkUnitSegmentRuleConfig,
) error {

	return h.cli.upsertNetworkUnitSegmentRules(nCtx, cfg)
}
