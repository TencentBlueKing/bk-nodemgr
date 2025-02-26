/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package trigengine means trigger engine...
package trigengine

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	// TriggerPrefix ...
	TriggerPrefix = "T"
)

// Trigger ...
type Trigger struct {
	TriggerID        string
	Category         Category
	MetadataOnce     MetadataOnce
	MetadataPeriodic MetadataPeriodic
	MetadataOrdered  MetadataOrdered
	State            State
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Validate ...
func (t *Trigger) Validate() error {
	if t.TriggerID == "" {
		return errors.New("trigger id is empty")
	}

	switch t.State {
	case StateInit, StateRunning, StateTerminated:
	default:
		return fmt.Errorf("unknown state, state(%s)", t.State)
	}

	switch t.Category {
	case CategoryOnce:
		if err := t.MetadataOnce.Validate(); err != nil {
			return err
		}
	case CategoryPeriodic:
		if err := t.MetadataPeriodic.Validate(); err != nil {
			return err
		}
	case CategoryOrdered:
		if err := t.MetadataOrdered.Validate(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown category, category(%s)", t.Category)
	}

	return nil
}

// CheckFeasibility ...
func (t *Trigger) CheckFeasibility(ctx context.Context) error {
	var err error
	switch t.Category {
	case CategoryPeriodic:
		err = t.checkPeriodicTriggerFeasibility(ctx)
	case CategoryOnce:
	case CategoryOrdered:
	default:
		err = fmt.Errorf("unknown category, category(%s)", t.Category)
	}

	return err
}

// checkOnceTriggerFeasibility ...
func (t *Trigger) checkOnceTriggerFeasibility(ctx context.Context) error {
	// TODO: implement me
	return errors.New("not implemented")
}

// checkPeriodicTriggerFeasibility
func (t *Trigger) checkPeriodicTriggerFeasibility(ctx context.Context) error {
	now := time.Now()
	interval := t.MetadataPeriodic.GetInterval()
	if now.Sub(t.UpdatedAt) < interval {
		return nil
	}

	return nil
}

// checkOrderedTriggerFeasibility ...
func (t *Trigger) checkOrderedTriggerFeasibility(ctx context.Context) error {
	// TODO: implement me
	return errors.New("not implemented")
}

// Execute ...
func (t *Trigger) Execute(ctx context.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("trigger panic, recover(%v)", r)
		}

		t.UpdatedAt = time.Now()
	}()

	t.State = StateRunning

	switch t.Category {
	case CategoryOnce:
		err = t.executeOnceTrigger(ctx)
	case CategoryOrdered:
		err = t.executeOrderedTrigger(ctx)
	case CategoryPeriodic:
		err = t.executePeriodicTrigger(ctx)
	default:
		err = fmt.Errorf("unknown category, category(%s)", t.Category)
	}
	if err != nil {
		return err
	}

	return nil
}

// executeOnceTrigger ...
func (t *Trigger) executeOnceTrigger(ctx context.Context) error {
	// TODO: implement me
	return errors.New("implement me")
}

// executeOrderedTrigger ...
func (t *Trigger) executeOrderedTrigger(ctx context.Context) error {
	// TODO: implement me
	return errors.New("implement me")
}

// executePeriodicTrigger ...
func (t *Trigger) executePeriodicTrigger(ctx context.Context) error {
	// TODO: implement me
	return errors.New("implement me")
}

// Terminate ...
func (t *Trigger) Terminate() error {
	if t.State == StateTerminated {
		return errors.New("trigger already terminated")
	}

	t.UpdatedAt = time.Now()
	t.State = StateTerminated

	return nil
}

// TriggerOptionFunc ...
type TriggerOptionFunc func(*Trigger)

// WithTriggerID this option func is used in scenarios where triggers are read from the database.
func WithTriggerID(triggerID string) TriggerOptionFunc {
	return func(t *Trigger) {
		t.TriggerID = triggerID
	}
}

// NewTrigger create a new trigger.
func NewTrigger(category Category, metadata MetadataPeriodic, opts ...TriggerOptionFunc) *Trigger {
	trigger := &Trigger{
		TriggerID:        fmt.Sprintf("%s-%s", TriggerPrefix, uuid.NewString()),
		Category:         category,
		MetadataPeriodic: metadata,
		State:            StateInit,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	for _, opt := range opts {
		opt(trigger)
	}

	return trigger
}
