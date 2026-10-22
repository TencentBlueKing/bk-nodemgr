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

package firewall

import (
	"context"
	"reflect"
	"runtime"
	"testing"
)

func TestParseFeatures(t *testing.T) {
	features, err := parseFeatures(" nodemgr,gse,gse ")
	if err != nil {
		t.Fatalf("parseFeatures() error = %v", err)
	}

	if want := []string{"gse", "nodemgr"}; !reflect.DeepEqual(features, want) {
		t.Fatalf("parseFeatures() = %v, want %v", features, want)
	}
}

func TestParseRemoteIPs(t *testing.T) {
	remoteIPs, err := parseRemoteIPs("10.0.0.1/24, fd00::1/8, 10.0.0.0/24")
	if err != nil {
		t.Fatalf("parseRemoteIPs() error = %v", err)
	}

	if want := []string{"10.0.0.0/24", "fd00::/8"}; !reflect.DeepEqual(remoteIPs, want) {
		t.Fatalf("parseRemoteIPs() = %v, want %v", remoteIPs, want)
	}
}

func TestBuildRules(t *testing.T) {
	rules, err := buildRules("prod", []string{featureGSE, featureNodeMgr}, []string{"10.0.0.0/24", "fd00::/8"})
	if err != nil {
		t.Fatalf("buildRules() error = %v", err)
	}

	if len(rules) != 10 {
		t.Fatalf("buildRules() returned %d rules, want 10", len(rules))
	}
	if rules[0].Name != "prod_Inbound_GSE2_Rule_TCP_IPv4" || rules[0].RemoteIPs != "10.0.0.0/24" {
		t.Fatalf("unexpected first rule: %+v", rules[0])
	}
	if rules[1].Name != "prod_Inbound_GSE2_Rule_TCP_IPv6" || rules[1].RemoteIPs != "fd00::/8" {
		t.Fatalf("unexpected second rule: %+v", rules[1])
	}
}

func TestBuildRulesRequiresRemoteIPs(t *testing.T) {
	_, err := buildRules("prod", []string{featureGSE}, nil)
	if err != nil {
		return
	}
	t.Fatal("buildRules() did not return an error")
}

func TestRunFirewallSkipsInvalidConfigOnNonWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("firewall configuration is applied on Windows")
	}

	if err := NewStep(StepArgs{Features: "unsupported", RemoteIPs: "invalid"}).Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v, want nil on non-Windows", err)
	}
}
