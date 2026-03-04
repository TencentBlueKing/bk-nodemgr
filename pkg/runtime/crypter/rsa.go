/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package crypter provides encryption utilities based on RSA-OAEP
package crypter

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"hash"
)

const (
	// RSADefaultLabel is the default label used for RSA-OAEP.
	RSADefaultLabel = "com.example.crypto.rsa.v1"

	// RSAVersion is the version prefix for RSA-OAEP ciphertext.
	RSAVersion = 1

	// RSAVersionHybrid is the version prefix for hybrid ciphertext.
	// The output layout is: RSAVersionHybrid(1) + WrappedKey(RSA ciphertext, keySize bytes) + Nonce(12) + AES-GCM ciphertext.
	RSAVersionHybrid = 2

	// rsaHybridKeySize is the symmetric key size used for hybrid encryption (AES-256).
	rsaHybridKeySize = 32

	// rsaGCMNonceSize is the nonce size used for AES-GCM.
	rsaGCMNonceSize = 12

	// PEMBlockTypeRSAPrivateKeyType is the PEM block type for RSA private keys.
	PEMBlockTypeRSAPrivateKeyType string = "RSA PRIVATE KEY"

	// PEMBlockTypeRSAPublicKeyType is the PEM block type for RSA public keys.
	PEMBlockTypeRSAPublicKeyType string = "RSA PUBLIC KEY"

	// PEMBlockTypePrivateKeyType is the PEM block type for generic private keys.
	PEMBlockTypePrivateKeyType string = "PRIVATE KEY"

	// PEMBlockTypePublicKeyType is the PEM block type for generic public keys.
	PEMBlockTypePublicKeyType string = "PUBLIC KEY"
)

// RSAKeySize defines the RSA key size.
type RSAKeySize int

const (
	// RSAKeySize2048 represents a 2048-bit RSA key size.
	RSAKeySize2048 RSAKeySize = 2048
	// RSAKeySize3072 represents a 3072-bit RSA key size.
	RSAKeySize3072 RSAKeySize = 3072
	// RSAKeySize4096 represents a 4096-bit RSA key size.
	RSAKeySize4096 RSAKeySize = 4096
)

// Validate checks if the RSAKeySize is valid.
func (r RSAKeySize) Validate() error {
	switch r {
	case RSAKeySize2048, RSAKeySize3072, RSAKeySize4096:
		return nil
	default:
		return fmt.Errorf("invalid RSA key size: %d", r)
	}
}

// ToInt converts RSAKeySize to int.
func (r RSAKeySize) ToInt() int {
	return int(r)
}

// RSA defines the crypter instance based on RSA-OAEP.
type RSA struct {
	pub   *rsa.PublicKey
	priv  *rsa.PrivateKey
	label []byte
}

// rsaOAEPHash defines the hash algorithm used by RSA-OAEP in this package.
// Keep it centralized so changing OAEP hash requires a single edit.
// nolint: gochecknoglobals
var rsaOAEPHash = struct {
	New  func() hash.Hash
	Size int
}{
	New:  sha256.New,
	Size: sha256.Size,
}

func (r *RSA) oaepMaxPlaintextSize() int {
	if r.pub == nil {
		return 0
	}

	// See crypto/rsa for OAEP constraints: https://pkg.go.dev/crypto/rsa#EncryptOAEP.
	k := r.pub.Size()
	maxLength := k - 2*rsaOAEPHash.Size - 2 // nolint: mnd
	if maxLength < 0 {
		return 0
	}

	return maxLength
}

