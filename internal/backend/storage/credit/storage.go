/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package credit

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/basestorage"
)

// IStorage defines the Storage interface.
type IStorage interface {
	basestorage.Interface

	IStorageHostCredit
}

// IStorageHostCredit defines the StorageHostCredit interface.
type IStorageHostCredit interface {
	basestorage.Interface

	// CreateHostCredit create host credit data.
	CreateHostCredit(
		ctx contextx.ITenantContext,
		creditData []byte,
	) (string, error)

	// LoadHostCredit load host credit data.
	LoadHostCredit(
		ctx contextx.ITenantContext,
		creditID string,
	) ([]byte, error)

	// CheckHostCreditValid check host credit valid.
	CheckHostCreditValid(
		ctx contextx.ITenantContext,
		creditIDList ...string,
	) (map[string]bool, error)
}
