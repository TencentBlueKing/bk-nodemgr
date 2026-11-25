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

// SystemConfig is the explicit business-facing view of loaded Backend configuration.
// It does not include credentials, client endpoints, or internal observability settings.
type SystemConfig struct {
	TenantMode               string
	System                   SystemConfigSystem
	IAM                      SystemConfigIAM
	NodeEventDataID          SystemConfigNodeEventDataID
	DefaultDirectNetworkUnit SystemConfigDefaultDirectNetworkUnit
	GSEDeployConfigs         []SystemConfigGSEDeploy
}

// SystemConfigSystem identifies the deployment environment and edition.
type SystemConfigSystem struct {
	Env     string
	Edition string
}

// SystemConfigIAM describes the configured permission-management switches.
type SystemConfigIAM struct {
	V3Enabled bool
	V4Enabled bool
}

// SystemConfigNodeEventDataID contains the global fallback node event data IDs.
// Business-specific overrides are not resolved by this view.
type SystemConfigNodeEventDataID struct {
	AgentBaseAlarmEventDataID int64
	TaskProcEventDataID       int64
}

// SystemConfigDefaultDirectNetworkUnit describes automatic direct-unit creation.
type SystemConfigDefaultDirectNetworkUnit struct {
	Enabled bool
	Name    string
}

// SystemConfigGSEDeploy describes the configured installation defaults for one generation and OS.
// Per-deployment overrides and target-host state are not resolved by this view.
type SystemConfigGSEDeploy struct {
	Generation       int64
	OSType           string
	BaseWorkDir      string
	BaseDeployDir    string
	ExtraConfigDir   string
	HostIDPath       string
	V2RunDir         string
	V2DataDir        string
	ZoneID           string
	CityID           string
	HTTPProxyMode    SystemConfigHTTPProxyMode
	FirewallFeatures []FirewallFeature
}

// SystemConfigHTTPProxyMode describes installer proxy selection without exposing the proxy address.
type SystemConfigHTTPProxyMode string

const (
	// SystemConfigHTTPProxyModeInherit keeps the target host's environment proxy settings.
	SystemConfigHTTPProxyModeInherit SystemConfigHTTPProxyMode = "inherit"
	// SystemConfigHTTPProxyModeDirect explicitly skips the target host's environment proxy settings.
	SystemConfigHTTPProxyModeDirect SystemConfigHTTPProxyMode = "direct"
	// SystemConfigHTTPProxyModeSpecified uses the configured proxy without exposing its address.
	SystemConfigHTTPProxyModeSpecified SystemConfigHTTPProxyMode = "specified"
)
