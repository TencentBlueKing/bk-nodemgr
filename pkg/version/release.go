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

package version

import "regexp"

const (
	// PhaseAlpha identifies the alpha release phase.
	PhaseAlpha = "alpha"
	// PhaseBeta identifies the beta release phase.
	PhaseBeta = "beta"
	// PhaseRC identifies the release candidate phase.
	PhaseRC = "rc"
	// PhaseRelease identifies the release phase.
	PhaseRelease = "release"
)

var releaseVersionRegex = regexp.MustCompile(
	`^v(\d+)\.(\d+)\.(\d+)-([a-zA-Z]+)\.(\d+)(?:-([a-zA-Z0-9][a-zA-Z0-9.-]*))?$`,
)

// ReleaseVersion describes the components defined by the project's release version format.
// Numeric components remain strings so parsing does not impose machine integer limits.
type ReleaseVersion struct {
	Major  string
	Minor  string
	Patch  string
	Phase  string
	Serial string
	Mark   string
}

// ParseRelease parses v{major}.{minor}.{patch}-{phase}.{serial}[-{mark}].
// It returns false for development identifiers or other non-release formats.
func ParseRelease(raw string) (ReleaseVersion, bool) {
	matches := releaseVersionRegex.FindStringSubmatch(raw)
	if len(matches) == 0 {
		return ReleaseVersion{}, false
	}

	return ReleaseVersion{
		Major:  matches[1],
		Minor:  matches[2],
		Patch:  matches[3],
		Phase:  matches[4],
		Serial: matches[5],
		Mark:   matches[6],
	}, true
}
