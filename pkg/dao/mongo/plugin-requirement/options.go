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

package pluginrequirement

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
)

// OptFn provides requirement metadata filtering options.
type OptFn = base.OptFn

// WithRequirementFileID filters by stable file record IDs.
func WithRequirementFileID(id ...string) OptFn {
	return base.WithValues(FieldKeyRequirementFileID, id...)
}

// WithPluginName filters by exact owning plugin names.
func WithPluginName(name ...string) OptFn {
	return base.WithValues(FieldKeyPluginName, name...)
}

// WithName filters by exact business dependency names.
func WithName(name ...string) OptFn {
	return base.WithValues(FieldKeyName, name...)
}

// WithTargetFileName filters by exact downlink filenames.
func WithTargetFileName(name ...string) OptFn {
	return base.WithValues(FieldKeyTargetFileName, name...)
}

// WithOS filters stored OS values by exact administrative intersection.
func WithOS(os ...criteria.OSType) OptFn {
	return base.WithValues(FieldKeyOS, os...)
}

// WithArch filters stored architecture values by exact administrative intersection.
func WithArch(arch ...criteria.CPUArch) OptFn {
	return base.WithValues(FieldKeyArch, arch...)
}

// WithVersion filters stored package versions by exact administrative intersection.
func WithVersion(version ...string) OptFn {
	return base.WithValues(FieldKeyVersion, version...)
}
