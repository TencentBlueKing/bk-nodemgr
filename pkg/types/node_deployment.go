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

import (
	"errors"
	"fmt"
)

// NodeDeployment this is the info for node deployment.
type NodeDeployment struct {
	Token    string
	Info     *DeploymentInfo
	NodeConf *NodeConf
}

// DeploymentInfo this is the info for node deployment.
type DeploymentInfo struct {
	OperInstID string
	ActionName string
	Host
}

// Validate this is the validate for node deployment.
// nolint: gocognit,gocyclo,cyclop
func (info DeploymentInfo) Validate() error {
	if info.OperInstID == "" {
		return errors.New("oper_inst_id shouldn't not be empty")
	}

	if info.ActionName == "" {
		return errors.New("action_name shouldn't not be empty")
	}

	if info.Host.HostID < 0 {
		return errors.New("host_id should be equal or greater than 0")
	}

	if info.Host.Static.OSType == "" {
		return errors.New("os_type shouldn't not be empty")
	}

	if info.TenantID == "" {
		return errors.New("tenant_id shouldn't not be empty")
	}

	if err := info.Host.Dynamic.NodeRole.Validate(); err != nil {
		return fmt.Errorf("node_role validate failed, err: %w", err)
	}

	if err := info.Host.Dynamic.NodeStatus.Validate(); err != nil {
		return fmt.Errorf("node_status validate failed, err: %w", err)
	}

	if info.Dynamic.NodeVersion == "" {
		return errors.New("node_version shouldn't not be empty")
	}
	if info.Dynamic.NodeGeneration != 1 && info.Dynamic.NodeGeneration != 2 {
		return errors.New("node_generation should be 1 or 2")
	}
	if info.Dynamic.NetworkUnitID < 0 {
		return errors.New("network_unit_id should be equal or greater than 0")
	}
	if info.Static.BizID < 0 {
		return errors.New("biz_id should be equal or greater than 0")
	}
	if info.Static.NetworkAreaID < 0 {
		return errors.New("network_area_id should be equal or greater than 0")
	}
	if info.Static.InnerIP == "" {
		return errors.New("inner_ip shouldn't not be empty")
	}
	if info.Static.Addressing == "" {
		return errors.New("addressing shouldn't not be empty")
	}

	// nolint: nestif
	if info.Dynamic.NodeRole == NodeRoleProxy {
		if info.Dynamic.ProxyClusterPort > 0 {
			return errors.New("proxy_cluster_port should be 0")
		}
		if info.Dynamic.ProxyDataPort > 0 {
			return errors.New("proxy_data_port should be 0")
		}
		if info.Dynamic.ProxyFilePort > 0 {
			return errors.New("proxy_file_port should be 0")
		}
	} else {
		if info.Dynamic.ProxyClusterPort != 0 {
			return errors.New("node_role is not proxy, proxy_cluster_port should be 0")
		}

		if info.Dynamic.ProxyDataPort != 0 {
			return errors.New("node_role is not proxy, proxy_data_port should be 0")
		}

		if info.Dynamic.ProxyFilePort != 0 {
			return errors.New("node_role is not proxy, proxy_file_port should be 0")
		}
	}

	return nil
}

// NodeConf this is the node conf for node deployment.
type NodeConf struct {
	PreSetting    map[string]any
	CustomSetting map[string]any
}
