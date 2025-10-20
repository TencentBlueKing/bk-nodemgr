/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package release

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
)

// OptFn provides filtering options.
type OptFn = base.OptFn

// WithName provides filtering by name.
func WithName(name ...string) OptFn {
	return base.WithValues(FieldKeyName, name...)
}

// WithGeneration provides filtering by generation.
func WithGeneration(gen ...types.Generation) OptFn {
	return base.WithValues(FieldKeyGeneration, types.GenerationListToInt64List(gen)...)
}

// WithType provides filtering by release type.
func WithType(releaseType ...types.ReleaseType) OptFn {
	return base.WithValues(FieldKeyType, types.ReleaseTypeListToStringList(releaseType)...)
}

// WithVersion provides filtering by version.
func WithVersion(version ...string) OptFn {
	return base.WithValues(FieldKeyVersion, version...)
}

// WithCPUArch provides filtering by cpu arch.
func WithCPUArch(cpuArch ...string) OptFn {
	return base.WithValues(FieldKeyCPUArch, cpuArch...)
}

// WithOSType provides filtering by os type.
func WithOSType(osType ...string) OptFn {
	return base.WithValues(FieldKeyOSType, osType...)
}

// WithAsDefault provides filtering by as default.
func WithAsDefault(asDefault ...bool) OptFn {
	return base.WithValues(FieldKeyAsDefault, asDefault...)
}

// WithEnabled provides filtering by enabled.
func WithEnabled(enabled ...bool) OptFn {
	return base.WithValues(FieldKeyEnabled, enabled...)
}

// WithFileName provides filtering by file name.
func WithFileName(filename ...string) OptFn {
	return base.WithValues(FieldKeyFileName, filename...)
}

// WithPlatform provides filtering by platform.
func WithPlatform(platform ...platfmt.Platform) OptFn {
	if len(platform) == 0 {
		return func(f bson.D) bson.D {
			return f
		}
	}

	if len(platform) == 1 {
		return func(f bson.D) bson.D {
			return append(f,
				bson.E{Key: FieldKeyCPUArch, Value: platform[0].Arch},
				bson.E{Key: FieldKeyOSType, Value: platform[0].OS})
		}
	}

	values := make(bson.A, 0)
	for _, plat := range platform {
		values = append(values, bson.D{
			{Key: FieldKeyCPUArch, Value: plat.Arch},
			{Key: FieldKeyOSType, Value: plat.OS},
		})
	}

	return func(f bson.D) bson.D {
		return append(f, bson.E{
			Key:   "$or",
			Value: values,
		})
	}
}
