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

// ServerPushEventType defines the event type.
type ServerPushEventType string

const (
	// EventTypeEcho describes the echo event type.
	EventTypeEcho ServerPushEventType = "echo"
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

// EchoEvent describes the echo request.
type EchoEvent struct {
	// Message describes the message.
	Message string `json:"message"`
}
