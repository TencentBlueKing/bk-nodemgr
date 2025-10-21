/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package identifier provides identifier definition for runtime.
package identifier

import (
	"strings"

	"github.com/google/uuid"
)

const (
	tagRequestID           = "rid"
	tagMessageID           = "mid"
	tagWorkflowID          = "wf"
	tagTriggerID           = "trig"
	tagOperationID         = "oper"
	tagOperationInstanceID = "oper-inst"
	tagActionInstanceID    = "act-inst"
	tagServiceID           = "svc"
	tagUploadID            = "up"
	tagPluginID            = "plugin"
	tagProcessID           = "proc"
)

func generateID(tag string) string {
	return tag + ":" + strings.ReplaceAll(uuid.New().String(), "-", "")
}

// GenRequestID generates a request id.
func GenRequestID() string {
	return generateID(tagRequestID)
}

// GenMessageID generates a message id.
func GenMessageID() string {
	return generateID(tagMessageID)
}

// GenWorkflowID generates a workflow id.
func GenWorkflowID() string {
	return generateID(tagWorkflowID)
}

// GenTriggerID generates a trigger id.
func GenTriggerID() string {
	return generateID(tagTriggerID)
}

// GenOperationID generates a operation id.
func GenOperationID() string {
	return generateID(tagOperationID)
}

// GenOperationInstanceID generates a operation instance id.
func GenOperationInstanceID() string {
	return generateID(tagOperationInstanceID)
}

// GenActionInstanceID generates a action instance id.
func GenActionInstanceID() string {
	return generateID(tagActionInstanceID)
}

// GenServiceID generates a service id.
func GenServiceID() string {
	return generateID(tagServiceID)
}

// GenUploadID generates a upload id.
func GenUploadID() string {
	return generateID(tagUploadID)
}

// GenPluginID generates a plugin id.
func GenPluginID() string {
	return generateID(tagPluginID)
}

// GenProcessID generates a process id.
func GenProcessID() string {
	return generateID(tagProcessID)
}
