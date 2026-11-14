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

package wiki

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/version"
)

// GitHubConfig configures the repository used for product Wiki links.
type GitHubConfig struct {
	RepositoryURL string
}

// Validate checks the repository URL without network access.
func (conf GitHubConfig) Validate() error {
	repository, err := url.Parse(conf.RepositoryURL)
	if err != nil {
		return fmt.Errorf("invalid wiki github repository url: %w", err)
	}
	if (repository.Scheme != "https" && repository.Scheme != "http") || repository.Host == "" ||
		repository.User != nil || repository.RawQuery != "" || repository.Fragment != "" ||
		len(strings.Split(strings.Trim(repository.Path, "/"), "/")) != 2 {

		return fmt.Errorf("wiki github repository url must be an http(s) repository address without credentials, query or fragment")
	}

	return nil
}

var _ IHandler = (*githubHandler)(nil)

type githubHandler struct {
	prefix string
}

func newGitHubHandler(conf GitHubConfig) *githubHandler {
	tag := "master"
	if release, ok := version.ParseRelease(version.VERSION); ok {
		tag = fmt.Sprintf("v%s.%s.%s-%s.%s", release.Major, release.Minor, release.Patch, release.Phase, release.Serial)
	}

	return &githubHandler{
		prefix: strings.TrimRight(conf.RepositoryURL, "/") + "/blob/" + tag + "/support-files/wiki/",
	}
}

// URL returns a localized link under the release tag or master for non-release builds.
func (h *githubHandler) URL(document Document, language Language) string {
	return h.prefix + string(language) + "/" + string(document)
}
