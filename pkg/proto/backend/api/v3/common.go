/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package v3 defines the backend v3 protocols
package v3

import (
	"errors"
	"fmt"
	"time"

	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	// backendPagingListTimeout defines the timeout for paging list requests
	backendPagingListTimeout = 30 * time.Minute
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

func convPageToTypes(reqPage *Page, maxLimit int) (types.Page, error) {
	page := types.Page{}
	if reqPage != nil {
		page.Offset = int(reqPage.GetOffset())
		page.Limit = int(reqPage.GetLimit())
	}

	if page.Offset < 0 {
		page.Offset = 0
	}

	if page.Limit <= 0 || page.Limit > maxLimit {
		return page, fmt.Errorf("page.limit must be in (0, %d]", maxLimit)
	}

	if maxLimit <= 0 {
		page.Limit = 0
	}

	return page, nil
}

// formatRespSlice formats the response slice.
func formatRespSlice[T ~bool | ~string | ~int64](values []T) []T {
	if values == nil {
		return make([]T, 0)
	}

	return values
}

// convertTimeRangeToTypes convert time range to types with validation.
func convertTimeRangeToTypes(timeRange *TimeRange) (*types.TimeRange, error) {
	if timeRange == nil {
		return nil, nil
	}

	startTimestampSec := timeRange.GetStartTimestampSec()
	endTimestampSec := timeRange.GetEndTimestampSec()

	// Validate timestamp range: Unix timestamp should be non-negative for practical use cases
	// Negative timestamps represent dates before 1970-01-01, which are typically not expected
	if startTimestampSec < 0 || endTimestampSec < 0 {
		return nil, fmt.Errorf("invalid time range: timestamps must be non-negative, got start=%d end=%d", startTimestampSec, endTimestampSec)
	}

	startTime := time.Unix(startTimestampSec, 0)
	endTime := time.Unix(endTimestampSec, 0)

	// Validate that start time is not after end time
	if startTime.After(endTime) {
		return nil, fmt.Errorf("invalid time range: start time (%v) must be before or equal to end time (%v)", startTime, endTime)
	}

	return &types.TimeRange{
		StartTime: startTime,
		EndTime:   endTime,
	}, nil
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

	// Use convertTimeRangeToTypes to ensure timestamp validity
	validatedTimeRange, err := convertTimeRangeToTypes(timeRange)
	if err != nil {
		return err
	}

	timeDuration := validatedTimeRange.EndTime.Sub(validatedTimeRange.StartTime)

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
