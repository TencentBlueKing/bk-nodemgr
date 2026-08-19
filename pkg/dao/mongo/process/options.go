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

package process

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// OptFn provides filtering options.
type OptFn = base.OptFn

// WithHostID filters by host-id.
func WithHostID(hostIDs ...int64) OptFn {
	return base.WithValues(FieldKeyHostID, hostIDs...)
}

// WithBizID filters by biz-id.
func WithBizID(bizIDs ...int64) OptFn {
	return base.WithValues(FieldKeyBizID, bizIDs...)
}

// WithPluginName filters by name.
func WithPluginName(pluginNames ...string) OptFn {
	return base.WithValues(FieldKeyPluginName, pluginNames...)
}

// WithFuzzyName filters by fuzzy name.
func WithFuzzyName(names ...string) OptFn {
	return base.WithFuzzyValues(FieldKeyPluginName, names...)
}

// WithGroup filters by group.
func WithGroup(groups ...string) OptFn {
	return base.WithValues(FieldKeyGroup, groups...)
}

// WithPkgName filters by package name.
func WithPkgName(pkgNames ...string) OptFn {
	return base.WithValues(FieldKeyPkgName, pkgNames...)
}

// WithFuzzyPkgName filters by fuzzy package name.
func WithFuzzyPkgName(pkgNames ...string) OptFn {
	return base.WithFuzzyValues(FieldKeyPkgName, pkgNames...)
}

// WithGeneration filters by generation.
func WithGeneration(generations ...int64) OptFn {
	return base.WithValues(FieldKeyGeneration, generations...)
}

// WithPlatformOS filters by platform OS.
func WithPlatformOS(osList ...string) OptFn {
	return base.WithValues(FieldKeyPlatformOS, osList...)
}

// WithPlatformArch filters by platform architecture.
func WithPlatformArch(archs ...string) OptFn {
	return base.WithValues(FieldKeyPlatformArch, archs...)
}

// WithInfoStatus filters by info status.
func WithInfoStatus(statuses ...types.ProcessStatus) OptFn {
	return base.WithValues(FieldKeyInfoStatus, statuses...)
}

// WithInfoAgentID filters by info agent ID.
func WithInfoAgentID(agentIDs ...string) OptFn {
	return base.WithValues(FieldKeyInfoAgentID, agentIDs...)
}

// WithInfoVersion filters by info version.
func WithInfoVersion(versions ...string) OptFn {
	return base.WithValues(FieldKeyInfoVersion, versions...)
}
