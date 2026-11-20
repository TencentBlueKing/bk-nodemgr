/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package v3

import (
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/cloudsg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestConvertAccessPoint_CloudSecurityGroupPoolID(t *testing.T) {
	tests := []struct {
		name   string
		poolID string
	}{
		{name: "bound pool", poolID: "pool-a"},
		{name: "no pool configured", poolID: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origin := &types.AccessPoint{
				ID:                       1,
				Name:                     "ap-1",
				NetworkAreaID:            types.DefaultNetworkAreaID,
				CloudSecurityGroupPoolID: tt.poolID,
			}

			restored := convertAccessPointToTypes(convertAccessPointFromTypes(origin))
			assert.Equal(t, tt.poolID, restored.CloudSecurityGroupPoolID)
		})
	}
}

func TestConvertAccessPoint_OmittedPoolIDReadsAsEmpty(t *testing.T) {
	restored := convertAccessPointToTypes(&AccessPoint{})
	assert.Empty(t, restored.CloudSecurityGroupPoolID)
}

func TestTopoCloudSecurityGroupPoolListResp_PoolsRoundTrip(t *testing.T) {
	pools := []cloudsg.CloudSecurityGroupPoolSummary{
		{PoolID: "pool-a", PoolName: "Pool A"},
		{PoolID: "pool-b", PoolName: "Pool B"},
	}

	resp := &TopoCloudSecurityGroupPoolListResp{}
	resp.ConvertPoolsFromTypes(pools)

	assert.Equal(t, pools, resp.ConvertPoolsToTypes())
}
