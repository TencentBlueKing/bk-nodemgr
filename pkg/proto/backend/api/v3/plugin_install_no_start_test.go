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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestPluginInstallNoStartReqValidate(t *testing.T) {
	require.Error(t, new(PluginInstallNoStartReq).Validate())

	req := &PluginInstallNoStartReq{
		Plugin: []*PluginInstallNoStartReq_InstallInfo{
			{BkHostId: 0, PluginName: "bkmonitorbeat"},
		},
	}
	require.Error(t, req.Validate())

	req.Plugin[0].BkHostId = 1
	req.Plugin[0].PluginName = ""
	require.Error(t, req.Validate())

	req.Plugin[0].PluginName = "bkmonitorbeat"
	require.NoError(t, req.Validate())
}

func TestPluginInstallNoStartReqConvertParamToTypesWithHostTopoMapping(t *testing.T) {
	customConfigContext, err := structpb.NewStruct(map[string]any{"key": "value"})
	require.NoError(t, err)

	cpuLimitPercent, memLimitPercent := 50.0, 60.0
	startCheckSeconds := int64(10)

	req := &PluginInstallNoStartReq{
		Plugin: []*PluginInstallNoStartReq_InstallInfo{
			{
				BkHostId:            1,
				PluginName:          "bkmonitorbeat",
				Version:             "3.0.0",
				CustomConfigContext: customConfigContext,
				CustomSpec: &PluginInstallNoStartReq_DeploymentSpec{
					Resource: &ProcessResource{
						CpuLimitPercent: &cpuLimitPercent,
						MemLimitPercent: &memLimitPercent,
					},
					MonitorPolicy: &ProcessMonitorPolicy{
						StartCheckSeconds: &startCheckSeconds,
					},
				},
			},
			{BkHostId: 2, PluginName: "bkmonitorbeat"},
		},
	}

	params := req.ConvertParamToTypesWithHostTopoMapping(map[int64]types.HostTopoRelation{
		1: {BizID: 100},
		2: {BizID: 200},
	})

	require.Len(t, params, 2)

	assert.Equal(t, int64(1), params[0].HostID)
	assert.Equal(t, int64(100), params[0].BizID)
	assert.Equal(t, "bkmonitorbeat", params[0].PluginName)
	assert.Equal(t, "3.0.0", params[0].Version)
	assert.Equal(t, map[string]any{"key": "value"}, params[0].CustomConfigContext)
	require.NotNil(t, params[0].CustomSpec)
	assert.Equal(t, float64(50), params[0].CustomSpec.Resource.CPULimitPercent)
	assert.Equal(t, float64(60), params[0].CustomSpec.Resource.MemLimitPercent)
	assert.Equal(t, int64(10), params[0].CustomSpec.MonitorPolicy.StartCheckSecs)
	assert.True(t, params[0].RenewSubConfig)
	// the compatibility mode is not involved in the install without start operation.
	assert.False(t, params[0].EnableCompatibilityMode)

	assert.Equal(t, int64(2), params[1].HostID)
	assert.Equal(t, int64(200), params[1].BizID)
	assert.Nil(t, params[1].CustomSpec)
	assert.Empty(t, params[1].CustomConfigContext)

	assert.Equal(t, []int64{1, 2}, req.GetHostIDs())
}