func (r *RSA) encryptHybrid(plaintext []byte) ([]byte, error) {
	// Generate random symmetric key and nonce.
	aesKey := make([]byte, rsaHybridKeySize)
	if _, err := rand.Read(aesKey); err != nil {
		return nil, fmt.Errorf("failed to generate aes key: %w", err)
	}

	nonce := make([]byte, rsaGCMNonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate gcm nonce: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create aes cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create gcm: %w", err)
	}

	gcmCiphertext := aead.Seal(nil, nonce, plaintext, r.label)

	// Wrap the symmetric key with RSA-OAEP.
	wrappedKey, err := rsa.EncryptOAEP(rsaOAEPHash.New(), rand.Reader, r.pub, aesKey, r.label)
	if err != nil {
		return nil, fmt.Errorf("failed to wrap aes key with rsa-oaep: %w", err)
	}

	result := make([]byte, 1+len(wrappedKey)+len(nonce)+len(gcmCiphertext))
	result[0] = RSAVersionHybrid
	copy(result[1:], wrappedKey)
	copy(result[1+len(wrappedKey):], nonce)
	copy(result[1+len(wrappedKey)+len(nonce):], gcmCiphertext)

	return result, nil
}

func (r *RSA) encryptOAEP(plaintext []byte) ([]byte, error) {
	if r.pub == nil {
		return nil, errors.New("rsa public key is not set")
	}

	ciphertext, err := rsa.EncryptOAEP(rsaOAEPHash.New(), rand.Reader, r.pub, plaintext, r.label)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt with rsa-oaep: %w", err)
	}

	result := make([]byte, 1+len(ciphertext))
	result[0] = RSAVersion
	copy(result[1:], ciphertext)

	return result, nil
}

func (r *RSA) decryptHybrid(ciphertext []byte) ([]byte, error) {
	if r.priv == nil {
		return nil, errors.New("rsa private key is not set")
	}

	if len(ciphertext) == 0 {
		return nil, errors.New("ciphertext cannot be empty")
	}

	keySize := r.priv.Size()
	// 16 = GCM tag overhead size, so minimum ciphertext length should be: 1 (version) + keySize (wrapped key) + rsaGCMNonceSize (nonce) + 16 (GCM tag)
	minLen := 1 + keySize + rsaGCMNonceSize + 16 // nolint: mnd
	if len(ciphertext) < minLen {
		return nil, errors.New("invalid ciphertext length")
	}

	wrappedKey := ciphertext[1 : 1+keySize]
	nonce := ciphertext[1+keySize : 1+keySize+rsaGCMNonceSize]
	gcmCiphertext := ciphertext[1+keySize+rsaGCMNonceSize:]

	aesKey, err := rsa.DecryptOAEP(rsaOAEPHash.New(), rand.Reader, r.priv, wrappedKey, r.label)
	if err != nil {
		return nil, fmt.Errorf("failed to unwrap aes key with rsa-oaep: %w", err)
	}

	if len(aesKey) != rsaHybridKeySize {
		return nil, fmt.Errorf("invalid aes key size: %d", len(aesKey))
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create aes cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create gcm: %w", err)
	}

	plaintext, err := aead.Open(nil, nonce, gcmCiphertext, r.label)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt with aes-gcm: %w", err)
	}

	return plaintext, nil
}

func (r *RSA) decryptOAEP(ciphertext []byte) ([]byte, error) {
	if r.priv == nil {
		return nil, errors.New("rsa private key is not set")
	}

	if len(ciphertext) == 0 {
		return nil, errors.New("ciphertext cannot be empty")
	}

	// Remove version byte
	actualCiphertext := ciphertext[1:]
	plaintext, err := rsa.DecryptOAEP(rsaOAEPHash.New(), rand.Reader, r.priv, actualCiphertext, r.label)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt with rsa-oaep: %w", err)
	}

	return plaintext, nil
}

// RSAOption is an option for the RSA crypter.
type RSAOption func(*RSA)

// WithRSALabel sets the label for RSA-OAEP.
func WithRSALabel(label []byte) RSAOption {
	return func(r *RSA) {
		r.label = label
	}
}

