/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package node

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/common"
)

type mockActionInstancePrivateDataStorage struct {
	operInstID  string
	actionName  string
	privateData map[string]any
}

func (m *mockActionInstancePrivateDataStorage) UpsertActionInstancePrivateData(
	_ contextx.IContext,
	operInstID string,
	actionName string,
	privateData map[string]any,
) error {
	m.operInstID = operInstID
	m.actionName = actionName
	m.privateData = privateData

	return nil
}

func (m *mockActionInstancePrivateDataStorage) GetActionInstanceData(contextx.IContext, string, string) (*action.InstanceData, error) {
	return nil, nil
}

func (m *mockActionInstancePrivateDataStorage) GetActionInstanceLifecycle(contextx.IContext, string, string) (*action.Lifecycle, error) {
	return nil, nil
}

func (m *mockActionInstancePrivateDataStorage) UpdateActionInstanceLifecycle(contextx.IContext, string, string, *action.Lifecycle) error {
	return nil
}

func (m *mockActionInstancePrivateDataStorage) UpdateActionInstanceContent(contextx.IContext, string, string, map[string]any) error {
	return nil
}

func (m *mockActionInstancePrivateDataStorage) UpdateOperInstActionStatus(contextx.IContext, string, string, action.State) error {
	return nil
}

func (m *mockActionInstancePrivateDataStorage) PushActionInstanceMessage(contextx.IContext, string, string, ...common.Message) error {
	return nil
}

func (m *mockActionInstancePrivateDataStorage) GetActionInstancePrivateData(contextx.IContext, string, string) (map[string]any, error) {
	return nil, nil
}

func Test_actionInstallPreOrderedPlugins_saveSubWorkflowRefs(t *testing.T) {
	stg := new(mockActionInstancePrivateDataStorage)
	act := &actionInstallPreOrderedPlugins{
		storageActionInstance: stg,
	}

	serializedRefs, err := serializeSubWorkflowRefs([]types.SubWorkflowRef{{
		WorkflowID:     "wf-plugin-123",
		WorkflowDomain: types.WorkflowDomainPlugin,
	}})
	require.NoError(t, err)

	err = act.saveSubWorkflowRefs(contextx.New(context.Background()), "oper-inst-1", serializedRefs)
	require.NoError(t, err)
	require.Equal(t, "oper-inst-1", stg.operInstID)
	require.Equal(t, ActionNameInstallPreOrderedPlugins, stg.actionName)
	require.Equal(t, map[string]any{
		types.PDKeySubWorkflowRefs: `[{"workflow_id":"wf-plugin-123","workflow_domain":"plugin"}]`,
	}, stg.privateData)
}
