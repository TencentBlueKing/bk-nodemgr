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

import "github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"

// NetworkPolicyEndpoint network policy endpoint.
type NetworkPolicyEndpoint struct {
	Name          string
	Type          criteria.NetEndpointType
	Values        []string
	DescriptionEN string
	DescriptionZH string
}

// NetworkPolicyService network policy service.
type NetworkPolicyService struct {
	Protocol      criteria.NetType
	Ports         []string // single port 80 or port range 8000-9000
	DescriptionEN string
	DescriptionZH string
}

// NetworkPolicy network policy.
type NetworkPolicy struct {
	Name          string
	DescriptionEN string
	DescriptionZH string
	Source        NetworkPolicyEndpoint
	Target        NetworkPolicyEndpoint
	Service       NetworkPolicyService
}
