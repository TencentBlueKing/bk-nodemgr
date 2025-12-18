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

// NodeAgentInstallCheckInfo describes the node agent install check info.
type NodeAgentInstallCheckInfo struct {
	BizID         int64
	HostID        int64
	NetworkUnitID int64
	InnerIP       string
}

// NodeAgentInstallElig describes the node agent install eligibility.
type NodeAgentInstallElig string

const (
	// NodeAgentInstallEligConflictIP indicates there is a conflicting IP under the same business context.
	NodeAgentInstallEligConflictIP NodeAgentInstallElig = "conflict_ip"

	// NodeAgentInstallEligDuplicateIP indicates there are duplicate IPs in dynamic addressing mode.
	NodeAgentInstallEligDuplicateIP NodeAgentInstallElig = "duplicate_dynamic_ip"

	// NodeAgentInstallEligExistProxy indicates a proxy already exists on this node.
	NodeAgentInstallEligExistProxy NodeAgentInstallElig = "exist_proxy"

	// NodeAgentInstallEligNormalInstall indicates a normal node agent install; the host can be reused.
	NodeAgentInstallEligNormalInstall NodeAgentInstallElig = "normal_install"

	// NodeAgentInstallEligImportCmdbAndNormalInstall indicates a normal node agent install; the host can be reused.
	NodeAgentInstallEligImportCmdbAndNormalInstall NodeAgentInstallElig = "import_cmdb_and_normal_install"

	// NodeAgentInstallEligNotExistRelay indicates there is no relay host.
	NodeAgentInstallEligNotExistRelay NodeAgentInstallElig = "not_exist_relay"
)

// NodeAgentInstallCheckResult describes the node agent install check result.
type NodeAgentInstallCheckResult struct {
	InnerIP     string
	InstallElig NodeAgentInstallElig

	// The host IDs that are pending to be wait acked.
	PendingHostIDs []int64
}
