/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package versionlog

import (
	"io/fs"
	"regexp"
	"strconv"
	"strings"

	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/version"
)

const (
	languageCookieKey = "blueking_language"
	defaultLanguage   = "zh"
	englishLanguage   = "en"
	stableTag         = "stable"
	stableTagOrder    = 0
	alphaTagOrder     = 1
	betaTagOrder      = 2
	rcTagOrder        = 3
)

var (
	versionRegex       = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)(?:-(alpha|beta|rc)\.(\d+))?(?:-(\d+))?$`)
	changelogFileRegex = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)(?:-(alpha|beta|rc)\.(\d+))?(?:-(\d+))?_(\d{4}-\d{2}-\d{2})\.md$`)
)

type parsedVersion struct {
	major    int
	minor    int
	patch    int
	tag      string
	tagPatch int
	mark     int
}

func normalizeLanguage(lang string) string {
	switch {
	case strings.HasPrefix(lang, defaultLanguage):
		return defaultLanguage
	case strings.HasPrefix(lang, englishLanguage):
		return englishLanguage
	default:
		return defaultLanguage
	}
}

func isValidVersion(version string) bool {
	return versionRegex.MatchString(version)
}

func buildVersionLogEntry(filename string) (*protoApplication.VersionLogEntry, bool) {
	matches := changelogFileRegex.FindStringSubmatch(filename)
	if len(matches) == 0 {
		return nil, false
	}

	separatorIndex := strings.LastIndex(filename, "_")
	if separatorIndex < 0 {
		return nil, false
	}

	versionName := filename[:separatorIndex]

	return &protoApplication.VersionLogEntry{
		Version:   versionName,
		Date:      matches[7],
		IsCurrent: versionName == version.VERSION,
	}, true
}

func extractChangelogDate(filename string) string {
	matches := changelogFileRegex.FindStringSubmatch(filename)
	if len(matches) == 0 {
		return ""
	}

	return matches[7]
}

func findChangelogFilename(entries []fs.DirEntry, version string) string {
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if strings.HasPrefix(filename, version+"_") && strings.HasSuffix(filename, ".md") {
			return filename
		}
	}

	return ""
}

func compareVersions(leftVersion, rightVersion string) int {
	left := parseVersion(leftVersion)
	right := parseVersion(rightVersion)

	for _, result := range []int{
		compareIntDesc(left.major, right.major),
		compareIntDesc(left.minor, right.minor),
		compareIntDesc(left.patch, right.patch),
		compareIntDesc(versionTagOrder(left.tag), versionTagOrder(right.tag)),
		compareIntDesc(left.tagPatch, right.tagPatch),
		compareIntAsc(left.mark, right.mark),
	} {
		if result != 0 {
			return result
		}
	}

	return 0
}

func versionTagOrder(tag string) int {
	switch tag {
	case "rc":
		return rcTagOrder
	case "beta":
		return betaTagOrder
	case "alpha":
		return alphaTagOrder
	default:
		return stableTagOrder
	}
}

func parseVersion(version string) parsedVersion {
	parsed := parsedVersion{tag: stableTag}
	matches := versionRegex.FindStringSubmatch(version)
	if len(matches) == 0 {
		return parsed
	}

	parsed.major = parseVersionNumber(matches[1])
	parsed.minor = parseVersionNumber(matches[2])
	parsed.patch = parseVersionNumber(matches[3])
	if matches[4] != "" {
		parsed.tag = matches[4]
		parsed.tagPatch = parseVersionNumber(matches[5])
	}
	if matches[6] != "" {
		parsed.mark = parseVersionNumber(matches[6])
	}

	return parsed
}

func parseVersionNumber(raw string) int {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}

	return value
}

func compareIntDesc(left, right int) int {
	switch {
	case left > right:
		return -1
	case left < right:
		return 1
	default:
		return 0
	}
}

func compareIntAsc(left, right int) int {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}
