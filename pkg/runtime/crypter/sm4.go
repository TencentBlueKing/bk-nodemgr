/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

// Package crypter provides encryption utilities based on SM4-GCM mode.
package crypter

import (
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/emmansun/gmsm/sm4"
)

const (
	// SM4DefaultSalt is the default salt used to derive the SM4 key.
	SM4DefaultSalt = "com.example.crypto.sm4.v1"

	// SM4Version the version of SM4-GCM algorithm, distinct from AESVersion
	// so stored ciphertexts stay dispatchable if the symmetric suite evolves.
	SM4Version = 2

	// sm4KeyLen is the SM4 key length in bytes (128-bit).
	sm4KeyLen = 16

	// sm4GCMNonceLen is the GCM nonce length in bytes.
	sm4GCMNonceLen = 12

	// sm4GCMTagLen is the GCM authentication tag length in bytes.
	sm4GCMTagLen = 16
)

// SM4 defines the crypter instance based on SM4-GCM, backed by
// emmansun/gmsm (pure Go).
type SM4 struct {
	key  []byte
	salt []byte
	aead cipher.AEAD
}

// SM4Option is an option for the SM4 crypter.
type SM4Option func(*SM4)

// WithSM4Salt sets the salt for the SM4 crypter.
func WithSM4Salt(salt []byte) SM4Option {
	return func(crypt *SM4) {
		crypt.salt = salt
	}
}

// NewSM4Crypter creates a new SM4 crypter instance.
// It derives a fixed-length 16-byte key via SHA-256 over (key || salt).
func NewSM4Crypter(key []byte, opts ...SM4Option) (Crypter, error) {
	if len(key) == 0 {
		return nil, errors.New("key cannot be empty")
	}

	crypt := &SM4{
		salt: []byte(SM4DefaultSalt),
	}

	for _, opt := range opts {
		opt(crypt)
	}

	// reuse the AES key derivation and truncate to the SM4 key length.
	crypt.key = deriveKey(key, crypt.salt)[:sm4KeyLen]

	block, err := sm4.NewCipher(crypt.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create sm4 cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create sm4-gcm: %w", err)
	}

	crypt.aead = aead

	return crypt, nil
}

// Encrypt encrypts the plaintext using SM4-GCM with a random nonce.
func (crypt *SM4) Encrypt(plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return nil, errors.New("plaintext cannot be empty")
	}

	nonce := make([]byte, sm4GCMNonceLen)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate gcm nonce: %w", err)
	}

	// GCM ciphertext appends the 16-byte authentication tag.
	ciphertext := crypt.aead.Seal(nil, nonce, plaintext, nil)

	// assembling the final ciphertext: SM4Version(1) + Nonce(12) + Ciphertext(including tag)
	result := make([]byte, 1+len(nonce)+len(ciphertext))
	result[0] = SM4Version
	copy(result[1:], nonce)
	copy(result[1+len(nonce):], ciphertext)

	return result, nil
}

// Decrypt decrypts the SM4-GCM ciphertext. Authentication is verified:
// ciphertexts with a wrong key or tampered content fail with an error.
func (crypt *SM4) Decrypt(ciphertext []byte) ([]byte, error) {
	// at least version(1) + nonce(12) + one ciphertext block with tag(16+1).
	if len(ciphertext) < 1+sm4GCMNonceLen+sm4GCMTagLen+1 {
		return nil, errors.New("invalid ciphertext length")
	}

	// verification version.
	if ciphertext[0] != SM4Version {
		return nil, fmt.Errorf("unsupported version: %d", ciphertext[0])
	}

	// extract nonce and ciphertext.
	nonce := ciphertext[1 : 1+sm4GCMNonceLen]
	actualCiphertext := ciphertext[1+sm4GCMNonceLen:]

	plaintext, err := crypt.aead.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt with sm4-gcm: %w", err)
	}

	return plaintext, nil
}
