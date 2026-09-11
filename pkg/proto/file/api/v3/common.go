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

// Package v3 defines the file v3 protocols
package v3

import (
	"fmt"
	"math"

	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate validates page settings.
func (x *Page) Validate() error {
	if x == nil {
		return fmt.Errorf("page is required")
	}
	if x.GetOffset() < 0 {
		return fmt.Errorf("offset must be non-negative")
	}
	if x.GetOffset() > math.MaxInt {
		return fmt.Errorf("offset exceeds maximum value")
	}
	if x.GetLimit() <= 0 {
		return fmt.Errorf("limit must be positive")
	}
	if x.GetLimit() > math.MaxInt {
		return fmt.Errorf("limit exceeds maximum value")
	}

	return nil
}

// ConvertToTypes converts page settings to domain types.
func (x *Page) ConvertToTypes() (types.Page, error) {
	if err := x.Validate(); err != nil {
		return types.Page{}, err
	}

	return types.Page{Offset: int(x.GetOffset()), Limit: int(x.GetLimit()), Sort: x.GetSort()}, nil
}

// ConvertFromTypes converts domain page settings to the protocol message.
func (x *Page) ConvertFromTypes(page types.Page) {
	x.Offset = int64(page.Offset)
	x.Limit = int64(page.Limit)
	x.Sort = page.Sort
}

// ConvertPlatformToTypes convert platform to types.
func ConvertPlatformToTypes(plat *Platform) platfmt.Platform {
	return platfmt.Platform{
		OS:   criteria.OSType(plat.GetOsType()),
		Arch: criteria.CPUArch(plat.GetCpuArch()),
	}
}

// ConvertPlatformFromTypes convert platform from types.
func ConvertPlatformFromTypes(plat platfmt.Platform) *Platform {
	return &Platform{
		OsType:  string(plat.OS),
		CpuArch: string(plat.Arch),
	}
}
