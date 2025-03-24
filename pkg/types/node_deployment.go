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

// NodeDeployment this is the info for node deployment.
type NodeDeployment struct {
	Token    string
	Info     *DeploymentInfo
	NodeConf *NodeConf
}

// DeploymentInfo this is the info for node deployment.
type DeploymentInfo struct {
	OperInstID     string
	ActionName     string
	HostID         int64
	TenantID       string
	NodeRole       NodeRole
	NodeStatus     NodeStatus
	NodeVersion    string
	NodeGeneration int64
	AgentID        string
	NetworkUnitID  int64
}

// NodeConf this is the node conf for node deployment.
type NodeConf struct {
	PreSetting    map[string]any
	CustomSetting map[string]any
}
