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

// Package kms handles the secret from kms.
package kms

import (
	"encoding/json"
	"fmt"

	"github.com/TencentBlueKing/bk-kms-sdk/go/decrypt"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config/kms/envelope"
)

// EnvelopeLoadParam the load param.
type EnvelopeLoadParam struct {
	// global envelope and private-key.
	EnvelopeGlobal   string
	PrivateKeyGlobal string

	// module envelope and private-key.
	EnvelopeModule   string
	PrivateKeyModule string
}

// LoadGlobalEnvelope loads only the kms global secret.
func LoadGlobalEnvelope(param EnvelopeLoadParam) (*envelope.GlobalEnvelope, error) {
	return loadKMSGlobalSecret(param.EnvelopeGlobal, param.PrivateKeyGlobal)
}

// LoadApplicationEnvelopes load the kms application secret.
func LoadApplicationEnvelopes(param EnvelopeLoadParam) (*envelope.GlobalEnvelope, *envelope.ApplicationEnvelope, error) {
	dataGlobal, err := loadKMSGlobalSecret(param.EnvelopeGlobal, param.PrivateKeyGlobal)
	if err != nil {
		return nil, nil, err
	}

	dataModule, err := loadApplicationEnvelopes(param.EnvelopeModule, param.PrivateKeyModule)
	if err != nil {
		return nil, nil, err
	}

	return dataGlobal, dataModule, nil
}

// LoadBackendEnvelopes load the kms backend secret.
func LoadBackendEnvelopes(param EnvelopeLoadParam) (*envelope.GlobalEnvelope, *envelope.BackendEnvelope, error) {
	dataGlobal, err := loadKMSGlobalSecret(param.EnvelopeGlobal, param.PrivateKeyGlobal)
	if err != nil {
		return nil, nil, err
	}

	dataModule, err := loadBackendEnvelopes(param.EnvelopeModule, param.PrivateKeyModule)
	if err != nil {
		return nil, nil, err
	}

	return dataGlobal, dataModule, nil
}

// LoadFileEnvelopes load the kms file secret.
func LoadFileEnvelopes(param EnvelopeLoadParam) (*envelope.GlobalEnvelope, *envelope.FileEnvelope, error) {
	dataGlobal, err := loadKMSGlobalSecret(param.EnvelopeGlobal, param.PrivateKeyGlobal)
	if err != nil {
		return nil, nil, err
	}

	dataModule, err := loadFileEnvelopes(param.EnvelopeModule, param.PrivateKeyModule)
	if err != nil {
		return nil, nil, err
	}

	return dataGlobal, dataModule, nil
}

func loadKMSGlobalSecret(envelopeGlobal, key string) (*envelope.GlobalEnvelope, error) {
	plainText, err := decrypt.Decrypt(envelopeGlobal, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt global secret: %w", err)
	}

	var data envelope.GlobalEnvelope
	err = json.Unmarshal([]byte(plainText), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal global secret: %w", err)
	}

	return &data, nil
}

func loadApplicationEnvelopes(envelopeModule, key string) (*envelope.ApplicationEnvelope, error) {
	plainText, err := decrypt.Decrypt(envelopeModule, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt application secret: %w", err)
	}

	var data envelope.ApplicationEnvelope
	err = json.Unmarshal([]byte(plainText), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal application secret: %w", err)
	}

	return &data, nil
}

func loadBackendEnvelopes(envelopeModule, key string) (*envelope.BackendEnvelope, error) {
	plainText, err := decrypt.Decrypt(envelopeModule, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt backend secret: %w", err)
	}

	var data envelope.BackendEnvelope
	err = json.Unmarshal([]byte(plainText), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal backend secret: %w", err)
	}

	return &data, nil
}

func loadFileEnvelopes(envelopeModule, key string) (*envelope.FileEnvelope, error) {
	plainText, err := decrypt.Decrypt(envelopeModule, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt file secret: %w", err)
	}

	var data envelope.FileEnvelope
	err = json.Unmarshal([]byte(plainText), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal file secret: %w", err)
	}

	return &data, nil
}
