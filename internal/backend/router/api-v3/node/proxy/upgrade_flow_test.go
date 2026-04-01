package proxy

import (
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"testing"
)

func TestValidateUpgradeHostNetworkUnit_UsesTargetUnitWhenCurrentInvalid(t *testing.T) {
	hostID := int64(100)
	targetUnitID := int64(10)
	hosts := map[int64]*types.Host{
		hostID: &types.Host{
			HostID: hostID,
			Static: &types.HostStatic{BizID: 3},
			Dynamic: &types.HostDynamic{
				NetworkUnitID: -1,
			},
		},
	}

	err := validateUpgradeHostNetworkUnit([]*protoBackend.NodeProxyUpgradeReq_Host{
		{
			BkHostId:        hostID,
			BkNetworkunitId: &targetUnitID,
		},
	}, hosts)
	if err != nil {
		t.Fatalf("expected target network unit to satisfy validation, got %v", err)
	}
}
