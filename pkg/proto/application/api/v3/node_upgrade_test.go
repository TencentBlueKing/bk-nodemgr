/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

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
		TargetVersion: []*TargetVersion{
			{
				OsType:  "linux",
				CpuArch: "arm64",
				Version: "7.1.0",
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
