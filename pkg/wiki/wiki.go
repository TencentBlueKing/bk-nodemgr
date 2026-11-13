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

// Package wiki provides version-aware links to the product user Wiki.
package wiki

import (
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/version"
)

// Mode identifies the source of product user documentation.
type Mode string

// ModeGitHub serves documentation from the configured GitHub repository.
const ModeGitHub Mode = "github"

// Language identifies the language directory of a product Wiki document.
type Language string

const (
	// LanguageZh selects simplified Chinese documentation.
	LanguageZh Language = "zh"
	// LanguageEn selects English documentation.
	LanguageEn Language = "en"
)

// Document is a path relative to a language directory under support-files/wiki/.
// Paths are shared link contracts and must remain stable after publication.
type Document string

// SSHConnectionFailed explains how to troubleshoot SSH connection failures.
const SSHConnectionFailed Document = "troubleshooting/ssh-connection-failed.md"

// Config defines the inputs used to build product Wiki links.
type Config struct {
	Mode                Mode
	GitHubRepositoryURL string
}

// Validate checks the supported mode and repository URL without network access.
func (conf Config) Validate() error {
	if conf.Mode != ModeGitHub {
		return fmt.Errorf("unsupported wiki mode: %s", conf.Mode)
	}

	repository, err := url.Parse(conf.GitHubRepositoryURL)
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

// nolint: gochecknoglobals
var linkConfig struct {
	sync.Once
	prefix string
}

// SetConfig initializes Wiki links once, before serving requests or starting workers.
// Release builds use their parsed release tag; other versions use master.
func SetConfig(conf Config) error {
	if err := conf.Validate(); err != nil {
		return fmt.Errorf("failed to set wiki config: %w", err)
	}

	tag := "master"
	if release, ok := version.ParseRelease(version.VERSION); ok {
		tag = fmt.Sprintf("v%s.%s.%s-%s.%s", release.Major, release.Minor, release.Patch, release.Phase, release.Serial)
	}

	linkConfig.Do(func() {
		linkConfig.prefix = strings.TrimRight(conf.GitHubRepositoryURL, "/") + "/blob/" + tag + "/support-files/wiki/"
	})

	return nil
}

// URL returns the configured link to a document.
// An empty result means Wiki configuration has not been initialized.
func URL(document Document, language Language) string {
	if linkConfig.prefix == "" {
		return ""
	}

	return linkConfig.prefix + string(language) + "/" + string(document)
}
