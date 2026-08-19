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

// RelayInfo defines the relay info.
type RelayInfo struct {
	// HostID is the proxy host id.
	HostID int64
	// AgentID is the proxy host id.
	AgentID string
	// PackageDestDir is the package dest dir.
	PackageDestDir string

	// AdvertiseIP is the advertise ip.
	AdvertiseIP string
	// AdvertiseIPV6 is the advertise ipv6.
	AdvertiseIPV6 string
	// DownloadSvcPort is the file service port.
	DownloadSvcPort int64
	// CallbackSvcPort is the callback service port.
	CallbackSvcPort int64
}
