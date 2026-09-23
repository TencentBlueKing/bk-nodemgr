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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestTopoRecommendNetworkUnitByNetworkSegmentReqConvertItemsToTypesPreservesOrderAndDuplicatesForApplication(t *testing.T) {
	req := &TopoRecommendNetworkUnitByNetworkSegmentReq{
		Items: []*TopoRecommendNetworkUnitByNetworkSegmentReq_Item{
			{BkNetworkareaId: 1001, Ip: "127.0.0.1"},
			{BkNetworkareaId: 1001, Ip: "127.0.0.1"},
			{BkNetworkareaId: 2002, Ip: "127.0.0.2"},
		},
	}

	items := req.ConvertItemsToTypes()
	require.Len(t, items, 3)

	assert.Equal(t, int64(1001), items[0].NetworkAreaID)
	assert.Equal(t, "127.0.0.1", items[0].IP)
	assert.Equal(t, int64(1001), items[1].NetworkAreaID)
	assert.Equal(t, "127.0.0.1", items[1].IP)
	assert.Equal(t, int64(2002), items[2].NetworkAreaID)
	assert.Equal(t, "127.0.0.2", items[2].IP)
}

func TestTopoRecommendNetworkUnitByNetworkSegmentRespConvertResultsFromTypesPreservesOrderDuplicatesAndMessagesForApplication(t *testing.T) {
	resp := &TopoRecommendNetworkUnitByNetworkSegmentResp{}
	resp.ConvertResultsFromTypes([]*types.NetworkUnitSegmentRecommendationResult{
		{NetworkAreaID: 1001, IP: "127.0.0.1", NetworkUnitID: 200101, Message: "matched"},
		{NetworkAreaID: 1001, IP: "127.0.0.1", NetworkUnitID: -1, Message: "no match"},
		{NetworkAreaID: 2002, IP: "bad-ip", NetworkUnitID: -1, Message: "invalid ip"},
	})

	require.NotNil(t, resp.Data)
	require.Len(t, resp.Data.Items, 3)

	assert.Equal(t, int64(1001), resp.Data.Items[0].GetBkNetworkareaId())
	assert.Equal(t, "127.0.0.1", resp.Data.Items[0].GetIp())
	assert.Equal(t, int64(200101), resp.Data.Items[0].GetBkNetworkunitId())
	assert.Equal(t, "matched", resp.Data.Items[0].GetMessage())

	assert.Equal(t, int64(1001), resp.Data.Items[1].GetBkNetworkareaId())
	assert.Equal(t, "127.0.0.1", resp.Data.Items[1].GetIp())
	assert.Equal(t, int64(-1), resp.Data.Items[1].GetBkNetworkunitId())
	assert.Equal(t, "no match", resp.Data.Items[1].GetMessage())

	assert.Equal(t, int64(2002), resp.Data.Items[2].GetBkNetworkareaId())
	assert.Equal(t, "bad-ip", resp.Data.Items[2].GetIp())
	assert.Equal(t, int64(-1), resp.Data.Items[2].GetBkNetworkunitId())
	assert.Equal(t, "invalid ip", resp.Data.Items[2].GetMessage())
}
