package proxy

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// buildBizResources is deprecated. Use auth.BuildBizResources instead.
func buildBizResources(bizIDs []int64) []auth.Resource {
	return authRouter.BuildBizResources(bizIDs)
}

func buildBizIDsFromTypeHosts(hosts []*types.Host) []int64 {
	bizIDMap := make(map[int64]struct{})
	for _, host := range hosts {
		if host == nil || host.Static == nil {
			continue
		}
		bizIDMap[host.Static.BizID] = struct{}{}
	}

	return conv.MapKeyToSlice(bizIDMap)
}

// buildNetworkUnitResources is deprecated. Use auth.BuildNetworkUnitResources instead.
func buildNetworkUnitResources(networkUnitIDs []int64) []auth.Resource {
	return authRouter.BuildNetworkUnitResources(networkUnitIDs)
}
