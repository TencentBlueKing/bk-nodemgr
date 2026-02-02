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

	// BKDomain the front setting field.
	BKDomain() string

	// BKDocsCenterURL the front setting field.
	BKDocsCenterURL() string

	// BKAppNavOpenSourceURL the front setting field.
	BKAppNavOpenSourceURL() string

	// EnableNotice the front setting field.
	EnableNotice() bool
}

var _ IFrontSetting = &FrontSetting{}

// FrontSetting front setting.
type FrontSetting struct {
	bkloginURL            string
	bkRequestIDHeaderKEy  string
	bkPassAnalyticsScript string

	passwordVaultSwitch bool
	passwordVaultName   string

	bkUserWebURL string

	bkDomain              string
	bkDocsCenterURL       string
	bkAppNavOpenSourceURL string

	enableNotice bool
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

	// BKDomain is the domain of bk platform.
	BKDomain string

	// BKDocsCenterURL is the URL of bk docs center.
	BKDocsCenterURL string

	// BKAppNavOpenSourceURL is the URL of bk app nav open source.
	BKAppNavOpenSourceURL string

	// EnableNotice controls whether the notice feature is enabled.
	EnableNotice bool
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
		bkloginURL:            opt.BKLoginURL,
		bkRequestIDHeaderKEy:  opt.BKRequestIDHeaderKEy,
		bkPassAnalyticsScript: opt.BKPassAnalyticsScript,
		passwordVaultSwitch:   opt.PasswordVaultSwitch,
		passwordVaultName:     opt.PasswordVaultName,
		bkUserWebURL:          opt.BKUserWebURL,
		bkDomain:              opt.BKDomain,
		bkDocsCenterURL:       opt.BKDocsCenterURL,
		bkAppNavOpenSourceURL: opt.BKAppNavOpenSourceURL,
		enableNotice:          opt.EnableNotice,
	}, nil
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
