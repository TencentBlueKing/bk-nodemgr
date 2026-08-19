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

package backendadmin

import (
	"encoding/json"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/globalsettings"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (c *cli) getNetworkUnitSegmentRules(nCtx contextx.IContext) (types.NetworkUnitSegmentRuleConfig, error) {
	if c.getNetworkUnitSegmentRulesFn != nil {
		return c.getNetworkUnitSegmentRulesFn(nCtx)
	}

	header, err := c.getCommonHeader(nCtx)
	if err != nil {
		return nil, err
	}

	req := &protoBackend.GetGlobalSettingReq{SettingName: globalsettings.NetworkUnitSegmentRules}
	resp := new(protoBackend.GetGlobalSettingResp)
	err = c.client.Post().
		SubResourcef("/globalsettings/get").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if resp.GetCode() != 0 {
		return nil, fmt.Errorf("get networkunit segment rules failed, code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil || resp.GetData().GetValue() == "" {
		return types.NetworkUnitSegmentRuleConfig{}, nil
	}

	var cfg types.NetworkUnitSegmentRuleConfig
	if err := json.Unmarshal([]byte(resp.GetData().GetValue()), &cfg); err != nil {
		return nil, fmt.Errorf("parse networkunit segment rules: %w", err)
	}

	return cfg, nil
}

func (c *cli) upsertNetworkUnitSegmentRules(nCtx contextx.IContext, cfg types.NetworkUnitSegmentRuleConfig) error {
	if c.upsertNetworkUnitSegmentRulesFn != nil {
		return c.upsertNetworkUnitSegmentRulesFn(nCtx, cfg)
	}

	value, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal networkunit segment rules: %w", err)
	}

	header, err := c.getCommonHeader(nCtx)
	if err != nil {
		return err
	}

	req := &protoBackend.UpsertGlobalSettingsReq{Settings: []*protoBackend.GlobalSetting{{
		SettingName: globalsettings.NetworkUnitSegmentRules,
		Value:       string(value),
	}}}
	resp := new(protoBackend.UpsertGlobalSettingsResp)
	err = c.client.Post().
		SubResourcef("/globalsettings/upsertmany").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return err
	}

	if resp.GetCode() != 0 {
		return fmt.Errorf("upsert networkunit segment rules failed, code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}
