package agent

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
)

// buildBizResources is deprecated. Use auth.BuildBizResources instead.
func buildBizResources(bizIDs []int64) []auth.Resource {
	return authRouter.BuildBizResources(bizIDs)
}

// buildNetworkUnitResources is deprecated. Use auth.BuildNetworkUnitResources instead.
func buildNetworkUnitResources(networkUnitIDs []int64) []auth.Resource {
	return authRouter.BuildNetworkUnitResources(networkUnitIDs)
}
