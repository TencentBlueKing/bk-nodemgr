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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/retrier"

	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/manager"
	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/messagetracker"
)

// ClientMessagerConfig defines the config.
type ClientMessagerConfig struct {
	// PluginVersion is the plugin version.
	PluginVersion string `json:"plugin_version"`

	// PluginName is the plugin name.
	PluginName string `json:"plugin_name"`

	// DomainSocketPath is the domain socket path when in unix node.
	DomainSocketPath string `json:"domain_socket_path"`

	// MessageTrackerFullPath is the message tracker full path.
	MessageTrackerFullPath string `json:"message_tracker_full_path"`

	// LocalSocketPort is the local socket port when in windows node.
	LocalSocketPort int `json:"local_socket_port"`
}

const (
	clientCheckAckInterval = 10 * time.Millisecond
	clientCheckAckTimeout  = 3 * time.Second
)

// NewClientMessager creates a new client messager.
func NewClientMessager(conf ClientMessagerConfig) (IClientMessager, error) {
	fileMsgTracker, err := messagetracker.NewFileTracker(context.Background(), conf.MessageTrackerFullPath)
	if err != nil {
		return nil, err
	}

	return &clientMessager{
		config:          conf,
		messages:        make(map[string]*synchronousData),
		eventDispatcher: manager.NewDefaultEventDispatcher(),
		fileMsgTracker:  fileMsgTracker,
	}, nil
}

// clientMessager provides the managements for receiving and sending messages via gse agent.
type clientMessager struct {
	config ClientMessagerConfig

	client agentmessage.Client

	messagesMutex sync.RWMutex
	messages      map[string]*synchronousData

	eventDispatcher manager.EventDispatcher

	retrier        *retrier.ExpoBackoff
	fileMsgTracker messagetracker.IMessageTracker
}

// Start starts the messager.
func (m *clientMessager) Start(nCtx contextx.IContext) error {
	logger.G.Sys().With("config", m.config).Info("try to start messager")

	client, err := agentmessage.New(
		agentmessage.WithPluginName(m.config.PluginName),
		agentmessage.WithPluginVersion(m.config.PluginVersion),
		agentmessage.WithDomainSocketPath(m.config.DomainSocketPath),
		agentmessage.WithRecvCallback(m.messageCallback),
		agentmessage.WithLogger(logger.G.Sys()))
	if err != nil {
		return err
	}

	// hang until connected.
	if err = client.Launch(nCtx); err != nil {
		return err
	}

	m.retrier = retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault())
	m.client = client

	logger.G.Sys().Info("started messager")

	return nil
}

// Stop stops the messager.
func (m *clientMessager) Stop(nCtx contextx.IContext) error {
	logger.G.Sys().With("config", m.config).Info("try to stop messager")
	defer logger.G.Sys().Info("stopped messager")

	if m.client != nil {
		return m.client.Terminate(nCtx)
	}

	return nil
}

// EventDispatcher returns the event dispatcher.
func (m *clientMessager) EventDispatcher() manager.EventDispatcher {
	return m.eventDispatcher
}

// messageCallback receives messages from agent.
func (m *clientMessager) messageCallback(messageID string, content []byte) {
	logger.G.Sys().With("message-id", messageID).Info("receive message")

	var base protoRelay.Base
	if err := json.Unmarshal(content, &base); err != nil {
		logger.G.Sys().WithErr(err).With("message-id", messageID, "content", string(content)).Info("failed to unmarshal base message")

		return
	}

	logger.G.Sys().With("message-id", messageID, "type", base.MessageType).Info("begin to handle message")
	switch base.MessageType {
	case protoRelay.MessageTypeCallbackResp:
		go m.setSynchronousData(messageID, content)

		return
	case protoRelay.MessageTypeAckReq:
		go m.handleAck(contextx.New(context.Background()), content)

		return
	case protoRelay.MessageTypeServerPushReq:
		go m.handleServerPush(contextx.New(context.Background()), messageID, content)

		return
	default:
		logger.G.Sys().With("message-id", messageID, "type", base.MessageType).Error("unknown message type")

		return
	}
}

func (m *clientMessager) handleAck(nCtx contextx.IContext, content []byte) {
	var msg protoRelay.AckReq
	if err := json.Unmarshal(content, &msg); err != nil {
		logger.G.Sys().WithErr(err).With("content", string(content)).Error("failed to unmarshal ack message")

		return
	}

	if err := m.fileMsgTracker.MarkAcked(nCtx, msg.OriginalMessageID); err != nil {
		logger.G.Sys().WithErr(err).With("original-message-id", msg.OriginalMessageID).Error("failed to mark acked")

		return
	}

	logger.G.Sys().With("original-message-id", msg.OriginalMessageID).Info("marked messsage acked")
}

