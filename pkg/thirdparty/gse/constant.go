/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package gse

const (
	// queryMultiProcessInfoPageSize the page size of query multi process info.
	// notice: This is not a mandatory value, but a recommended value
	queryMultiProcessInfoPageSize = 30000

	// ListAgentStatePageSize the page size of list agent state.
	ListAgentStatePageSize = 1000

	// ListAgentInfoPageSize the page size of list agent info.
	ListAgentInfoPageSize = 1000
)

const (
	// WindowsOperateUser the operate user of Windows.
	// notice: In Windows use "system" user to do operations will not switch users,
	// which can avoid the problem caused by the need to re-enter the password in some environments.
	WindowsOperateUser = "system"
)
