/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package relay defines the relay protocol.
package relay

// MessageType defines the relay message type.
type MessageType string

const (
	// MessageTypeCallbackReq describes the callback request message type.
	MessageTypeCallbackReq MessageType = "callback_req"

	// MessageTypeCallbackResp describes the callback response message type.
	MessageTypeCallbackResp MessageType = "callback_resp"

	// MessageTypeServerPushReq describes the server push message type.
	MessageTypeServerPushReq MessageType = "server_push_req"

	// MessageTypeClientPushReq describes the client push data message type.
	MessageTypeClientPushReq MessageType = "client_push_req"

	// MessageTypeAckReq describes the ack request message type.
	MessageTypeAckReq MessageType = "ack_req"
)

// Base defines the base info.
type Base struct {
	// MessageID describes the message-id.
	MessageID string `json:"message_id"`

	// Type describes the message type.
	MessageType MessageType `json:"type"`
}

// CallbackReq describes the callback request.
type CallbackReq struct {
	Base

	// URL describes the url to callback server.
	URL string `json:"url"`

	// Body describes the body to callback server.
	Body []byte `json:"body"`
}

// CallbackResp describes the callback response.
type CallbackResp struct {
	Base

	HTTPCode int    `json:"http_code"`
	Body     []byte `json:"body"`
}

// ServerPushReq describes the server push request.
type ServerPushReq struct {
	Base

	// EventType describes the event type.
	EventType ServerPushEventType `json:"event_type"`

	// Payload describes the payload.
	Payload []byte `json:"payload"`
}

// ClientPushReq defines the client push data.
type ClientPushReq struct {
	Base

	// URL describes the url to ClientPushData.
	URL string `json:"url"`

	// Body describes the body to ClientPushData.
	Body []byte `json:"body"`
}

// AckReq describes the ack request.
type AckReq struct {
	Base
	OriginalMessageID string `json:"original_message_id"`
}

// ServerPushEventType defines the event type.
type ServerPushEventType string

const (
	// ServerPushEventTypeCheckPkgState describes the check pkg state event type.
	ServerPushEventTypeCheckPkgState ServerPushEventType = "check_pkg_state"

	// ServerPushEventTypeNotifyReceive describes the notify receive event type.
	ServerPushEventTypeNotifyReceive ServerPushEventType = "notify_receive"

	// ServerPushEventTypeDetectInfoBySSH describes the detect info by ssh event type.
	ServerPushEventTypeDetectInfoBySSH ServerPushEventType = "detect_info_by_ssh"

	// ServerPushEventTypeDetectInfoByWMI describes the detect info by wmi event type.
	ServerPushEventTypeDetectInfoByWMI ServerPushEventType = "detect_info_by_wmi"

	// ServerPushEventTypeInstallBySSH describes the install by ssh event type.
	ServerPushEventTypeInstallBySSH ServerPushEventType = "install_by_ssh"

	// ServerPushEventTypeInstallByWMI describes the install by wmi event type.
	ServerPushEventTypeInstallByWMI ServerPushEventType = "install_by_wmi"
)

// define server_push relay event struct.

// CheckPkgStateReq describes the check pkg state request.
type CheckPkgStateReq struct {
	// ActionName describes the action name.
	ActionName string `json:"action_name"`

	// OperInstID describes the operation instance id.
	OperInstID string `json:"oper_inst_id"`

	// FileStorageTmpDir describes backend need to transfer file to this dir.
	FileStorageTmpDir string `json:"file_storage_tmp_dir"`

	// FileList describes the file list.
	FileList []FileInfo `json:"file_list"`
}

// FileInfo defines the file info.
type FileInfo struct {
	FileName string `json:"file_name"`
	FileMD5  string `json:"file_md5"`
}

// NotifyReceiveReq describes the transfer pkg complete request.
type NotifyReceiveReq struct {
	// ActionName describes the action name.
	ActionName string `json:"action_name"`

	// OperInstID describes the operation instance id.
	OperInstID string `json:"oper_inst_id"`

	// PkgName describes the package name.
	PkgName []string `json:"pkg_name"`
}

// DetectInfoBySSHReq defines the detect info by ssh request.
type DetectInfoBySSHReq struct {
	ActionName string `json:"action_name"`
	OperInstID string `json:"oper_inst_id"`

	IP        string `json:"ip"`
	Port      int64  `json:"port"`
	User      string `json:"user"`
	Password  string `json:"password"`
	LoginMode string `json:"login_mode"`
}

// DetectInfoByWMIReq defines the detect info by wmi request.
type DetectInfoByWMIReq struct {
	ActionName string `json:"action_name"`
	OperInstID string `json:"oper_inst_id"`

	IP        string `json:"ip"`
	Port      int64  `json:"port"`
	User      string `json:"user"`
	Password  string `json:"password"`
	LoginMode string `json:"login_mode"`
}

// InstallPagentBySSHReq defines the install pagent by ssh request.
type InstallPagentBySSHReq struct {
	ActionName string `json:"action_name"`
	OperInstID string `json:"oper_inst_id"`

	IP        string `json:"ip"`
	Port      int64  `json:"port"`
	User      string `json:"user"`
	Password  string `json:"password"`
	LoginMode string `json:"login_mode"`

	InstallerWorkDir string   `json:"installer_work_dir"`
	ToolsName        string   `json:"tools_name"`
	InstallerCmd     []string `json:"installer_cmd"`
}

// InstallPagentByWMIReq defines the install pagent by wmi request.
type InstallPagentByWMIReq struct {
	ActionName string `json:"action_name"`
	OperInstID string `json:"oper_inst_id"`

	IP        string `json:"ip"`
	Port      int64  `json:"port"`
	User      string `json:"user"`
	Password  string `json:"password"`
	LoginMode string `json:"login_mode"`

	InstallerWorkDir string   `json:"installer_work_dir"`
	TargetWorkDir    string   `json:"target_work_dir"`
	ToolsName        string   `json:"tools_name"`
	InstallerBatName string   `json:"installer_bat_name"`
	InstallerCmd     []string `json:"installer_cmd"`
}
