package v3

import (
	"reflect"
	"testing"
)

func TestNodeAgentUpgradeReqConvertParamToTypesPreservesNewFields(t *testing.T) {
	networkUnitID := int64(2001)
	req := &NodeAgentUpgradeReq{
		Host: []*NodeAgentUpgradeHost{
			{
				BkHostId:        1001,
				TargetVersion:   "7.1.0",
				Force:           true,
				BkNetworkunitId: &networkUnitID,
				CpuArch:         "amd64",
			},
		},
	}

	param := req.ConvertParamToTypes()
	if param == nil || len(param.Hosts) != 1 {
		t.Fatalf("expected one converted host, got %#v", param)
	}

	hostValue := reflect.ValueOf(param.Hosts[0]).Elem()

	networkUnitField := hostValue.FieldByName("NetworkUnitID")
	if !networkUnitField.IsValid() {
		t.Fatalf("expected converted agent upgrade host to expose NetworkUnitID")
	}
	if got := networkUnitField.Int(); got != networkUnitID {
		t.Fatalf("expected NetworkUnitID %d, got %d", networkUnitID, got)
	}

	cpuArchField := hostValue.FieldByName("CPUArch")
	if !cpuArchField.IsValid() {
		t.Fatalf("expected converted agent upgrade host to expose CPUArch")
	}
	if got := cpuArchField.String(); got != "amd64" {
		t.Fatalf("expected CPUArch amd64, got %q", got)
	}
}

func TestNodeProxyUpgradeReqConvertParamToTypesPreservesNewFields(t *testing.T) {
	networkUnitID := int64(3001)
	req := &NodeProxyUpgradeReq{
		Host: []*NodeProxyUpgradeHost{
			{
				BkHostId:        1002,
				Force:           true,
				BkNetworkunitId: &networkUnitID,
				CpuArch:         "arm64",
			},
		},
	}

	param := req.ConvertParamToTypes()
	if param == nil || len(param.Hosts) != 1 {
		t.Fatalf("expected one converted host, got %#v", param)
	}

	hostValue := reflect.ValueOf(param.Hosts[0]).Elem()

	networkUnitField := hostValue.FieldByName("NetworkUnitID")
	if !networkUnitField.IsValid() {
		t.Fatalf("expected converted proxy upgrade host to expose NetworkUnitID")
	}
	if got := networkUnitField.Int(); got != networkUnitID {
		t.Fatalf("expected NetworkUnitID %d, got %d", networkUnitID, got)
	}

	cpuArchField := hostValue.FieldByName("CPUArch")
	if !cpuArchField.IsValid() {
		t.Fatalf("expected converted proxy upgrade host to expose CPUArch")
	}
	if got := cpuArchField.String(); got != "arm64" {
		t.Fatalf("expected CPUArch arm64, got %q", got)
	}
}
