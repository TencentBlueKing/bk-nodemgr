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

package relayconstant

// defines the default listen ports of the relay servers, they are used when the
// caller does not specify the ports explicitly.
const (
	// DefaultCallbackPort defines the default port of the relay callback server.
	DefaultCallbackPort int64 = 28302

	// DefaultDownloadPort defines the default port of the relay download server.
	DefaultDownloadPort int64 = 28303
)
