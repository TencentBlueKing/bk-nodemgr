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
	"fmt"
	"net/http"
	"sync"
	"time"

	agentmessage "github.com/TencentBlueKing/bk-gse-sdk/go/service/agent-message"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/retrier"

	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/manager"
	relayStorage "github.com/TencentBlueKing/bk-nodemgr/internal/relay/storage"
)

// ClientMessagerConfig defines the config.
type ClientMessagerConfig struct {
	// PluginVersion is the plugin version.
	PluginVersion string `json:"plugin_version"`

	PluginName string `json:"plugin_name"`

	// DomainSocketPath is the domain socket path when in unix node.
	DomainSocketPath string `json:"domain_socket_path"`

	// MessageIDFullPath is the full path for message ID storage.
	MessageIDPath string `json:"message_id_path"`

	// LocalSocketPort is the local socket port when in windows node.
	LocalSocketPort int `json:"local_socket_port"`

	// Logger is the logger.
	Logger logger.Logger
}

// NewClientMessager creates a new client messager.
func NewClientMessager(conf ClientMessagerConfig) *clientMessager {
	return &clientMessager{
		config:          conf,
		messages:        make(map[string]*synchronousData),
		eventDispatcher: manager.NewDefaultEventDispatcher(),
		fileStorage:     relayStorage.NewFileManager(conf.MessageIDPath),
	}
}

// clientMessager provides the managements for receiving and sending messages via gse agent.
type clientMessager struct {
	config ClientMessagerConfig

	client agentmessage.Client

	messagesMutex sync.RWMutex
	messages      map[string]*synchronousData

	eventDispatcher manager.EventDispatcher

	retrier     *retrier.ExpoBackoff
	fileStorage relayStorage.MessageStore
}

// Start starts the messager.
func (m *clientMessager) Start(ctx context.Context) error {
	m.config.Logger.Infof("try to start messager: %+v", m.config)

	client, err := agentmessage.New(
		agentmessage.WithPluginName(m.config.PluginName),
		agentmessage.WithPluginVersion(m.config.PluginVersion),
		agentmessage.WithDomainSocketPath(m.config.DomainSocketPath),
		agentmessage.WithRecvCallback(m.messageCallback),
		agentmessage.WithLogger(&loggerAdaptor{Logger: m.config.Logger}))
	if err != nil {
		return err
	}

	go func() {
		ticker := time.NewTicker(12 * time.Hour) //nolint: mnd
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := m.fileStorage.CleanupExpired(ctx); err != nil {
					m.config.Logger.Errorf("failed to cleanup expired messages, err: %v", err)
					return
				}
			}
		}
	}()

	// hang until connected.
	if err = client.Launch(ctx); err != nil {
		return err
	}

	m.retrier = retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault())

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

// EventDispatcher returns the event dispatcher.
func (m *clientMessager) EventDispatcher() manager.EventDispatcher {
	return m.eventDispatcher
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

		return
	case protoRelay.MessageTypeAckReq:
		go m.handleAck(context.Background(), content)

		return
	case protoRelay.MessageTypeServerPushReq:
		m.handleServerPush(context.Background(), messageID, content)

		return
	default:
		m.config.Logger.Errorf("unknown message type. type(%s)", base.MessageType)
		return
	}
}

func (m *clientMessager) handleAck(ctx context.Context, content []byte) {
	var msg protoRelay.AckReq
	if err := json.Unmarshal(content, &msg); err != nil {
		m.config.Logger.Errorf("invalid push format, err: %v", err)
		return
	}

	if err := m.fileStorage.MarkedAcked(ctx, msg.OriginalMessageID); err != nil {
		m.config.Logger.Errorf("failed to mark acked. original-message-id(%s), err: %v", msg.OriginalMessageID, err)
	}

	m.config.Logger.Infof("ack received for message. original-message-id(%s)", msg.OriginalMessageID)
}

func (m *clientMessager) handleServerPush(ctx context.Context, messageID string, content []byte) {
	go m.sendAck(ctx, messageID)

	processed, err := m.fileStorage.IsProcessed(ctx, messageID)
	if err != nil {
		m.config.Logger.Errorf("failed to check if message is processed. message-id(%s), err: %v", messageID, err)

		return
	}

	if processed {
		m.config.Logger.Infof("message already processed. message-id(%s)", messageID)
		return
	}

	if err := m.fileStorage.MarkProcessed(ctx, messageID); err != nil {
		m.config.Logger.Errorf("failed to mark message as processed. message-id(%s), error: %v", messageID, err)
		return
	}

	go m.dispatcherServerPushEvent(content)
}

func (m *clientMessager) dispatcherServerPushEvent(content []byte) {
	var push protoRelay.ServerPushReq
	if err := json.Unmarshal(content, &push); err != nil {
		m.config.Logger.Errorf("invalid push format, err: %v", err)
		return
	}

	if m.eventDispatcher == nil {
		m.config.Logger.Errorf("no event dispatcher registered for event. event-type(%s)", push.EventType)
		return
	}

	m.config.Logger.Infof("dispatching event. event-type(%s)", push.EventType)
	m.eventDispatcher.Dispatch(push.EventType, push.Payload)
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

// sendAck sends an ACK to the server for a processed message.
func (m *clientMessager) sendAck(ctx context.Context, originalMessageID string) {
	ackReq := &protoRelay.AckReq{
		Base: protoRelay.Base{
			MessageID:   identifier.GenMessageID(),
			MessageType: protoRelay.MessageTypeAckReq,
		},
		OriginalMessageID: originalMessageID,
	}

	ackData, err := json.Marshal(ackReq)
	if err != nil {
		m.config.Logger.Errorf("failed to marshal ack request, err: %v", err)
	}

	if err := m.client.SendMessage(ctx, ackReq.MessageID, ackData); err != nil {
		m.config.Logger.Errorf("failed to send ack request, err: %v", err)
	}

	m.config.Logger.Infof("ack sent for message. message-id(%s)", originalMessageID)
}

// ClientPushReq sends a client push request asynchronously and returns a channel for results.
func (m *clientMessager) ClientPushReq(ctx context.Context, callbackURL string, body []byte) <-chan error {
	resultChan := make(chan error, 1)

	if callbackURL == "" {
		resultChan <- errors.New("invalid url")
		return resultChan
	}

	messageID := identifier.GenMessageID()
	req := &protoRelay.CallbackReq{
		Base: protoRelay.Base{
			MessageID:   messageID,
			MessageType: protoRelay.MessageTypeClientPushReq,
		},
		URL:  callbackURL,
		Body: body,
	}
	reqData, err := json.Marshal(req)
	if err != nil {
		resultChan <- fmt.Errorf("marshal request failed: %w", err)
		return resultChan
	}

	go func() {
		defer close(resultChan)

		retryCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		retryErr := m.retrier.Do(retryCtx, func(attempt int) error {
			select {
			case <-retryCtx.Done():
				return retryCtx.Err()
			default:
			}

			m.config.Logger.Infof("sending client push request (attempt %d). message-id(%s)", attempt, messageID)

			if err := m.client.SendMessage(retryCtx, messageID, reqData); err != nil {
				return fmt.Errorf("send message failed, err: %w", err)
			}

			acked, err := m.fileStorage.IsAcked(retryCtx, messageID)
			if err != nil {
				return fmt.Errorf("check ack failed, err: %w", err)
			}
			if !acked {
				return errors.New("ack not received")
			}

			m.config.Logger.Infof("client push request acknowledged. message-id(%s)", messageID)

			return nil
		})

		resultChan <- retryErr
	}()

	return resultChan
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
