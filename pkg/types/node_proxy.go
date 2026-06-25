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

// NodeProxyInstallHost describes the node proxy install host.
type NodeProxyInstallHost struct {
	HostID                   int64
	BizID                    int64
	InnerIP                  []string
	InnerIPV6                []string
	Addressing               Addressing
	LoginIP                  string
	LoginPort                int64
	LoginUser                string
	LoginMode                LoginMode
	LoginPassword            string
	LoginKeyFile             string
	NetworkUnitID            int64
	OSType                   string
	ExportIP                 string
	AdvertiseIP              string
	ReRegister               bool
	RenewGSETask             bool
	RenewGSEProc             bool
	InstallPreOrderedPlugins bool
	ProxyTags                []ProxyTag
	ProxyInstallOriginUnitID int64
	CreditExpiredIntervalSec int64
	RelayDownloadPort        int64
	RelayCallbackPort        int64
	CPUArch                  string
	InstallMethod            NodeInstallMethod
}

// NodeProxyInstallParam describes the node proxy install parameter.
type NodeProxyInstallParam struct {
	Hosts                   []*NodeProxyInstallHost
	TargetVersion           []*TargetVersion
	IsManual                bool
	IsOffline               bool
	EnableCompatibilityMode bool
}

// NodeProxyUpgradeHost describes the node proxy upgrade host.
type NodeProxyUpgradeHost struct {
	HostID                 int64
	NetworkUnitID          int64
	CPUArch                string
	Force                  bool
	TargetVersion          string
	GracefulRestartTimeout time.Duration
}

// NodeProxyUpgradeParam describes the node proxy upgrade parameter.
type NodeProxyUpgradeParam struct {
	Hosts []*NodeProxyUpgradeHost
}

// NodeProxyRestartHost describes the node proxy restart host.
type NodeProxyRestartHost struct {
	HostID                 int64
	Force                  bool
	GracefulRestartTimeout time.Duration
}

// NodeProxyRestartParam describes the node proxy restart parameter.
type NodeProxyRestartParam struct {
	Hosts []*NodeProxyRestartHost
}

// NodeProxyReconfigHost describes the node proxy reconfig host.
type NodeProxyReconfigHost struct {
	HostID                 int64
	Force                  bool
	GracefulRestartTimeout time.Duration
}

// NodeProxyReconfigParam describes the node proxy reconfig parameter.
type NodeProxyReconfigParam struct {
	Hosts []*NodeProxyReconfigHost
}

// NodeProxyUpdateHost describes the node proxy update host.
type NodeProxyUpdateHost struct {
	HostID            int64
	LoginIP           string
	LoginPort         int64
	LoginUser         string
	LoginMode         LoginMode
	ExportIP          string
	ExportIPV6        string
	AdvertiseIP       string
	AdvertiseIPV6     string
	ProxyTags         []ProxyTag
	RelayDownloadPort int64
	RelayCallbackPort int64
}

// NodeProxyUpdateParam describes the node proxy update parameter.
type NodeProxyUpdateParam struct {
	Hosts []*NodeProxyUpdateHost
}

// NodeProxyUninstallHost describes the node proxy uninstall host.
type NodeProxyUninstallHost struct {
	HostID int64
}

// NodeProxyUninstallParam describes the node proxy uninstall parameter.
type NodeProxyUninstallParam struct {
	Hosts []*NodeProxyUninstallHost
}

// NodeProxyInstallCheckParam describes the node proxy install check param.
type NodeProxyInstallCheckParam struct {
	BizID         int64
	HostID        int64
	NetworkUnitID int64
	InnerIPList   []string
	InnerIPV6List []string
}

// NodeProxyInstallCheckStatus describes the node proxy install check status.
type NodeProxyInstallCheckStatus string

const (
	// NodeProxyInstallCheckStatusDuplicatedInnerIP indicates there is a duplicated inner IP.
	NodeProxyInstallCheckStatusDuplicatedInnerIP NodeProxyInstallCheckStatus = "duplicated_inner_ip"

	// NodeProxyInstallCheckStatusDuplicatedInnerIPV6 indicates there is a duplicated inner IPv6.
	NodeProxyInstallCheckStatusDuplicatedInnerIPV6 NodeProxyInstallCheckStatus = "duplicated_inner_ipv6"

	// NodeProxyInstallCheckStatusHostNotFound indicates the host is not found.
	NodeProxyInstallCheckStatusHostNotFound NodeProxyInstallCheckStatus = "host_not_found"

	// NodeProxyInstallCheckStatusNetworkUnitNotFound indicates the network unit is not found.
	NodeProxyInstallCheckStatusNetworkUnitNotFound NodeProxyInstallCheckStatus = "networkunit_not_found"

	// NodeProxyInstallCheckStatusMismatchedInnerIP indicates there is a mismatched inner IP.
	NodeProxyInstallCheckStatusMismatchedInnerIP NodeProxyInstallCheckStatus = "mismatched_inner_ip"

	// NodeProxyInstallCheckStatusMismatchedInnerIPV6 indicates there is a mismatched inner IPv6.
	NodeProxyInstallCheckStatusMismatchedInnerIPV6 NodeProxyInstallCheckStatus = "mismatched_inner_ipv6"

	// NodeProxyInstallCheckStatusMismatchedBizID indicates there is a mismatched biz ID.
	NodeProxyInstallCheckStatusMismatchedBizID NodeProxyInstallCheckStatus = "mismatched_biz_id"

	// NodeProxyInstallCheckStatusMismatchedNetworkAreaID indicates there is a mismatched networkarea ID.
	NodeProxyInstallCheckStatusMismatchedNetworkAreaID NodeProxyInstallCheckStatus = "mismatched_networkarea_id"

	// NodeProxyInstallCheckStatusInvalidNodeRole indicates there is an invalid node role.
	NodeProxyInstallCheckStatusInvalidNodeRole NodeProxyInstallCheckStatus = "invalid_node_role"

	// NodeProxyInstallCheckStatusNormalInstall indicates a normal node proxy install.
	NodeProxyInstallCheckStatusNormalInstall NodeProxyInstallCheckStatus = "normal_install"

	// NodeProxyInstallCheckStatusRegisterToCMDBAndInstall indicates the node proxy needs to register to CMDB and install.
	NodeProxyInstallCheckStatusRegisterToCMDBAndInstall NodeProxyInstallCheckStatus = "register_to_cmdb_and_install"
)

