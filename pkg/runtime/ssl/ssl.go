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

// Package ssl ...
package ssl

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// TLSConfig is inner tls config.
type TLSConfig struct {
	// InsecureSkipVerify controls whether a client verifies the
	// server's certificate chain and host name.
	// If InsecureSkipVerify is true, TLS accepts any certificate
	// presented by the server and any host name in that certificate.
	// In this mode, TLS is susceptible to man-in-the-middle attacks.
	// This should be used only for testing.
	InsecureSkipVerify bool

	// VerifyClient controls whether a server requests a certificate from a client.
	// If TLSConfig is used in client side, this field is ignored.
	VerifyClient bool

	// CA authentication root certificate content.
	CA []byte

	// Cert authentication certificate content.
	Cert []byte

	// Key authentication key content.
	Key []byte

	// EncCert TLCP (GM TLS, GB/T 38636) encryption certificate content,
	// paired with EncKey. When set together with Cert/Key (the signing
	// pair), the client dials with pure TLCP instead of standard TLS.
	EncCert []byte

	// EncKey TLCP (GM TLS) encryption key content.
	EncKey []byte

	// Password authentication key file password.
	Password string
}

// IsEnabled checks if ssl is enabled.
func (c TLSConfig) IsEnabled() bool {
	return len(c.CA) != 0 && len(c.Cert) != 0 && len(c.Key) != 0
}

// NewClientTLSConf load and verify CA/CERT/KEY files of ssl as client side.
func (c TLSConfig) NewClientTLSConf() (*tls.Config, error) {
	caPool, err := loadCa(c.CA)
	if err != nil {
		return nil, err
	}

	cert, err := loadCertificates(c.Cert, c.Key, c.Password)
	if err != nil {
		return nil, err
	}

	tlsConf := &tls.Config{
		InsecureSkipVerify: c.InsecureSkipVerify,
		RootCAs:            caPool,
		Certificates:       []tls.Certificate{*cert},
	}

	return tlsConf, nil
}

// NewServerTLSConf load and verify CA/CERT/KEY files of ssl as server side.
func (c TLSConfig) NewServerTLSConf() (*tls.Config, error) {
	caPool, err := loadCa(c.CA)
	if err != nil {
		return nil, err
	}

	cert, err := loadCertificates(c.Cert, c.Key, c.Password)
	if err != nil {
		return nil, err
	}

	conf := &tls.Config{
		ClientCAs:    caPool,
		Certificates: []tls.Certificate{*cert},
		MinVersion:   tls.VersionTLS12,
	}

	if c.VerifyClient {
		conf.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return conf, nil
}

func loadCa(ca []byte) (*x509.CertPool, error) {
	caPool := x509.NewCertPool()
	if ok := caPool.AppendCertsFromPEM(ca); !ok {
		return nil, fmt.Errorf("append ca cert failed")
	}

	return caPool, nil
}

func loadCertificates(cert, key []byte, passwd string) (*tls.Certificate, error) {
	if len(passwd) != 0 {
		priPem, _ := pem.Decode(key)
		if priPem == nil {
			return nil, fmt.Errorf("decode private key failed")
		}

		// legacy PEM encryption (RFC 1423), kept for backward compatibility
		// with operator-provided encrypted key files.
		// nolint: staticcheck // SA1019: see the comment above.
		priDecrPem, decErr := x509.DecryptPEMBlock(priPem, []byte(passwd))
		if decErr != nil {
			return nil, decErr
		}

		key = pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: priDecrPem,
		})
	}

	tlsCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	return &tlsCert, nil
}
