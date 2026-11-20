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

package tencentcloud

import "fmt"

const (
	// Vendor is the provider identifier of Tencent Cloud, used by pool configuration
	// to select this handler implementation.
	Vendor = "tencent"
	// ManagedRuleDescription marks rules managed by bk-nodemgr.
	ManagedRuleDescription = "bknm"
	ruleProtocolTCP        = "TCP"
	ruleActionAccept       = "ACCEPT"
)

// SecurityGroupPool contains the credential and target groups of one pool.
type SecurityGroupPool struct {
	Region string
	// Endpoint is the optional custom api endpoint, empty uses the vendor public default.
	Endpoint         string
	SecretID         string
	SecretKey        string
	SecurityGroupIDs []string
}

// QuotaExceededError reports that a security group rule quota is exhausted.
type QuotaExceededError struct {
	SecurityGroupID string
	Err             error
}

// Error implements error.
func (e *QuotaExceededError) Error() string {
	return fmt.Sprintf("security group %s rule quota exceeded: %v", e.SecurityGroupID, e.Err)
}

// Unwrap implements errors.Unwrap.
func (e *QuotaExceededError) Unwrap() error {
	return e.Err
}
