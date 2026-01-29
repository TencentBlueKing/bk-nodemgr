/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package cmdb provides mock CMDB server types.
package cmdb

import "fmt"

// CMDB-standard error codes. Reference: CMDB API error codes.
const (
	// CodeOK is the success code.
	CodeOK = 0

	// CodeInvalidParameter is the invalid parameter code.
	CodeInvalidParameter = 1199000

	// CodeServerError is the server internal error code.
	CodeServerError = 1199001
)

// Config holds the configuration for CMDB mock data.
type Config struct {
	// Businesses is the list of mock business data.
	Businesses []BusinessConfig `yaml:"businesses"`
	// TODO: add more data here. such as areas, hosts, etc.
}

// BusinessConfig holds the configuration for a single business.
type BusinessConfig struct {
	// BKBizID is the business ID.
	BKBizID int64 `yaml:"bk_biz_id"`
	// BKBizName is the business name.
	BKBizName string `yaml:"bk_biz_name"`
}

// Validate validates the BusinessConfig.
func (b *BusinessConfig) Validate() error {
	if b.BKBizID < 0 {
		return fmt.Errorf("bk_biz_id must be >= 0, bk_biz_id(%d)", b.BKBizID)
	}
	if b.BKBizName == "" {
		return fmt.Errorf("bk_biz_name cannot be empty")
	}

	return nil
}

// Validate validates the Config.
func (c *Config) Validate() error {
	for _, biz := range c.Businesses {
		if err := biz.Validate(); err != nil {
			return err
		}
	}

	return nil
}
