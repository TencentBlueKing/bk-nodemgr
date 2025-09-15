/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package types

import (
	"fmt"
)

// NodeAgentInstallHost describes the node agent install host.
type NodeAgentInstallHost struct {
	BizID         int64
	InnerIP       string
	InnerIPV6     string
	Addressing    Addressing
	LoginIP       string
	LoginPort     int64
	LoginUser     string
	LoginMode     LoginMode
	LoginPassword string
	LoginKeyFile  string
	NetworkUnitID int64
	OSType        string
}

// NodeAgentInstallParam describes the node agent install parameter.
type NodeAgentInstallParam struct {
	NodeAgentInstallHosts       []*NodeAgentInstallHost
	NodeInstallTargetVersion    []*TargetVersion
	DisableDefaultTargetVersion bool
}

// NodeOperationRetryParam validates the node install parameter.
type NodeOperationRetryParam struct {
	WorkflowID   string
	OperationIDs []string
	RetryMode    NodeOperationRetryMode
}

// NodeOperationRetryMode describes the node operation mode.
type NodeOperationRetryMode string

const (
	// OperationRetryModeFull is the full node instance retry mode.
	OperationRetryModeFull NodeOperationRetryMode = "full_node_instance_retry"

	// OperationRetryModePartial is the partial node instance retry mode.
	OperationRetryModePartial NodeOperationRetryMode = "partial_node_instance_retry"
)

// Validate validates the node operation retry mode.
func (mode NodeOperationRetryMode) Validate() error {
	switch mode {
	case OperationRetryModeFull, OperationRetryModePartial:
		return nil
	default:
		return fmt.Errorf("invalid node operation retry mode. mode(%s)", mode)
	}
}

// NodeAgentInstallEligibility describes the node agent install eligibility.
type NodeAgentInstallEligibility string

const (
	// NodeAgentInstallEligibilityConflictIP indicates there is a conflicting IP under the same business context.
	NodeAgentInstallEligibilityConflictIP NodeAgentInstallEligibility = "conflict_ip"

	// NodeAgentInstallEligibilityDuplicateIP indicates there are duplicate IPs in dynamic addressing mode.
	NodeAgentInstallEligibilityDuplicateIP NodeAgentInstallEligibility = "duplicate_dynamic_ip"

	// NodeAgentInstallEligibilityExistProxy indicates a proxy already exists on this node.
	NodeAgentInstallEligibilityExistProxy NodeAgentInstallEligibility = "exist_proxy"

	// NodeAgentInstallEligibilityExistAgent indicates an agent already exists on this node.
	NodeAgentInstallEligibilityExistAgent NodeAgentInstallEligibility = "exist_agent"

	// NodeAgentInstallEligibilityNormal indicates a normal node agent installation; the host can be reused.
	NodeAgentInstallEligibilityNormal NodeAgentInstallEligibility = "normal_install"

	// NodeAgentInstallEligibilityClean indicates a clean node agent installation and can be imported into CMDB.
	NodeAgentInstallEligibilityClean NodeAgentInstallEligibility = "clean_install"
)

// NodeAgentInstallCheckResult describes the node agent install check result.
type NodeAgentInstallCheckResult struct {
	InnerIP             string
	InstallEligibilitiy NodeAgentInstallEligibility
	DuplicateHostIDs    []int64
}
