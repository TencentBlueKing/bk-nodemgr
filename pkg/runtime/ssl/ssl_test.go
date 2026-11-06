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
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"
)

// stdTestCerts is a self-signed CA plus a server keypair signed by it, kept
// as PEM content mirroring the in-memory TLSConfig contract.
type stdTestCerts struct {
	ca   []byte
	cert []byte
	key  []byte
}

// generateStdTestCerts builds the RSA CA and the server certificate signed by
// it. RSA-2048 keeps the test fast; the algorithms matter, not the key size.
func generateStdTestCerts(t *testing.T) *stdTestCerts {
	t.Helper()

	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	caTpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "nodemgr-test-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	caDER, err := x509.CreateCertificate(rand.Reader, caTpl, caTpl, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}

	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		t.Fatal(err)
	}

	srvKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	srvTpl := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "nodemgr-test-server"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		DNSNames:     []string{"nodemgr-test-server"},
	}

	srvDER, err := x509.CreateCertificate(rand.Reader, srvTpl, caCert, &srvKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}

	return &stdTestCerts{
		ca:   pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER}),
		cert: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: srvDER}),
		key: pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(srvKey),
		}),
	}
}

// TestNewClientTLSConf checks the CA/cert/key content is loaded into a client
// side tls.Config, and that broken content is reported.
func TestNewClientTLSConf(t *testing.T) {
	certs := generateStdTestCerts(t)

	conf := TLSConfig{
		InsecureSkipVerify: true,
		CA:                 certs.ca,
		Cert:               certs.cert,
		Key:                certs.key,
	}

	tlsConf, err := conf.NewClientTLSConf()
	if err != nil {
		t.Fatalf("NewClientTLSConf() error = %v", err)
	}

	if tlsConf.RootCAs == nil {
		t.Error("RootCAs should be loaded from the CA content")
	}

	if len(tlsConf.Certificates) != 1 {
		t.Errorf("Certificates len = %d, want 1", len(tlsConf.Certificates))
	}

	if !tlsConf.InsecureSkipVerify {
		t.Error("InsecureSkipVerify should be propagated to the tls.Config")
	}

	// CA content without any PEM certificate must be rejected.
	if _, err := (TLSConfig{
		CA:   []byte("not a pem file"),
		Cert: certs.cert,
		Key:  certs.key,
	}).NewClientTLSConf(); err == nil {
		t.Error("CA content without a PEM certificate should fail")
	}

	// broken key content must surface the parse error.
	if _, err := (TLSConfig{
		CA:   certs.ca,
		Cert: certs.cert,
		Key:  []byte("not a pem file"),
	}).NewClientTLSConf(); err == nil {
		t.Error("broken key content should fail")
	}
}

// TestNewServerTLSConf checks the server side config: minimum TLS version and
// the optional mutual TLS switch driven by VerifyClient.
func TestNewServerTLSConf(t *testing.T) {
	certs := generateStdTestCerts(t)

	base := TLSConfig{
		CA:   certs.ca,
		Cert: certs.cert,
		Key:  certs.key,
	}

	tlsConf, err := base.NewServerTLSConf()
	if err != nil {
		t.Fatalf("NewServerTLSConf() error = %v", err)
	}

	if tlsConf.MinVersion != tls.VersionTLS12 {
		t.Errorf("MinVersion = %d, want %d", tlsConf.MinVersion, tls.VersionTLS12)
	}

	if tlsConf.ClientAuth != tls.NoClientCert {
		t.Errorf("ClientAuth = %v, want NoClientCert when VerifyClient is false", tlsConf.ClientAuth)
	}

	conf := base
	conf.VerifyClient = true

	tlsConf, err = conf.NewServerTLSConf()
	if err != nil {
		t.Fatalf("NewServerTLSConf(VerifyClient) error = %v", err)
	}

	if tlsConf.ClientAuth != tls.RequireAndVerifyClientCert {
		t.Errorf("ClientAuth = %v, want RequireAndVerifyClientCert", tlsConf.ClientAuth)
	}

	if tlsConf.ClientCAs == nil {
		t.Error("ClientCAs should be loaded from the CA content")
	}

	// broken certificate content must surface the parse error.
	if _, err := (TLSConfig{
		CA:   certs.ca,
		Cert: []byte("not a pem file"),
		Key:  certs.key,
	}).NewServerTLSConf(); err == nil {
		t.Error("broken cert content should fail")
	}
}

// TestLoadCertificatesPassword checks the legacy RFC 1423 encrypted key
// path: the right password decrypts, a wrong one is rejected, and a non-PEM
// key is reported instead of being silently accepted.
func TestLoadCertificatesPassword(t *testing.T) {
	certs := generateStdTestCerts(t)

	block, _ := pem.Decode(certs.key)
	if block == nil {
		t.Fatal("failed to decode the generated key pem")
	}

	// nolint: staticcheck // SA1019: mirrors the legacy RFC 1423 path under test.
	encBlock, err := x509.EncryptPEMBlock(rand.Reader, block.Type, block.Bytes, []byte("passwd"), x509.PEMCipherAES256)
	if err != nil {
		t.Fatal(err)
	}

	encKey := pem.EncodeToMemory(encBlock)

	if _, err := loadCertificates(certs.cert, encKey, "passwd"); err != nil {
		t.Errorf("loadCertificates() with the right password error = %v", err)
	}

	if _, err := loadCertificates(certs.cert, encKey, "wrong-passwd"); err == nil {
		t.Error("loadCertificates() with a wrong password should fail")
	}

	// a key that is not PEM at all must be rejected.
	if _, err := loadCertificates(certs.cert, []byte("not a pem file"), "passwd"); err == nil {
		t.Error("loadCertificates() with a non-PEM key should fail")
	}
}
