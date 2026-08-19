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

package globalsettings

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandler defines the interface for global settings.
type IHandler interface {
	// ListAll list all global settings.
	ListAll(ctx contextx.IContext) ([]*types.GlobalSettings, int64, error)

	// Get gets a global settings by setting name return default value if not exist.
	Get(ctx contextx.IContext, name string, defaultValue string) string

	// Upsert update or insert a global settings.
	Upsert(ctx contextx.IContext, name, value string) error

	// Delete delete global settings.
	Delete(ctx contextx.IContext, names ...string) error
}
