/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package asymmetricencryption defines the interface for asymmetric encryption storage handlers.
package asymmetricencryption

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IStorage defines the interface for asymmetric encryption storage.
type IStorage interface {
	basestorage.Interface

	// GetAsymmetricEncryption gets asymmetric encryption.
	GetAsymmetricEncryption(nCtx contextx.IContext, keyType types.AsymmetricKeyType, cipherType types.AsymmetricCipherType) (
		*types.AsymmetricEncryption, error)

	// CreateAsymmetricEncryption creates asymmetric encryption records.
	CreateAsymmetricEncryption(nCtx contextx.IContext, encryption ...*types.AsymmetricEncryption) error

	// UpsertAsymmetricEncryption updates or inserts asymmetric encryption.
	UpsertAsymmetricEncryption(nCtx contextx.IContext, encryption *types.AsymmetricEncryption) error

	// Exist checks if asymmetric encryption exists by key-type and cipher-type.
	ExistAsymmetricEncryption(nCtx contextx.IContext, keyType types.AsymmetricKeyType, cipherType types.AsymmetricCipherType) (bool, error)

	// DeleteAsymmetricEncryption deletes asymmetric encryption by key-type and cipher-type.
	DeleteAsymmetricEncryption(nCtx contextx.IContext, keyType types.AsymmetricKeyType, cipherType types.AsymmetricCipherType) error
}
