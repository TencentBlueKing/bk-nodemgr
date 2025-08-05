/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package file defines the file manager interface.
package file

import (
	"context"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
)

// IFileManager defines the file manager interface.
type IFileManager interface {
	// GetFileInfo get file info.
	GetFileInfo(ctx context.Context, filename string) (*iface.FileInfo, error)

	// DownloadFile returnn file content.
	DownloadFile(ctx context.Context, filename string) (iface.FileContent, *iface.FileInfo, error)

	// StoreFile store file form srcPath.
	StoreFile(ctx context.Context, srcPath, filename string) (*iface.FileInfo, error)

	// FileExists check file exists.
	FileExists(ctx context.Context, filename string) bool
}
