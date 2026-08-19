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

// Package relayconstant ...
package relayconstant

// defines the field name stored in the database.
const (

	// FileStateKey defines the relay file state key.
	FileStateKey = "file_state"
	// FileStateStorageKey defines the relay file state storage key.
	FileStateStorageKey = "file_state_storage_dir"

	// DetectResultKey defines the detect result key.
	DetectResultKey = "detect_result"
	// DetectResultOsTypeKey defines the detect result key.
	DetectResultOsTypeKey = "os_type"
	// DetectResultCPUArchKey defines the detect result key.
	DetectResultCPUArchKey = "cpu_arch"
	// DetectResultConnectionDirKey defines the detect result key.
	DetectResultConnectionDirKey = "connection_dir"
	// DetectResultErrMsg defines the detect result key.
	DetectResultErrMsgKey = "err_msg"

	// StorageResultKey defines the storage result key.
	StorageResultKey = "storage_result"
	// StorageResultErrMsgKey defines the storage result key.
	StorageResultErrMsgKey = "err_msg"

	// InstallResultKey defines the install result key.
	InstallResultKey = "install_result"
	// InstallResultOutStrKey defines the install result key.
	InstallResultOutStrKey = "out_str"
	// InstallResultErrMsgKey defines the install result key.
	InstallResultErrMsgKey = "err_msg"
)
