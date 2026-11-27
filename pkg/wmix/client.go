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

package wmix

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tracing"
)

type httpClient struct {
	url   url.URL
	token string
	http  *http.Client
	rest  restclient.IClient
}

// REST metrics are registered once, independently of embedded process restarts.
// nolint: gochecknoglobals
var sharedClient struct {
	sync.Mutex
	http *http.Client
	rest restclient.IClient
}

// SetTraceService initializes the shared REST client before WMI operations start.
func SetTraceService(traceSvc tracing.IService) error {
	if traceSvc == nil {
		return fmt.Errorf("WMI trace service is required")
	}
	sharedClient.Lock()
	defer sharedClient.Unlock()
	if sharedClient.rest != nil {
		return fmt.Errorf("WMI REST client already initialized")
	}

	return initRESTClient(traceSvc)
}

func initRESTClient(traceSvc tracing.IService) error {
	const idleTimeout = 30 * time.Second
	// Remote WMI calls wait on network I/O; do not serialize independent hosts
	// behind the renderer's small connection limit.
	const maxIdleConnections = 300
	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        maxIdleConnections,
			MaxIdleConnsPerHost: maxIdleConnections,
			IdleConnTimeout:     idleTimeout,
		},
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
	}
	rest, err := restclient.NewClient(&restclient.Capability{
		Name:                 "wmix",
		HTTPClient:           httpClient,
		TraceSvc:             traceSvc,
		ToleranceLatencyTime: restclient.ToleranceLatencyTimeDefault,
	}, "/", restclient.WithCustomHeaderMasker("Authorization", func(string) string { return "hidden" }))
	if err != nil {
		return fmt.Errorf("failed to create WMI REST client: %w", err)
	}
	sharedClient.http = httpClient
	sharedClient.rest = rest

	return nil
}

func newHTTPClient(port int, token string) (*httpClient, error) {
	sharedClient.Lock()
	defer sharedClient.Unlock()
	if sharedClient.rest == nil {
		traceSvc, err := tracing.G().NewService(tracing.ServiceConfig{
			ServiceName: "wmix", ServiceCategory: tracing.ServiceCategoryHTTP,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create WMI trace service: %w", err)
		}
		if err := initRESTClient(traceSvc); err != nil {
			return nil, err
		}
	}

	return &httpClient{
		url:   url.URL{Scheme: "http", Host: fmt.Sprintf("127.0.0.1:%d", port)},
		token: token,
		http:  sharedClient.http,
		rest:  sharedClient.rest,
	}, nil
}

// The REST request builder carries the inherited context via WithContext.
// nolint: contextcheck
func (cli *httpClient) request(ctx context.Context, method, path string, body []byte, result any) error {
	endpoint := cli.url
	endpoint.Path = path
	var req *restclient.Request
	switch method {
	case http.MethodGet:
		req = cli.rest.Get()
	case http.MethodPost:
		// POST is not retried: a lost response does not imply a command was not executed.
		req = cli.rest.Post().BodyReader(bytes.NewReader(body))
	default:
		return fmt.Errorf("unsupported WMI HTTP method %s", method)
	}
	resp := req.WithContext(contextx.FromContext(ctx)).WithURL(&endpoint).
		WithHeaders(http.Header{"Authorization": {cli.token}}).Do()
	if resp.Err != nil {
		return fmt.Errorf("WMI server request failed: %w", resp.Err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("WMI server returned HTTP %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode WMI response: %w", err)
	}

	return nil
}

func (cli *httpClient) execute(ctx context.Context, args, envs []string) (string, string, error) {
	var password string
	for _, env := range envs {
		if value, ok := strings.CutPrefix(env, WMIEnvPassword+"="); ok {
			password = value
		}
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		return "", "", fmt.Errorf("WMI operation requires a deadline")
	}
	if err := ctx.Err(); err != nil {
		return "", "", err
	}
	body, err := json.Marshal(struct {
		Args     []string `json:"args"`
		Password string   `json:"password"`
		Timeout  float64  `json:"timeout"`
	}{Args: args, Password: password, Timeout: time.Until(deadline).Seconds()})
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal WMI request: %w", err)
	}
	var result struct {
		Stdout string `json:"stdout"`
		Stderr string `json:"stderr"`
		Error  string `json:"error"`
	}
	if err := cli.request(ctx, http.MethodPost, "/execute", body, &result); err != nil {
		return "", "", err
	}
	if result.Error != "" {
		return "", "", fmt.Errorf("commands cannot be executed using wmiexec, stdOut(%s), stdErr(%s): %s",
			result.Stdout, result.Stderr, result.Error)
	}

	return result.Stdout, result.Stderr, nil
}
