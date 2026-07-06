/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package action describes the basic step in operation.
package action

import (
	"time"
)

// Tag represents a tag of an action.
type Tag string

const (
	// TagNeedManualExecInstallScript tag need manual execute install script.
	TagNeedManualExecInstallScript Tag = "need_manual_exec_install_script"

	// TagNeedOfflineManualInstall tag for offline install: user downloads the offline package,
	// executes it manually, then submits the result data via the guide UI.
	TagNeedOfflineManualInstall Tag = "need_offline_manual_install"
)

// Definition represents an action, which is a single basic step of work.
type Definition interface {
	// Name returns the name of the action.
	Name() string

	// DisplayNameZh returns the Chinese display name of the action.
	DisplayNameZh() string

	// DisplayNameEn returns the English display name of the action.
	DisplayNameEn() string

	// Version returns the version of the action.
	Version() string

	// Description returns the description of the action.
	Description() string

	// Timeout returns the timeout of the action.
	Timeout() time.Duration

	// Tags returns the tags of the action.
	Tags() []Tag

	// MaxRetryCount returns the max retry count of the action.
	MaxRetryCount() uint

	// DelayFn returns the delay function of the action.
	DelayFn(attempt int) func()

	// Do executes the action, with specified context.
	Do(iCtx *InstanceContext) error
}
