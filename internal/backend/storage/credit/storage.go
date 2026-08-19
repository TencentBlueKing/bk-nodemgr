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

package credit

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

// IStorage defines the Storage interface.
type IStorage interface {
	basestorage.Interface

	IStorageHostCredit
}

// IStorageHostCredit defines the StorageHostCredit interface.
type IStorageHostCredit interface {
	basestorage.Interface

	// CreateHostCreditWithExpiredAt create host credit data with custom expired time.
	CreateHostCredit(
		nCtx contextx.IContext,
		creditData []byte,
		expiredAt time.Time,
	) (string, error)

	// LoadHostCredit load host credit data.
	LoadHostCredit(
		nCtx contextx.IContext,
		creditID string,
	) ([]byte, error)

	// CheckHostCreditValid check host credit valid.
	CheckHostCreditValid(
		nCtx contextx.IContext,
		creditIDList ...string,
	) (map[string]bool, error)
}
