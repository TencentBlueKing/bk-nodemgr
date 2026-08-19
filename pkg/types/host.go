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

package types

import (
	"errors"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
)

// Addressing represents an addressing type.
type Addressing string

const (
	// AddressingDynamic means the addressing type is dynamic.
	AddressingDynamic Addressing = "dynamic"

	// AddressingStatic means the addressing type is static.
	AddressingStatic Addressing = "static"

	// IPSeparator means the ip separator.
	IPSeparator = ","

	// AreaIPSeparator means the separator between area-id and ip.
	AreaIPSeparator = ":"
)

// Validate validates the addressing type.
func (addr Addressing) Validate() error {
	switch addr {
	case AddressingDynamic, AddressingStatic:
		return nil
	default:
		return errors.New("addressing must be dynamic or static")
	}
}

// NodeStatus represents a node status when node role is not blank.
type NodeStatus string

// Validate validates the node status.
func (nodeStatus NodeStatus) Validate() error {
	switch nodeStatus {
	case NodeStatusInit, NodeStatusRunning, NodeStatusDamaged, NodeStatusBusy,
		NodeStatusStarting, NodeStatusUpgrade, NodeStatusStopping, NodeStatusUninit, NodeStatusUnknown:
		return nil
	default:
		return errors.New("invalid node status")
	}
}

// NodeStatusListToStringList converts a node status list to a string list.
func NodeStatusListToStringList(nodeStatusList []NodeStatus) []string {
	data := make([]string, len(nodeStatusList))
	for idx, nodeStatus := range nodeStatusList {
		data[idx] = string(nodeStatus)
	}

	return data
}

// StringListToNodeStatusList converts a string list to a node status list.
func StringListToNodeStatusList(stringList []string) []NodeStatus {
	data := make([]NodeStatus, len(stringList))
	for idx, nodeStatus := range stringList {
		data[idx] = NodeStatus(nodeStatus)
	}

	return data
}

const (
	// NodeStatusInit means this node status is initialized.
	NodeStatusInit NodeStatus = "init"

	// NodeStatusRunning means this node status is running.
	NodeStatusRunning NodeStatus = "running"

	// NodeStatusDamaged means this node status is damaged.
	NodeStatusDamaged NodeStatus = "damaged"

	// NodeStatusBusy means this node status is busy.
	NodeStatusBusy NodeStatus = "busy"

	// NodeStatusStarting means this node status is starting.
	NodeStatusStarting NodeStatus = "starting"

	// NodeStatusUpgrade means this node status is upgrading.
	NodeStatusUpgrade NodeStatus = "upgrade"

	// NodeStatusStopping means this node status is stopping.
	NodeStatusStopping NodeStatus = "stopping"

	// NodeStatusUninit means this node status is uninitialized.
	NodeStatusUninit NodeStatus = "uninit"

	// NodeStatusUnknown means this node status is unknown.
	NodeStatusUnknown NodeStatus = "unknown"
)

// HostDynamic represents a dynamic host under a host.
// dynamic means it is set by user.
type HostDynamic struct {
	NodeRole       NodeRole
	NodeStatus     NodeStatus
	NodeVersion    string
	NodeGeneration Generation
	NodeCPUArch    criteria.CPUArch
	NodeOsType     criteria.OSType
	AgentID        string
	NetworkUnitID  int64

	// LoginIP represents the ip when SSH login to install.
	// ExportIP and ExportIPV6 represents the ip that the outgoing IP in NAT.
	// AdvertiseIP and AdvertiseIPV6 represents the ip that the incoming IP in NAT.
	LoginIP          string
	LoginPort        int64
	LoginUser        string
	LoginMode        LoginMode
	LoginCreditID    string
	LoginCreditValid bool
	ExportIP         string
	ExportIPV6       string
	AdvertiseIP      string
	AdvertiseIPV6    string

	// ProxyAccessDisabled This means that there will be no new proxy access connection establishment for this node.
	// ! This setting does not affect the established connections.
	ProxyAccessDisabled bool

	// ProxyTags represents the tags of this proxy.
	// ! Please make sure to use these tags on a whitelist basis.
	ProxyTags        []ProxyTag
	ProxyClusterPort int64
	ProxyDataPort    int64
	ProxyFilePort    int64

	// RelayDownloadPort represents the port of relay file server.
	// RelayCallbackPort represents the port of relay callback server.
	RelayDownloadPort int64
	RelayCallbackPort int64

	// ProxyInstallOriginUnitID represents the origin network unit id where this proxy is installed.
	ProxyInstallOriginUnitID int64

	// ConnCycleTime represents the connection cycle time of the agent reported by GSE.
	ConnCycleTime string

	// Ops represents the fields related to ops console and outband management.
	OpsConsoleHostID   int64
	OpsOutBandType     string
	OpsOutBandProtocol string
	OpsBMCIP           string
	OpsBMCPort         int64
}

