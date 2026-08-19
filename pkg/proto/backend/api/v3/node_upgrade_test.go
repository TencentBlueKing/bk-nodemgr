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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestNodeAgentUpgradeReqConvertParamFromTypesPreservesNewFields(t *testing.T) {
	upgradeHost := &types.NodeAgentUpgradeHost{
		HostID:                 2001,
		TargetVersion:          "7.1.0",
		Force:                  true,
		GracefulRestartTimeout: 30 * time.Second,
	}
	setInt64Field(t, upgradeHost, "NetworkUnitID", 4001)
	setStringField(t, upgradeHost, "CPUArch", "amd64")

	req := &NodeAgentUpgradeReq{}
	req.ConvertParamFromTypes(&types.NodeAgentUpgradeParam{
		Hosts: []*types.NodeAgentUpgradeHost{upgradeHost},
	})

	if got := req.GetHost()[0].GetBkNetworkunitId(); got != 4001 {
		t.Fatalf("expected bk_networkunit_id 4001, got %d", got)
	}
	if got := req.GetHost()[0].GetCpuArch(); got != "amd64" {
		t.Fatalf("expected cpu_arch amd64, got %q", got)
	}
}

func TestNodeProxyUpgradeReqConvertParamFromTypesPreservesNewFields(t *testing.T) {
	upgradeHost := &types.NodeProxyUpgradeHost{
		HostID:                 2002,
		Force:                  true,
		GracefulRestartTimeout: 45 * time.Second,
	}
	setInt64Field(t, upgradeHost, "NetworkUnitID", 5001)
	setStringField(t, upgradeHost, "CPUArch", "arm64")

	req := &NodeProxyUpgradeReq{}
	req.ConvertParamFromTypes(&types.NodeProxyUpgradeParam{
		Hosts: []*types.NodeProxyUpgradeHost{upgradeHost},
		TargetVersion: []*types.TargetVersion{
			{
				OsType:  "linux",
				CPUArch: "arm64",
				Version: "7.1.0",
			},
		},
	})

	if got := req.GetHost()[0].GetBkNetworkunitId(); got != 5001 {
		t.Fatalf("expected bk_networkunit_id 5001, got %d", got)
	}
	if got := req.GetHost()[0].GetCpuArch(); got != "arm64" {
		t.Fatalf("expected cpu_arch arm64, got %q", got)
	}
}

func setInt64Field(t *testing.T, target any, fieldName string, value int64) {
	t.Helper()

	field := reflect.ValueOf(target).Elem().FieldByName(fieldName)
	if !field.IsValid() {
		t.Fatalf("expected field %s to exist", fieldName)
	}
	field.SetInt(value)
}

func setStringField(t *testing.T, target any, fieldName string, value string) {
	t.Helper()

	field := reflect.ValueOf(target).Elem().FieldByName(fieldName)
	if !field.IsValid() {
		t.Fatalf("expected field %s to exist", fieldName)
	}
	field.SetString(value)
}
