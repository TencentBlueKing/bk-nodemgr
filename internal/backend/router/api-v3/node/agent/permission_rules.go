package agent

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
)

func buildBizResources(bizIDs []int64) []auth.Resource {
	resources := make([]auth.Resource, 0, len(bizIDs))
	for _, bizID := range bizIDs {
		resources = append(resources, auth.Resource{SystemID: auth.SystemIDCMDB, Type: auth.ResourceTypeBiz, ID: fmt.Sprintf("%d", bizID)})
	}

	return resources
}

func buildNetworkUnitResources(networkUnitID int64) []auth.Resource {
	return []auth.Resource{{
		SystemID: auth.SystemIDNodeMgr,
		Type:     auth.ResourceTypeNetworkUnit,
		ID:       fmt.Sprintf("%d", networkUnitID),
	}}
}