func (m *clientMessager) handleServerPush(nCtx contextx.IContext, messageID string, content []byte) {
	go m.sendAck(nCtx, messageID)

	exists, err := m.fileMsgTracker.TryMarkProcessed(nCtx, messageID)
	if err != nil {
		logger.G.Sys().WithErr(err).With("message-id", messageID).Error("failed to mark message process")

		return
	}

	// already processed
	if !exists {
		return
	}

	m.dispatcherServerPushEvent(nCtx, content)
}

func (m *clientMessager) dispatcherServerPushEvent(nCtx contextx.IContext, content []byte) {
	var push protoRelay.ServerPushReq
	if err := json.Unmarshal(content, &push); err != nil {
		logger.G.Sys().WithErr(err).With("content", string(content)).Error("failed to unmarshal server push")

		return
	}

	logger.G.Sys().With("message-id", push.Base.MessageID, "event-type", push.EventType).Info("begin to dispatch event")
	if m.eventDispatcher == nil {
		logger.G.Sys().With("message-id", push.Base.MessageID, "event-type", push.EventType).Error("no event dispatcher registered for event")

		return
	}

	m.eventDispatcher.Dispatch(nCtx, push.EventType, push.Payload)
}

// RequestCallback sends request to url. only transfer the response body to callback.
func (m *clientMessager) RequestCallback(nCtx contextx.IContext, url string, content []byte) ([]byte, int, error) {
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
		return nil, http.StatusInternalServerError, fmt.Errorf("marshal request failed: %w", err)
	}

	logger.G.Sys().With("message-id", messageID).Info("sending request to callback")
	if err = m.client.SendMessage(nCtx, messageID, reqData); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	for {
		select {
		case <-nCtx.Done():
			return nil, http.StatusInternalServerError, nCtx.Err()
		case respData := <-ch:
			var resp protoRelay.CallbackResp
			if err := json.Unmarshal(respData, &resp); err != nil {
				return nil, http.StatusInternalServerError, fmt.Errorf(
					"unmarshal response failed: %w", err)
			}

			return resp.Body, resp.HTTPCode, nil
		}
	}
}

// sendAck sends an ACK to the server for a processed message.
func (m *clientMessager) sendAck(nCtx contextx.IContext, originalMessageID string) {
	ackReq := &protoRelay.AckReq{
		Base: protoRelay.Base{
			MessageID:   identifier.GenMessageID(),
			MessageType: protoRelay.MessageTypeAckReq,
		},
		OriginalMessageID: originalMessageID,
	}

	ackData, err := json.Marshal(ackReq)
	if err != nil {
		logger.G.Sys().WithErr(err).With("original-message-id", originalMessageID).Error("failed to marshal ack request")

		return
	}

	if err := m.client.SendMessage(nCtx, ackReq.MessageID, ackData); err != nil {
		logger.G.Sys().WithErr(err).With("original-message-id", originalMessageID).Error("failed to send ack request")

		return
	}

	logger.G.Sys().With("original-message-id", originalMessageID).Info("ack sent for message")
}

// ClientPushReq sends a client push request asynchronously and returns a channel for results.
func (m *clientMessager) ClientPushReq(nCtx contextx.IContext, callbackURL string, body []byte) <-chan error {
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

		retryErr := m.retrier.Do(nCtx, func(attempt int) error {
			if nCtx.Err() != nil {
				return nil
			}

			logger.G.Sys().With("attempt", attempt, "message-id", messageID, "callback-url", callbackURL).Info("sending client push request")
			if err := m.client.SendMessage(nCtx, messageID, reqData); err != nil {
				return fmt.Errorf("sending client push request failed: %w", err)
			}

			checkAckTicker := time.NewTicker(clientCheckAckInterval)
			defer checkAckTicker.Stop()

			select {
			case <-nCtx.Done():
				return nil

			case <-time.After(clientCheckAckTimeout):
				return fmt.Errorf("wait client push request ack timeout. message-id(%s)", messageID)

			case <-checkAckTicker.C:
				if acked, _ := m.fileMsgTracker.IsAcked(nCtx, messageID); acked {
					logger.G.Sys().With("message-id", messageID, "callback-url", callbackURL).Info("client push request acked successfully")

					return nil
				}
			}

			return nil
		})

		if ctxErr := nCtx.Err(); ctxErr != nil {
			resultChan <- ctxErr
			return
		}

		resultChan <- retryErr
	}()

	return resultChan
}

func (m *clientMessager) newSyncronousData(messageID string) <-chan []byte {
	m.messagesMutex.Lock()
	defer m.messagesMutex.Unlock()

	ch := make(chan []byte, 1)
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
