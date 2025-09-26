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
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	serverapi "github.com/TencentBlueKing/bk-gse-sdk/go/service/server-api"
	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/messagetracker"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rediscache"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/retrier"
	"github.com/redis/go-redis/v9"
)

// ServerMessagerConfig defines the config.
type ServerMessagerConfig struct {
	// SlotID is the gse slot id.
	SlotID int

	// Token is the gse slot token.
	Token string

	// AppCode defines the app code for apigw.
	AppCode string

	// AppSecret defines the app secret for apigw.
	AppSecret string

	// GSEBaseURL defines the gse base url.
	GSEBaseURL string

	// SkipTLSVerify defines the skip tls verify.
	SkipTLSVerify bool

	// RedisClient is the redis client for storing pending messages.
	RedisClient *redis.Client
}

const (
	serverCheckAckInterval = 10 * time.Millisecond
	serverCheckAckTimeout  = 3 * time.Second
)

// NewServerMessager creates a new server messager.
func NewServerMessager(conf ServerMessagerConfig) IServerMessager {
	return &serverMessager{
		config: conf,
	}
}

var _ IServerMessager = &serverMessager{}

var (
	// nolint: revive
	errDispatchPartialFailed = errors.New("dispatch message partial failed")
)

// serverMessager provides the managements for receiving and sending messages via gse cluster.
type serverMessager struct {
	config ServerMessagerConfig

	client serverapi.Client

	retrier *retrier.ExpoBackoff

	redisMsgTracker messagetracker.IMessageTracker
}

// Start starts the messager.
func (m *serverMessager) Start(_ contextx.IContext) error {
	logger.G.Sys().With("config", m.config).Info("try to start messager")

	// initialize http client.
	httpClient := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: m.config.SkipTLSVerify},
	}}

	client, err := serverapi.New(
		serverapi.WithBaseURL(m.config.GSEBaseURL),
		serverapi.WithClient(httpClient),
		serverapi.WithClusterAuth(m.config.SlotID, m.config.Token),
		serverapi.WithAPIGwAuth(m.config.AppCode, m.config.AppSecret),
		serverapi.WithLogger(logger.G.Sys()))
	if err != nil {
		return err
	}

	m.client = client

	m.redisMsgTracker = messagetracker.NewRedisTracker(
		rediscache.NewRedisCache(m.config.RedisClient, rediscache.DefaultTimeout))
	m.retrier = retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault())

	logger.G.Sys().Info("started messager")

	return nil
}

// Stop stops the messager.
func (m *serverMessager) Stop(_ contextx.IContext) error {
	logger.G.Sys().With("config", m.config).Info("try to stop messager")

	logger.G.Sys().Info("stopped messager")

	return nil
}

// DecodeBaseRequest decodes the base request.
func (m *serverMessager) DecodeBaseRequest(req []byte) (*ServerReceivedData, error) {
	data, err := m.client.Cluster().EncoderDecoder().DecodePluginRespondMessageCallback(req)
	if err != nil {
		return nil, err
	}

	base := new(protoRelay.Base)
	if err = json.Unmarshal([]byte(data.Content), base); err != nil {
		return nil, fmt.Errorf("failed to decode base request content. data(%s): %w", data.Content, err)
	}

	return &ServerReceivedData{
		MessageID:   base.MessageID,
		MessageType: base.MessageType,
		AgentID:     data.AgentID,
		Content:     []byte(data.Content),
	}, nil
}

// DecodeCallbackRequest decodes the callback request.
func (m *serverMessager) DecodeCallbackRequest(data *ServerReceivedData) (*protoRelay.CallbackReq, error) {
	result := new(protoRelay.CallbackReq)
	if err := json.Unmarshal(data.Content, result); err != nil {
		return nil, fmt.Errorf("failed to decode callback request content. data(%s): %w", data.Content, err)
	}

	return result, nil
}

// DecodeCallbackRequest decodes the callback request.
func (m *serverMessager) DecodeAckRequest(data *ServerReceivedData) (*protoRelay.AckReq, error) {
	result := new(protoRelay.AckReq)
	if err := json.Unmarshal(data.Content, result); err != nil {
		return nil, fmt.Errorf("failed to decode ack request content. data(%s): %w", data.Content, err)
	}

	return result, nil
}

// DecodeClientPushRequest decodes the client push request.
func (m *serverMessager) DecodeClientPushRequest(data *ServerReceivedData) (
	*protoRelay.ClientPushReq, error) {

	result := new(protoRelay.ClientPushReq)
	if err := json.Unmarshal(data.Content, result); err != nil {
		return nil, fmt.Errorf("failed to decode client push request content. data(%s): %w", data.Content, err)
	}

	return result, nil
}

