/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package proxy provides the proxy API.
package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/gin-gonic/gin"
)

// TODO: This url bind with the backend node install workflow.
const (
	// backendCallbackUrlPrefix is the prefix of the callback URL for the backend.
	backendCallbackURLPrefix = "callback/workflow/node_install/"
)

type handler struct {
	rg             *gin.RouterGroup
	provider       discover.Provider
	proxyMessanger relayhandler.ServerMessager
	logger         logger.Logger
}

// newHandler ...
func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:             rg.Group("/proxy"),
		provider:       capability.DiscoverProvider,
		proxyMessanger: capability.ProxyMessager,
		logger:         capability.Logger,
	}
}

// Load ter register the proxy router.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.Any("", h.generalHandler)
	h.rg.Any("/*path", h.generalHandler)
}

func (h *handler) generalHandler(gCtx *gin.Context) {
	gCtx.Status(http.StatusOK)
	ctx := gCtx.Request.Context()

	body, err := io.ReadAll(gCtx.Request.Body)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to read request body. err: %v", err)
		return
	}

	h.logger.InfoCtxf(ctx, "received proxy request: %s", string(body))

	data, err := h.proxyMessanger.DecodeBaseRequest(body)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to decode plugin respond message. err: %v", err)
		return
	}

	switch data.MessageType {
	case protoRelay.MessageTypeCallbackReq:
		h.handleCallback(ctx, data)

		return
	case protoRelay.MessageTypeAckReq:
		h.handleAck(ctx, data)

		return
	case protoRelay.MessageTypeClientPushReq:
		h.handleClientPush(ctx, data)

		return
	default:
		h.logger.ErrorCtxf(ctx, "unknown message type: %s", data.MessageType)
		return
	}
}

func (h *handler) handleAck(ctx context.Context, data *relayhandler.ServerReceivedData) {
	h.logger.InfoCtxf(ctx, "received ack request. agent-id(%s)", data.AgentID)
	msg, err := h.proxyMessanger.DecodeAckRequest(data)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to decode plugin respond message. agent-id(%s), err: %v",
			data.AgentID, err)

		return
	}

	if err := h.proxyMessanger.MarkAcked(ctx, msg.OriginalMessageID); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to handle ack. agent-id(%s), err: %v",
			data.AgentID, err)

		return
	}
	h.logger.InfoCtxf(ctx, "ack handled successfully. agent-id(%s), original-message-id(%s)",
		data.AgentID, msg.OriginalMessageID)
}

func (h *handler) handleCallback(ctx context.Context, data *relayhandler.ServerReceivedData) {
	msg, err := h.proxyMessanger.DecodeCallbackRequest(data)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to decode plugin respond message. agent-id(%s), err: %v",
			data.AgentID, err)

		return
	}

	callbackEndpoint, err := h.provider.GetEndpoint(
		discover.ServiceNameBackend,
		discover.EndpointNameBackendCallback,
		discover.NewRandomSelector())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to get callback endpoint. agent-id(%s), err: %v",
			data.AgentID, err)

		return
	}

	h.logger.InfoCtxf(ctx, "try to redirect request to callback endpoint(%s), agent-id(%s)",
		callbackEndpoint.GetIPV4Address(), data.AgentID)
	resp, err := http.Post(
		fmt.Sprintf("http://%s/%s", callbackEndpoint.GetIPV4Address(), strings.TrimLeft(msg.URL, "/")),
		"application/json",
		bytes.NewReader(msg.Body))
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to send request to callback. agent-id(%s), err: %v",
			data.AgentID, err)

		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to read response body. agent-id(%s), err: %v",
			data.AgentID, err)

		return
	}

	if err := h.proxyMessanger.RespondCallback(
		ctx,
		msg.MessageID,
		resp.StatusCode,
		body,
		data.AgentID,
	); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to respond proxy callback. agent-id(%s), err: %v",
			data.AgentID, err)

		return
	}

	h.logger.InfoCtxf(ctx, "responded proxy callback. agent-id(%s)", data.AgentID)
}

func (h *handler) handleClientPush(ctx context.Context, data *relayhandler.ServerReceivedData) {
	go h.proxyMessanger.SendAck(ctx, data.MessageID, data.AgentID)

	marked, err := h.proxyMessanger.TryMarkProcessed(ctx, data.MessageID)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to mark message process. message-id(%s), err: %v", data.MessageID, err)
		return
	}

	if !marked {
		h.logger.InfoCtxf(ctx, "message already processed. message-id(%s)", data.MessageID)
		return
	}

	msg, err := h.proxyMessanger.DecodeClientPushRequest(data)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to decode plugin respond message. agent-id(%s), err: %v",
			data.AgentID, err)

		return
	}

	go h.callbackBackend(ctx, msg, data.AgentID)
}

func (h *handler) callbackBackend(ctx context.Context, msg *protoRelay.ClientPushReq, agentID string) {
	callbackEndpoint, err := h.provider.GetEndpoint(
		discover.ServiceNameBackend,
		discover.EndpointNameBackendCallback,
		discover.NewRandomSelector())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to get callback endpoint. agent-id(%s), err: %v",
			agentID, err)

		return
	}

	h.logger.InfoCtxf(ctx, "try to redirect request to callback endpoint(%s), agent-id(%s)",
		callbackEndpoint.GetIPV4Address(), agentID)

	url := backendCallbackURLPrefix + msg.URL

	resp, err := http.Post(
		fmt.Sprintf("http://%s/%s", callbackEndpoint.GetIPV4Address(), url),
		"application/json",
		bytes.NewReader(msg.Body))

	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to send request to callback. agent-id(%s), err: %v",
			agentID, err)

		return
	}

	if resp.StatusCode != http.StatusOK {
		h.logger.ErrorCtxf(ctx, "failed to send request to callback. agent-id(%s), status-code(%d)",
			agentID, resp.StatusCode)

		return
	}
}
