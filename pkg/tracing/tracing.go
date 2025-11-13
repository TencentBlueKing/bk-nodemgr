/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package tracing

import (
	"fmt"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

// nolint: gochecknoglobals
var globalHandler struct {
	sync.Once
	IHandler
}

// Init init the global tracing globalHandler.
func Init(conf Config) error {
	var err error
	globalHandler.Do(func() {
		globalHandler.IHandler, err = New(contextx.Background(), conf)
		if err != nil {
			err = fmt.Errorf("failed to init tracing globalHandler, err: %w", err)
		}
	})

	return err
}

// G get the global tracing globalHandler.
func G() IHandler {
	globalHandler.Once.Do(func() {
		_ = Init(DefaultConfig())
	})

	return globalHandler.IHandler
}
