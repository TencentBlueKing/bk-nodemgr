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

// Package crypter provides encryption utilities based on AES-CBC mode
package crypter

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"errors"
	"fmt"
)

const (
	// AESDefaultSalt is the default salt used to derive the AES key.
	AESDefaultSalt = "com.example.crypto.v1"
	// AESVersion the version of AES-GCM algorithm.
	AESVersion = 1
)

// AES defines the crypter instance based on AES algorithm.
type AES struct {
	key   []byte
	block cipher.Block
	salt  []byte
}

// Option is an option for the AES crypter.
type Option func(*AES)

// WithSalt sets the salt for the AES crypter.
func WithSalt(salt []byte) Option {
	return func(c *AES) {
		c.salt = salt
	}
}

// NewAESCrypter creates a new AES crypter instance.
func NewAESCrypter(key []byte, opts ...Option) (Crypter, error) {
	if len(key) == 0 {
		return nil, errors.New("key cannot be empty")
	}

	a := &AES{
		salt: []byte(AESDefaultSalt),
	}

	for _, opt := range opts {
		opt(a)
	}

	// use hkdf derived fixed length keys
	derivedKey := deriveKey(key, a.salt)

	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	a.block = block
	a.key = derivedKey

	return a, nil
}

// deriveKey define key derivation function.
func deriveKey(key, salt []byte) []byte {
	// Use SHA-256 to ensure that the output key length is fixed at 32 bytes.
	h := sha256.New()
	h.Write(key)
	h.Write(salt)

	return h.Sum(nil)
}

// generating deterministic ivs.
func (a *AES) generateIV(plaintext []byte) []byte {
	// Using the hash of the plaintext as part of the IV ensures that the same plaintext produces the same IV
	h := sha256.New()
	h.Write(plaintext)
	h.Write(a.key)
	hash := h.Sum(nil)

	// take the first 16 bytes as iv.
	return hash[:16]
}

// Encrypt encrypts the plaintext.
func (a *AES) Encrypt(plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return nil, errors.New("plaintext cannot be empty")
	}

	// generating deterministic ivs.
	iv := a.generateIV(plaintext)

	// encryption using cbc mode.
	mode := cipher.NewCBCEncrypter(a.block, iv)

	// PKCS7 padding.
	paddedPlaintext := pkcs7Pad(plaintext, aes.BlockSize)
	ciphertext := make([]byte, len(paddedPlaintext))
	mode.CryptBlocks(ciphertext, paddedPlaintext)

	// assembling the final ciphertext: AESVersion(1) + IV(16) + Ciphertext
	result := make([]byte, 1+len(iv)+len(ciphertext))
	result[0] = AESVersion
	copy(result[1:], iv)
	copy(result[1+len(iv):], ciphertext)

	return result, nil
}

// Decrypt decrypts the ciphertext.
func (a *AES) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < 1+16+aes.BlockSize {
		return nil, errors.New("invalid ciphertext length")
	}

	// verification version.
	if ciphertext[0] != AESVersion {
		return nil, fmt.Errorf("unsupported version: %d", ciphertext[0])
	}

	// extract iv and ciphertext.
	iv := ciphertext[1:17]
	actualCiphertext := ciphertext[17:]

	// decryption using cbc mode.
	mode := cipher.NewCBCDecrypter(a.block, iv)
	plaintext := make([]byte, len(actualCiphertext))
	mode.CryptBlocks(plaintext, actualCiphertext)

	// PKCS7 unpadding.
	unpaddedPlaintext, err := pkcs7Unpad(plaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to unpad: %w", err)
	}

	return unpaddedPlaintext, nil
}

// PKCS7 padding.
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

// PKCS7 unpadding.
func pkcs7Unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("invalid padding size")
	}

	padding := int(data[length-1])
	if padding > length {
		return nil, errors.New("invalid padding size")
	}

	return data[:length-padding], nil
}
