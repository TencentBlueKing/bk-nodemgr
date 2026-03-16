/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package iegtjj provides handlers to operate on iegtjj api.
package iegtjj

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
)

// cli client for iegtjj.
type cli struct {
	client restclient.IClient
}

// newClient initialize a new iegtjj restclient.
func newClient(c *restclient.Capability) (*cli, error) {
	restCli, err := restclient.NewClient(c, "/")
	if err != nil {
		return nil, err
	}

	return &cli{
		client: restCli,
	}, nil
}

// getCommonHeader get a common header.
func (c *cli) getCommonHeader() (http.Header, error) {
	header := http.Header{}

	return header, nil
}

func (c *cli) getDevicePassword(nCtx contextx.IContext, req *GetDevicePasswordReq) (*GetDevicePasswordResp, error) {
	resp := new(BaseBroker[json.RawMessage])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	result := c.client.Post().
		SubResourcef("/pwd/getDevicePassword").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do()

	if result.StatusCode != 0 && (result.StatusCode < 200 || result.StatusCode >= 300) {
		return nil, fmt.Errorf("iegtjj api returned non-2xx status. status(%d)", result.StatusCode)
	}

	if err := result.Into(resp); err != nil {
		return nil, err
	}

	if resp.HasError {
		return nil, fmt.Errorf("iegtjj api error (requestId: %s): %s",
			resp.RequestID, extractErrorMessage(resp.Message, resp.Data))
	}

	var data GetDevicePasswordResp
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse device password response: %w", err)
	}

	return &data, nil
}

// extractErrorMessage extracts human-readable error messages from TJJ error responses.
// When HasError is true, ResponseItems is typically map[string]string (e.g. {"4": "调用IP未被授权"}).
func extractErrorMessage(message string, rawItems json.RawMessage) string {
	var errItems map[string]string
	if err := json.Unmarshal(rawItems, &errItems); err == nil && len(errItems) > 0 {
		msgs := make([]string, 0, len(errItems))
		for _, msg := range errItems {
			msgs = append(msgs, msg)
		}

		return strings.Join(msgs, "; ")
	}

	if message != "" {
		return message
	}

	return string(rawItems)
}
