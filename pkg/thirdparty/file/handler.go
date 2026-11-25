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

package file

import (
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
)

// IHandler is interface for file handler.
type IHandler interface {
	IPkgUploadHandler
	IPkgPublishHandler
	IPkgDownloadHandler
	IPkgInfoHandler
	IPkgExportHandler
	IPkgReleaseHandler
	IPackageEventHandler
	ITransfer
	ISharedReleaseSyncHandler
	IPluginRequirementHandler
}

type handler struct {
	cli *cli
}

// New initialize a new nodeman backend handler.
func New(c *restclient.Capability, conf *Config) (IHandler, error) {
	cli, err := newClient(c, conf)
	if err != nil {
		return nil, err
	}

	return &handler{cli: cli}, nil
}
