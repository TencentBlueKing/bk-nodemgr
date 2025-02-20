/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operengine ...
package operengine

import (
	"context"
	"time"
)

// ActionTag represents a tag of an action.
type ActionTag string

const (
	// ActionTagDisabledAutoRefreshMsg represents an action which is disabled auto refresh msg.
	ActionTagDisabledAutoRefreshMsg ActionTag = "disabled_auto_refresh_msg"
)

// ActionInstContext represents the context of an action.
type ActionInstContext struct {
	// Context is the context of the action
	Ctx context.Context
	// Data is the action being executed
	Data *ActionInstData
}

// ActionDef represents an operation inst operInstMgr action, which is a single basic step of work.
type ActionDef interface {
	// Name returns the name of the action.
	Name() string

	// Version returns the version of the action.
	Version() string

	// Description returns the description of the action.
	Description() string

	// Timeout returns the timeout of the action.
	Timeout() time.Duration

	// Tags returns the tags of the action.
	Tags() []ActionTag

	// MaxRetryCount returns the max retry count of the action.
	MaxRetryCount() uint

	// DelayFn returns the delay function of the action.
	DelayFn() func()

	// Do executes the action, with specified context.
	Do(*ActionInstContext) error
}
