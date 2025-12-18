/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package criteria

import "fmt"

// NetType define the network type.
type NetType string

const (
	// NetTypeTCP this defines the network type of tcp.
	NetTypeTCP NetType = "tcp"

	// NetTypeTCP6 this defines the network type of tcp6.
	NetTypeTCP6 NetType = "tcp6"

	// NetTypeUDP this defines the network type of udp.
	NetTypeUDP NetType = "udp"

	// NetTypeUDP6 this defines the network type of udp6.
	NetTypeUDP6 NetType = "udp6"
)

// Validate validate the network type.
func (netType NetType) Validate() error {
	switch netType {
	case NetTypeTCP, NetTypeTCP6, NetTypeUDP, NetTypeUDP6:
		return nil
	default:
		return fmt.Errorf("invalid network type: %s", netType)
	}
}

// NetEndpointType define the network endpoint type.
type NetEndpointType string

const (
	// NetEndpointTypeIP this defines the network endpoint type of ip.
	NetEndpointTypeIP NetEndpointType = "ip"

	// NetEndpointTypeDomain this defines the network endpoint type of domain.
	NetEndpointTypeDomain NetEndpointType = "domain"

	// NetEndpointTypeCIDR this defines the network endpoint type of cidr.
	NetEndpointTypeCIDR NetEndpointType = "cidr"

	// NetEndpointTypeSecureGroup this defines the network endpoint type of secure_group.
	NetEndpointTypeSecureGroup NetEndpointType = "secure_group"
)

// Validate validate the network endpoint type.
func (netEndpointType NetEndpointType) Validate() error {
	switch netEndpointType {
	case NetEndpointTypeIP, NetEndpointTypeDomain, NetEndpointTypeCIDR, NetEndpointTypeSecureGroup:
		return nil
	default:
		return fmt.Errorf("invalid network endpoint type: %s", netEndpointType)
	}
}
