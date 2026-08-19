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

package globalsettings

import (
	"encoding/json"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/compatibility"
	pkgglobalsettings "github.com/TencentBlueKing/bk-nodemgr/pkg/globalsettings"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

type pluginCompatibilityModePolicyGetReq struct{}

func (r *pluginCompatibilityModePolicyGetReq) Validate() error {
	return nil
}

func (r *pluginCompatibilityModePolicyGetReq) AutoConvert() {}

type pluginCompatibilityModePolicyUpsertReq struct {
	EnabledPlugins []string                              `json:"enabled_plugins"`
	DisabledBiz    []pluginCompatibilityModePolicyBizReq `json:"disabled_biz"`
}

func (r *pluginCompatibilityModePolicyUpsertReq) Validate() error {
	return nil
}

func (r *pluginCompatibilityModePolicyUpsertReq) AutoConvert() {}

func (r *pluginCompatibilityModePolicyUpsertReq) toPolicy() compatibility.Policy {
	var disabledBiz []compatibility.DisabledBiz
	if r.DisabledBiz != nil {
		disabledBiz = make([]compatibility.DisabledBiz, 0, len(r.DisabledBiz))
		for _, item := range r.DisabledBiz {
			disabledBiz = append(disabledBiz, compatibility.DisabledBiz{
				TenantID: item.TenantID,
				BKBizID:  item.BKBizID,
			})
		}
	}

	return compatibility.Policy{
		EnabledPlugins: r.EnabledPlugins,
		DisabledBiz:    disabledBiz,
	}
}

type pluginCompatibilityModePolicyBizReq struct {
	TenantID string `json:"tenant_id"`
	BKBizID  int64  `json:"bk_biz_id"`
}

// GetPluginCompatibilityModePolicy gets the plugin compatibility mode policy.
func (h *handler) GetPluginCompatibilityModePolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(pluginCompatibilityModePolicyGetReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error(
			"failed to get plugin compatibility mode policy, failed to decode request body",
		)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	value, err := h.storage.GetGlobalSetting(rCtx, pkgglobalsettings.PluginCompatibilityModePolicy)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Warn("failed to get plugin compatibility mode policy, fallback to default")
		value = ""
	}
	policy := compatibility.ParseStoredPolicy(value, rCtx)

	return policy, nil
}

// UpsertPluginCompatibilityModePolicy upserts the plugin compatibility mode policy.
func (h *handler) UpsertPluginCompatibilityModePolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(pluginCompatibilityModePolicyUpsertReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error(
			"failed to upsert plugin compatibility mode policy, failed to decode request body",
		)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	policy := req.toPolicy()
	if err := compatibility.ValidateForWrite(policy); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upsert plugin compatibility mode policy, invalid request")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	policy = policy.Normalize()
	value, err := json.Marshal(policy)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error(
			"failed to upsert plugin compatibility mode policy, failed to encode policy",
		)
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	setting := &types.GlobalSettings{
		SettingName: pkgglobalsettings.PluginCompatibilityModePolicy,
		Value:       string(value),
	}
	if err := h.storage.UpsertGlobalSettings(rCtx, setting); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upsert plugin compatibility mode policy")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	return policy, nil
}
