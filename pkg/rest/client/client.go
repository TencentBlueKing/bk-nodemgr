/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package client

import (
	"strings"
	"time"

	restmetrics "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/metrics"
)

const (
	// maxRetryCycleDefault max retry cycle.
	maxRetryCycleDefault = 3

	// toleranceLatencyTimeDefault tolerance latency time.
	toleranceLatencyTimeDefault = 2 * time.Second
)

// IClient http client interface.
type IClient interface {
	Post() *Request
	Put() *Request
	Get() *Request
	Delete() *Request
	Patch() *Request
	Head() *Request
}

// Opt define options.
type Opt func(client *Client)

// WithHeaderMasker set header masker.
func WithHeaderMasker(header ...string) Opt {
	return func(client *Client) {
		for _, h := range header {
			client.headerMasker[h] = defaultHeaderMasker
		}
	}
}

// nolint: mnd
func defaultHeaderMasker(value string) string {
	var maskedValues string
	if len(value) > 6 {
		maskedValues = value[:3] + "***" + value[len(value)-3:]
	} else {
		maskedValues = strings.Repeat("*", len(value))
	}

	return maskedValues
}

// WithCustomHeaderMasker set custom header masker.
func WithCustomHeaderMasker(header string, headerMasker func(string) string) Opt {
	return func(client *Client) {
		client.headerMasker[header] = headerMasker
	}
}

// WithExclusionURL set exclusion url.
func WithExclusionURL(url ...string) Opt {
	return func(client *Client) {
		for _, u := range url {
			client.exclusionURL[u] = struct{}{}
		}
	}
}

// NewClient get rest client.
func NewClient(capability *Capability, baseURL string, opts ...Opt) (IClient, error) {
	if baseURL != "/" {
		baseURL = strings.Trim(baseURL, "/")
		baseURL = "/" + baseURL + "/"
	}

	if capability.ToleranceLatencyTime <= 0 {
		// set default tolerance latency time
		capability.ToleranceLatencyTime = toleranceLatencyTimeDefault
	}

	restClient := &Client{
		baseURL:    baseURL,
		capability: capability,
		metrics: restmetrics.NewMonitor(capability.Name).
			WithDurationMSBuckets(capability.MetricOpts.DurationMSBuckets).
			WithSlowTime(capability.ToleranceLatencyTime).
			Enable(),
		exclusionURL:  make(map[string]struct{}),
		maxRetryCycle: maxRetryCycleDefault,
		headerMasker:  make(map[string]func(string) string),
	}

	for _, opt := range opts {
		opt(restClient)
	}

	return restClient, nil
}

// Client http client.
type Client struct {
	// base url.
	baseURL string

	// client capability.
	capability *Capability

	// client metrics monitor.
	metrics *restmetrics.Monitor

	// exclusionURL define the url that does not need to be monitored.
	exclusionURL map[string]struct{}

	// maxRetryCycle define the max retry cycle.
	maxRetryCycle int

	// headerMasker will be used to mask header value.
	headerMasker map[string]func(string) string
}

// verb get request.
func (client *Client) verb(verb VerbType) *Request {
	return &Request{
		client:       client,
		verb:         verb,
		baseURL:      client.baseURL,
		capability:   client.capability,
		headerMasker: client.headerMasker,
	}
}

// Post method.
func (client *Client) Post() *Request {
	return client.verb(VerbTypePOST)
}

// Put method.
func (client *Client) Put() *Request {
	return client.verb(VerbTypePUT)
}

// Get method.
func (client *Client) Get() *Request {
	return client.verb(VerbTypeGET)
}

// Delete delete method.
func (client *Client) Delete() *Request {
	return client.verb(VerbTypeDELETE)
}

// Patch method.
func (client *Client) Patch() *Request {
	return client.verb(VerbTypePATCH)
}

// Head method.
func (client *Client) Head() *Request {
	return client.verb(VerbTypeHEAD)
}
