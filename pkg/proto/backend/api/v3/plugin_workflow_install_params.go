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

package v3

import (
	"fmt"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"google.golang.org/protobuf/types/known/structpb"
)

// AutoConvert keeps historical identifiers unchanged.
func (x *PluginWorkflowOperationInstallParamsGetReq) AutoConvert() {}

// Validate checks the selected workflow and operation identifiers.
func (x *PluginWorkflowOperationInstallParamsGetReq) Validate() error {
	if strings.TrimSpace(x.GetWorkflowId()) == "" {
		return fmt.Errorf("workflow_id is required")
	}
	if len(x.GetOperationIds()) == 0 {
		return fmt.Errorf("operation_ids is required")
	}
	seen := make(map[string]struct{}, len(x.GetOperationIds()))
	for index, operationID := range x.GetOperationIds() {
		if strings.TrimSpace(operationID) == "" {
			return fmt.Errorf("operation_ids[%d] is required", index)
		}
		if _, exists := seen[operationID]; exists {
			return fmt.Errorf("operation_ids[%d] is duplicated", index)
		}
		seen[operationID] = struct{}{}
	}
	return nil
}

// ConvertResultFromTypes converts recovered installation parameters to response data.
func (x *PluginWorkflowOperationInstallParamsGetResp) ConvertResultFromTypes(result *types.PluginWorkflowInstallParamsResult) error {
	x.Data = &PluginWorkflowOperationInstallParamsGetResp_Data{
		Items:                   make([]*PluginWorkflowOperationInstallParams, 0),
		UnavailableOperationIds: make([]string, 0),
	}
	if result == nil {
		return nil
	}
	x.Data.Type = string(result.Type)
	x.Data.UnavailableOperationIds = formatRespSlice(result.UnavailableOperationIDs)
	items, err := conv.SliceToSliceWithError(result.Items, convertPluginInstallParamsFromTypes)
	if err != nil {
		return fmt.Errorf("failed to convert plugin installation parameters: %w", err)
	}
	x.Data.Items = items
	return nil
}

func convertPluginInstallParamsFromTypes(item *types.PluginWorkflowInstallParams) (*PluginWorkflowOperationInstallParams, error) {
	if item == nil {
		return nil, nil
	}
	result := &PluginWorkflowOperationInstallParams{
		OperationId: item.OperationID, BkHostId: item.HostID, PluginName: item.PluginName,
		Version: item.Version, MissingFields: formatRespSlice(item.MissingFields),
	}
	if item.CustomConfigContext != nil {
		context, err := structpb.NewStruct(item.CustomConfigContext)
		if err != nil {
			return nil, fmt.Errorf("invalid historical custom_config_context")
		}
		result.CustomConfigContext = context
	}
	if spec := item.CustomSpec; spec != nil {
		restartType := spec.MonitorPolicy.RestartType.String()
		result.CustomSpec = &PluginInstallReq_DeploymentSpec{
			Resource: &ProcessResource{
				CpuLimitPercent: &spec.Resource.CPULimitPercent,
				MemLimitPercent: &spec.Resource.MemLimitPercent,
			},
			MonitorPolicy: &ProcessMonitorPolicy{
				RestartType:           &restartType,
				StartCheckSeconds:     &spec.MonitorPolicy.StartCheckSecs,
				StopCheckSeconds:      &spec.MonitorPolicy.StopCheckSecs,
				OperateTimeoutSeconds: &spec.MonitorPolicy.OpTimeoutSecs,
			},
		}
	}
	return result, nil
}

// ConvertResultToTypes converts backend response data to domain installation parameters.
func (x *PluginWorkflowOperationInstallParamsGetResp) ConvertResultToTypes() *types.PluginWorkflowInstallParamsResult {
	data := x.GetData()
	if data == nil {
		return nil
	}
	return &types.PluginWorkflowInstallParamsResult{
		Type:                    types.PluginWorkflowType(data.GetType()),
		UnavailableOperationIDs: data.GetUnavailableOperationIds(),
		Items: conv.SliceToSlice(data.GetItems(), func(item *PluginWorkflowOperationInstallParams) *types.PluginWorkflowInstallParams {
			if item == nil {
				return nil
			}
			result := &types.PluginWorkflowInstallParams{
				OperationID: item.GetOperationId(), HostID: item.GetBkHostId(), PluginName: item.GetPluginName(),
				Version: item.GetVersion(), MissingFields: item.GetMissingFields(),
			}
			if item.GetCustomConfigContext() != nil {
				result.CustomConfigContext = item.GetCustomConfigContext().AsMap()
			}
			if spec := item.GetCustomSpec(); spec != nil {
				result.CustomSpec = &types.PluginSpec{
					Resource: types.ProcessResource{
						CPULimitPercent: spec.GetResource().GetCpuLimitPercent(),
						MemLimitPercent: spec.GetResource().GetMemLimitPercent(),
					},
					MonitorPolicy: types.ProcessMonitorPolicy{
						RestartType:    types.ProcessRestartType(spec.GetMonitorPolicy().GetRestartType()),
						StartCheckSecs: spec.GetMonitorPolicy().GetStartCheckSeconds(),
						StopCheckSecs:  spec.GetMonitorPolicy().GetStopCheckSeconds(),
						OpTimeoutSecs:  spec.GetMonitorPolicy().GetOperateTimeoutSeconds(),
					},
				}
			}
			return result
		}),
	}
}
