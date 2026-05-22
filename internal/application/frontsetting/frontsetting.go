/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package frontsetting provides the frontend settings.
package frontsetting

import (
	"fmt"
	"html/template"
	"net/url"
	"strings"
)

type frontValueKind int

const (
	frontValueKindURL frontValueKind = iota
	frontValueKindURI
	frontValueKindHost
)

// notice: this file use interface to avoid this setting changed by other package.

// IFrontSetting front setting.
type IFrontSetting interface {
	// BKLoginURL the front setting field.
	BKLoginURL() string

	// BKRequestIDHeaderKey the front setting field.
	BKRequestIDHeaderKey() string

	// BKPassAnalyticsScript the front setting field.
	BKPassAnalyticsScript() template.HTML

	// PasswordVaultSwitch the front setting field.
	PasswordVaultSwitch() bool

	// PasswordVaultName the front setting field.
	PasswordVaultName() string

	// BKUserWebURL the front setting field.
	BKUserWebURL() string

	// BKIamSaaSHost the front setting field.
	BKIamSaaSHost() string

	// BKUserSaaSHost the front setting field.
	BKUserSaaSHost() string

	// BKDomain the front setting field.
	BKDomain() string

	// BKDocsCenterURL the front setting field.
	BKDocsCenterURL() string

	// BKAppNavOpenSourceURL the front setting field.
	BKAppNavOpenSourceURL() string

	// WindowsWMIPortDefault the front setting field.
	WindowsWMIPortDefault() int

	// UnixSSHPortDefault the front setting field.
	UnixSSHPortDefault() int

	// EnableNotice the front setting field.
	EnableNotice() bool

	// BKIamSystemIDBKNodemgr the front setting field.
	BKIamSystemIDBKNodemgr() string

	// BKIamSystemIDBKCmdb the front setting field.
	BKIamSystemIDBKCmdb() string
}

var _ IFrontSetting = &FrontSetting{}

// FrontSetting front setting.
type FrontSetting struct {
	bkloginURL            string
	bkRequestIDHeaderKEy  string
	bkPassAnalyticsScript string

	passwordVaultSwitch bool
	passwordVaultName   string

	bkUserWebURL   string
	bkIamSaaSHost  string
	bkUserSaaSHost string

	bkDomain              string
	bkDocsCenterURL       string
	bkAppNavOpenSourceURL string

	windowsWMIPortDefault int
	unixSSHPortDefault    int

	enableNotice bool

	bkIamSystemIDBKNodemgr string
	bkIamSystemIDBKCmdb    string
}

// Option front setting option.
type Option struct {
	BKLoginURL            string
	BKRequestIDHeaderKEy  string
	BKPassAnalyticsScript string

	// PasswordVault Options.
	PasswordVaultSwitch bool
	PasswordVaultName   string

	// BKUserWebURL is the URL of bk user web service.
	BKUserWebURL string

	// BKIamSaaSHost is the host of bk iam saas.
	BKIamSaaSHost string

	// BKUserSaaSHost is the host of bk user saas.
	BKUserSaaSHost string

	// BKDomain is the domain of bk platform.
	BKDomain string

	// BKDocsCenterURL is the URL of bk docs center.
	BKDocsCenterURL string

	// BKAppNavOpenSourceURL is the URL of bk app nav open source.
	BKAppNavOpenSourceURL string

	// WindowsWMIPortDefault is the default port for Windows WMI connection.
	WindowsWMIPortDefault int

	// UnixSSHPortDefault is the default port for Unix-like OS (Linux, AIX, Darwin, etc.) SSH connection.
	UnixSSHPortDefault int

	// EnableNotice controls whether the notice feature is enabled.
	EnableNotice bool

	// BKIamSystemIDBKNodemgr is the bk iam system id.
	BKIamSystemIDBKNodemgr string

	// BKIamSystemIDBKCmdb is the bk iam system id for cmdb.
	BKIamSystemIDBKCmdb string
}

// Validate validate.
func (option *Option) Validate() error {
	if option.BKLoginURL == "" {
		return fmt.Errorf("bk login url is empty")
	}

	if option.BKRequestIDHeaderKEy == "" {
		return fmt.Errorf("bk request id header key is empty")
	}

	return nil
}

