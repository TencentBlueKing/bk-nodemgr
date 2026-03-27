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

func buildNetworkUnitResources(networkUnitIDs []int64) []auth.Resource {
	resources := make([]auth.Resource, 0, len(networkUnitIDs))
	for _, id := range networkUnitIDs {
		resources = append(resources, auth.Resource{
			SystemID: auth.SystemIDNodeMgr,
			Type:     auth.ResourceTypeNetworkUnit,
			ID:       fmt.Sprintf("%d", id),
		})
	}

	return resources
}
