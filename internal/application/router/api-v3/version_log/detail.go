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

package versionlog

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/support-files/changelog"
)

// VersionLogDetail returns the changelog for a specific version.
func (h *handler) VersionLogDetail(rCtx restserver.IContext) (interface{}, error) {
	version := rCtx.GContext().Param("version")
	if !isValidVersion(version) {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("invalid version format: %s", version))
	}

	lang, _ := rCtx.GetCookie(languageCookieKey)
	locale := normalizeLanguage(lang)
	entries, err := changelog.FS.ReadDir(locale)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to read changelog directory")
		return nil, resterrf.ErrWrap(resterrf.RecordNotFound, fmt.Errorf("changelog not found for version: %s", version))
	}

	filename := findChangelogFilename(entries, version)
	if filename == "" {
		return nil, resterrf.ErrWrap(resterrf.RecordNotFound, fmt.Errorf("changelog not found for version: %s", version))
	}

	content, err := changelog.FS.ReadFile(locale + "/" + filename)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to read changelog file")
		return nil, resterrf.ErrWrap(resterrf.RecordNotFound, err)
	}

	return &protoApplication.VersionLogDetailResp_Data{
		Version: version,
		Date:    extractChangelogDate(filename),
		Content: string(content),
	}, nil
}