// ProxySupportInstaller returns whether this node support installer.
func (dynamic *HostDynamic) ProxySupportInstaller() bool {
	for _, tag := range dynamic.ProxyTags {
		if tag == ProxyTagDedicatedInstaller {
			return true
		}
	}

	return false
}

// ProxySupportCluster returns whether this node support cluster.
func (dynamic *HostDynamic) ProxySupportCluster() bool {
	for _, tag := range dynamic.ProxyTags {
		if tag == ProxyTagClusterTunnel {
			return true
		}
	}

	return false
}

// ProxySupportFile returns whether this node support file.
func (dynamic *HostDynamic) ProxySupportFile() bool {
	for _, tag := range dynamic.ProxyTags {
		if tag == ProxyTagFileTunnel {
			return true
		}
	}

	return false
}

// ProxySupportData returns whether this node support data.
func (dynamic *HostDynamic) ProxySupportData() bool {
	for _, tag := range dynamic.ProxyTags {
		if tag == ProxyTagDataTunnel {
			return true
		}
	}

	return false
}

// HostDynamicFields represents the fields of HostDynamic fields.
type HostDynamicFields struct {
	NodeRole       bool
	NodeStatus     bool
	NodeVersion    bool
	NodeGeneration bool
	NodeCPUArch    bool
	NodeOsType     bool
	AgentID        bool
	NetworkUnitID  bool

	LoginIP       bool
	LoginPort     bool
	LoginUser     bool
	LoginMode     bool
	LoginCreditID bool
	ExportIP      bool
	ExportIPV6    bool
	AdvertiseIP   bool
	AdvertiseIPV6 bool

	ProxyTags                bool
	ProxyClusterPort         bool
	ProxyDataPort            bool
	ProxyFilePort            bool
	ProxyInstallOriginUnitID bool

	RelayDownloadPort bool
	RelayCallbackPort bool

	ConnCycleTime bool

	// Ops fields related to out-of-band management.
	OpsConsoleHostID   bool
	OpsOutBandType     bool
	OpsOutBandProtocol bool
	OpsBMCIP           bool
	OpsBMCPort         bool
}

// ProxyTag represents a proxy tag.
type ProxyTag string

const (
	// ProxyTagDedicatedInstaller means this node is a dedicated installer.
	ProxyTagDedicatedInstaller = "dedicated_installer"

	// ProxyTagClusterTunnel means this node is a cluster tunnel.
	ProxyTagClusterTunnel = "cluster_tunnel"

	// ProxyTagFileTunnel means this node is a file tunnel.
	ProxyTagFileTunnel = "file_tunnel"

	// ProxyTagDataTunnel means this node is a data tunnel.
	ProxyTagDataTunnel = "data_tunnel"
)

// Validate validates the proxy tag.
func (tag ProxyTag) Validate() error {
	switch tag {
	case ProxyTagDedicatedInstaller, ProxyTagClusterTunnel, ProxyTagFileTunnel, ProxyTagDataTunnel:
		return nil
	default:
		return errors.New("invalid proxy tag")
	}
}

// AllProxyTag returns all proxy tags.
func AllProxyTag() []ProxyTag {
	return []ProxyTag{ProxyTagDedicatedInstaller, ProxyTagClusterTunnel, ProxyTagFileTunnel, ProxyTagDataTunnel}
}

// ProxyTagListToStringList converts a proxy tag list to a string list.
func ProxyTagListToStringList(tagList []ProxyTag) []string {
	stringList := make([]string, 0, len(tagList))
	for _, tag := range tagList {
		stringList = append(stringList, string(tag))
	}

	return stringList
}

// StringListToProxyTagList converts a string list to a proxy tag list.
func StringListToProxyTagList(stringList []string) []ProxyTag {
	tagList := make([]ProxyTag, 0, len(stringList))
	for _, tag := range stringList {
		tagList = append(tagList, ProxyTag(tag))
	}

	return tagList
}

// NewBlankNodeDynamic returns a blank node dynamic.
func NewBlankNodeDynamic() *HostDynamic {
	return &HostDynamic{
		NodeRole:                 NodeRoleBlank,
		NodeStatus:               NodeStatusUnknown,
		NodeVersion:              "",
		NodeGeneration:           0,
		AgentID:                  "",
		NetworkUnitID:            -1,
		ProxyInstallOriginUnitID: -1,
	}
}

// Host represents a host.
type Host struct {
	HostID   int64
	TenantID string

	Static  *HostStatic
	Dynamic *HostDynamic

	// OperationUpdatedAt tracks the last time a user-initiated or API-driven
	// business operation modified this host. Zero value means no operation has
	// been recorded yet (legacy data).
	OperationUpdatedAt time.Time
}

// HostTopo represents a host's set and module relation in CMDB.
type HostTopo struct {
	SetID    int64
	ModuleID int64
}

