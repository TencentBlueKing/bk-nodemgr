package proxy

import (
	"slices"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestBuildBizResources(t *testing.T) {
	resources := buildBizResources([]int64{2, 4})
	if len(resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(resources))
	}

	for idx, resource := range resources {
		if resource.SystemID != types.SystemIDCMDB {
			t.Fatalf("resource %d system id = %s, want %s", idx, resource.SystemID, types.SystemIDCMDB)
		}
		if resource.Type != types.AuthResourceTypeBiz {
			t.Fatalf("resource %d type = %s, want %s", idx, resource.Type, types.AuthResourceTypeBiz)
		}
	}

	if resources[0].ID != "2" || resources[1].ID != "4" {
		t.Fatalf("unexpected resource ids: %+v", resources)
	}
}

func TestBuildBizIDsFromTypeHosts(t *testing.T) {
	hosts := []*types.Host{
		{Static: &types.HostStatic{BizID: 3}},
		{Static: &types.HostStatic{BizID: 1}},
		{Static: &types.HostStatic{BizID: 3}},
	}

	bizIDs := buildBizIDsFromTypeHosts(hosts)
	slices.Sort(bizIDs)

	if !slices.Equal(bizIDs, []int64{1, 3}) {
		t.Fatalf("unexpected biz ids: %v", bizIDs)
	}
}

func TestBuildNetworkUnitResources(t *testing.T) {
	resources := buildNetworkUnitResources([]int64{10, 20})
	if len(resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(resources))
	}

	for idx, resource := range resources {
		if resource.SystemID != types.SystemIDNodeMgr {
			t.Fatalf("resource %d system id = %s, want %s", idx, resource.SystemID, types.SystemIDNodeMgr)
		}
		if resource.Type != types.AuthResourceTypeNetworkUnit {
			t.Fatalf("resource %d type = %s, want %s", idx, resource.Type, types.AuthResourceTypeNetworkUnit)
		}
	}

	if resources[0].ID != "10" || resources[1].ID != "20" {
		t.Fatalf("unexpected resource ids: %+v", resources)
	}
}
