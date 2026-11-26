/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package jinja2x

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// client communicates only with the embedded renderer.
type client struct {
	url   string
	token string
	http  *http.Client
}

func newClient(port int, token string) *client {
	const maxConnections = 16
	const idleTimeout = 30 * time.Second

	return &client{
		url:   fmt.Sprintf("http://127.0.0.1:%d", port),
		token: token,
		http: &http.Client{
			Transport: &http.Transport{
				MaxConnsPerHost:     maxConnections,
				MaxIdleConnsPerHost: maxConnections,
				IdleConnTimeout:     idleTimeout,
			},
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
		},
	}
}

func (cli *client) request(ctx context.Context, method, path string, body []byte, result any) error {
	req, err := http.NewRequestWithContext(ctx, method, cli.url+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", cli.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := cli.http.Do(req)
	if err != nil {
		return fmt.Errorf("jinja2 server request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("jinja2 server returned HTTP %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode jinja2 response: %w", err)
	}

	return nil
}

func (cli *client) render(ctx context.Context, tmpl string, content map[string]any) (string, error) {
	body, err := json.Marshal(struct {
		Template string         `json:"template"`
		Context  map[string]any `json:"context"`
	}{Template: tmpl, Context: content})
	if err != nil {
		return "", fmt.Errorf("failed to marshal render request: %w", err)
	}
	var result struct {
		Result string `json:"result"`
	}
	if err := cli.request(ctx, http.MethodPost, "/render", body, &result); err != nil {
		return "", err
	}

	return result.Result, nil
}