// NodeProxyInstallCheckMatchedItem describes the node proxy install check matched item.
type NodeProxyInstallCheckMatchedItem struct {
	HostID        int64
	BizID         int64
	NetworkAreaID int64
	NetworkUnitID int64
	OsType        criteria.OSType
	NodeRole      NodeRole
	InnerIPList   []string
	InnerIPV6List []string
}

// ConvertHostToNodeProxyInstallCheckMatchedItem converts host to node proxy install check matched item.
func ConvertHostToNodeProxyInstallCheckMatchedItem(host *Host) *NodeProxyInstallCheckMatchedItem {
	if host == nil {
		return nil
	}

	return &NodeProxyInstallCheckMatchedItem{
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

// NodeProxyInstallCheckResult describes the node proxy install check result.
type NodeProxyInstallCheckResult struct {
	Status    NodeProxyInstallCheckStatus
	Matched   *NodeProxyInstallCheckMatchedItem
	MessageEn string
	MessageZh string
	Category  string
}

// NodeProxyUpgradeCheckStatus describes the node proxy upgrade check status.
type NodeProxyUpgradeCheckStatus string

const (
	// NodeProxyUpgradeCheckStatusHostNotFound indicates the host is not found.
	NodeProxyUpgradeCheckStatusHostNotFound NodeProxyUpgradeCheckStatus = "host_not_found"

	// NodeProxyUpgradeCheckStatusNetworkUnitNotFound indicates the network unit is not found.
	NodeProxyUpgradeCheckStatusNetworkUnitNotFound NodeProxyUpgradeCheckStatus = "networkunit_not_found"

	// NodeProxyUpgradeCheckStatusNetworkUnitMismatch indicates the network unit does not belong to the host's
	// current network area.
	NodeProxyUpgradeCheckStatusNetworkUnitMismatch NodeProxyUpgradeCheckStatus = "networkunit_mismatch"

	// NodeProxyUpgradeCheckStatusVersionNotFound indicates no matching version is found for the host's
	// os_type and cpu_arch.
	NodeProxyUpgradeCheckStatusVersionNotFound NodeProxyUpgradeCheckStatus = "version_not_found"

	// NodeProxyUpgradeCheckStatusNodeStatusNotAllowed indicates the host's current node status does not
	// allow upgrade.
	NodeProxyUpgradeCheckStatusNodeStatusNotAllowed NodeProxyUpgradeCheckStatus = "node_status_not_allowed"

	// NodeProxyUpgradeCheckStatusNetworkUnitChanged indicates the network unit will change after upgrade.
	NodeProxyUpgradeCheckStatusNetworkUnitChanged NodeProxyUpgradeCheckStatus = "networkunit_changed"

	// NodeProxyUpgradeCheckStatusNormalUpgrade indicates a normal node proxy upgrade.
	NodeProxyUpgradeCheckStatusNormalUpgrade NodeProxyUpgradeCheckStatus = "normal_upgrade"

	// NodeProxyUpgradeCheckStatusCPUArchMissing indicates the host's cpu_arch is missing.
	NodeProxyUpgradeCheckStatusCPUArchMissing NodeProxyUpgradeCheckStatus = "cpu_arch_missing"
)

// NodeProxyUpgradeCheckMatchedItem describes the node proxy upgrade check matched item.
type NodeProxyUpgradeCheckMatchedItem struct {
	HostID        int64
	BizID         int64
	NetworkAreaID int64
	NetworkUnitID int64
	OsType        criteria.OSType
	NodeRole      NodeRole
	InnerIPList   []string
	InnerIPV6List []string
}

// NodeProxyUpgradeCheckResult describes the node proxy upgrade check result.
type NodeProxyUpgradeCheckResult struct {
	Status  NodeProxyUpgradeCheckStatus
	Matched *NodeProxyUpgradeCheckMatchedItem
}

// NodeProxyUpgradeCheckParam describes the node proxy upgrade check parameter.
type NodeProxyUpgradeCheckParam struct {
	HostID        int64
	NetworkUnitID int64
	CPUArch       string
}

// NodeProxyAssignUnitParam describes the node proxy assign unit parameter.
type NodeProxyAssignUnitParam struct {
	HostIDs       []int64
	NetworkUnitID int64
}

// NodeProxyAssignUnitResult describes the node proxy assign unit result.
type NodeProxyAssignUnitResult struct {
	SuccessCount  int64
	FailedCount   int64
	FailedReasons []string
	WorkflowID    string
}
