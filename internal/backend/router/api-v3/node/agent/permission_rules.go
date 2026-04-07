package agent

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	routerAuth "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
)

// buildBizResources is deprecated. Use auth.BuildBizResources instead.
func buildBizResources(bizIDs []int64) []auth.Resource {
	return routerAuth.BuildBizResources(bizIDs)
}

// buildNetworkUnitResources is deprecated. Use auth.BuildNetworkUnitResources instead.
func buildNetworkUnitResources(networkUnitIDs []int64) []auth.Resource {
	return routerAuth.BuildNetworkUnitResources(networkUnitIDs)
}
