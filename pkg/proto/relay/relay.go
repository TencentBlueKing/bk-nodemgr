/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package proxy defines the proxy protocol.
package proxy

// MessageType defines the proxy message type.
type MessageType string

const (
	// MessageTypeCallbackReq describes the callback request message type.
	MessageTypeCallbackReq MessageType = "callback_req"

	// MessageTypeCallbackResp describes the callback response message type.
	MessageTypeCallbackResp MessageType = "callback_resp"

	// MessageTypeServerPush describes the server push message type.
	MessageTypeServerPush MessageType = "server_push"
)

// EventType defines the event type.
type EventType string

const (
	// EventTypeCheckReleaseExist describes the check release exist event type.
	EventTypeCheckReleaseExist EventType = "check_release_exist"
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

// ServerPush describes the server push.
type ServerPush struct {
	Base

	// EventType describes the event type.
	EventType EventType `json:"event_type"`

	// Payload describes the payload.
	Payload []byte `json:"payload"`
}

// CheckReleaseExistEvent describes the check release exist event.
type CheckReleaseExistEvent struct {
	// Filename describes the filename.
	Filename string `json:"filename"`

	// MD5 describes the md5.
	MD5 string `json:"md5"`
}
