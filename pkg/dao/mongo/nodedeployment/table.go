/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodedeployment

import "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"

// TableName node deployment table name.
const TableName = "node_deployment"

// NodeDeployment represents the table of node deployment.
// Token should be the unique key.
type NodeDeployment struct {
	Token    string    `json:"token" bson:"token"`
	Info     *Info     `json:"info" bson:"info"`
	NodeConf *NodeConf `json:"node_conf" bson:"node_conf"`
}

// Info this is the info of this node deployment.
type Info struct {
	OperInstID string `json:"oper_inst_id" bson:"oper_inst_id"`
	ActionName string `json:"action_name" bson:"action_name"`
}

// NodeConf this is the node conf for node deployment.
type NodeConf struct {
	PreSetting    map[string]any `json:"pre_setting" bson:"pre_setting"`
	CustomSetting map[string]any `json:"custom_setting" bson:"custom_setting"`
}

// UniqueKey unique key of the table.
func (deploy *NodeDeployment) UniqueKey() string {
	return deploy.Token
}

// TableNodeDeployment represent the complete db structures of node deployment.
type TableNodeDeployment base.TableBroker[*NodeDeployment]
