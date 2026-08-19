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

package client

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
)

const (
	// ResponseHeaderTimeout response header timeout.
	ResponseHeaderTimeout = 30 * time.Minute
	// MaxIdleConnsPerHost max idle conns per host.
	MaxIdleConnsPerHost = 1000
	// DialKeepAlive dial keep alive.
	DialKeepAlive = 30 * time.Second
	// DialTimeout dial timeout.
	DialTimeout = 5 * time.Second
	// TLSHandshakeTimeout tls handshake timeout.
	TLSHandshakeTimeout = 5 * time.Second
)

// NewHTTPClient new http client.
func NewHTTPClient(c *ssl.TLSConfig) (*http.Client, error) {
	tlsConf := new(tls.Config)
	if c != nil {
		tlsConf.InsecureSkipVerify = c.InsecureSkipVerify
		if len(c.CAFile) != 0 && len(c.CertFile) != 0 && len(c.KeyFile) != 0 {
			var err error
			tlsConf, err = c.NewClientTLSConf()
			if err != nil {
				return nil, err
			}
		}
	}

	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		TLSHandshakeTimeout: TLSHandshakeTimeout,
		TLSClientConfig:     tlsConf,
		Dial: (&net.Dialer{
			Timeout:   DialTimeout,
			KeepAlive: DialKeepAlive,
		}).Dial,
		MaxIdleConnsPerHost: MaxIdleConnsPerHost,
		// TODO: 同步如果调整为异步，则调整为10min
		ResponseHeaderTimeout: ResponseHeaderTimeout,
	}

	client := new(http.Client)
	client.Transport = transport

	return client, nil
}

// HTTPClient http client interface.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
