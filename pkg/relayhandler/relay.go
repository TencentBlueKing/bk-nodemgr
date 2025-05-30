/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package relayhandler provides the handler for both server and client side.
package relayhandler

import (
	"context"

	protoProxy "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/proxy"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
)

const (
	pluginName = "bk-nodemgr-proxy"
)

// ClientMessager provides the managements for receiving and sending messages via client side gse agent.
type ClientMessager interface {
	// Start starts the messager.
	Start(ctx context.Context) error

	// Stop stops the messager.
	Stop(ctx context.Context) error

	CallbackClient
}

// CallbackClient defines the callback client.
type CallbackClient interface {
	// RequestCallback sends request to url. returns the response body and http code.
	RequestCallback(ctx context.Context, url string, content []byte) ([]byte, int, error)
}

// ServerMessager provides the managements for receiving and sending messages via server side gse api.
type ServerMessager interface {
	// Start starts the messager.
	Start(ctx context.Context) error

	// Stop stops the messager.
	Stop(ctx context.Context) error

	// DecodeBaseRequest decodes the base request.
	DecodeBaseRequest(req []byte) (*ServerReceivedData, error)

	CallbackServer
}

// ServerReceivedData defines the server received data.
type ServerReceivedData struct {
	MessageID   string
	MessageType protoProxy.MessageType
	AgentID     string
	Content     []byte
}

// CallbackServer defines the callback server.
type CallbackServer interface {
	// DecodeCallbackRequest decodes the callback request.
	DecodeCallbackRequest(data *ServerReceivedData) (*protoProxy.CallbackReq, error)

	// RespondCallback sends the callback response.
	RespondCallback(ctx context.Context, messageID string, httpCode int, content []byte, agentIDs ...string) error
}

type loggerAdaptor struct {
	Logger logger.Logger
}

// Debug logs to DEBUG log.
func (l *loggerAdaptor) Debug(format string, args ...interface{}) { l.Logger.Debugf(format, args...) }

// Info logs to INFO log.
func (l *loggerAdaptor) Info(format string, args ...interface{}) { l.Logger.Infof(format, args...) }

// Warn logs to WARNING log.
func (l *loggerAdaptor) Warn(format string, args ...interface{}) { l.Logger.Warnf(format, args...) }

// Error logs to ERROR log.
func (l *loggerAdaptor) Error(format string, args ...interface{}) { l.Logger.Errorf(format, args...) }
