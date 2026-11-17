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

package wiki

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/support-files/wiki"
)

// ReadFile returns an embedded document or resource at <language>/<relative path>.
// Invalid paths, unsupported languages and missing files wrap fs.ErrNotExist.
// Content access does not depend on the configured link source.
func ReadFile(name string) ([]byte, error) {
	language, _, _ := strings.Cut(name, "/")
	if !fs.ValidPath(name) || (language != string(LanguageZh) && language != string(LanguageEn)) {
		return nil, fmt.Errorf("wiki file %q: %w", name, fs.ErrNotExist)
	}

	content, err := wiki.FS.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("failed to read wiki file %q: %w", name, err)
	}

	return content, nil
}
