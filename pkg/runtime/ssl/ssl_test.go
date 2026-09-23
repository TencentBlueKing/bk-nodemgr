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
	"os"
	"path/filepath"
	"testing"
	"time"
)

// stdTestCerts is a self-signed CA plus a server keypair signed by it, written
// as PEM files under a temp directory.
type stdTestCerts struct {
	dir      string
	caFile   string
	certFile string
	keyFile  string
}

// generateStdTestCerts builds the RSA CA and the server certificate signed by
// it. RSA-2048 keeps the test fast; the algorithms matter, not the key size.
func generateStdTestCerts(t *testing.T) *stdTestCerts {
	t.Helper()

	dir := t.TempDir()

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

	writePEM := func(name string, block *pem.Block) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, pem.EncodeToMemory(block), 0o600); err != nil {
			t.Fatal(err)
		}

		return path
	}

	return &stdTestCerts{
		dir:    dir,
		caFile: writePEM("ca.crt", &pem.Block{Type: "CERTIFICATE", Bytes: caDER}),
		certFile: writePEM("server.crt", &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: srvDER,
		}),
		keyFile: writePEM("server.key", &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(srvKey),
		}),
	}
}

// TestTLSConfigValidate checks the passthrough validation contract: an empty
// TLS config is valid, and a configured one is accepted as-is for now.
func TestTLSConfigValidate(t *testing.T) {
	if err := (TLSConfig{}).Validate(); err != nil {
		t.Errorf("empty TLSConfig should validate, got %v", err)
	}

	conf := TLSConfig{
		CAFile:      "ca.crt",
		CertFile:    "server.crt",
		KeyFile:     "server.key",
		EncCertFile: "server_enc.crt",
		EncKeyFile:  "server_enc.key",
	}
	if err := conf.Validate(); err != nil {
		t.Errorf("configured TLSConfig should validate, got %v", err)
	}
}

// TestNewClientTLSConf checks the CA/cert/key files are loaded into a client
// side tls.Config, and that a missing or broken file is reported.
func TestNewClientTLSConf(t *testing.T) {
	certs := generateStdTestCerts(t)

	conf := TLSConfig{
		InsecureSkipVerify: true,
		CAFile:             certs.caFile,
		CertFile:           certs.certFile,
		KeyFile:            certs.keyFile,
	}

	tlsConf, err := conf.NewClientTLSConf()
	if err != nil {
		t.Fatalf("NewClientTLSConf() error = %v", err)
	}

	if tlsConf.RootCAs == nil {
		t.Error("RootCAs should be loaded from the CA file")
	}

	if len(tlsConf.Certificates) != 1 {
		t.Errorf("Certificates len = %d, want 1", len(tlsConf.Certificates))
	}

	if !tlsConf.InsecureSkipVerify {
		t.Error("InsecureSkipVerify should be propagated to the tls.Config")
	}

	// a missing CA file must surface the read error.
	if _, err := (TLSConfig{
		CAFile:   filepath.Join(certs.dir, "missing.crt"),
		CertFile: certs.certFile,
		KeyFile:  certs.keyFile,
	}).NewClientTLSConf(); err == nil {
		t.Error("missing CA file should fail")
	}

	// a CA file without any PEM certificate must be rejected.
	badCA := filepath.Join(certs.dir, "bad-ca.crt")
	if err := os.WriteFile(badCA, []byte("not a pem file"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := (TLSConfig{
		CAFile:   badCA,
		CertFile: certs.certFile,
		KeyFile:  certs.keyFile,
	}).NewClientTLSConf(); err == nil {
		t.Error("CA file without a PEM certificate should fail")
	}

	// a missing key file must surface the read error.
	if _, err := (TLSConfig{
		CAFile:   certs.caFile,
		CertFile: certs.certFile,
		KeyFile:  filepath.Join(certs.dir, "missing.key"),
	}).NewClientTLSConf(); err == nil {
		t.Error("missing key file should fail")
	}
}

// TestNewServerTLSConf checks the server side config: minimum TLS version and
// the optional mutual TLS switch driven by VerifyClient.
func TestNewServerTLSConf(t *testing.T) {
	certs := generateStdTestCerts(t)

	base := TLSConfig{
		CAFile:   certs.caFile,
		CertFile: certs.certFile,
		KeyFile:  certs.keyFile,
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
		t.Error("ClientCAs should be loaded from the CA file")
	}

	// a missing certificate file must surface the read error.
	if _, err := (TLSConfig{
		CAFile:   certs.caFile,
		CertFile: filepath.Join(certs.dir, "missing.crt"),
		KeyFile:  certs.keyFile,
	}).NewServerTLSConf(); err == nil {
		t.Error("missing cert file should fail")
	}
}

// TestLoadCertificatesPassword checks the legacy RFC 1423 encrypted key file
// path: the right password decrypts, a wrong one is rejected, and a non-PEM
// key file is reported instead of being silently accepted.
func TestLoadCertificatesPassword(t *testing.T) {
	certs := generateStdTestCerts(t)

	keyPEM, err := os.ReadFile(certs.keyFile)
	if err != nil {
		t.Fatal(err)
	}

	block, _ := pem.Decode(keyPEM)
	if block == nil {
		t.Fatal("failed to decode the generated key pem")
	}

	// nolint: staticcheck // SA1019: mirrors the legacy RFC 1423 path under test.
	encBlock, err := x509.EncryptPEMBlock(rand.Reader, block.Type, block.Bytes, []byte("passwd"), x509.PEMCipherAES256)
	if err != nil {
		t.Fatal(err)
	}

	encKeyFile := filepath.Join(certs.dir, "server.enc.key")
	if err := os.WriteFile(encKeyFile, pem.EncodeToMemory(encBlock), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := loadCertificates(certs.certFile, encKeyFile, "passwd"); err != nil {
		t.Errorf("loadCertificates() with the right password error = %v", err)
	}

	if _, err := loadCertificates(certs.certFile, encKeyFile, "wrong-passwd"); err == nil {
		t.Error("loadCertificates() with a wrong password should fail")
	}

	// a key file that is not PEM at all must be rejected.
	badKey := filepath.Join(certs.dir, "bad.key")
	if err := os.WriteFile(badKey, []byte("not a pem file"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := loadCertificates(certs.certFile, badKey, "passwd"); err == nil {
		t.Error("loadCertificates() with a non-PEM key file should fail")
	}
}
