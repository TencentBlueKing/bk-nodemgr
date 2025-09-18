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
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/scheduler"
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
	Interval           string
	AllowedConcurrency bool
}

// Validate validates the metadata.
func (m *MetadataPeriodic) Validate() error {
	_, err := scheduler.CronParser().Parse(m.Interval)
	if err != nil {
		return fmt.Errorf("invalid interval: %w", err)
	}

	return nil
}

// NewMetadataPeriodic creates a new MetadataPeriodic instance.
// The interval can be a time.Duration or a cron expression string.
// If a time.Duration is provided, it will be converted to a cron expression using "@every <duration>" format.
func NewMetadataPeriodic[T time.Duration | string](interval T, allowedConcurrency bool) (*MetadataPeriodic, error) {
	var cronExpr string
	switch t := any(interval).(type) {
	case time.Duration:
		cronExpr = scheduler.Every + t.String()
	case string:
		cronExpr = t
	}

	meta := &MetadataPeriodic{
		Interval:           cronExpr,
		AllowedConcurrency: allowedConcurrency,
	}

	if err := meta.Validate(); err != nil {
		return nil, fmt.Errorf("invalid metadata periodic: %w", err)
	}

	return meta, nil
}
