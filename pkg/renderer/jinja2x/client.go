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
	"net/url"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tracing"
)

// client holds the endpoint and credentials of a single embedded server.
type client struct {
	url   url.URL
	token string
	http  *http.Client
	rest  restclient.IClient
}

// The REST client and its registered metrics outlive individual server processes.
// nolint: gochecknoglobals
var sharedClient struct {
	sync.Mutex
	http *http.Client
	rest restclient.IClient
}

// SetTraceService initializes the shared REST client before rendering starts.
func SetTraceService(traceSvc tracing.IService) error {
	if traceSvc == nil {
		return fmt.Errorf("jinja2 trace service is required")
	}
	sharedClient.Lock()
	defer sharedClient.Unlock()
	if sharedClient.rest != nil {
		return fmt.Errorf("jinja2 REST client already initialized")
	}

	return initRESTClient(traceSvc)
}

func initRESTClient(traceSvc tracing.IService) error {
	const maxConnections = 16
	const idleTimeout = 30 * time.Second

	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxConnsPerHost:     maxConnections,
			MaxIdleConnsPerHost: maxConnections,
			IdleConnTimeout:     idleTimeout,
		},
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
	}
	rest, err := restclient.NewClient(&restclient.Capability{
		Name:                 "jinja2x",
		HTTPClient:           httpClient,
		TraceSvc:             traceSvc,
		ToleranceLatencyTime: restclient.ToleranceLatencyTimeDefault,
	}, "/", restclient.WithCustomHeaderMasker("Authorization", func(string) string { return "hidden" }))
	if err != nil {
		return fmt.Errorf("failed to create jinja2 REST client: %w", err)
	}
	sharedClient.http = httpClient
	sharedClient.rest = rest

	return nil
}

func newClient(port int, token string) (*client, error) {
	sharedClient.Lock()
	defer sharedClient.Unlock()
	if sharedClient.rest == nil {
		// Standalone renderers use an unsampled service unless configured by backend startup.
		traceSvc, err := tracing.G().NewService(tracing.ServiceConfig{
			ServiceName: "jinja2x", ServiceCategory: tracing.ServiceCategoryHTTP,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create jinja2 trace service: %w", err)
		}
		if err := initRESTClient(traceSvc); err != nil {
			return nil, err
		}
	}

	return &client{
		url:   url.URL{Scheme: "http", Host: fmt.Sprintf("127.0.0.1:%d", port)},
		token: token,
		http:  sharedClient.http,
		rest:  sharedClient.rest,
	}, nil
}

// The REST request builder carries the inherited context via WithContext.
// nolint: contextcheck
func (cli *client) request(ctx context.Context, method, path string, body []byte, result any) error {
	endpoint := cli.url
	endpoint.Path = path
	var req *restclient.Request
	switch method {
	case http.MethodGet:
		req = cli.rest.Get()
	case http.MethodPost:
		req = cli.rest.Post().BodyReader(bytes.NewReader(body))
	default:
		return fmt.Errorf("unsupported jinja2 HTTP method %s", method)
	}
	resp := req.WithContext(contextx.FromContext(ctx)).WithURL(&endpoint).
		WithHeaders(http.Header{"Authorization": {cli.token}}).Do()
	if resp.Err != nil {
		return fmt.Errorf("jinja2 server request failed: %w", resp.Err)
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
