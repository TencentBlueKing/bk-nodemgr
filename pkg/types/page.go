/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package types ...
package types

import (
	"fmt"
	"math"
)

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

// UnlimitedPage is an unlimited page.
func UnlimitedPage() Page {
	return Page{
		Offset: 0,
		Limit:  math.MaxInt32,
	}
}
