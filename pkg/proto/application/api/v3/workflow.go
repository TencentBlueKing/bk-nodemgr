/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package v3

import (
	"encoding/json"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// subWorkflowRefsFromPrivateData maps private_data[sub_workflow_refs] to API sub_workflow_refs.
func subWorkflowRefsFromPrivateData(privateData map[string]any) []*SubWorkflowRef {
	serializedRefs, ok := privateData[types.PDKeySubWorkflowRefs].(string)
	if !ok || serializedRefs == "" {
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
