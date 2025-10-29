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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
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
	proxyMessanger relayhandler.IServerMessager
}

// newHandler ...
func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:             rg.Group("/proxy"),
		provider:       capability.DiscoverProvider,
		proxyMessanger: capability.ProxyMessager,
	}
}

// Load ter register the proxy router.
func Load(rg *gin.RouterGroup, capability *options.Capability) restserver.IMiddlewareChain {
	h := newHandler(rg, capability)

	h.rg.Any("", h.generalHandler)
	h.rg.Any("/*path", h.generalHandler)

	return restserver.NewMiddlewareChain(h.rg)
}

func (h *handler) generalHandler(gCtx *gin.Context) {
	gCtx.Status(http.StatusOK)
	nCtx := contextx.New(gCtx.Request.Context())

	body, err := io.ReadAll(gCtx.Request.Body)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to read request body")

		return
	}

	logger.G.Biz(nCtx).Info("received proxy request: %s", string(body))

	data, err := h.proxyMessanger.DecodeBaseRequest(body)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to decode plugin respond message")

		return
	}

	switch data.MessageType {
	case protoRelay.MessageTypeCallbackReq:
		h.handleCallback(nCtx, data)

		return
	case protoRelay.MessageTypeAckReq:
		h.handleAck(nCtx, data)

		return
	case protoRelay.MessageTypeClientPushReq:
		h.handleClientPush(nCtx, data)

		return
	default:
		logger.G.Biz(nCtx).With("type", data.MessageType).Error("unknown message type")

		return
	}
}

func (h *handler) handleAck(nCtx contextx.IContext, data *relayhandler.ServerReceivedData) {
	logger.G.Biz(nCtx).With("agent-id", data.AgentID).Info("received ack request")

	msg, err := h.proxyMessanger.DecodeAckRequest(data)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("agent-id", data.AgentID).Error("failed to decode ack request message")

		return
	}

	if err := h.proxyMessanger.MarkAcked(nCtx, msg.OriginalMessageID); err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("agent-id", data.AgentID).Error("failed to handle ack")

		return
	}

	logger.G.Biz(nCtx).With("agent-id", data.AgentID, "original-message-id", msg.OriginalMessageID).Info("ack handled successfully")
}

func (h *handler) handleCallback(nCtx contextx.IContext, data *relayhandler.ServerReceivedData) {
	msg, err := h.proxyMessanger.DecodeCallbackRequest(data)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("agent-id", data.AgentID).Error("failed to decode callback request message")

		return
	}

	callbackEndpoint, err := h.provider.GetEndpoint(
		discover.ServiceNameBackend,
		discover.EndpointNameBackendCallback,
		discover.NewRandomSelector())
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("agent-id", data.AgentID).Error("failed to get callback endpoint")

		return
	}

	url := fmt.Sprintf("http://%s/%s", callbackEndpoint.GetIPV4Address(), strings.TrimLeft(msg.URL, "/"))
	logger.G.Biz(nCtx).With("agent-id", data.AgentID, "callback-url", url).Info("try to redirect request to callback")
	resp, err := http.Post(
		url,
		"application/json",
		bytes.NewReader(msg.Body))
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("agent-id", data.AgentID, "callback-url", url).Error("failed to send request to callback")

		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("agent-id", data.AgentID, "callback-url", url).Error("failed to read response body")

		return
	}

	if err := h.proxyMessanger.RespondCallback(
		nCtx,
		msg.MessageID,
		resp.StatusCode,
		body,
		data.AgentID,
	); err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("agent-id", data.AgentID, "callback-url", url).Error("failed to respond proxy callback")

		return
	}

	logger.G.Biz(nCtx).With("agent-id", data.AgentID, "callback-url", url).Info("responded proxy callback")
}

func (h *handler) handleClientPush(nCtx contextx.IContext, data *relayhandler.ServerReceivedData) {
	go h.proxyMessanger.SendAck(nCtx, data.MessageID, data.AgentID)

	marked, err := h.proxyMessanger.TryMarkProcessed(nCtx, data.MessageID)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("message-id", data.MessageID, "agent-id", data.AgentID).Error("failed to mark message process")

		return
	}

	if !marked {
		logger.G.Biz(nCtx).With("message-id", data.MessageID, "agent-id", data.AgentID).Info("message already processed")

		return
	}

	msg, err := h.proxyMessanger.DecodeClientPushRequest(data)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("message-id", data.MessageID, "agent-id", data.AgentID).Error("failed to decode client push request message")

		return
	}

	go h.callbackBackend(nCtx, msg, data.AgentID)
}

func (h *handler) callbackBackend(nCtx contextx.IContext, msg *protoRelay.ClientPushReq, agentID string) {
	callbackEndpoint, err := h.provider.GetEndpoint(
		discover.ServiceNameBackend,
		discover.EndpointNameBackendCallback,
		discover.NewRandomSelector())
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("agent-id", agentID).Error("failed to get callback endpoint")

		return
	}

	url, err := url.JoinPath(backendCallbackURLPrefix, msg.URL)
	if err != nil {
		logger.G.Biz(nCtx).
			WithErr(err).
			With("agent-id", agentID, "prefix-url", backendCallbackURLPrefix, "msg-url", msg.URL).
			Error("failed to format callback endpoint url")

		return
	}

	logger.G.Biz(nCtx).With("agent-id", agentID, "callback-url", url).Info("try to redirect request to callback")
	resp, err := http.Post(
		fmt.Sprintf("http://%s/%s", callbackEndpoint.GetIPV4Address(), url),
		"application/json",
		bytes.NewReader(msg.Body))

	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("agent-id", agentID).Error("failed to send request to callback")

		return
	}

	if resp.StatusCode != http.StatusOK {
		logger.G.Biz(nCtx).WithErr(err).With("agent-id", agentID, "status-code", resp.StatusCode).Error("failed to send request to callback")

		return
	}
}
