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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	// IAMCallbackMaxPageLimit is the maximum page limit for IAM callback APIs.
	// This prevents performance issues when querying large datasets.
	IAMCallbackMaxPageLimit = 1000

	// IAMCallbackDefaultPageLimit is the default page limit for IAM callback APIs.
	IAMCallbackDefaultPageLimit = 100
)

// ValidateIAMCallbackPage validates IAM callback page parameters.
func ValidateIAMCallbackPage(page *IAMResourceCallbackPage) error {
	if page == nil {
		return nil
	}

	if page.GetOffset() < 0 {
		return fmt.Errorf("\"page.offset\" field must be >= 0, got %d", page.GetOffset())
	}

	if page.GetLimit() < 0 {
		return fmt.Errorf("\"page.limit\" field must be >= 0, got %d", page.GetLimit())
	}

	return nil
}

// ConvIAMCallbackPageToTypes converts IAM callback page to types.Page with validation and normalization.
// It enforces max limit to prevent performance issues.
func ConvIAMCallbackPageToTypes(reqPage *IAMResourceCallbackPage) (types.Page, error) {
	page := types.Page{}

	if reqPage != nil {
		page.Offset = int(reqPage.GetOffset())
		page.Limit = int(reqPage.GetLimit())
	}

	// Normalize offset
	if page.Offset < 0 {
		page.Offset = 0
	}

	// Normalize limit
	if page.Limit <= 0 {
		page.Limit = IAMCallbackDefaultPageLimit
	}

	// Enforce max limit
	if page.Limit > IAMCallbackMaxPageLimit {
		return page, fmt.Errorf("page.limit must not exceed %d, got %d", IAMCallbackMaxPageLimit, page.Limit)
	}

	return page, nil
}
