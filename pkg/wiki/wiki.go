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
	"sync"
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

// IURL provides a document URL for every supported Wiki language.
// Adding a language requires every implementation to provide its URL method.
type IURL interface {
	// Zh returns the simplified Chinese document URL.
	Zh() string
	// En returns the English document URL.
	En() string
}

var _ IURL = Document("")

// Zh returns the configured simplified Chinese URL, or an empty string before initialization.
func (document Document) Zh() string {
	return URL(document, LanguageZh)
}

// En returns the configured English URL, or an empty string before initialization.
func (document Document) En() string {
	return URL(document, LanguageEn)
}

// UnixSSHConnectionFailed explains how to troubleshoot Unix SSH connection failures.
const UnixSSHConnectionFailed Document = "troubleshooting/unix-ssh-connection-failed.md"

// Config defines the inputs used to build product Wiki links.
type Config struct {
	Mode   Mode
	GitHub GitHubConfig
}

// Validate checks the supported mode and its source-specific inputs without network access.
func (conf Config) Validate() error {
	switch conf.Mode {
	case ModeGitHub:
		return conf.GitHub.Validate()
	default:
		return fmt.Errorf("unsupported wiki mode: %s", conf.Mode)
	}
}

// IHandler builds localized links to product Wiki documents.
type IHandler interface {
	// URL returns the link to a document in the requested language.
	URL(document Document, language Language) string
}

// New creates a Wiki handler for the configured source.
func New(conf Config) (IHandler, error) {
	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("failed to create wiki handler: %w", err)
	}

	switch conf.Mode {
	case ModeGitHub:
		return newGitHubHandler(conf.GitHub), nil
	default:
		return nil, fmt.Errorf("unsupported wiki mode: %s", conf.Mode)
	}
}

// nolint: gochecknoglobals
var linkConfig struct {
	sync.Once
	handler IHandler
}

// SetConfig initializes Wiki links once, before serving requests or starting workers.
// Release builds use their parsed release tag; other versions use master.
func SetConfig(conf Config) error {
	handler, err := New(conf)
	if err != nil {
		return fmt.Errorf("failed to set wiki config: %w", err)
	}

	linkConfig.Do(func() {
		linkConfig.handler = handler
	})

	return nil
}

// URL returns the configured link to a document.
// An empty result means Wiki configuration has not been initialized.
func URL(document Document, language Language) string {
	if linkConfig.handler == nil {
		return ""
	}

	return linkConfig.handler.URL(document, language)
}
