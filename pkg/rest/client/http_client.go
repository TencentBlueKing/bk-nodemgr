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
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"gitee.com/Trisia/gotlcp/tlcp"
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
func NewHTTPClient(tlsConfig *ssl.TLSConfig) (*http.Client, error) {
	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		TLSHandshakeTimeout: TLSHandshakeTimeout,
		Dial: (&net.Dialer{
			Timeout:   DialTimeout,
			KeepAlive: DialKeepAlive,
		}).Dial,
		MaxIdleConnsPerHost: MaxIdleConnsPerHost,
		// TODO: 同步如果调整为异步，则调整为10min
		ResponseHeaderTimeout: ResponseHeaderTimeout,
	}

	if err := setupTransportSecurity(transport, tlsConfig); err != nil {
		return nil, err
	}

	client := new(http.Client)
	client.Transport = transport

	return client, nil
}

// setupTransportSecurity configures the transport security layer: TLCP (GM
// TLS) when the encryption certificate pair is set, standard TLS otherwise.
//
// With DialTLSContext set, https requests always go through the TLCP dialer
// while plain http requests keep the transport-level Dial, so the standard
// dialer configured above stays untouched.
func setupTransportSecurity(transport *http.Transport, conf *ssl.TLSConfig) error {
	if conf == nil {
		return nil
	}

	if conf.GMEnabled() {
		tlcpConf, err := conf.NewClientTLCPConf()
		if err != nil {
			return err
		}

		transport.DialTLSContext = tlcpDialTLSContext(tlcpConf)

		return nil
	}

	tlsConf := new(tls.Config)
	tlsConf.InsecureSkipVerify = conf.InsecureSkipVerify
	if len(conf.CAFile) != 0 && len(conf.CertFile) != 0 && len(conf.KeyFile) != 0 {
		var err error
		tlsConf, err = conf.NewClientTLSConf()
		if err != nil {
			return err
		}
	}

	transport.TLSClientConfig = tlsConf

	return nil
}

// tlcpDialTLSContext builds the TLCP dial func for http.Transport.
//
// A custom DialTLSContext owns the whole TLS setup, so ServerName cannot stay
// empty when verification is enabled: ResolveTLCPDialConf derives it from the
// dial address per connection, otherwise gotlcp silently skips the hostname
// check.
//
// Known limitation: once DialTLSContext is set, net/http takes the
// hasCustomTLSDialer branch for https and skips both Transport.Dial and the
// CONNECT proxy setup (net/http/transport.go), so TLCP dials always go direct
// and ignore Transport.Proxy. Supporting TLCP through an HTTP proxy would
// require issuing the CONNECT request inside this dialer before the TLCP
// handshake; it is deliberately not implemented because no current deployment
// proxies the GSE server-api endpoint.
//
// Note the NetDialer.Timeout bounds both the TCP connect and the TLCP
// handshake: gotlcp's dial applies it to the context wrapping
// HandshakeContext (gotlcp v1.5.0 tlcp.go dial). This matters because
// http.Transport.TLSHandshakeTimeout is not applied to custom dialers.
func tlcpDialTLSContext(conf *tlcp.Config) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return (&tlcp.Dialer{
			NetDialer: &net.Dialer{
				Timeout:   DialTimeout,
				KeepAlive: DialKeepAlive,
			},
			Config: ssl.ResolveTLCPDialConf(conf, addr),
		}).DialContext(ctx, network, addr)
	}
}

// HTTPClient http client interface.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
