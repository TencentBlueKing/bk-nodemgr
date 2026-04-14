package agent

import (
	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// buildBizResources is deprecated. Use auth.BuildBizResources instead.
func buildBizResources(bizIDs []int64) []types.AuthResource {
	return authRouter.BuildBizResources(bizIDs...)
}

// buildNetworkUnitResources is deprecated. Use auth.BuildNetworkUnitResources instead.
func buildNetworkUnitResources(networkUnitIDs []int64) []types.AuthResource {
	return authRouter.BuildNetworkUnitResources(networkUnitIDs)
}
