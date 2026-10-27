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

package ssl

import (
	"crypto"
	"encoding/pem"
	"errors"
	"fmt"
	"net"

	"gitee.com/Trisia/gotlcp/tlcp"
	"github.com/emmansun/gmsm/pkcs8"
	"github.com/emmansun/gmsm/smx509"
)

// GMEnabled reports whether the config carries a TLCP (GM TLS) encryption
// certificate pair, in which case Cert/Key are the signing pair and the
// client dials with pure TLCP instead of standard TLS.
func (c TLSConfig) GMEnabled() bool {
	return len(c.EncCert) != 0 && len(c.EncKey) != 0
}

// NewClientTLCPConf loads CA and the optional double certificates of ssl as
// a TLCP (GB/T 38636) client side config.
func (c TLSConfig) NewClientTLCPConf() (*tlcp.Config, error) {
	conf := &tlcp.Config{
		InsecureSkipVerify: c.InsecureSkipVerify,
	}

	if len(c.CA) != 0 {
		caPool, err := loadTLCPCa(c.CA)
		if err != nil {
			return nil, err
		}

		conf.RootCAs = caPool
	}

	if len(c.Cert) != 0 && len(c.Key) != 0 {
		signCert, err := loadTLCPCertificates(c.Cert, c.Key, c.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to load sign certificate: %w", err)
		}

		conf.Certificates = append(conf.Certificates, *signCert)

		// the TLCP client certificate is also a double pair.
		if c.GMEnabled() {
			encCert, err := loadTLCPCertificates(c.EncCert, c.EncKey, c.Password)
			if err != nil {
				return nil, fmt.Errorf("failed to load encrypt certificate: %w", err)
			}

			conf.Certificates = append(conf.Certificates, *encCert)
		}
	}

	return conf, nil
}

// ResolveTLCPDialConf resolves the per-dial TLCP client config for the target
// address.
//
// gotlcp never derives the server name from the dial address, and a custom
// http DialTLSContext owns the whole handshake: with an empty ServerName the
// client still verifies the certificate chain but silently skips the hostname
// check (verifyServerCertificate passes an empty DNSName). Mirroring what the
// standard net/http transport does for its own TLS dialing, the host part of
// addr fills ServerName so the hostname is actually verified. The config is
// cloned before mutation so concurrent dials sharing the base config stay
// race-free.
//
// IP endpoints are exempt: an IP would only ever match an IP SAN while
// net/http leaves ServerName empty for them, so deriving it here would turn a
// working dial into a hostname mismatch.
func ResolveTLCPDialConf(conf *tlcp.Config, addr string) *tlcp.Config {
	if conf.ServerName != "" || conf.InsecureSkipVerify {
		return conf
	}

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return conf
	}

	if host == "" || net.ParseIP(host) != nil {
		return conf
	}

	cloned := conf.Clone()
	cloned.ServerName = host

	return cloned
}

// loadTLCPCertificates parses a certificate/key PEM pair for TLCP. A
// non-empty passwd decrypts a PKCS#8 (PEM) encrypted private key; with an
// empty passwd the key must be plaintext PEM.
func loadTLCPCertificates(certPEM, keyPEM []byte, passwd string) (*tlcp.Certificate, error) {
	if len(certPEM) == 0 || len(keyPEM) == 0 {
		return nil, errors.New("tlcp certificate and key are both required")
	}

	if passwd == "" {
		cert, err := tlcp.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to parse tlcp key pair: %w", err)
		}

		return &cert, nil
	}

	key, err := loadTLCPPrivateKey(keyPEM, passwd)
	if err != nil {
		return nil, err
	}

	cert, err := smx509.ParseCertificatePEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tlcp certificate: %w", err)
	}

	return &tlcp.Certificate{
		Certificate: [][]byte{cert.Raw},
		PrivateKey:  key,
		Leaf:        cert,
	}, nil
}

// loadTLCPPrivateKey parses a password-encrypted PKCS#8 SM2 private key.
func loadTLCPPrivateKey(keyPEM []byte, passwd string) (crypto.PrivateKey, error) {
	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return nil, errors.New("failed to decode tlcp private key pem")
	}

	key, err := pkcs8.ParsePKCS8PrivateKeySM2(block.Bytes, []byte(passwd))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt tlcp private key: %w", err)
	}

	return key, nil
}

// loadTLCPCa parses CA PEM content into the smx509 certificate pool.
func loadTLCPCa(ca []byte) (*smx509.CertPool, error) {
	caPool := smx509.NewCertPool()
	if ok := caPool.AppendCertsFromPEM(ca); !ok {
		return nil, fmt.Errorf("append ca cert failed")
	}

	return caPool, nil
}
