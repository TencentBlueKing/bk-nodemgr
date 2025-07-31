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
	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/messagetracke"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rediscache"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
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

	// Logger is the logger.
	Logger logger.Logger
}

// NewServerMessager creates a new server messager.
func NewServerMessager(conf ServerMessagerConfig) *serverMessager {
	return &serverMessager{
		config: conf,
	}
}

// serverMessager provides the managements for receiving and sending messages via gse cluster.
type serverMessager struct {
	config ServerMessagerConfig

	client serverapi.Client

	retrier *retrier.ExpoBackoff

	redisStore messagetracke.MessageTracker
}

// Start starts the messager.
func (m *serverMessager) Start(_ context.Context) error {
	m.config.Logger.Infof("try to start messager: %+v", m.config)

	// initialize http client.
	httpClient := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: m.config.SkipTLSVerify},
	}}

	client, err := serverapi.New(
		serverapi.WithBaseURL(m.config.GSEBaseURL),
		serverapi.WithClient(httpClient),
		serverapi.WithClusterAuth(m.config.SlotID, m.config.Token),
		serverapi.WithAPIGwAuth(m.config.AppCode, m.config.AppSecret),
		serverapi.WithLogger(&loggerAdaptor{Logger: m.config.Logger}))
	if err != nil {
		return err
	}

	m.client = client

	m.redisStore = messagetracke.NewRedisStore(rediscache.NewRedisCache(m.config.RedisClient, 12*time.Hour)) //nolint: mnd
	m.retrier = retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault())

	m.config.Logger.Infof("started messager")

	return nil
}

// Stop stops the messager.
func (m *serverMessager) Stop(_ context.Context) error {
	m.config.Logger.Infof("try to stop messager: %+v", m.config)

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
		return nil, err
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
		return nil, err
	}

	return result, nil
}

// DecodeCallbackRequest decodes the callback request.
func (m *serverMessager) DecodeAckRequest(data *ServerReceivedData) (*protoRelay.AckReq, error) {
	result := new(protoRelay.AckReq)
	if err := json.Unmarshal(data.Content, result); err != nil {
		return nil, err
	}

	return result, nil
}

// DecodeClientPushRequest decodes the client push request.
func (m *serverMessager) DecodeClientPushRequest(data *ServerReceivedData) (
	*protoRelay.ClientPushReq, error) {

	result := new(protoRelay.ClientPushReq)
	if err := json.Unmarshal(data.Content, result); err != nil {
		return nil, err
	}

	return result, nil
}

// RespondCallback sends the callback resp.
func (m *serverMessager) RespondCallback(ctx context.Context,
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
		return err
	}

	result, err := m.client.Cluster().PluginDispatchMessage(ctx, messageID, respData, agentIDs...)
	if err != nil {
		return err
	}

	if result.Code != 0 || len(result.AgentResults) > 0 {
		err = fmt.Errorf("failed to send callback resp to agents. code(%d), agent-results(%v)",
			result.Code, conv.MapKeyToSlice(result.AgentResults))
		m.config.Logger.WarnCtxf(ctx, "%v", err)

		return err
	}

	return nil
}

// PushToClient sends the server push to client asynchronously and returns a channel for results.
func (m *serverMessager) PushToClient(ctx context.Context,
	eventType protoRelay.ServerPushEventType, payload []byte, agentIDs ...string) <-chan error {

	resultChan := make(chan error, 1)

	messageID := identifier.GenMessageID()
	push := &protoRelay.ServerPushReq{
		Base: protoRelay.Base{
			MessageID:   messageID,
			MessageType: protoRelay.MessageTypeServerPushReq,
		},
		EventType: eventType,
		Payload:   payload,
	}
	pushData, err := json.Marshal(push)
	if err != nil {
		resultChan <- fmt.Errorf("marshal push data failed: %w", err)
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

			m.config.Logger.Infof("sending message (attempt %d). message-id(%s)", attempt, messageID)

			result, err := m.client.Cluster().PluginDispatchMessage(
				retryCtx, push.MessageID, pushData, agentIDs...)
			if err != nil {
				return fmt.Errorf("dispatch message failed, err: %w", err)
			}

			if result.Code != 0 || len(result.AgentResults) > 0 {
				errMsg := fmt.Sprintf("failed to send to agents. code(%d), agent-results(%v)",
					result.Code, conv.MapKeyToSlice(result.AgentResults))
				m.config.Logger.WarnCtxf(retryCtx, "%s", errMsg)

				return errors.New(errMsg)
			}

			acked, err := m.isMessageAcked(retryCtx, messageID)
			if err != nil {
				return fmt.Errorf("check ack failed, err: %w", err)
			}
			if !acked {
				return errors.New("ack not received")
			}

			m.config.Logger.Infof("message acknowledged. message-id(%s)", messageID)

			return nil
		})

		resultChan <- retryErr
	}()

	return resultChan
}

// SendAck sends the ack to client.
func (m *serverMessager) SendAck(ctx context.Context, originalMessageID string, agentIDs ...string) {
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
		m.config.Logger.Errorf("failed to marshal ack request, err: %v", err)
	}

	result, err := m.client.Cluster().PluginDispatchMessage(ctx, messageID, respData, agentIDs...)
	if err != nil {
		m.config.Logger.Errorf("failed to send ack, err: %v", err)
		return
	}

	if result.Code != 0 || len(result.AgentResults) > 0 {
		err = fmt.Errorf("failed to send ack to agents. code(%d), agent-results(%v)",
			result.Code, conv.MapKeyToSlice(result.AgentResults))
		m.config.Logger.WarnCtxf(ctx, "%v", err)

		return
	}

	m.config.Logger.Infof("ack sent for message. original-message-id(%s)", originalMessageID)
}

func (m *serverMessager) isMessageAcked(ctx context.Context, mid string) (bool, error) {
	acked, err := m.redisStore.IsAcked(ctx, mid)
	if err != nil {
		return false, err
	}

	return acked, nil
}

// MarkProcessed marks a message ID as processed.
func (m *serverMessager) MarkProcessed(ctx context.Context, mid string) error {
	if err := m.redisStore.MarkProcessed(ctx, mid); err != nil {
		m.config.Logger.Errorf("mark processed failed: %s, %v", mid, err)
		return fmt.Errorf("failed to mark processed, err: %w", err)
	}

	return nil
}

// MarkAcked marks a message ID as acked.
func (m *serverMessager) MarkAcked(ctx context.Context, mid string) error {
	if err := m.redisStore.MarkedAcked(ctx, mid); err != nil {
		m.config.Logger.Errorf("mark ack failed: %s, %v", mid, err)
		return fmt.Errorf("failed to mark acked, err: %w", err)
	}

	m.config.Logger.Infof("ACK received. message-id(%s)", mid)

	return nil
}
