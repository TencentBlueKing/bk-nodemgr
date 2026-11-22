/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package cloudsg

import "testing"

func TestCloudSecurityGroupRuleCovers(t *testing.T) {
	target := CloudSecurityGroupRule{CIDR: "1.2.3.4/32", Port: "28625,28668", Protocol: "TCP"}

	tests := []struct {
		name     string
		covering CloudSecurityGroupRule
		want     bool
	}{
		{
			name:     "exact match covers",
			covering: CloudSecurityGroupRule{CIDR: "1.2.3.4/32", Port: "28625,28668", Protocol: "TCP"},
			want:     true,
		},
		{
			name:     "drop rule does not cover accept target",
			covering: CloudSecurityGroupRule{CIDR: "1.2.3.4/32", Port: "28625,28668", Protocol: "TCP", Action: "DROP"},
			want:     false,
		},
		{
			name:     "broader cidr covers",
			covering: CloudSecurityGroupRule{CIDR: "1.2.3.0/24", Port: "28625,28668", Protocol: "TCP"},
			want:     true,
		},
		{
			name:     "narrower cidr does not cover",
			covering: CloudSecurityGroupRule{CIDR: "1.2.3.5/32", Port: "28625,28668", Protocol: "TCP"},
			want:     false,
		},
		{
			name:     "all port covers discrete ports",
			covering: CloudSecurityGroupRule{CIDR: "1.2.3.4/32", Port: "ALL", Protocol: "TCP"},
			want:     true,
		},
		{
			name:     "port range covers discrete ports",
			covering: CloudSecurityGroupRule{CIDR: "1.2.3.4/32", Port: "28000-29000", Protocol: "TCP"},
			want:     true,
		},
		{
			name:     "partial port coverage does not cover",
			covering: CloudSecurityGroupRule{CIDR: "1.2.3.4/32", Port: "28625", Protocol: "TCP"},
			want:     false,
		},
		{
			name:     "all protocol covers tcp",
			covering: CloudSecurityGroupRule{CIDR: "1.2.3.4/32", Port: "28625,28668", Protocol: "ALL"},
			want:     true,
		},
		{
			name:     "udp does not cover tcp",
			covering: CloudSecurityGroupRule{CIDR: "1.2.3.4/32", Port: "28625,28668", Protocol: "UDP"},
			want:     false,
		},
		{
			name:     "empty protocol defaults to tcp",
			covering: CloudSecurityGroupRule{CIDR: "1.2.3.4/32", Port: "28625,28668"},
			want:     true,
		},
		{
			name:     "all target is only covered by all",
			covering: CloudSecurityGroupRule{CIDR: "0.0.0.0/0", Port: "ALL", Protocol: "ALL"},
			want:     true,
		},
		{
			name:     "invalid cidr never covers",
			covering: CloudSecurityGroupRule{CIDR: "not-a-cidr", Port: "ALL", Protocol: "ALL"},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CloudSecurityGroupRuleCovers(tt.covering, target); got != tt.want {
				t.Fatalf("covers = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCloudSecurityGroupRuleCoversAllTarget(t *testing.T) {
	target := CloudSecurityGroupRule{CIDR: "1.2.3.4/32", Port: "ALL", Protocol: "TCP"}
	covering := CloudSecurityGroupRule{CIDR: "1.2.3.4/32", Port: "28000-29000", Protocol: "TCP"}
	if CloudSecurityGroupRuleCovers(covering, target) {
		t.Fatal("port range should not cover ALL target")
	}
}

func TestCloudSecurityGroupRuleCoveredByUnion(t *testing.T) {
	target := CloudSecurityGroupRule{CIDR: "1.2.3.4/32", Port: "28625,28668,28925", Protocol: "TCP"}

	tests := []struct {
		name     string
		existing []CloudSecurityGroupRule
		want     bool
	}{
		{
			name: "scattered single port rules jointly cover combined target",
			existing: []CloudSecurityGroupRule{
				{CIDR: "1.2.3.4/32", Port: "28625", Protocol: "TCP"},
				{CIDR: "1.2.3.4/32", Port: "28668", Protocol: "TCP"},
				{CIDR: "1.2.3.4/32", Port: "28925", Protocol: "TCP"},
			},
			want: true,
		},
		{
			name: "drop rule does not contribute to accept coverage",
			existing: []CloudSecurityGroupRule{
				{CIDR: "1.2.3.4/32", Port: "28625,28668,28925", Protocol: "TCP", Action: "DROP"},
			},
			want: false,
		},
		{
			name: "union missing one port does not cover",
			existing: []CloudSecurityGroupRule{
				{CIDR: "1.2.3.4/32", Port: "28625", Protocol: "TCP"},
				{CIDR: "1.2.3.4/32", Port: "28668", Protocol: "TCP"},
			},
			want: false,
		},
		{
			name: "union with wrong cidr does not cover",
			existing: []CloudSecurityGroupRule{
				{CIDR: "1.2.3.5/32", Port: "28625", Protocol: "TCP"},
				{CIDR: "1.2.3.4/32", Port: "28668", Protocol: "TCP"},
				{CIDR: "1.2.3.4/32", Port: "28925", Protocol: "TCP"},
			},
			want: false,
		},
		{
			name: "union across broader cidr and single port",
			existing: []CloudSecurityGroupRule{
				{CIDR: "1.2.3.0/24", Port: "28625", Protocol: "TCP"},
				{CIDR: "1.2.3.4/32", Port: "28668,28925", Protocol: "TCP"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CloudSecurityGroupRuleCoveredBy(tt.existing, target); got != tt.want {
				t.Fatalf("covered by union = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCloudSecurityGroupRuleEffectiveAction(t *testing.T) {
	if got := (CloudSecurityGroupRule{}).EffectiveAction(); got != CloudSecurityGroupRuleActionAccept {
		t.Fatalf("empty action = %q, want %q", got, CloudSecurityGroupRuleActionAccept)
	}
	if got := (CloudSecurityGroupRule{Action: "drop"}).EffectiveAction(); got != "DROP" {
		t.Fatalf("lowercase action = %q, want DROP", got)
	}
}

func TestCloudSecurityGroupRuleCoveredByUnionRanges(t *testing.T) {
	tests := []struct {
		name     string
		covering CloudSecurityGroupRule
		target   CloudSecurityGroupRule
		want     bool
	}{
		{
			name:     "contiguous ranges within one rule jointly cover",
			covering: CloudSecurityGroupRule{CIDR: "1.2.3.4/32", Port: "80-81,82", Protocol: "TCP"},
			target:   CloudSecurityGroupRule{CIDR: "1.2.3.4/32", Port: "80-82", Protocol: "TCP"},
			want:     true,
		},
		{
			name:     "overlapping ranges within one rule jointly cover",
			covering: CloudSecurityGroupRule{CIDR: "1.2.3.4/32", Port: "80-82,82-84", Protocol: "TCP"},
			target:   CloudSecurityGroupRule{CIDR: "1.2.3.4/32", Port: "81-83", Protocol: "TCP"},
			want:     true,
		},
		{
			name:     "gap between ranges does not cover",
			covering: CloudSecurityGroupRule{CIDR: "1.2.3.4/32", Port: "80-81,83-84", Protocol: "TCP"},
			target:   CloudSecurityGroupRule{CIDR: "1.2.3.4/32", Port: "80-84", Protocol: "TCP"},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CloudSecurityGroupRuleCovers(tt.covering, tt.target); got != tt.want {
				t.Fatalf("covers = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCloudSecurityGroupRuleCoveredByScatteredRanges(t *testing.T) {
	target := CloudSecurityGroupRule{CIDR: "1.2.3.4/32", Port: "80-82", Protocol: "TCP"}

	tests := []struct {
		name     string
		existing []CloudSecurityGroupRule
		want     bool
	}{
		{
			name: "scattered ranges across rules jointly cover",
			existing: []CloudSecurityGroupRule{
				{CIDR: "1.2.3.4/32", Port: "80-81", Protocol: "TCP"},
				{CIDR: "1.2.3.4/32", Port: "82", Protocol: "TCP"},
			},
			want: true,
		},
		{
			name: "bridging range with wrong cidr does not cover",
			existing: []CloudSecurityGroupRule{
				{CIDR: "1.2.3.4/32", Port: "80-81", Protocol: "TCP"},
				{CIDR: "9.9.9.9/32", Port: "82", Protocol: "TCP"},
			},
			want: false,
		},
		{
			name: "overlapping ranges across rules jointly cover",
			existing: []CloudSecurityGroupRule{
				{CIDR: "1.2.3.4/32", Port: "80-81", Protocol: "TCP"},
				{CIDR: "1.2.3.4/32", Port: "81-82", Protocol: "TCP"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CloudSecurityGroupRuleCoveredBy(tt.existing, target); got != tt.want {
				t.Fatalf("covered by = %v, want %v", got, tt.want)
			}
		})
	}
}
