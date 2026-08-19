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
	"sort"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/support-files/changelog"
)

// VersionLogsList returns the list of available versions.
func (h *handler) VersionLogsList(rCtx restserver.IContext) (interface{}, error) {
	lang, _ := rCtx.GetCookie(languageCookieKey)
	entries, err := changelog.FS.ReadDir(normalizeLanguage(lang))
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to read changelog directory")
		return &protoApplication.VersionLogsListResp_Data{VersionLogs: []*protoApplication.VersionLogEntry{}}, nil
	}

	versionLogs := make([]*protoApplication.VersionLogEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		versionLog, ok := buildVersionLogEntry(entry.Name())
		if !ok {
			continue
		}

		versionLogs = append(versionLogs, versionLog)
	}

	sort.Slice(versionLogs, func(i, j int) bool {
		return compareVersions(versionLogs[i].GetVersion(), versionLogs[j].GetVersion()) < 0
	})

	return &protoApplication.VersionLogsListResp_Data{VersionLogs: versionLogs}, nil
}
