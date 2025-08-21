/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package constance ...
package constance

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
	// StorageResultMsgKey defines the storage result key.
	StorageResultMsgKey = "err_msg"

	// InstallBySSHResultKey defines the install by ssh result key.
	InstallBySSHResultKey = "install_by_ssh_result"
	// InstallBySSHResultStdOutKey defines the install by ssh result key.
	InstallBySSHResultStdOutKey = "std_out"
	// InstallBySSHResultErrMsgKey defines the install by ssh result key.
	InstallBySSHResultErrMsgKey = "err_msg"
)
