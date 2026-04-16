package v3

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestTopoRecommendNetworkUnitByNetworkSegmentReqConvertItemsToTypesPreservesOrderAndDuplicates(t *testing.T) {
	req := &TopoRecommendNetworkUnitByNetworkSegmentReq{
		Items: []*TopoRecommendNetworkUnitByNetworkSegmentReq_Item{
			{BkNetworkareaId: 1001, Ip: "10.0.0.1"},
			{BkNetworkareaId: 1001, Ip: "10.0.0.1"},
			{BkNetworkareaId: 2002, Ip: "172.16.0.8"},
		},
	}

	items := req.ConvertItemsToTypes()
	require.Len(t, items, 3)

	assert.Equal(t, int64(1001), items[0].NetworkAreaID)
	assert.Equal(t, "10.0.0.1", items[0].IP)
	assert.Equal(t, int64(1001), items[1].NetworkAreaID)
	assert.Equal(t, "10.0.0.1", items[1].IP)
	assert.Equal(t, int64(2002), items[2].NetworkAreaID)
	assert.Equal(t, "172.16.0.8", items[2].IP)
}

func TestTopoRecommendNetworkUnitByNetworkSegmentRespConvertResultsFromTypesPreservesOrderDuplicatesAndMessages(t *testing.T) {
	resp := &TopoRecommendNetworkUnitByNetworkSegmentResp{}
	resp.ConvertResultsFromTypes([]*types.NetworkUnitSegmentRecommendationResult{
		{NetworkAreaID: 1001, IP: "10.0.0.1", NetworkUnitID: 200101, Message: "matched"},
		{NetworkAreaID: 1001, IP: "10.0.0.1", NetworkUnitID: -1, Message: "no match"},
		{NetworkAreaID: 2002, IP: "bad-ip", NetworkUnitID: -1, Message: "invalid ip"},
	})

	require.NotNil(t, resp.Data)
	require.Len(t, resp.Data.Items, 3)

	assert.Equal(t, int64(1001), resp.Data.Items[0].GetBkNetworkareaId())
	assert.Equal(t, "10.0.0.1", resp.Data.Items[0].GetIp())
	assert.Equal(t, int64(200101), resp.Data.Items[0].GetBkNetworkunitId())
	assert.Equal(t, "matched", resp.Data.Items[0].GetMessage())

	assert.Equal(t, int64(1001), resp.Data.Items[1].GetBkNetworkareaId())
	assert.Equal(t, "10.0.0.1", resp.Data.Items[1].GetIp())
	assert.Equal(t, int64(-1), resp.Data.Items[1].GetBkNetworkunitId())
	assert.Equal(t, "no match", resp.Data.Items[1].GetMessage())

	assert.Equal(t, int64(2002), resp.Data.Items[2].GetBkNetworkareaId())
	assert.Equal(t, "bad-ip", resp.Data.Items[2].GetIp())
	assert.Equal(t, int64(-1), resp.Data.Items[2].GetBkNetworkunitId())
	assert.Equal(t, "invalid ip", resp.Data.Items[2].GetMessage())
}
