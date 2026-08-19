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

package v3

import (
	"encoding/json"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// subWorkflowRefsFromPrivateData maps private_data[sub_workflow_refs] to API sub_workflow_refs.
func subWorkflowRefsFromPrivateData(privateData map[string]any) []*SubWorkflowRef {
	serializedRefs := conv.ToStringDefault(privateData[types.PDKeySubWorkflowRefs], "")
	if serializedRefs == "" {
		return []*SubWorkflowRef{}
	}

	var refs []types.SubWorkflowRef
	if err := json.Unmarshal([]byte(serializedRefs), &refs); err != nil {
		return []*SubWorkflowRef{}
	}

	out := make([]*SubWorkflowRef, 0, len(refs))
	for _, ref := range refs {
		out = append(out, &SubWorkflowRef{
			WorkflowId:     ref.WorkflowID,
			WorkflowDomain: string(ref.WorkflowDomain),
		})
	}

	return out
}
