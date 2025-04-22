/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package local support the realize of filex to control local LocalFile.
package local

import (
	"sync"

	"github.com/spf13/afero"
)

var readFS struct {
	once sync.Once
	fs   afero.Fs
}

func rFs() afero.Fs {
	readFS.once.Do(func() {
		readFS.fs = afero.NewOsFs()
	})

	return readFS.fs
}

var writeFS struct {
	once sync.Once
	fs   afero.Fs
}

func wFs() afero.Fs {
	writeFS.once.Do(func() {
		writeFS.fs = afero.NewOsFs()
	})

	return writeFS.fs
}
