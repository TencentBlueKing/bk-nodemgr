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

package packageevent

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// OptFn provides filtering options.
type OptFn = base.OptFn

// WithGeneration provides filtering by generation.
func WithGeneration(gen ...types.Generation) OptFn {
	return base.WithValues(FieldKeyGeneration, types.GenerationListToInt64List(gen)...)
}

// WithoutGeneration provides filtering by generation.
func WithoutGeneration(gen ...types.Generation) OptFn {
	return base.WithoutValues(FieldKeyGeneration, types.GenerationListToInt64List(gen)...)
}

// WithReleaseType provides filtering by release type.
func WithReleaseType(releaseType ...types.ReleaseType) OptFn {
	return base.WithValues(FieldKeyReleaseType, types.ReleaseTypeListToStringList(releaseType)...)
}

// WithoutReleaseType provides filtering by release type.
func WithoutReleaseType(releaseType ...types.ReleaseType) OptFn {
	return base.WithoutValues(FieldKeyReleaseType, types.ReleaseTypeListToStringList(releaseType)...)
}

// WithEventType provides filtering by release type.
func WithEventType(eventType ...types.PackageEventType) OptFn {
	return base.WithValues(FieldKeyEventType, types.PackageEventTypeListToStringList(eventType)...)
}

// WithoutEventType provides filtering by release type.
func WithoutEventType(eventType ...types.PackageEventType) OptFn {
	return base.WithoutValues(FieldKeyEventType, types.PackageEventTypeListToStringList(eventType)...)
}

// WithVersion provides filtering by version.
func WithVersion(version ...string) OptFn {
	return base.WithValues(FieldKeyVersion, version...)
}

// WithoutVersion provides filtering by version.
func WithoutVersion(version ...string) OptFn {
	return base.WithoutValues(FieldKeyVersion, version...)
}

// WithCPUArch provides filtering by cpu arch.
func WithCPUArch(cpuArch ...criteria.CPUArch) OptFn {
	return base.WithValues(FieldKeyCPUArch, criteria.CPUArchListToStringList(cpuArch)...)
}

// WithoutCPUArch provides filtering by cpu arch.
func WithoutCPUArch(cpuArch ...criteria.CPUArch) OptFn {
	return base.WithoutValues(FieldKeyCPUArch, criteria.CPUArchListToStringList(cpuArch)...)
}

// WithOSType provides filtering by os type.
func WithOSType(osType ...criteria.OSType) OptFn {
	return base.WithValues(FieldKeyOSType, criteria.OSTypeListToStringList(osType)...)
}

// WithoutOSType provides filtering by os type.
func WithoutOSType(osType ...criteria.OSType) OptFn {
	return base.WithoutValues(FieldKeyOSType, criteria.OSTypeListToStringList(osType)...)
}

// WithOperator filters by operator.
func WithOperator(operator ...string) OptFn {
	return base.WithValues(FieldKeyOperator, operator...)
}

// WithoutOperator filters by operator.
func WithoutOperator(operator ...string) OptFn {
	return base.WithoutValues(FieldKeyOperator, operator...)
}

// WithOperateTimeRange filters by operate-time.
func WithOperateTimeRange(timeRange types.TimeRange) OptFn {
	return base.WithTimeRange(FieldKeyOperateTime, timeRange.StartTime, timeRange.EndTime)
}
