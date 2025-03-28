/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operinstdata ...
package operinstdata

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler this define the handler interface.
type IHandler interface {
	ICrud
	IActionInstData

	// FindOneWithoutActionData find one OperInstData without action data.
	FindOneWithoutActionData(ctx context.Context, opts ...OptFn) (*operengine.OperInstData, error)

	// UpdateLifecycle updates or inserts an OperInstData's lifecycle.
	UpdateLifecycle(ctx context.Context, operInstID string, lifecycle *operengine.Lifecycle) error
}

type handler struct {
	dao *dao
}

// New create a new host handler.
func New(client *mongo.Database, logger logger.Logger) IHandler {
	return &handler{
		dao: newDao(client, logger),
	}
}

// FindOneWithoutActionData find operinstdata without action data.
func (h *handler) FindOneWithoutActionData(ctx context.Context, opts ...OptFn) (*operengine.OperInstData, error) {
	if ctx == nil {
		return nil, errors.New("ctx is nil")
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	field := "action_data"
	operInstDatas, err := h.dao.findWithoutFields(ctx, filter, field)
	if err != nil {
		return nil, err
	}

	if len(operInstDatas) == 0 {
		return nil, errors.New("not found")
	}

	operInstData := operInstDatas[0]

	data := &operengine.OperInstData{
		TriggerID:         operInstData.TriggerID,
		OperInstID:        operInstData.OperInstID,
		OperDefName:       operInstData.OperDefName,
		ActionNames:       operInstData.ActionNames,
		ParentOperationID: operInstData.ParentOperationID,
		Timeout:           operInstData.Timeout,
		Lifecycle:         convLifecycleToCommon(operInstData.Lifecycle),
	}

	if len(operInstData.InitContent) == 0 {
		return nil, errors.New("invalid init content")
	}

	err = json.Unmarshal([]byte(operInstData.InitContent), &data.InitContent)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// UpdateLifecycle update operation instance's lifecycle.
func (h *handler) UpdateLifecycle(ctx context.Context, operInstID string, lifecycle *operengine.Lifecycle) error {
	if ctx == nil {
		return errors.New("ctx is nil")
	}

	if operInstID == "" {
		return errors.New("operation instance id is empty")
	}

	if lifecycle == nil {
		return errors.New("lifecycle is nil")
	}

	filter := base.AliveFilter()
	opts := []OptFn{
		WithOperInstID(operInstID),
	}
	for _, opt := range opts {
		filter = opt(filter)
	}

	err := h.dao.updateField(ctx, filter, FieldKeyLifeCycle.String(), convLifecycleToDB(lifecycle))
	if err != nil {
		return err
	}

	return nil
}

// convLifecycleToDB convert lifecycle to db.
func convLifecycleToDB(lifecycle *operengine.Lifecycle) *Lifecycle {
	if lifecycle == nil {
		return nil
	}

	return &Lifecycle{
		CreatedAt: lifecycle.CreatedAt,
		StartedAt: lifecycle.StartedAt,
		EndedAt:   lifecycle.EndedAt,
		State:     string(lifecycle.State),
		StoppedAt: lifecycle.StoppedAt,
	}
}

// convActInstLifeCycleToDB convert action instance lifecycle to db.
func convActInstLifeCycleToDB(lifecycle *operengine.ActInstLifeCycle) *ActInstLifeCycle {
	if lifecycle == nil {
		return nil
	}

	return &ActInstLifeCycle{
		StartedAt: lifecycle.StartedAt,
		EndedAt:   lifecycle.EndedAt,
		State:     string(lifecycle.State),
		StoppedAt: lifecycle.StoppedAt,
	}
}

// convLifecycleToCommon convert lifecycle to common.
func convLifecycleToCommon(lifecycle *Lifecycle) *operengine.Lifecycle {
	if lifecycle == nil {
		return nil
	}

	return &operengine.Lifecycle{
		CreatedAt: lifecycle.CreatedAt,
		StartedAt: lifecycle.StartedAt,
		EndedAt:   lifecycle.EndedAt,
		State:     operengine.OperInstState(lifecycle.State),
		StoppedAt: lifecycle.StoppedAt,
	}
}

// convActInstLifeCycleToCommon convert action instance lifecycle to common.
func convActInstLifeCycleToCommon(lifecycle *ActInstLifeCycle) *operengine.ActInstLifeCycle {
	dbData := &operengine.ActInstLifeCycle{
		State:     operengine.ActionInstState(lifecycle.State),
		StartedAt: lifecycle.StartedAt,
		EndedAt:   lifecycle.EndedAt,
		StoppedAt: lifecycle.StoppedAt,
	}

	return dbData
}

func convActionInstDataToDB(actionInstData *operengine.ActionInstData) (*ActionInstData, error) {
	data := &ActionInstData{
		TriggerID:  actionInstData.TriggerID,
		OperInstID: actionInstData.OperInstID,
		Name:       actionInstData.Name,
		Index:      actionInstData.Index,
		Messages:   make([]Message, 0, len(actionInstData.Messages)),
		Lifecycle:  convActInstLifeCycleToDB(actionInstData.Lifecycle),
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

// convOperInstDataToDB convert oper inst data to db.
func convOperInstDataToDB(data *operengine.OperInstData) (*OperInstData, error) {
	dbData := &OperInstData{
		TriggerID:         data.TriggerID,
		OperInstID:        data.OperInstID,
		ActionNames:       data.ActionNames,
		OperDefName:       data.OperDefName,
		ParentOperationID: data.ParentOperationID,
		Timeout:           data.Timeout,
		Lifecycle:         convLifecycleToDB(data.Lifecycle),
	}

	if data.InitContent == nil {
		return nil, errors.New("invalid init content")
	}

	bytes, err := json.Marshal(data.InitContent)
	if err != nil {
		return nil, err
	}

	dbData.InitContent = string(bytes)

	dbData.ActionInstDataMap, err = convActionInstDataMapToDB(data.ActionInstDataMap)
	if err != nil {
		return nil, err
	}

	return dbData, nil
}

// convActionInstDataMapToDB convert action inst data map to db.
func convActionInstDataMapToDB(commData map[string]*operengine.ActionInstData) (map[string]*ActionInstData, error) {
	var err error

	dbData := make(map[string]*ActionInstData, len(commData))
	for k, v := range commData {
		dbData[k], err = convActionInstDataToDB(v)
		if err != nil {
			return nil, err
		}
	}

	return dbData, nil
}

// convMessageToDB convert message to db.
func convMessageToDB(msgs []operengine.Message) []Message {
	dbData := make([]Message, len(msgs))
	for idx, msg := range msgs {
		dbData[idx] = Message{
			Time: msg.Time,
			Text: msg.Text,
		}
	}

	return dbData
}
