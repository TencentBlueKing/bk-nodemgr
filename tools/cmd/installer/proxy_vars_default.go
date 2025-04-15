//go:build !linux

/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package main

// GetGseDataProxyPath get gse data path.
func GetGseDataProxyPath() string {
	panic("data proxy only support linux")
}

// GetGseFileProxyPath get gse file path.
func GetGseFileProxyPath() string {
	panic("file proxy only support linux")
}

// GetGseDataProxyConfPath get gse data proxy config file path.
func GetGseDataProxyConfPath() string {
	panic("data proxy only support linux")
}

// GetGseFileProxyConfPath get gse file proxy config file path.
func GetGseFileProxyConfPath() string {
	panic("file proxy only support linux")
}
