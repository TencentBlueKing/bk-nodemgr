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

// Package v3 defines the backend v3 protocols
package v3

import (
	"fmt"
	"math"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	// IAMCallbackMaxPageLimit is the maximum page limit for IAM callback APIs.
	// This prevents performance issues when querying large datasets.
	IAMCallbackMaxPageLimit = 1000

	// IAMCallbackDefaultPageLimit is the default page limit for IAM callback APIs.
	IAMCallbackDefaultPageLimit = 100

	// IAMV4CallbackMaxPageSize is the V4 list callback page size limit.
	IAMV4CallbackMaxPageSize = 1000
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

// Validate validates IAM V4 callback fields without applying V3 defaults.
func (x *IAMV4ResourceCallbackReq) Validate() error {
	if strings.TrimSpace(x.GetType()) == "" || strings.TrimSpace(x.GetMethod()) == "" {
		return fmt.Errorf("type and method are required")
	}
	fields := x.GetFilter().AsMap()
	switch x.GetMethod() {
	case "list_instance":
		if _, err := x.GetPage().ConvertToTypes(); err != nil {
			return err
		}
		if keyword, exists := fields["keyword"]; exists {
			if _, ok := keyword.(string); !ok {
				return fmt.Errorf("filter.keyword must be a string")
			}
		}
		if parent, exists := fields["parent"]; exists {
			object, ok := parent.(map[string]interface{})
			if !ok {
				return fmt.Errorf("filter.parent must be an object")
			}
			parentType, typeOK := object["type"].(string)
			parentID, idOK := object["id"].(string)
			if !typeOK || !idOK || parentType == "" || parentID == "" {
				return fmt.Errorf("filter.parent.type and filter.parent.id must be nonempty strings")
			}
		}
	case "fetch_instance_info":
		ids, ok := fields["ids"].([]interface{})
		if !ok {
			return fmt.Errorf("filter.ids must be a string array")
		}
		for _, id := range ids {
			if _, ok := id.(string); !ok {
				return fmt.Errorf("filter.ids must contain only strings")
			}
		}
	}

	return nil
}

// ConvertToTypes converts V4 page numbers to offsets without defaults.
func (x *IAMV4ResourceCallbackPage) ConvertToTypes() (types.Page, error) {
	page, pageSize := x.GetPage(), x.GetPageSize()
	if page < 1 || pageSize < 1 || pageSize > IAMV4CallbackMaxPageSize {
		return types.Page{}, fmt.Errorf("page must be >= 1 and page_size must be between 1 and %d", IAMV4CallbackMaxPageSize)
	}
	if page-1 > int64(math.MaxInt)/pageSize {
		return types.Page{}, fmt.Errorf("page offset overflows int")
	}

	return types.Page{Offset: int((page - 1) * pageSize), Limit: int(pageSize)}, nil
}
