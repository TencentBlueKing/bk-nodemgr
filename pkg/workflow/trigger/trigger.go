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

// Package trigger describes the trigger.
package trigger

import (
	"errors"
	"fmt"
	"time"
)

// Trigger defines the trigger, which will handle the workflow.
type Trigger struct {
	TriggerID       string
	Category        Category
	Metadata        Metadata
	Active          bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	LastTriggeredAt time.Time
}

// Validate validates the trigger.
func (t *Trigger) Validate() error {
	if t.TriggerID == "" {
		return errors.New("trigger_id should not be empty")
	}

	switch t.Category {
	case CategoryOnce:
		_, ok := t.Metadata.(*MetadataOnce)
		if !ok {
			return errors.New("metadata not match category once")
		}

	case CategoryOrdered:
		_, ok := t.Metadata.(*MetadataOrdered)
		if !ok {
			return errors.New("metadata not match category ordered")
		}

	case CategoryPeriodic:
		_, ok := t.Metadata.(*MetadataPeriodic)
		if !ok {
			return errors.New("metadata not match category periodic")
		}

	default:
		return fmt.Errorf("category should be one of %s, %s, %s",
			CategoryOnce, CategoryPeriodic, CategoryOrdered)
	}

	if err := t.Metadata.Validate(); err != nil {
		return err
	}

	return nil
}

// Category represents the category of a trigger.
type Category string

const (
	// CategoryOnce represents a trigger that is executed only once.
	CategoryOnce Category = "once"

	// CategoryPeriodic represents a trigger that is executed periodically.
	CategoryPeriodic Category = "periodic"

	// CategoryOrdered represents a trigger that is executed in order.
	CategoryOrdered Category = "ordered"
)

// CategoryListToStringList converts a category list to a string list.
func CategoryListToStringList(categories []Category) []string {
	data := make([]string, len(categories))
	for idx, category := range categories {
		data[idx] = string(category)
	}

	return data
}

// StringListToCategoryList converts a string list to a category list.
func StringListToCategoryList(categories []string) []Category {
	data := make([]Category, len(categories))
	for idx, category := range categories {
		data[idx] = Category(category)
	}

	return data
}