// HostStatic represents a static host under a host.
// static means it is synced from CMDB.
// or sometimes it will be insert first into database in case of syncing latency.
type HostStatic struct {
	// belongs to
	BizID         int64
	Topo          []*HostTopo
	NetworkAreaID int64
	ZoneID        int64
	CityID        string

	// host information.
	HostName      string
	DeptName      string
	InnerIPList   []string
	InnerIPV6List []string
	OuterIPList   []string
	OuterIPV6List []string
	Operator      string
	Mac           string
	OSTypeCCID    string
	OSType        string
	Arch          string
	Addressing    Addressing

	CPUNum float64
	MemCap float64

	// synced types, do not use this for processing.
	// just use it for comparing and checking.
	SyncedAgentID            string
	SyncedOpsConsoleHostID   int64
	SyncedOpsOutBandType     string
	SyncedOpsOutBandProtocol string
	SyncedOpsBMCIP           string
	SyncedOpsBMCPort         int64
}

// GetInnerIPList returns a list of inner ip.
func (static *HostStatic) GetInnerIPList() []string {
	return static.InnerIPList
}

// GetOuterIPList returns a list of outer ip.
func (static *HostStatic) GetOuterIPList() []string {
	return static.OuterIPList
}

// HostStaticFields represents the fields of HostStatic fields.
type HostStaticFields struct {
	BizID                    bool
	Topo                     bool
	NetworkAreaID            bool
	ZoneID                   bool
	CityID                   bool
	HostName                 bool
	DeptName                 bool
	InnerIPList              bool
	InnerIPV6List            bool
	OuterIPList              bool
	OuterIPV6List            bool
	Operator                 bool
	Mac                      bool
	OSTypeCCID               bool
	OSType                   bool
	Arch                     bool
	Addressing               bool
	CPUNum                   bool
	MemCap                   bool
	SyncedAgentID            bool
	SyncedOpsConsoleHostID   bool
	SyncedOpsOutBandType     bool
	SyncedOpsOutBandProtocol bool
	SyncedOpsBMCIP           bool
	SyncedOpsBMCPort         bool
}

// UpdateAllHostStaticFields returns a HostStaticFields with all fields set to true.
func UpdateAllHostStaticFields() HostStaticFields {
	return HostStaticFields{
		BizID:                    true,
		Topo:                     true,
		NetworkAreaID:            true,
		ZoneID:                   true,
		CityID:                   true,
		HostName:                 true,
		DeptName:                 true,
		InnerIPList:              true,
		InnerIPV6List:            true,
		OuterIPList:              true,
		OuterIPV6List:            true,
		Operator:                 true,
		Mac:                      true,
		OSTypeCCID:               true,
		OSType:                   true,
		Arch:                     true,
		Addressing:               true,
		CPUNum:                   true,
		MemCap:                   true,
		SyncedAgentID:            true,
		SyncedOpsConsoleHostID:   true,
		SyncedOpsOutBandType:     true,
		SyncedOpsOutBandProtocol: true,
		SyncedOpsBMCIP:           true,
		SyncedOpsBMCPort:         true,
	}
}

// NodeRole represents a node role.
type NodeRole string

// Validate validates the node role.
func (nodeRole NodeRole) Validate() error {
	switch nodeRole {
	case NodeRoleBlank, NodeRoleAgent, NodeRoleProxy:
		return nil
	default:
		return errors.New("invalid node role")
	}
}

// NodeRoleListToStringList converts a node role list to a string list.
func NodeRoleListToStringList(nodeRoleList []NodeRole) []string {
	data := make([]string, len(nodeRoleList))
	for idx, nodeRole := range nodeRoleList {
		data[idx] = string(nodeRole)
	}

	return data
}

// StringListToNodeRoleList converts a string list to a node role list.
func StringListToNodeRoleList(stringList []string) []NodeRole {
	data := make([]NodeRole, len(stringList))
	for idx, nodeRole := range stringList {
		data[idx] = NodeRole(nodeRole)
	}

	return data
}

const (
	// NodeRoleBlank means this node is blank. nothing installed.
	NodeRoleBlank NodeRole = "blank"

	// NodeRoleAgent means this node is an agent.
	NodeRoleAgent NodeRole = "agent"

	// NodeRoleProxy means this node is a proxy.
	NodeRoleProxy NodeRole = "proxy"
)

// HostFieldSelection represents field selection options for host queries.
type HostFieldSelection struct {
	HostID bool
	// Static fields
	BizID         bool
	Topo          bool
	NetworkAreaID bool
	InnerIPList   bool
	InnerIPV6List bool

	// Dynamic fields
	NodeRole       bool
	NodeStatus     bool
	NodeVersion    bool
	NodeGeneration bool
	LoginUser      bool
	AgentID        bool
	AdvertiseIP    bool
	AdvertiseIPV6  bool
}

// HostTopoRelation represents the relationship between a host and its topology.
type HostTopoRelation struct {
	HostID int64
	BizID  int64
	Topo   []*HostTopo
}
