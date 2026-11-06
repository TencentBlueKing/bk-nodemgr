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
	"context"
	"crypto/rand"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"testing"
	"time"

	"gitee.com/Trisia/gotlcp/tlcp"
	"github.com/emmansun/gmsm/pkcs8"
	"github.com/emmansun/gmsm/sm2"
	"github.com/emmansun/gmsm/smx509"
)

// tlcpTestCerts is a set of self-generated TLCP double certificates, kept as
// PEM content mirroring the in-memory TLSConfig contract.
type tlcpTestCerts struct {
	ca             []byte
	serverSignCert []byte
	serverSignKey  []byte
	serverEncCert  []byte
	serverEncKey   []byte
	clientSignCert []byte
	clientSignKey  []byte
	clientEncCert  []byte
	clientEncKey   []byte
}

// generateTLCPTestCerts builds a CA plus server/client double certificate
// pairs with SM2 keys, returned as PEM content.
func generateTLCPTestCerts(t *testing.T) *tlcpTestCerts {
	t.Helper()

	certs := &tlcpTestCerts{}

	caKey, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	caTemplate := &smx509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "tlcp-test-ca", Organization: []string{"bk-nodemgr-test"}},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              smx509.KeyUsageCertSign | smx509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	caDER, err := smx509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}

	caCert, err := smx509.ParseCertificate(caDER)
	if err != nil {
		t.Fatal(err)
	}

	encodePEM := func(blockType string, der []byte) []byte {
		return pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: der})
	}

	encodeKey := func(key *sm2.PrivateKey) []byte {
		der, err := smx509.MarshalPKCS8PrivateKey(key)
		if err != nil {
			t.Fatal(err)
		}

		return encodePEM("PRIVATE KEY", der)
	}

	certs.ca = encodePEM("CERTIFICATE", caDER)

	newLeaf := func(cn string, usage smx509.KeyUsage) (*smx509.Certificate, *sm2.PrivateKey) {
		key, err := sm2.GenerateKey(rand.Reader)
		if err != nil {
			t.Fatal(err)
		}

		template := &smx509.Certificate{
			SerialNumber: big.NewInt(time.Now().UnixNano()),
			Subject:      pkix.Name{CommonName: cn, Organization: []string{"bk-nodemgr-test"}},
			NotBefore:    time.Now().Add(-time.Hour),
			NotAfter:     time.Now().Add(24 * time.Hour),
			KeyUsage:     usage,
			ExtKeyUsage:  []smx509.ExtKeyUsage{smx509.ExtKeyUsageServerAuth, smx509.ExtKeyUsageClientAuth},
			IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		}

		der, err := smx509.CreateCertificate(rand.Reader, template, caCert, &key.PublicKey, caKey)
		if err != nil {
			t.Fatal(err)
		}

		cert, err := smx509.ParseCertificate(der)
		if err != nil {
			t.Fatal(err)
		}

		return cert, key
	}

	serverSignCert, serverSignKey := newLeaf("tlcp-test-server-sign", smx509.KeyUsageDigitalSignature)
	certs.serverSignCert = encodePEM("CERTIFICATE", serverSignCert.Raw)
	certs.serverSignKey = encodeKey(serverSignKey)

	serverEncCert, serverEncKey := newLeaf("tlcp-test-server-enc", smx509.KeyUsageKeyEncipherment|smx509.KeyUsageDigitalSignature)
	certs.serverEncCert = encodePEM("CERTIFICATE", serverEncCert.Raw)
	certs.serverEncKey = encodeKey(serverEncKey)

	clientSignCert, clientSignKey := newLeaf("tlcp-test-client-sign", smx509.KeyUsageDigitalSignature)
	certs.clientSignCert = encodePEM("CERTIFICATE", clientSignCert.Raw)
	certs.clientSignKey = encodeKey(clientSignKey)

	clientEncCert, clientEncKey := newLeaf("tlcp-test-client-enc", smx509.KeyUsageKeyEncipherment|smx509.KeyUsageDigitalSignature)
	certs.clientEncCert = encodePEM("CERTIFICATE", clientEncCert.Raw)
	certs.clientEncKey = encodeKey(clientEncKey)

	return certs
}

