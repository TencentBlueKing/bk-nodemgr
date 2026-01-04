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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
)

// NodeAgentInstallHost describes the node agent install host.
type NodeAgentInstallHost struct {
	HostID        int64
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
	NodeAgentInstallHosts    []*NodeAgentInstallHost
	NodeInstallTargetVersion []*TargetVersion
	IsManual                 bool
}

// NodeAgentUpgradeHost describes the node agent upgrade host.
type NodeAgentUpgradeHost struct {
	HostID                 int64
	Force                  bool
	TargetVersion          string
	GracefulRestartTimeout time.Duration
}

// NodeAgentUpgradeParam describes the node agent upgrade parameter.
type NodeAgentUpgradeParam struct {
	Hosts []*NodeAgentUpgradeHost
}

// NodeAgentRestartHost describes the node agent restart host.
type NodeAgentRestartHost struct {
	HostID                 int64
	Force                  bool
	GracefulRestartTimeout time.Duration
}

// NodeAgentRestartParam describes the node agent restart parameter.
type NodeAgentRestartParam struct {
	Hosts []*NodeAgentRestartHost
}

// NodeAgentReconfigHost describes the node agent reconfig host.
type NodeAgentReconfigHost struct {
	HostID                 int64
	Force                  bool
	GracefulRestartTimeout time.Duration
}

// NodeAgentReconfigParam describes the node agent reconfig parameter.
type NodeAgentReconfigParam struct {
	Hosts []*NodeAgentReconfigHost
}

// NodeAgentUninstallHost describes the node agent uninstall host.
type NodeAgentUninstallHost struct {
	HostID int64
}

// NodeAgentUninstallParam describes the node agent uninstall parameter.
type NodeAgentUninstallParam struct {
	Hosts []*NodeAgentUninstallHost
}

// NodeAgentInstallCheckParam describes the node agent install check param.
type NodeAgentInstallCheckParam struct {
	BizID         int64
	HostID        int64
	NetworkUnitID int64
	InnerIPList   []string
	InnerIPV6List []string
}

// NodeAgentInstallCheckStatus describes the node agent install check status.
type NodeAgentInstallCheckStatus string

const (
	// NodeAgentInstallCheckStatusDuplicatedInnerIP indicates there is a duplicated inner IP.
	NodeAgentInstallCheckStatusDuplicatedInnerIP NodeAgentInstallCheckStatus = "duplicated_inner_ip"

	// NodeAgentInstallCheckStatusDuplicatedInnerIPV6 indicates there is a duplicated inner IPv6.
	NodeAgentInstallCheckStatusDuplicatedInnerIPV6 NodeAgentInstallCheckStatus = "duplicated_inner_ipv6"

	// NodeAgentInstallCheckStatusHostNotFound indicates the host is not found.
	NodeAgentInstallCheckStatusHostNotFound NodeAgentInstallCheckStatus = "host_not_found"

	// NodeAgentInstallCheckStatusNetworkUnitNotFound indicates the network unit is not found.
	NodeAgentInstallCheckStatusNetworkUnitNotFound NodeAgentInstallCheckStatus = "networkunit_not_found"

	// NodeAgentInstallCheckStatusMismatchedInnerIP indicates there is a mismatched inner IP.
	NodeAgentInstallCheckStatusMismatchedInnerIP NodeAgentInstallCheckStatus = "mismatched_inner_ip"

	// NodeAgentInstallCheckStatusMismatchedInnerIPV6 indicates there is a mismatched inner IPv6.
	NodeAgentInstallCheckStatusMismatchedInnerIPV6 NodeAgentInstallCheckStatus = "mismatched_inner_ipv6"

	// NodeAgentInstallCheckStatusMismatchedBizID indicates there is a mismatched biz ID.
	NodeAgentInstallCheckStatusMismatchedBizID NodeAgentInstallCheckStatus = "mismatched_biz_id"

	// NodeAgentInstallCheckStatusMismatchedNetworkAreaID indicates there is a mismatched networkarea ID.
	NodeAgentInstallCheckStatusMismatchedNetworkAreaID NodeAgentInstallCheckStatus = "mismatched_networkarea_id"

	// NodeAgentInstallCheckStatusInvalidNodeRole indicates there is an invalid node role.
	NodeAgentInstallCheckStatusInvalidNodeRole NodeAgentInstallCheckStatus = "invalid_node_role"

	// NodeAgentInstallCheckStatusNetworkUnitNotSupportInstall indicates the network unit does not support install.
	NodeAgentInstallCheckStatusNetworkUnitNotSupportInstall NodeAgentInstallCheckStatus = "networkunit_not_support_install"

	// NodeAgentInstallCheckStatusNormalInstall indicates a normal node agent install.
	NodeAgentInstallCheckStatusNormalInstall NodeAgentInstallCheckStatus = "normal_install"

	// NodeAgentInstallCheckStatusRegisterToCMDBAndInstall indicates the node agent needs to register to CMDB and install.
	NodeAgentInstallCheckStatusRegisterToCMDBAndInstall NodeAgentInstallCheckStatus = "register_to_cmdb_and_install"
)

// NodeAgentInstallCheckMatchedItem describes the node agent install check matched item.
type NodeAgentInstallCheckMatchedItem struct {
	HostID        int64
	BizID         int64
	NetworkAreaID int64
	NetworkUnitID int64
	OsType        criteria.OSType
	NodeRole      NodeRole
	InnerIPList   []string
	InnerIPV6List []string
}

// ConvertHostToNodeAgentInstallCheckMatchedItem converts host to node agent install check matched item.
func ConvertHostToNodeAgentInstallCheckMatchedItem(host *Host) *NodeAgentInstallCheckMatchedItem {
	if host == nil {
		return nil
	}

	return &NodeAgentInstallCheckMatchedItem{
		HostID:        host.HostID,
		BizID:         host.Static.BizID,
		NetworkAreaID: host.Static.NetworkAreaID,
		NetworkUnitID: host.Dynamic.NetworkUnitID,
		OsType:        host.Dynamic.NodeOsType,
		NodeRole:      host.Dynamic.NodeRole,
		InnerIPList:   host.Static.InnerIPList,
		InnerIPV6List: host.Static.InnerIPV6List,
	}
}

// NodeAgentInstallCheckResult describes the node agent install check result.
type NodeAgentInstallCheckResult struct {
	Status  NodeAgentInstallCheckStatus
	Matched *NodeAgentInstallCheckMatchedItem
}
