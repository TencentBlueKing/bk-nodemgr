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
	"fmt"

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

func parseProjectionParam(ctx *action.InstanceContext) (*OperParamComposeProjectedPluginSubConfigs, error) {
	param := new(OperParamComposeProjectedPluginSubConfigs)
	if err := conv.MapToStruct(ctx.Data.Content, param); err != nil {
		return nil, fmt.Errorf("failed to parse projection parameters: %w", err)
	}
	if param.TenantID == "" || param.HostID <= 0 || len(param.SourceTokens) == 0 {
		return nil, fmt.Errorf("projection tenant, execution host and source tokens are required")
	}
	if err := param.SourceOperation.Validate(); err != nil {
		return nil, err
	}

	return param, nil
}

func projectionSource(ctx *action.InstanceContext, storage pluginStg.IStorage,
	param *OperParamComposeProjectedPluginSubConfigs, token string) (*pluginUtils.PluginActionStandarder, error) {

	if token == "" {
		return nil, fmt.Errorf("projection source token is required")
	}
	nCtx := contextx.From(ctx.Ctx, contextx.WithTenantID(param.TenantID), contextx.WithBKUsername(param.Operator))
	std := pluginUtils.NewPluginActionStandarder(storage)
	if err := std.Initialize(&action.InstanceContext{Ctx: nCtx, Data: ctx.Data}, pluginUtils.PluginActionStandardParam{
		TenantID: param.TenantID, Operator: param.Operator, Token: token,
	}); err != nil {
		return nil, err
	}
	if current := std.DeployInfo(); current == nil || current.Process.TenantID != param.TenantID ||
		current.Process.HostID != param.HostID || current.Process.PluginName == "" {

		return nil, fmt.Errorf("projection source identity changed")
	}

	return std, nil
}

func saveProjectionChildren(storage workflow.IStorageActionInstance, ctx *action.InstanceContext,
	actionName string, refs []types.SubWorkflowRef) error {

	serialized, err := types.SerializeSubWorkflowRefs(refs)
	if err != nil {
		return fmt.Errorf("failed to serialize projection children: %w", err)
	}
	if err := storage.UpsertActionInstancePrivateData(ctx.Ctx, ctx.Data.OperationInstanceID,
		actionName, map[string]any{types.PDKeySubWorkflowRefs: serialized}); err != nil {
		return fmt.Errorf("failed to save projection children: %w", err)
	}
	if ctx.Data.PrivateData == nil {
		ctx.Data.PrivateData = make(map[string]any)
	}
	ctx.Data.PrivateData[types.PDKeySubWorkflowRefs] = serialized

	return nil
}
