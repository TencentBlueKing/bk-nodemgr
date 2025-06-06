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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// OptFn provides filtering options.
type OptFn = base.OptFn

// WithGeneration provides filtering by generation.
func WithGeneration(gen ...types.Generation) OptFn {
	return base.WithValues(FieldKeyGeneration, func() []int64 {
		res := make([]int64, len(gen))
		for idx, v := range gen {
			res[idx] = int64(v)
		}

		return res
	}()...)
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