// NewRSACrypterFromPrivateKey creates a new RSA crypter from a private key.
// It enables both encryption and decryption.
func NewRSACrypterFromPrivateKey(pemBytes []byte, opts ...RSAOption) (Crypter, error) {
	if pemBytes == nil {
		return nil, errors.New("rsa pem private key cannot be nil")
	}

	priv, err := parseRSAPrivateKeyFromPEM(pemBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rsa private key: %w", err)
	}

	r := &RSA{
		priv:  priv,
		pub:   &priv.PublicKey,
		label: []byte(RSADefaultLabel),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r, nil
}

// Encrypt encrypts the plaintext using RSA-OAEP.
// For plaintext larger than RSA-OAEP limit, it falls back to hybrid encryption (AES-256-GCM + RSA-OAEP wrapped key).
func (r *RSA) Encrypt(plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return nil, errors.New("plaintext cannot be empty")
	}

	if r.pub == nil {
		return nil, errors.New("rsa public key is not set")
	}

	if len(plaintext) > r.oaepMaxPlaintextSize() {
		return r.encryptHybrid(plaintext)
	}

	return r.encryptOAEP(plaintext)
}

// Decrypt decrypts the ciphertext.
// It supports both RSA-OAEP ciphertext (v1) and hybrid ciphertext (v2).
func (r *RSA) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) <= 1 {
		return nil, errors.New("invalid ciphertext length")
	}

	if r.priv == nil {
		return nil, errors.New("rsa private key is not set")
	}

	switch ciphertext[0] {
	case RSAVersion:
		return r.decryptOAEP(ciphertext)
	case RSAVersionHybrid:
		return r.decryptHybrid(ciphertext)
	default:
		return nil, fmt.Errorf("unsupported rsa version: %d", ciphertext[0])
	}
}

// parseRSAPrivateKeyFromPEM parses an RSA private key from a PEM-encoded block.
// It supports PKCS#1 and PKCS#8 formats.
func parseRSAPrivateKeyFromPEM(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("invalid pem data for rsa private key")
	}

	// PKCS1 first if possible
	if priv, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return priv, nil
	}

	// then try PKCS8
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rsa private key: %w", err)
	}

	priv, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("parsed private key is not RSA")
	}

	return priv, nil
}

// GenerateRSAKeyPairPEM generates an RSA key pair and returns the PEM-encoded private key and public key.
func GenerateRSAKeyPairPEM(bits RSAKeySize) ([]byte, []byte, error) {
	priv, err := generateRSAKeyPair(bits)
	if err != nil {
		return nil, nil, err
	}

	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	privPEM := pem.EncodeToMemory(&pem.Block{Type: PEMBlockTypeRSAPrivateKeyType, Bytes: privBytes})

	pubBytes, err := x509.MarshalPKIXPublicKey(priv.Public())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal rsa public key: %w", err)
	}

	pubPEM := pem.EncodeToMemory(&pem.Block{Type: PEMBlockTypePublicKeyType, Bytes: pubBytes})

	return privPEM, pubPEM, nil
}

// generateRSAKeyPair generates an RSA private/public key pair with the given bit size.
func generateRSAKeyPair(bits RSAKeySize) (*rsa.PrivateKey, error) {
	if err := bits.Validate(); err != nil {
		return nil, fmt.Errorf("validate rsa key size failed, bits(%d): %w", bits, err)
	}

	priv, err := rsa.GenerateKey(rand.Reader, bits.ToInt())
	if err != nil {
		return nil, fmt.Errorf("failed to generate rsa key: %w", err)
	}

	return priv, nil
}

// DecryptRSABase64Ciphertext decrypts RSA ciphertext encoded with standard base64.
// Expected format: base64(RSAVersion + raw RSA ciphertext).
func DecryptRSABase64Ciphertext(cry Crypter, base64Ciphertext string) (string, error) {
	if cry == nil {
		return "", errors.New("nil crypter")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(base64Ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 ciphertext: %w", err)
	}

	plaintext, err := cry.Decrypt(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt ciphertext: %w", err)
	}

	return string(plaintext), nil
}

// GetRSAPublicKeyPEMByPrivateKeyPEM extracts the public key PEM from the given private key PEM.
func GetRSAPublicKeyPEMByPrivateKeyPEM(privPEM []byte) ([]byte, error) {
	priv, err := parseRSAPrivateKeyFromPEM(privPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rsa private key: %w", err)
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(priv.Public())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal rsa public key: %w", err)
	}

	pubPEM := pem.EncodeToMemory(&pem.Block{Type: PEMBlockTypePublicKeyType, Bytes: pubBytes})

	return pubPEM, nil
}