// RespondCallback sends the callback resp.
func (m *serverMessager) RespondCallback(nCtx contextx.IContext,
	messageID string, httpCode int, content []byte, agentIDs ...string) error {

	resp := &protoRelay.CallbackResp{
		Base: protoRelay.Base{
			MessageID:   messageID,
			MessageType: protoRelay.MessageTypeCallbackResp,
		},
		HTTPCode: httpCode,
		Body:     content,
	}
	respData, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("marshal callback resp failed: %w", err)
	}

	result, err := m.client.Cluster().PluginDispatchMessage(nCtx, messageID, respData, agentIDs...)
	if err != nil {
		return err
	}

	if result.Code != 0 || len(result.AgentResults) > 0 {
		err = errDispatchPartialFailed

		logger.G.Biz(nCtx).
			WithErr(err).
			With("code", result.Code, "failed-agent-results", conv.MapKeyToSlice(result.AgentResults)).
			Warn("failed to send callback resp to agents")

		return err
	}

	return nil
}

// PushToClient sends the server push to client asynchronously and returns a channel for results.
func (m *serverMessager) PushToClient(
	nCtx contextx.IContext, eventType protoRelay.ServerPushEventType, payload []byte, agentIDs ...string) <-chan error {

	resultChan := make(chan error, 1)

	messageID := identifier.GenMessageID()
	req := &protoRelay.ServerPushReq{
		Base: protoRelay.Base{
			MessageType: protoRelay.MessageTypeServerPushReq,
			MessageID:   messageID,
		},
		EventType: eventType,
		Payload:   payload,
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

			logger.G.Biz(nCtx).With("attempt", attempt, "message-id", messageID).Info("sending message to client")

			result, err := m.client.Cluster().PluginDispatchMessage(
				nCtx, messageID, reqData, agentIDs...)
			if err != nil {
				return fmt.Errorf("sending message to client failed: %w", err)
			}

			if result.Code != 0 || len(result.AgentResults) > 0 {
				err = errDispatchPartialFailed

				logger.G.Biz(nCtx).
					WithErr(err).
					With("code", result.Code, "failed-agent-results", conv.MapKeyToSlice(result.AgentResults)).
					Warn("failed to send message to client")

				return err
			}

			checkAckTicker := time.NewTicker(serverCheckAckInterval)
			defer checkAckTicker.Stop()

			select {
			case <-nCtx.Done():
				return nil

			case <-time.After(serverCheckAckTimeout):
				return fmt.Errorf("wait message to client ack timeout. message-id(%s)", messageID)

			case <-checkAckTicker.C:
				if acked, _ := m.isMessageAcked(nCtx, messageID); acked {
					logger.G.Biz(nCtx).With("message-id", messageID).Info("message to client acked successfully")

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

// SendAck sends the ack to client.
func (m *serverMessager) SendAck(nCtx contextx.IContext, originalMessageID string, agentIDs ...string) {
	messageID := identifier.GenMessageID()
	resp := &protoRelay.AckReq{
		Base: protoRelay.Base{
			MessageID:   messageID,
			MessageType: protoRelay.MessageTypeAckReq,
		},
		OriginalMessageID: originalMessageID,
	}
	respData, err := json.Marshal(resp)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to marshal ack request")
	}

	result, err := m.client.Cluster().PluginDispatchMessage(nCtx, messageID, respData, agentIDs...)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to send ack")

		return
	}

	if result.Code != 0 || len(result.AgentResults) > 0 {
		err = errDispatchPartialFailed

		logger.G.Biz(nCtx).
			WithErr(err).
			With("code", result.Code, "failed-agent-results", conv.MapKeyToSlice(result.AgentResults)).
			Warn("failed to send ack to agents")

		return
	}

	logger.G.Biz(nCtx).With("original-message-id", originalMessageID).Info("ack sent to client")
}

func (m *serverMessager) isMessageAcked(ctx context.Context, mid string) (bool, error) {
	acked, err := m.redisMsgTracker.IsAcked(ctx, mid)
	if err != nil {
		return false, err
	}

	return acked, nil
}

// MarkProcessed marks a message ID as processed if it has been processed, return false.
func (m *serverMessager) TryMarkProcessed(nCtx contextx.IContext, mid string) (bool, error) {
	marked, err := m.redisMsgTracker.TryMarkProcessed(nCtx, mid)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("message-id", mid).Error("failed to mark processed")

		return false, fmt.Errorf("failed to mark processed: %w", err)
	}

	return marked, nil
}

// MarkAcked marks a message ID as acked.
func (m *serverMessager) MarkAcked(nCtx contextx.IContext, mid string) error {
	if err := m.redisMsgTracker.MarkAcked(nCtx, mid); err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("message-id", mid).Error("failed to mark ack")

		return fmt.Errorf("failed to mark acked: %w", err)
	}

	logger.G.Biz(nCtx).With("message-id", mid).Info("ack received")

	return nil
}