// TestTLSConfigGMEnabled checks the TLCP mode detection.
func TestTLSConfigGMEnabled(t *testing.T) {
	conf := TLSConfig{Cert: []byte("a"), Key: []byte("b")}
	if conf.GMEnabled() {
		t.Error("standard tls config should not be GM enabled")
	}

	conf.EncCert = []byte("c")
	if conf.GMEnabled() {
		t.Error("enc cert without enc key should not be GM enabled")
	}

	conf.EncKey = []byte("d")
	if !conf.GMEnabled() {
		t.Error("full tlcp config should be GM enabled")
	}
}

// TestResolveTLCPDialConf checks the per-dial ServerName derivation: without
// it gotlcp verifies the chain but silently skips the hostname check.
func TestResolveTLCPDialConf(t *testing.T) {
	base := &tlcp.Config{}

	// insecure mode: keep the config untouched.
	resolved := ResolveTLCPDialConf(&tlcp.Config{InsecureSkipVerify: true}, "gse.example.com:8080")
	if resolved.ServerName != "" {
		t.Errorf("insecure config should not derive ServerName, got %q", resolved.ServerName)
	}

	// verification mode: derive from the dial address.
	resolved = ResolveTLCPDialConf(base, "gse.example.com:8080")
	if resolved.ServerName != "gse.example.com" {
		t.Errorf("ServerName = %q, want %q", resolved.ServerName, "gse.example.com")
	}

	// the base config must stay unmutated (cloned per dial).
	if base.ServerName != "" {
		t.Errorf("base config mutated, ServerName = %q", base.ServerName)
	}

	// a pinned ServerName wins over the address.
	resolved = ResolveTLCPDialConf(&tlcp.Config{ServerName: "pinned.example.com"}, "other.example.com:8080")
	if resolved.ServerName != "pinned.example.com" {
		t.Errorf("ServerName = %q, want %q", resolved.ServerName, "pinned.example.com")
	}

	// an address without port is kept as-is without derivation.
	resolved = ResolveTLCPDialConf(base, "not-a-host-port")
	if resolved != base {
		t.Error("unparsable addr should return the base config unchanged")
	}

	// IP endpoints: an IP only matches an IP SAN, and net/http leaves
	// ServerName empty for them, so deriving it would break the handshake.
	for _, addr := range []string{"127.0.0.1:443", "[::1]:443", ":443"} {
		resolved = ResolveTLCPDialConf(base, addr)
		if resolved.ServerName != "" {
			t.Errorf("addr %q should not derive ServerName, got %q", addr, resolved.ServerName)
		}
	}
}

// TestTLCPEncryptedPrivateKey checks the PKCS#8 password-encrypted keys are
// loadable on the client TLCP config.
func TestTLCPEncryptedPrivateKey(t *testing.T) {
	certs := generateTLCPTestCerts(t)

	const passwd = "test-passwd"

	// re-encrypt the server sign/enc keys as password-encrypted PKCS#8.
	encryptKey := func(keyPEM []byte) []byte {
		block, _ := pem.Decode(keyPEM)
		if block == nil {
			t.Fatal("failed to decode key pem")
		}

		key, err := pkcs8.ParsePKCS8PrivateKeySM2(block.Bytes)
		if err != nil {
			t.Fatal(err)
		}

		encDER, err := pkcs8.MarshalPrivateKey(key, []byte(passwd), nil)
		if err != nil {
			t.Fatal(err)
		}

		return pem.EncodeToMemory(&pem.Block{Type: "ENCRYPTED PRIVATE KEY", Bytes: encDER})
	}

	conf := TLSConfig{
		Cert:         certs.serverSignCert,
		Key:          encryptKey(certs.serverSignKey),
		EncCert:      certs.serverEncCert,
		EncKey:       encryptKey(certs.serverEncKey),
		CA:           certs.ca,
		VerifyClient: true,
		Password:     passwd,
	}

	if _, err := conf.NewClientTLCPConf(); err != nil {
		t.Fatalf("client tlcp config with encrypted keys failed: %v", err)
	}

	clientConf := TLSConfig{
		CA:      certs.ca,
		Cert:    certs.clientSignCert,
		Key:     certs.clientSignKey,
		EncCert: certs.clientEncCert,
		EncKey:  certs.clientEncKey,
	}

	if _, err := clientConf.NewClientTLCPConf(); err != nil {
		t.Fatalf("client tlcp config with plaintext keys failed: %v", err)
	}

	// a wrong password must fail explicitly.
	conf.Password = "wrong-passwd"
	if _, err := conf.NewClientTLCPConf(); err == nil {
		t.Error("client tlcp config with wrong password expected error, got nil")
	}
}

