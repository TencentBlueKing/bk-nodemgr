/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

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
}

var _ IFrontSetting = &FrontSetting{}

// FrontSetting front setting.
type FrontSetting struct {
	bkloginURL            string
	bkRequestIDHeaderKEy  string
	bkPassAnalyticsScript string
}

// Option front setting option.
type Option struct {
	BKLoginURL            string
	BKRequestIDHeaderKEy  string
	BKPassAnalyticsScript string
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

	setting := &FrontSetting{
		bkloginURL:            opt.BKLoginURL,
		bkRequestIDHeaderKEy:  opt.BKRequestIDHeaderKEy,
		bkPassAnalyticsScript: opt.BKPassAnalyticsScript,
	}

	return setting, nil
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
