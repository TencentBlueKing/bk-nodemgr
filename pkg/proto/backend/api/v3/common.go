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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func validateTopoPage(reqPage *Page) error {
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
