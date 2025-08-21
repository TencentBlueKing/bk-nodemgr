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

import "time"

// NodeProxyInstallHost describes the node proxy install host.
type NodeProxyInstallHost struct {
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
	ExportIP      string
	AdvertiseIP   string
	ReRegister    bool
	ProxyTags     []ProxyTag
}

// NodeProxyInstallParam describes the node proxy install parameter.
type NodeProxyInstallParam struct {
	Hosts         []*NodeProxyInstallHost
	TargetVersion []*TargetVersion
}

// NodeProxyUpgradeHost describes the node proxy upgrade host.
type NodeProxyUpgradeHost struct {
	HostID                 int64
	Force                  bool
	GracefulRestartTimeout time.Duration
}

// NodeProxyUpgradeParam describes the node proxy upgrade parameter.
type NodeProxyUpgradeParam struct {
	Hosts         []*NodeProxyUpgradeHost
	TargetVersion []*TargetVersion
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
