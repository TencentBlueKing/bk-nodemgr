## 文件定位

| 文件名                       | 功能描述                | 备注               |
|---------------------------|---------------------|------------------|
| action_def.go             | 存储所有 action_def 的名称 | 需要找 action 从此处开始 |
| oper_def.go               |                     |                  |
| action_{{action_name}}.go | 用于存储 action 的逻辑     |                  |                          |
| oper_{{oper_name}}.go     | 用于存储 operation 的逻辑  |                  |                          |

## action 模板

```go
/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package workflowdef ...
package workflowdef

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameActionName defines the action name.
	ActionNameActionName = "action_name"
)

// NewActionActionName ...
func NewActionActionName() action.Definition {
	return &ActionName{}
}

// ActParamActionName ...
type ActParamActionName struct {
}

// ActionName ...
type ActionName struct {
}

// Name returns the name of the action.
func (act *ActionName) Name() string {
	return ActionNameActionName
}

// Version returns the version of the action.
func (act *ActionName) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *ActionName) Description() string {
	return ""
}

// Timeout returns the timeout of the action.
func (act *ActionName) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *ActionName) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *ActionName) MaxRetryCount() uint {
	return 3
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *ActionName) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *ActionName) Do(ctx *action.InstanceContext) error {
	param := new(ActParamActionName)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	/*
		do something
	*/

	return nil
}
```

## Action 日志注意事项

1. err 无需手动使用 ctx.Data.Log 进行记录，框架会自动记录
2. 框架会自动记录 Action 执行的开始和结束日志
