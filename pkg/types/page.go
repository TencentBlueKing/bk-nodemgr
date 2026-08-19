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

// Package types ...
package types

import (
	"fmt"
	"math"
	"strings"
)

// SortSeparator is the separator for sort keys.
const SortSeparator = ","

// Page describe the page data in request.
type Page struct {
	// Offset is the offset of the query result.
	Offset int

	// Limit is the size of the query result.
	Limit int

	// Sort is the field used to sort the query result.
	Sort string
}

// Validate Page.
func (p *Page) Validate() error {
	if p.Offset < 0 {
		return fmt.Errorf("offset must be non-negative")
	}

	if p.Limit <= 0 {
		return fmt.Errorf("limit must be positive")
	}

	return nil
}

// WithFieldDesc returns the field with descending order.
func WithFieldDesc(field string) string {
	return "-" + field
}

// WithFieldAsc returns the field with ascending order.
func WithFieldAsc(field string) string {
	return field
}

// WithSortFields returns the sort fields as a comma-separated string.
func WithSortFields(fields ...string) string {
	validFields := make([]string, 0, len(fields))

	for _, field := range fields {
		if field != "" {
			validFields = append(validFields, field)
		}
	}

	return strings.Join(validFields, SortSeparator)
}

// UnlimitedPage is an unlimited page.
func UnlimitedPage() Page {
	return Page{
		Offset: 0,
		Limit:  math.MaxInt32,
	}
}

// SingleItemPage is a single item page.
func SingleItemPage() Page {
	return Page{
		Offset: 0,
		Limit:  1,
	}
}
