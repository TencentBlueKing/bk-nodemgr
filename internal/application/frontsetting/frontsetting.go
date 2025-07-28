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

// notice: this file use interface to avoid this setting changed by other package.

// IFrontSetting front setting.
type IFrontSetting interface {
	// BKSharedResBaseJsURL the front setting field.
	BKSharedResBaseJsURL() string
	// BKLoginURL the front setting field.
	BKLoginURL() string
	// SiteUrl the front setting field.
	SiteUrl() string
}

// FrontSetting front setting.
type FrontSetting struct {
	bkloginURL           string
	bkSharedResBaseJsUrl string
	bkAPIPrefix          string
	siteURL              string
}

// NewFrontSetting new front setting.
func NewFrontSetting(bkloginURL string, bkSharedResBaseJsUrl string, siteURL string) *FrontSetting {
	return &FrontSetting{
		bkloginURL:           bkloginURL,
		bkSharedResBaseJsUrl: bkSharedResBaseJsUrl,
		siteURL:              siteURL,
	}
}

// BKSharedResBaseJsURL get bk shared res base js url.
func (f *FrontSetting) BKSharedResBaseJsURL() string {
	return f.bkSharedResBaseJsUrl
}

// BKLoginURL get bk login url.
func (f *FrontSetting) BKLoginURL() string {
	return f.bkloginURL
}

// SiteUrl get site url.
func (f *FrontSetting) SiteUrl() string {
	return f.siteURL
}
