/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package operinstdata

import (
	"encoding/json"
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// ConvOperaLifeCycleToDB convert LifeCycle to db.
func ConvOperaLifeCycleToDB(lifeCycle *operation.Lifecycle) *LifeCycle {
	if lifeCycle == nil {
		return nil
	}

	return &LifeCycle{
		CreatedAt: lifeCycle.CreatedAt,
		StartedAt: lifeCycle.StartedAt,
		EndedAt:   lifeCycle.EndedAt,
		State:     string(lifeCycle.State),
		StoppedAt: lifeCycle.StoppedAt,
	}
}

// ConvActInstLifeCycleToDB convert action instance LifeCycle to db.
func ConvActInstLifeCycleToDB(lifeCycle *action.Lifecycle) *LifeCycle {
	if lifeCycle == nil {
		return nil
	}

	return &LifeCycle{
		CreatedAt: lifeCycle.CreatedAt,
		StartedAt: lifeCycle.StartedAt,
		EndedAt:   lifeCycle.EndedAt,
		State:     string(lifeCycle.State),
		StoppedAt: lifeCycle.StoppedAt,
	}
}

// ConvOperaLifeCycleFromDB convert LifeCycle to common.
func ConvOperaLifeCycleFromDB(lifeCycle *LifeCycle) *operation.Lifecycle {
	if lifeCycle == nil {
		return nil
	}

	return &operation.Lifecycle{
		CreatedAt: lifeCycle.CreatedAt,
		StartedAt: lifeCycle.StartedAt,
		EndedAt:   lifeCycle.EndedAt,
		State:     operation.State(lifeCycle.State),
		StoppedAt: lifeCycle.StoppedAt,
	}
}

// ConvActInstLifeCycleFromDB convert action instance LifeCycle to common.
func ConvActInstLifeCycleFromDB(lifeCycle *LifeCycle) *action.Lifecycle {
	if lifeCycle == nil {
		return &action.Lifecycle{}
	}
	dbData := &action.Lifecycle{
		State:     action.State(lifeCycle.State),
		CreatedAt: lifeCycle.CreatedAt,
		StartedAt: lifeCycle.StartedAt,
		EndedAt:   lifeCycle.EndedAt,
		StoppedAt: lifeCycle.StoppedAt,
	}

	return dbData
}

// ConvActionInstDataToDB convert action inst data to db.
func ConvActionInstDataToDB(actionInstData *action.InstanceData) (*ActionInstData, error) {
	data := &ActionInstData{
		TriggerID:   actionInstData.TriggerID,
		OperInstID:  actionInstData.OperationInstanceID,
		OperationID: actionInstData.OperationID,
		OperDefName: actionInstData.OperationDefName,
		Name:        actionInstData.Name,
		Index:       actionInstData.Index,
		TotalIndex:  actionInstData.TotalIndex,

		PrivateData: make(map[string]any, len(actionInstData.PrivateData)),

		Messages:  make([]Message, 0, len(actionInstData.Messages)),
		Lifecycle: ConvActInstLifeCycleToDB(actionInstData.Lifecycle),
	}

	for k, v := range actionInstData.PrivateData {
		data.PrivateData[k] = v
	}

	for _, message := range actionInstData.Messages {
		data.Messages = append(data.Messages, Message{
			Time: message.Time,
			Text: message.Text,
		})
	}

	bytes, err := json.Marshal(actionInstData.Content)
	if err != nil {
		return nil, err
	}

	data.Content = string(bytes)

	return data, nil
}

// ConvOpInstanceDataToDB convert oper inst data to db.
func ConvOpInstanceDataToDB(data *operation.InstanceData) (*OperInstData, error) {
	dbData := &OperInstData{
		TriggerID:         data.Metadata.TriggerID,
		OperInstID:        data.Metadata.OperationInstanceID,
		ActionNames:       data.Metadata.ActionNames,
		OperationID:       data.Metadata.OperationID,
		OperDefName:       data.Metadata.OperationDefName,
		ParentOperationID: data.Metadata.ParentOperationID,
		Timeout:           data.Metadata.Timeout,
		Lifecycle:         ConvOperaLifeCycleToDB(data.Lifecycle),
	}

	if data.Metadata.InitContent == nil {
		return nil, errors.New("invalid init content")
	}

	bytes, err := json.Marshal(data.Metadata.InitContent)
	if err != nil {
		return nil, err
	}

	dbData.InitContent = string(bytes)

	dbData.ActionInstDataMap, err = ConvActionInstDataMapToDB(data.ActionInstanceDataMap)
	if err != nil {
		return nil, err
	}

	return dbData, nil
}

// ConvActionInstDataMapToDB convert action inst data map to db.
func ConvActionInstDataMapToDB(commData map[string]*action.InstanceData) (map[string]*ActionInstData, error) {
	var err error

	dbData := make(map[string]*ActionInstData, len(commData))
	for k, v := range commData {
		dbData[k], err = ConvActionInstDataToDB(v)
		if err != nil {
			return nil, err
		}
	}

	return dbData, nil
}

// convMessageToDB convert message to db.
func convMessageToDB(msgs []action.Message) []Message {
	dbData := make([]Message, len(msgs))
	for idx, msg := range msgs {
		dbData[idx] = Message{
			Time: msg.Time,
			Text: msg.Text,
		}
	}

	return dbData
}

// ConvAOperaInstDataWithoutActionFromDB convert oper inst data without action data to common.
func ConvAOperaInstDataWithoutActionFromDB(opear *OperInstData) (*operation.InstanceData, error) {
	if opear == nil {
		return nil, errors.New("oper inst data is nil")
	}

	data := &operation.InstanceData{
		InstanceBriefData: operation.InstanceBriefData{
			Metadata: &operation.InstanceMetadata{
				TriggerID:           opear.TriggerID,
				OperationInstanceID: opear.OperInstID,
				OperationID:         opear.OperationID,
				OperationDefName:    opear.OperDefName,
				ActionNames:         opear.ActionNames,

				ParentOperationID: opear.ParentOperationID,
				Timeout:           opear.Timeout,
			},
			Lifecycle: ConvOperaLifeCycleFromDB(opear.Lifecycle),
		},
	}

	if opear.InitContent != "" {
		err := json.Unmarshal([]byte(opear.InitContent), &data.Metadata.InitContent)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

// ConvAOpeInstBreiefDataFromDB convert oper inst data without action data to common.
func ConvAOpeInstBreiefDataFromDB(opear *OperInstData) (*operation.InstanceBriefData, error) {
	if opear == nil {
		return nil, errors.New("oper inst data is nil")
	}

	data := &operation.InstanceBriefData{
		Metadata: &operation.InstanceMetadata{
			TriggerID:           opear.TriggerID,
			OperationInstanceID: opear.OperInstID,
			OperationID:         opear.OperationID,
			OperationDefName:    opear.OperDefName,
			ActionNames:         opear.ActionNames,

			ParentOperationID: opear.ParentOperationID,
			Timeout:           opear.Timeout,
		},
		Lifecycle: ConvOperaLifeCycleFromDB(opear.Lifecycle),
	}

	if len(opear.InitContent) == 0 {
		return nil, errors.New("invalid init content")
	}

	err := json.Unmarshal([]byte(opear.InitContent), &data.Metadata.InitContent)
	if err != nil {
		return nil, err
	}

	return data, nil
}
