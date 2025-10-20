/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package v3 defines the application v3 protocols
package v3

import (
	"errors"
	"fmt"
	"time"

	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func validatePage(reqPage *Page) error {
	if reqPage == nil {
		return nil
	}

	if reqPage.GetOffset() < 0 {
		return errors.New("\"page.offset\" field must be >= 0")
	}

	if reqPage.GetLimit() < 0 {
		return errors.New("\"page.limit\" field must be >= 0")
	}

	return nil
}

// generatePage generates list page.
func generatePage(reqPage *Page, maxLimit int) types.Page {
	page := types.Page{}
	if reqPage != nil {
		page.Offset = int(reqPage.GetOffset())
		page.Limit = int(reqPage.GetLimit())
	}

	if page.Offset < 0 {
		page.Offset = 0
	}

	if page.Limit <= 0 || page.Limit > maxLimit {
		page.Limit = maxLimit
	}

	if maxLimit <= 0 {
		page.Limit = 0
	}

	return page
}

// formatRespSlice formats the response slice.
func formatRespSlice[T bool | string | int64](values []T) []T {
	if values == nil {
		return make([]T, 0)
	}

	return values
}

// validateTimeRange validates the time range.
func validateTimeRange(timeRange *TimeRange, maxDuration time.Duration) error {
	// no limit.
	if maxDuration <= 0 {
		return nil
	}

	if timeRange == nil {
		return nil
	}

	timeDuration := time.Unix(timeRange.GetEndTimestampSec(), 0).
		Sub(time.Unix(timeRange.GetStartTimestampSec(), 0))

	if timeDuration > maxDuration {
		return fmt.Errorf("time range %s is too long, max allowed is %s", timeDuration.String(), maxDuration.String())
	}

	return nil
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
