/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package cipher defines the interface for cipher storage handlers.
package cipher

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IStorage defines the interface for asymmetric encryption storage.
type IStorage interface {
	basestorage.Interface

	IDaoStorage
	IDomainStorage
}

// IDaoStorage defines single-table (cipher collection) data access interfaces.
type IDaoStorage interface {
	// GetCipher gets asymmetric encryption by name and key type.
	GetCipher(nCtx contextx.IContext, name string, keyType types.CipherKeyType) (*types.Cipher, error)

	// Exist checks if asymmetric encryption exists by name and key type.
	ExistCipher(nCtx contextx.IContext, name string, keyType types.CipherKeyType) (bool, error)

	// CreateCipher creates asymmetric encryption records.
	CreateCipher(nCtx contextx.IContext, encryption ...*types.Cipher) error
}

// IDomainStorage defines business domain interfaces for cipher storage.
type IDomainStorage interface {
	// EnsureDefaultCipher ensures the default RSA cipher.
	EnsureDefaultCipher(nCtx contextx.IContext) error
}