// NewFrontSetting new front setting.
func NewFrontSetting(opt Option) (*FrontSetting, error) {
	if err := opt.Validate(); err != nil {
		return nil, fmt.Errorf("failed to new front setting: %v", err)
	}

	return &FrontSetting{
		bkloginURL:             normalizeFrontValue(opt.BKLoginURL, frontValueKindURL),
		bkRequestIDHeaderKEy:   opt.BKRequestIDHeaderKEy,
		bkPassAnalyticsScript:  opt.BKPassAnalyticsScript,
		passwordVaultSwitch:    opt.PasswordVaultSwitch,
		passwordVaultName:      opt.PasswordVaultName,
		bkUserWebURL:           normalizeFrontValue(opt.BKUserWebURL, frontValueKindURL),
		bkIamSaaSHost:          opt.BKIamSaaSHost,
		bkUserSaaSHost:         opt.BKUserSaaSHost,
		bkDomain:               normalizeFrontValue(opt.BKDomain, frontValueKindHost),
		bkDocsCenterURL:        normalizeFrontValue(opt.BKDocsCenterURL, frontValueKindURL),
		bkAppNavOpenSourceURL:  normalizeFrontValue(opt.BKAppNavOpenSourceURL, frontValueKindURL),
		windowsWMIPortDefault:  opt.WindowsWMIPortDefault,
		unixSSHPortDefault:     opt.UnixSSHPortDefault,
		enableNotice:           opt.EnableNotice,
		bkIamSystemIDBKNodemgr: opt.BKIamSystemIDBKNodemgr,
		bkIamSystemIDBKCmdb:    opt.BKIamSystemIDBKCmdb,
	}, nil
}

func normalizeFrontValue(raw string, kind frontValueKind) string {
	switch kind {
	case frontValueKindURL:
		return normalizeURL(raw)
	case frontValueKindURI:
		return normalizeURI(raw)
	case frontValueKindHost:
		return normalizeHost(raw)
	default:
		return strings.TrimRight(raw, "/")
	}
}

func normalizeURL(raw string) string {
	if strings.HasSuffix(raw, "://") {
		return raw
	}

	parsedURL, err := url.Parse(raw)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return strings.TrimRight(raw, "/")
	}

	parsedURL.Path = strings.TrimRight(parsedURL.Path, "/")
	parsedURL.RawPath = strings.TrimRight(parsedURL.RawPath, "/")

	return parsedURL.String()
}

func normalizeURI(raw string) string {
	if raw == "/" {
		return raw
	}

	return strings.TrimRight(raw, "/")
}

func normalizeHost(raw string) string {
	host := raw
	if strings.Contains(host, "://") {
		parsedURL, err := url.Parse(host)
		if err == nil && parsedURL.Host != "" {
			host = parsedURL.Host
		}
	}

	if slashIndex := strings.Index(host, "/"); slashIndex >= 0 {
		host = host[:slashIndex]
	}

	return strings.TrimRight(host, "/")
}

// BKLoginURL get bk login url.
func (setting *FrontSetting) BKLoginURL() string {
	return setting.bkloginURL
}

// BKRequestIDHeaderKey get bk request id header key.
func (setting *FrontSetting) BKRequestIDHeaderKey() string {
	return setting.bkRequestIDHeaderKEy
}

// BKPassAnalyticsScript get bk pass analytics script.
func (setting *FrontSetting) BKPassAnalyticsScript() template.HTML {
	return template.HTML(setting.bkPassAnalyticsScript)
}

// PasswordVaultSwitch get password vault switch.
func (setting *FrontSetting) PasswordVaultSwitch() bool {
	return setting.passwordVaultSwitch
}

// PasswordVaultName get password vault name.
func (setting *FrontSetting) PasswordVaultName() string {
	return setting.passwordVaultName
}

// BKUserWebURL get bk user web url.
func (setting *FrontSetting) BKUserWebURL() string {
	return setting.bkUserWebURL
}

// BKIamSaaSHost get bk iam saas host.
func (setting *FrontSetting) BKIamSaaSHost() string {
	return setting.bkIamSaaSHost
}

// BKUserSaaSHost get bk user saas host.
func (setting *FrontSetting) BKUserSaaSHost() string {
	return setting.bkUserSaaSHost
}

// BKDomain get bk domain.
func (setting *FrontSetting) BKDomain() string {
	return setting.bkDomain
}

// BKDocsCenterURL get bk docs center url.
func (setting *FrontSetting) BKDocsCenterURL() string {
	return setting.bkDocsCenterURL
}

// BKAppNavOpenSourceURL get bk app nav open source url.
func (setting *FrontSetting) BKAppNavOpenSourceURL() string {
	return setting.bkAppNavOpenSourceURL
}

// EnableNotice get enable notice.
func (setting *FrontSetting) EnableNotice() bool {
	return setting.enableNotice
}

// WindowsWMIPortDefault get windows WMI port default.
func (setting *FrontSetting) WindowsWMIPortDefault() int {
	return setting.windowsWMIPortDefault
}

// UnixSSHPortDefault get Unix-like OS SSH port default.
func (setting *FrontSetting) UnixSSHPortDefault() int {
	return setting.unixSSHPortDefault
}

// BKIamSystemIDBKNodemgr get bk_nodemgr system id in bkiam.
func (setting *FrontSetting) BKIamSystemIDBKNodemgr() string {
	return setting.bkIamSystemIDBKNodemgr
}

// BKIamSystemIDBKCmdb get bk_cmdb system id in bkiam.
func (setting *FrontSetting) BKIamSystemIDBKCmdb() string {
	return setting.bkIamSystemIDBKCmdb
}
