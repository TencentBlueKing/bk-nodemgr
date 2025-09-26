/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package creditvault

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

// IHostPasswordVault defines the password vault interface.
type IHostPasswordVault interface {
	// LoadPassword load password.
	LoadPassword(nCtx contextx.IContext, rtx string, networkAreaID int64, ip string, loginUser string) (string, error)
}

// WithHostPasswordVault set host password vault.
func WithHostPasswordVault(vault IHostPasswordVault) Option {
	return func(cv *CreditVault) {
		if vault != nil {
			cv.IHostPasswordVault = vault
		}
	}
}

// DisabledHostPasswordVault this is a disabled host password vault.
type DisabledHostPasswordVault struct {
}

// LoadPassword load password.
func (v *DisabledHostPasswordVault) LoadPassword(
	_ contextx.IContext,
	_ string,
	_ int64,
	_ string,
	_ string,
) (string, error) {

	return "", errors.New("disabled host password vault")
}
