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

// Package token creates compact, authenticated tokens with embedded expiration times.
package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

const (
	formatVersion  = byte(1)
	headerSize     = 5
	tagSize        = 12
	minimumSize    = headerSize + 1 + tagSize
	defaultTTL     = 24 * time.Hour
	defaultKey     = "bk-nodemgr/runtime/token/default-key"
	maxUnixSeconds = int64(1<<32 - 1)
	signingDomain  = "bk-nodemgr/runtime/token/v1"
)

var (
	// ErrInvalid indicates that a token is malformed, unsupported, or has an invalid signature.
	ErrInvalid = errors.New("token: invalid")
	// ErrExpired indicates that a valid token has reached its embedded expiration time.
	ErrExpired = errors.New("token: expired")
)

// Option configures a Generator during construction.
type Option func(*Generator) error

// WithTTL sets the lifetime embedded in newly generated tokens.
// The TTL must be positive and is stored with one-second precision.
func WithTTL(ttl time.Duration) Option {
	return func(generator *Generator) error {
		if ttl <= 0 {
			return errors.New("ttl must be positive")
		}

		generator.ttl = ttl

		return nil
	}
}

// Generator creates and parses authenticated tokens.
// A Generator is safe for concurrent use after construction.
type Generator struct {
	key []byte
	ttl time.Duration
}

// New creates a Generator with a default TTL of 24 hours.
// An empty key uses the package default key; a non-empty key overrides it.
func New(key []byte, opts ...Option) (*Generator, error) {
	if len(key) == 0 {
		key = []byte(defaultKey)
	}

	generator := &Generator{
		key: append([]byte(nil), key...),
		ttl: defaultTTL,
	}
	for _, opt := range opts {
		if opt == nil {
			return nil, errors.New("token: option cannot be nil")
		}
		if err := opt(generator); err != nil {
			return nil, fmt.Errorf("token: applying option: %w", err)
		}
	}

	return generator, nil
}

// Generate authenticates payload and embeds its expiration time in a URL-safe token.
// The payload must not be empty. The returned expiration time exactly matches the
// second-precision value stored in the token.
func (generator *Generator) Generate(payload []byte) (string, time.Time, error) {
	if len(payload) == 0 {
		return "", time.Time{}, errors.New("token: payload cannot be empty")
	}

	expiresUnix := time.Now().Add(generator.ttl).Unix()
	if expiresUnix < 0 || expiresUnix > maxUnixSeconds {
		return "", time.Time{}, errors.New("token: expiration is outside the supported range")
	}

	message := make([]byte, headerSize+len(payload))
	message[0] = formatVersion
	binary.BigEndian.PutUint32(message[1:headerSize], uint32(expiresUnix))
	copy(message[headerSize:], payload)

	raw := make([]byte, len(message)+tagSize)
	copy(raw, message)
	copy(raw[len(message):], generator.sign(message))

	expiresAt := time.Unix(expiresUnix, 0).UTC()

	return base64.RawURLEncoding.EncodeToString(raw), expiresAt, nil
}

// Parse verifies token integrity and expiration, then returns a copy of its payload.
// It returns ErrInvalid for untrusted or unsupported input and ErrExpired for an
// authentic token whose expiration time has passed.
func (generator *Generator) Parse(value string) ([]byte, error) {
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("%w: malformed encoding", ErrInvalid)
	}
	if len(raw) < minimumSize {
		return nil, fmt.Errorf("%w: unexpected length", ErrInvalid)
	}
	if raw[0] != formatVersion {
		return nil, fmt.Errorf("%w: unsupported version", ErrInvalid)
	}

	tagOffset := len(raw) - tagSize
	message := raw[:tagOffset]
	if !hmac.Equal(raw[tagOffset:], generator.sign(message)) {
		return nil, fmt.Errorf("%w: signature mismatch", ErrInvalid)
	}

	expiresAt := time.Unix(int64(binary.BigEndian.Uint32(message[1:headerSize])), 0)
	if !time.Now().Before(expiresAt) {
		return nil, ErrExpired
	}

	return append([]byte(nil), message[headerSize:]...), nil
}

func (generator *Generator) sign(message []byte) []byte {
	mac := hmac.New(sha256.New, generator.key)
	_, _ = mac.Write([]byte(signingDomain))
	_, _ = mac.Write(message)

	return mac.Sum(nil)[:tagSize]
}
