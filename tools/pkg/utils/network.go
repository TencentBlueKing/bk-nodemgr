/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package utils ...
package utils

import "fmt"

// NetworkType defines network type.
type NetworkType string

const (
	// NetTCP define the network type is tcp.
	// notice: this means both tcp4 and tcp6.
	NetTCP NetworkType = "tcp"

	// NetTCP4 define the network type is tcp4.
	NetTCP4 NetworkType = "tcp4"

	// NetTCP6 define the network type is tcp6.
	NetTCP6 NetworkType = "tcp6"

	// NetUDP define the network type is udp.
	// notice: this means both udp4 and udp6.
	NetUDP NetworkType = "udp"

	// NetUDP4 define the network type is udp4.
	NetUDP4 NetworkType = "udp4"

	// NetUDP6 define the network type is udp6.
	NetUDP6 NetworkType = "udp6"
)

// String returns the string representation of the network type.
func (n NetworkType) String() string {
	return string(n)
}

// Validate validates the network type.
func (n NetworkType) Validate() error {
	switch n {
	case NetTCP, NetUDP, NetTCP4, NetTCP6, NetUDP4, NetUDP6:
		return nil
	default:
		return fmt.Errorf("invalid network type: %s", n)
	}
}
