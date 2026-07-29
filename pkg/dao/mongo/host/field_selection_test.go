package host

import (
	"reflect"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestConvertHostFieldSelectionToFields_DynamicAgentState(t *testing.T) {
	fields := convertHostFieldSelectionToFields(&types.HostFieldSelection{
		HostID:         true,
		AgentID:        true,
		NodeRole:       true,
		NodeStatus:     true,
		NodeVersion:    true,
		NodeGeneration: true,
	})
	expected := []string{
		FieldKeyHostID,
		FieldKeyDynamicNodeRole,
		FieldKeyDynamicNodeStatus,
		FieldKeyDynamicNodeVersion,
		FieldKeyDynamicNodeGeneration,
		FieldKeyDynamicAgentID,
	}

	if !reflect.DeepEqual(fields, expected) {
		t.Fatalf("fields = %#v, want %#v", fields, expected)
	}
}
