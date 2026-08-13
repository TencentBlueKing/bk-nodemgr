/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package downloader

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
)

// IHandler is the handler of downloader.
type IHandler interface {
	IDownloader
}

// IDownloader defines the interface of downloader.
type IDownloader interface {
	// Download retrieves rawURL, verifies it against opts.Checksum, and returns
	// the verified file. The file is backed by a temporary file; close the
	// reader returned by Content to remove it.
	Download(nCtx contextx.IContext, rawURL string, opts DownloadOptions) (fileiface.File, error)
}
