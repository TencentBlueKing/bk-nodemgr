/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package relayhandler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	agentmessage "github.com/TencentBlueKing/bk-gse-sdk/go/service/agent-message"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
)

// ClientMessagerConfig defines the config.
type ClientMessagerConfig struct {
	// PluginVersion is the plugin version.
	PluginVersion string `json:"plugin_version"`

	// DomainSocketPath is the domain socket path when in unix node.
	DomainSocketPath string `json:"domain_socket_path"`

	// LocalSocketPort is the local socket port when in windows node.
	LocalSocketPort int `json:"local_socket_port"`

	// Logger is the logger.
	Logger logger.Logger
}

// HandlerFunc defines the handler.
type HandlerFunc func([]byte)

// NewClientMessager creates a new client messager.
func NewClientMessager(conf ClientMessagerConfig) *clientMessager {
	return &clientMessager{
		config:   conf,
		messages: make(map[string]*synchronousData),
		handlers: make(map[protoRelay.EventType]HandlerFunc),
	}
}

// clientMessager provides the managements for receiving and sending messages via gse agent.
type clientMessager struct {
	config ClientMessagerConfig

	client agentmessage.Client

	messagesMutex sync.RWMutex
	messages      map[string]*synchronousData

	handlers   map[protoRelay.EventType]HandlerFunc
	handlerMux sync.RWMutex
}

// Start starts the messager.
func (m *clientMessager) Start(ctx context.Context) error {
	m.config.Logger.Infof("try to start messager: %+v", m.config)

	client, err := agentmessage.New(
		agentmessage.WithPluginName(pluginName),
		agentmessage.WithPluginVersion(m.config.PluginVersion),
		agentmessage.WithDomainSocketPath(m.config.DomainSocketPath),
		agentmessage.WithRecvCallback(m.messageCallback),
		agentmessage.WithLogger(&loggerAdaptor{Logger: m.config.Logger}))
	if err != nil {
		return err
	}

	// hang until connected.
	if err = client.Launch(ctx); err != nil {
		return err
	}

	m.client = client
	m.config.Logger.Infof("started messager")

	return nil
}

// Stop stops the messager.
func (m *clientMessager) Stop(ctx context.Context) error {
	m.config.Logger.Infof("try to stop messager: %+v", m.config)
	if m.client != nil {
		return m.client.Terminate(ctx)
	}

	return nil
}

// messageCallback receives messages from agent.
func (m *clientMessager) messageCallback(messageID string, content []byte) {
	m.config.Logger.Infof("receive message. message-id(%s), content(%s)", messageID, string(content))

	var base protoRelay.Base
	if err := json.Unmarshal(content, &base); err != nil {
		return
	}

	switch base.MessageType {
	case protoRelay.MessageTypeCallbackResp:
		go m.setSynchronousData(messageID, content)
	case protoRelay.MessageTypeServerPush:
		var push protoRelay.ServerPush
		if err := json.Unmarshal(content, &push); err != nil {
			m.config.Logger.Warnf("Invalid push format: %v", err)
			return
		}

		m.handlerMux.RLock()
		handler, exists := m.handlers[push.EventType]
		m.handlerMux.RUnlock()

		if !exists {
			m.config.Logger.Warnf("Unhandled event type: %s", push.EventType)
		}
		go handler(push.Payload)
	}
}

// RequestCallback sends request to url. returns the response body and http code.
func (m *clientMessager) RequestCallback(ctx context.Context, url string, content []byte) ([]byte, int, error) {
	if url == "" {
		return nil, http.StatusInternalServerError, errors.New("invalid url")
	}

	messageID := identifier.GenMessageID()
	ch := m.newSyncronousData(messageID)

	req := &protoRelay.CallbackReq{
		Base: protoRelay.Base{
			MessageID:   messageID,
			MessageType: protoRelay.MessageTypeCallbackReq,
		},
		URL:  url,
		Body: content,
	}
	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	if err = m.client.SendMessage(ctx, messageID, reqData); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	for {
		select {
		case <-ctx.Done():
			return nil, http.StatusInternalServerError, ctx.Err()
		case respData := <-ch:
			var resp protoRelay.CallbackResp
			if err := json.Unmarshal(respData, &resp); err != nil {
				return nil, http.StatusInternalServerError, err
			}

			return resp.Body, resp.HTTPCode, nil
		}
	}
}

// RegisterHandler registers the handler.
func (m *clientMessager) RegisterHandler(eventType protoRelay.EventType, handler HandlerFunc) {
	m.handlerMux.Lock()
	defer m.handlerMux.Unlock()
	m.handlers[eventType] = handler
}

func (m *clientMessager) newSyncronousData(messageID string) <-chan []byte {
	m.messagesMutex.Lock()
	defer m.messagesMutex.Unlock()

	ch := make(chan []byte)
	m.messages[messageID] = &synchronousData{
		content:  ch,
		createAt: time.Now(),
	}

	return ch
}

func (m *clientMessager) setSynchronousData(messageID string, content []byte) {
	m.messagesMutex.RLock()
	defer m.messagesMutex.RUnlock()

	data, ok := m.messages[messageID]
	if !ok {
		return
	}

	data.content <- content
}

type synchronousData struct {
	content  chan []byte
	createAt time.Time
}
