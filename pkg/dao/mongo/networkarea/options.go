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

package networkarea

import "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"

// OptFn provides filtering options.
type OptFn = base.OptFn

// WithNetworkAreaID filters by networkarea-id.
func WithNetworkAreaID(networkAreaIDs ...int64) OptFn {
	return base.WithValues(FieldKeyNetworkAreaID, networkAreaIDs...)
}

// WithoutNetworkAreaID filters by not contains networkarea-id.
func WithoutNetworkAreaID(networkAreaIDs ...int64) OptFn {
	return base.WithoutValues(FieldKeyNetworkAreaID, networkAreaIDs...)
}

// WithCloudVendor filters by cloud_vendor.
func WithCloudVendor(cloudVendors ...string) OptFn {
	return base.WithStringValues(FieldKeyCloudVendor, cloudVendors...)
}

// WithoutCloudVendor filters by not contains cloud_vendor.
func WithoutCloudVendor(cloudVendors ...string) OptFn {
	return base.WithoutStringValues(FieldKeyCloudVendor, cloudVendors...)
}

// WithFuzzyNetworkAreaName filters by networkarea-name.
func WithFuzzyNetworkAreaName(networkAreaNames ...string) OptFn {
	return base.WithFuzzyValues(FieldKeyNetworkAreaName, networkAreaNames...)
}

// WithoutFuzzyNetworkAreaName filters by not contains networkarea-name.
func WithoutFuzzyNetworkAreaName(networkAreaNames ...string) OptFn {
	return base.WithoutFuzzyValues(FieldKeyNetworkAreaName, networkAreaNames...)
}
