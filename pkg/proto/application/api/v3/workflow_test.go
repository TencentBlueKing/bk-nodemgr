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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestSubWorkflowRefsFromPrivateData_WithJSONEncodedRefs(t *testing.T) {
	got := subWorkflowRefsFromPrivateData(map[string]any{
		types.PDKeySubWorkflowRefs: `[{"workflow_id":"wf-plugin-123","workflow_domain":"plugin"}]`,
	})

	if assert.Len(t, got, 1) {
		assert.Equal(t, "wf-plugin-123", got[0].WorkflowId)
		assert.Equal(t, string(types.WorkflowDomainPlugin), got[0].WorkflowDomain)
	}
}

func TestSubWorkflowRefsFromPrivateData_WithUnexpectedType(t *testing.T) {
	got := subWorkflowRefsFromPrivateData(map[string]any{
		types.PDKeySubWorkflowRefs: []types.SubWorkflowRef{{
			WorkflowID:     "wf-plugin-123",
			WorkflowDomain: types.WorkflowDomainPlugin,
		}},
	})

	assert.Empty(t, got)
}
