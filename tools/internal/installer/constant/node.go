/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package constant

import "fmt"

// NodeType define node type.
type NodeType string

// Validate validate node type.
func (nodeType NodeType) Validate() error {
	switch nodeType {
	case NodeTypeAgent, NodeTypeProxy:
		return nil
	}

	return fmt.Errorf("unknown node type, node-type(%s)", nodeType)
}

const (
	// NodeTypeAgent means this is an agent node.
	NodeTypeAgent NodeType = "agent"

	// NodeTypeProxy means this is a proxy node.
	NodeTypeProxy NodeType = "proxy"
)
