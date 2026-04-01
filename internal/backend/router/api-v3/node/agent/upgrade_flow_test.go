package agent

import (
	"context"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestValidateUpgradeHostNetworkUnit_UsesTargetUnitWhenCurrentInvalid(t *testing.T) {
	hostID := int64(100)
	targetUnitID := int64(10)
	hosts := map[int64]*types.Host{
		hostID: newTestHost(hostID, 3, -1),
	}

	err := validateUpgradeHostNetworkUnit([]*protoBackend.NodeAgentUpgradeReq_Host{
		{
			BkHostId:        hostID,
			BkNetworkunitId: &targetUnitID,
		},
	}, hosts)
	if err != nil {
		t.Fatalf("expected target network unit to satisfy validation, got %v", err)
	}
}

func TestCheckUpgrade_CurrentNetworkUnitInvalidWithoutTargetReturnsNotFound(t *testing.T) {
	hostID := int64(100)
	host := newTestHost(hostID, 3, -1)
	host.Static.NetworkAreaID = 1
	host.Dynamic.NodeOsType = "linux"
	host.Dynamic.NodeRole = types.NodeRoleAgent
	h := &handler{
		storageHost:        &fakeStorageHost{listHosts: []*types.Host{host}},
		storageNetworkUnit: &fakeStorageNetworkUnit{},
	}

	results, err := h.checkUpgrade(
		contextx.New(context.Background()),
		[]*protoBackend.NodeAgentUpgradeCheckReq_Host{{BkHostId: &hostID}},
		nil,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != types.NodeAgentUpgradeCheckStatusNetworkUnitNotFound {
		t.Fatalf("expected status %q, got %q", types.NodeAgentUpgradeCheckStatusNetworkUnitNotFound, results[0].Status)
	}
}

func TestCheckUpgrade_CurrentNetworkUnitInvalidWithTargetAllowsChange(t *testing.T) {
	hostID := int64(100)
	targetUnitID := int64(10)
	host := newTestHost(hostID, 3, -1)
	host.Static.NetworkAreaID = 1
	host.Dynamic.NodeOsType = "linux"
	host.Dynamic.NodeRole = types.NodeRoleAgent
	h := &handler{
		storageHost: &fakeStorageHost{listHosts: []*types.Host{host}},
		storageNetworkUnit: &fakeStorageNetworkUnit{listUnits: []*types.NetworkUnit{
			{ID: targetUnitID, NetworkAreaID: 1},
		}},
	}

	results, err := h.checkUpgrade(
		contextx.New(context.Background()),
		[]*protoBackend.NodeAgentUpgradeCheckReq_Host{{
			BkHostId:        &hostID,
			BkNetworkunitId: &targetUnitID,
		}},
		nil,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != types.NodeAgentUpgradeCheckStatusNetworkUnitChanged {
		t.Fatalf("expected status %q, got %q", types.NodeAgentUpgradeCheckStatusNetworkUnitChanged, results[0].Status)
	}
}
