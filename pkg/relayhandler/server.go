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
	"fmt"
	"net/http"

	serverapi "github.com/TencentBlueKing/bk-gse-sdk/go/service/server-api"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
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
}

// Start starts the messager.
func (m *serverMessager) Start(ctx context.Context) error {
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
	m.config.Logger.Infof("started messager")

	return nil
}

// Stop stops the messager.
func (m *serverMessager) Stop(ctx context.Context) error {
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

// RespondCallback sends the callback resp.
func (m *serverMessager) RespondCallback(ctx context.Context, messageID string, httpCode int, content []byte, agentIDs ...string) error {
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

// PushToClient sends the server push to client.
func (m *serverMessager) PushToClient(ctx context.Context, eventType protoRelay.EventType,
	payload []byte, agentIDs ...string) error {

	push := &protoRelay.ServerPush{
		Base: protoRelay.Base{
			MessageID:   identifier.GenMessageID(),
			MessageType: protoRelay.MessageTypeServerPush,
		},
		EventType: eventType,
		Payload:   payload,
	}
	pushData, err := json.Marshal(push)
	if err != nil {
		return err
	}

	result, err := m.client.Cluster().PluginDispatchMessage(
		ctx, push.MessageID, pushData, agentIDs...)
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
