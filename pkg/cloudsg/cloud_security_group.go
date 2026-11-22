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

import "strings"

// CloudSecurityGroupVendor is the cloud vendor identifier of a security group pool.
// The pool configuration uses it to select the concrete provider implementation.
type CloudSecurityGroupVendor string

// CloudSecurityGroupRuleAction is the decision of a security group rule.
type CloudSecurityGroupRuleAction string

const (
	// CloudSecurityGroupRuleActionAccept allows matching traffic.
	CloudSecurityGroupRuleActionAccept CloudSecurityGroupRuleAction = "ACCEPT"
	// CloudSecurityGroupRuleActionDrop denies matching traffic.
	CloudSecurityGroupRuleActionDrop CloudSecurityGroupRuleAction = "DROP"
)

// CloudSecurityGroupRule is one managed security group ingress rule.
type CloudSecurityGroupRule struct {
	// CIDR is the authorized source CIDR.
	CIDR string
	// Port is the destination port or port range.
	Port string
	// Protocol is the IP protocol.
	Protocol string
	// Action is the rule decision, ACCEPT or DROP. Empty defaults to ACCEPT.
	Action CloudSecurityGroupRuleAction
	// Remark is an optional remark appended to the managed rule description,
	// for example the network area id, to help operators locate rules in the cloud console.
	Remark string
}

// EffectiveAction returns the uppercase rule decision, defaulting to ACCEPT when empty.
func (rule CloudSecurityGroupRule) EffectiveAction() CloudSecurityGroupRuleAction {
	if rule.Action == "" {
		return CloudSecurityGroupRuleActionAccept
	}

	return CloudSecurityGroupRuleAction(strings.ToUpper(string(rule.Action)))
}

// CloudSecurityGroupInspection is the live inspection result of one security group.
type CloudSecurityGroupInspection struct {
	SecurityGroupID string
	// Rules are all live ingress rules of the security group. The whole pool is managed,
	// user rules and platform written rules are not distinguished.
	Rules        []CloudSecurityGroupRule
	TotalIngress int
}

// CloudSecurityGroupPoolSummary is the public metadata of one global pool.
type CloudSecurityGroupPoolSummary struct {
	PoolID   string
	PoolName string
}
