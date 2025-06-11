/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package types

import (
	"errors"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
)

// ReleaseType defines the type of release.
type ReleaseType string

const (
	// ReleaseTypeOriginAgent defines the release of origin gse agent package.
	ReleaseTypeOriginAgent ReleaseType = "origin_agent"

	// ReleaseTypeOriginServer defines the release of origin gse server package.
	ReleaseTypeOriginServer ReleaseType = "origin_server"

	// ReleaseTypeAgent defines the release of nodemgr agent package transformed from origin gse agent package.
	ReleaseTypeAgent ReleaseType = "agent"

	// ReleaseTypeProxy defines the release of nodemgr proxy package transformed from origin gse agent+server package.
	ReleaseTypeProxy ReleaseType = "proxy"
)

// Validate validates the release type.
func (rt ReleaseType) Validate() error {
	switch rt {
	case ReleaseTypeOriginAgent, ReleaseTypeOriginServer, ReleaseTypeAgent, ReleaseTypeProxy:
		return nil
	default:
		return errors.New("invalid release type")
	}
}

// ReleaseTypeListToStringList converts a release type list to string list.
func ReleaseTypeListToStringList(releaseTypeList []ReleaseType) []string {
	data := make([]string, len(releaseTypeList))
	for idx, releaseType := range releaseTypeList {
		data[idx] = string(releaseType)
	}

	return data
}

// StringListToReleaseTypeList converts a string list to release type list.
func StringListToReleaseTypeList(releaseTypeList []string) []ReleaseType {
	data := make([]ReleaseType, len(releaseTypeList))
	for idx, releaseType := range releaseTypeList {
		data[idx] = ReleaseType(releaseType)
	}

	return data
}

// Release defines the release package information.
type Release struct {
	Generation  Generation
	Type        ReleaseType
	Version     string
	Platform    platform.Platform
	Labels      []string
	ChangeLogEN string
	ChangeLogZH string
	FileName    string
	UpstreamDir string
	LocalDir    string
	MD5         string
	UpdatedAt   time.Time
}

// OriginPkgDetail defines the detail of origin package.
type OriginPkgDetail struct {
	iface.FileInfo

	Version     string
	Platforms   []platform.Platform
	ChangeLogEN string
	ChangeLogZH string
}
