/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package globalsettings

import (
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

var handler struct {
	sync.Once
	IHandler
}

// Register registers the global settings handler.
func Register(nCtx contextx.IContext, stg IStorage) error {
	var err error
	handler.Do(func() {
		handler.IHandler, err = NewHandler(nCtx, stg)
	})

	return err
}

// ListAll lists all global settings.
func ListAll(nCtx contextx.IContext) ([]*types.GlobalSettings, int64, error) {
	return handler.ListAll(nCtx)
}

// Get gets a global settings.
func Get(nCtx contextx.IContext, name, defaultValue string) string {
	return handler.Get(nCtx, name, defaultValue)
}

// Upsert updates or inserts a global settings.
func Upsert(nCtx contextx.IContext, name, value string) error {
	return handler.Upsert(nCtx, name, value)
}

// Delete deletes global settings by names.
func Delete(nCtx contextx.IContext, names ...string) error {
	return handler.Delete(nCtx, names...)
}