// TestTLCPConfigErrors checks invalid TLCP configurations fail explicitly.
func TestTLCPConfigErrors(t *testing.T) {
	certs := generateTLCPTestCerts(t)

	// the encrypt pair must be parseable PEM.
	conf := TLSConfig{
		Cert:    certs.serverSignCert,
		Key:     certs.serverSignKey,
		EncCert: []byte("not a pem cert"),
		EncKey:  []byte("not a pem key"),
	}
	if _, err := conf.NewClientTLCPConf(); err == nil {
		t.Error("client tlcp config with broken pem content expected error, got nil")
	}

	// the encrypt pair is required once the TLCP mode is entered.
	conf = TLSConfig{
		Cert: certs.serverSignCert,
		Key:  certs.serverSignKey,
	}
	tlcpConf, err := conf.NewClientTLCPConf()
	if err != nil {
		t.Fatalf("sign-only client tlcp config failed: %v", err)
	}
	if len(tlcpConf.Certificates) != 1 {
		t.Errorf("sign-only config Certificates len = %d, want 1", len(tlcpConf.Certificates))
	}
}

// TestTLCPHTTPRoundTrip starts a TLCP HTTP server and accesses it with a
// TLCP client, verifying the double-certificate handshake and CA validation.
// The server side config is built inline from the shared cert loaders: the
// production code only ships the TLCP client, the server exists solely to
// exercise the handshake.
func TestTLCPHTTPRoundTrip(t *testing.T) {
	certs := generateTLCPTestCerts(t)

	signCert, err := loadTLCPCertificates(certs.serverSignCert, certs.serverSignKey, "")
	if err != nil {
		t.Fatal(err)
	}

	encCert, err := loadTLCPCertificates(certs.serverEncCert, certs.serverEncKey, "")
	if err != nil {
		t.Fatal(err)
	}

	caPool, err := loadTLCPCa(certs.ca)
	if err != nil {
		t.Fatal(err)
	}

	tlcpConf := &tlcp.Config{
		Certificates: []tlcp.Certificate{*signCert, *encCert},
		ClientCAs:    caPool,
		ClientAuth:   tlcp.RequireAndVerifyClientCert,
	}

	listener, err := tlcp.Listen("tcp", "127.0.0.1:0", tlcpConf)
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("pong"))
	})

	server := &http.Server{Handler: mux}
	go func() {
		_ = server.Serve(listener)
	}()
	defer func() {
		_ = server.Close()
	}()

	clientConf := TLSConfig{
		CA:      certs.ca,
		Cert:    certs.clientSignCert,
		Key:     certs.clientSignKey,
		EncCert: certs.clientEncCert,
		EncKey:  certs.clientEncKey,
	}

	clientTLCPConf, err := clientConf.NewClientTLCPConf()
	if err != nil {
		t.Fatal(err)
	}

	// route the dial through ResolveTLCPDialConf like the rest client does,
	// so the certificate hostname (IP SAN 127.0.0.1) is genuinely verified
	// instead of being silently skipped.
	client := &http.Client{
		Transport: &http.Transport{
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return (&tlcp.Dialer{
					NetDialer: &net.Dialer{Timeout: 5 * time.Second},
					Config:    ResolveTLCPDialConf(clientTLCPConf, addr),
				}).DialContext(ctx, network, addr)
			},
		},
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(fmt.Sprintf("https://%s/ping", listener.Addr().String()))
	if err != nil {
		t.Fatalf("tlcp http request failed: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status code = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	buf := make([]byte, 4)
	if _, err := resp.Body.Read(buf); err != nil && len(buf) != 4 {
		t.Fatalf("read body failed: %v", err)
	}

	if string(buf) != "pong" {
		t.Errorf("body = %q, want %q", string(buf), "pong")
	}
}
