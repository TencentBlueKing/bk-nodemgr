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

import "fmt"

// NodeRole define node role.
type NodeRole string

// Validate validate node role.
func (nodeRole NodeRole) Validate() error {
	switch nodeRole {
	case NodeRoleAgent, NodeRoleProxy:
		return nil
	}

	return fmt.Errorf("unknown node role, node-type(%s)", nodeRole)
}

const (
	// NodeRoleAgent means this is an agent node.
	NodeRoleAgent NodeRole = "agent"

	// NodeRoleProxy means this is a proxy node.
	NodeRoleProxy NodeRole = "proxy"
)

// Generation define node generation.
type Generation int

const (
	// Generation1 means this is a generation 1 node.
	Generation1 Generation = 1

	// Generation2 means this is a generation 2 node.
	Generation2 Generation = 2
)

// Validate validate node generation.
func (gen Generation) Validate() error {
	switch gen {
	case Generation1, Generation2:
		return nil
	}

	return fmt.Errorf("unknown node generation, generation(%d)", gen)
}
