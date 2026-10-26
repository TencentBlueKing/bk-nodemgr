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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestReleasePluginResponsePreservesInstallerLimits(t *testing.T) {
	item := &types.ReleasePlugin{
		Release: types.Release{
			AdditionInfo: map[string]any{
				"legacy_field": "preserved",
			},
		},
		ReleaseAdditionInfoPlugin: types.ReleaseAdditionInfoPlugin{
			InstallCPULimit:    0.1,
			InstallDiskIOLimit: 10,
		},
	}

	getResp := new(ReleasePluginGetResp)
	require.NoError(t, getResp.ConvertFromTypes(item))
	assertPluginInstallerLimits(t, getResp.GetData().GetItem())

	listResp := new(ReleasePluginListResp)
	require.NoError(t, listResp.ConvertFromTypes(1, []*types.ReleasePlugin{item}))
	assertPluginInstallerLimits(t, listResp.GetData().GetItems()[0])
}

func assertPluginInstallerLimits(t *testing.T, item *ReleaseInfo) {
	t.Helper()

	additionInfo := item.GetAdditionInfo().AsMap()
	require.Equal(t, 0.1, additionInfo["InstallCPULimit"])
	require.Equal(t, float64(10), additionInfo["InstallDiskIOLimit"])
	require.Equal(t, "preserved", additionInfo["legacy_field"])
}
