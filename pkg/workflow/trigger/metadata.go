/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package trigger

import (
	"errors"
	"time"
)

// Metadata defines the trigger metadata.
type Metadata interface {
	// Validate validates the metadata.
	Validate() error
}

// MetadataOnce will store the metadata of once trigger.
type MetadataOnce struct {
}

// Validate validates the metadata.
func (m *MetadataOnce) Validate() error {
	return nil
}

// MetadataOrdered will store the metadata of ordered trigger.
type MetadataOrdered struct {
	MaxConcurrencyNum int
}

// Validate validates the metadata.
func (m *MetadataOrdered) Validate() error {
	return nil
}

// MetadataPeriodic will store the metadata of periodic trigger.
type MetadataPeriodic struct {
	Interval           time.Duration
	AllowedConcurrency bool
}

// Validate validates the metadata.
func (m *MetadataPeriodic) Validate() error {
	if m.Interval <= 0 {
		return errors.New("interval must be greater than 0")
	}

	return nil
}
