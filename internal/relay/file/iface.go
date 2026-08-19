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

// Package file defines the file manager interface.
package file

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
)

// IFileManager defines the file manager interface.
type IFileManager interface {
	// GetFile get file info.
	GetFile(nCtx contextx.IContext, filename string) (iface.File, error)

	// StoreFile store file form srcPath.
	StoreFile(nCtx contextx.IContext, srcPath, filename string) (*iface.FileInfo, error)

	// FileExists check file exists.
	FileExists(nCtx contextx.IContext, filename, md5 string) bool
}
