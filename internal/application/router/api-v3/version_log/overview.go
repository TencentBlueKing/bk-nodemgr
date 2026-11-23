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
	"errors"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/version"
	"github.com/TencentBlueKing/bk-nodemgr/support-files/changelog"
)

var overviewVersionRegex = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

func currentOverviewVersion() string {
	if !isValidVersion(version.VERSION) {
		return ""
	}

	baseVersion, _, _ := strings.Cut(version.VERSION, "-")

	return baseVersion
}

// VersionOverviewList returns the available overviews and the running base version.
func (h *handler) VersionOverviewList(rCtx restserver.IContext) (interface{}, error) {
	lang, _ := rCtx.GetCookie(languageCookieKey)
	entries, err := changelog.FS.ReadDir(normalizeLanguage(lang) + "/overview")
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to read version overview directory")
		return nil, resterrf.ErrWrap(resterrf.RecordNotFound, err)
	}

	currentVersion := currentOverviewVersion()
	versionLogs := make([]*protoApplication.VersionOverviewEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		baseVersion, ok := strings.CutSuffix(entry.Name(), ".md")
		if !ok || !overviewVersionRegex.MatchString(baseVersion) {
			continue
		}

		versionLogs = append(versionLogs, &protoApplication.VersionOverviewEntry{
			Version:   baseVersion,
			IsCurrent: baseVersion == currentVersion,
		})
	}

	sort.Slice(versionLogs, func(i, j int) bool {
		return compareVersions(versionLogs[i].GetVersion(), versionLogs[j].GetVersion()) < 0
	})

	return &protoApplication.VersionOverviewListResp_Data{
		VersionLogs:    versionLogs,
		CurrentVersion: currentVersion,
	}, nil
}

// VersionOverviewDetail returns the independently authored overview for a base version.
func (h *handler) VersionOverviewDetail(rCtx restserver.IContext) (interface{}, error) {
	baseVersion := rCtx.GContext().Param("version")
	if !overviewVersionRegex.MatchString(baseVersion) {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("invalid base version format: %s", baseVersion))
	}

	lang, _ := rCtx.GetCookie(languageCookieKey)
	directory := normalizeLanguage(lang) + "/overview"
	content, err := changelog.FS.ReadFile(directory + "/" + baseVersion + ".md")
	if errors.Is(err, fs.ErrNotExist) {
		return nil, resterrf.ErrWrap(resterrf.RecordNotFound, fmt.Errorf("version overview not found: %s", baseVersion))
	}
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to read version overview file")
		return nil, resterrf.ErrWrap(resterrf.RecordNotFound, err)
	}

	return &protoApplication.VersionOverviewDetailResp_Data{
		Version: baseVersion,
		Content: string(content),
	}, nil
}
