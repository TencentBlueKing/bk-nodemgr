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

	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
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

	// ReleaseTypeCert defines the release of nodemgr cert package.
	ReleaseTypeCert ReleaseType = "cert"

	// ReleaseTypeBinTool defines the release of nodemgr bin tool package.
	ReleaseTypeBinTool ReleaseType = "bintool"
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
	MD5         string
	Enabled     bool
	AsDefault   bool
	UpdatedAt   time.Time
	Operator    string
}

// ReleaseCert defines the cert, it is kind of Release.
type ReleaseCert struct {
	FileName string
	MD5      string
}

// ReleaseBinTool defines the bin tool, it is kind of Release.
type ReleaseBinTool struct {
	Generation Generation
	FileName   string
	MD5        string
}

// OriginPkgDetail defines the detail of origin package.
type OriginPkgDetail struct {
	fileiface.FileInfo

	UploadID    string
	Existed     bool
	Version     string
	Platforms   []platform.Platform
	ChangeLogEN string
	ChangeLogZH string
}

// TargetPkgDetail defines the detail of target package.
type TargetPkgDetail struct {
}

// OriginCertPkgDetail defines the detail of cert package.
type OriginCertPkgDetail struct {
	fileiface.FileInfo

	UploadID  string
	Existed   bool
	CertFiles []string
}

// OriginBinToolPkgDetail defines the detail of bin tool package.
type OriginBinToolPkgDetail struct {
	fileiface.FileInfo

	UploadID       string
	Existed        bool
	AgentPlatforms []platform.Platform
	ProxyPlatforms []platform.Platform
}
