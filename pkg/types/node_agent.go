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

import "errors"

// NodeAgentInstallParam describes the node agent install parameter.
type NodeAgentInstallParam struct {
	BizID         int64
	InnerIP       string
	InnerIPV6     string
	Addressing    Addressing
	LoginIP       string
	LoginPort     int64
	LoginUser     string
	LoginMode     LoginMode
	LoginPassword string
	LoginKeyFile  []byte
	NetworkUnitID int64
	OSType        string
	TargetVersion string
}

// NodeOperationRetryParam validates the node agent install parameter.
type NodeOperationRetryParam struct {
	WorkflowID   string
	OperationIDs []string
	RetryMode    NodeOperationRetryMode
}

// NodeOperationRetryMode describes the node agent operation mode.
type NodeOperationRetryMode string

const (
	// FullNodeInstanceRetry is the full node instance retry mode.
	FullNodeInstanceRetry NodeOperationRetryMode = "full_node_instance_retry"

	// PartialNodeInstanceRetry is the partial node instance retry mode.
	PartialNodeInstanceRetry NodeOperationRetryMode = "partial_node_instance_retry"
)

// Validate validates the node operation retry mode.
func (tag NodeOperationRetryMode) Validate() error {
	switch tag {
	case FullNodeInstanceRetry, PartialNodeInstanceRetry:
		return nil
	default:
		return errors.New("invalid node operation retry mode")
	}
}
