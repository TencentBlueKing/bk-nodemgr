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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
	"go.mongodb.org/mongo-driver/mongo"
)

// Handler ...
type Handler interface {
	// Upsert updates or inserts an OperInstData.
	Upsert(ctx context.Context, data *operengine.OperInstData) error

	// FindOne find one OperInstData.
	FindOne(ctx context.Context, opts ...OptFn) (*operengine.OperInstData, error)

	// FindOneWithoutActionData find one OperInstData without action data.
	FindOneWithoutActionData(ctx context.Context, opts ...OptFn) (*operengine.OperInstData, error)

	// UpdateActInstLifecycle updates the action_inst_data's lifecycle.
	UpdateActInstLifecycle(ctx context.Context, operInstID, actionName string,
		lifecycle *operengine.ActInstLifeCycle) error

	// UpdateActionInstData update tge action_inst_data content.
	UpdateActionInstData(ctx context.Context, data *operengine.ActionInstData) error

	// GetActionInstData find one ActionInstData.
	GetActionInstData(ctx context.Context, operInstID string, actionName string) (*operengine.ActionInstData, error)

	// UpdateLifecycle updates the OperInst's Lifecycle.
	UpdateLifecycle(ctx context.Context, operInstID string, lifecycle *operengine.Lifecycle) error

	// PushActInstMsgs push a message to the action_inst_data's msg queue.
	PushActInstMsgs(ctx context.Context, operInstID string, actionName string, msgs ...operengine.Message) error

	// UpdateActionInstContent update the action_inst_data's content.
	UpdateActionInstContent(ctx context.Context, operInstID string, actionName string, content map[string]any) error

	// AddActInstPrivateData add action inst data private data.
	AddActInstPrivateData(ctx context.Context, operInstID string, actionName string, data map[string]any) error
}

type handler struct {
	dao *dao
}

// New create a new host handler.
func New(client *mongo.Database, logger logger.Logger) Handler {
	return &handler{
		dao: newDao(client, logger),
	}
}

// Upsert updates or inserts an OperInstData.
func (h *handler) Upsert(ctx context.Context, data *operengine.OperInstData) error {
	if ctx == nil {
		return errors.New("ctx is nil")
	}

	if data == nil {
		return errors.New("data is nil")
	}

	operInstData, err := convOperInstDataToDB(data)
	if err != nil {
		return fmt.Errorf("convert oper inst data to db data error: %v", err)
	}

	err = h.dao.upsert(ctx, operInstData)
	if err != nil {
		return err
	}

	return nil
}

// FindOne find one OperInstData.
func (h *handler) FindOne(ctx context.Context, opts ...OptFn) (*operengine.OperInstData, error) {
	if ctx == nil {
		return nil, errors.New("ctx is nil")
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	operInstDatas, err := h.dao.find(ctx, filter)
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
		ActionInstDataMap: make(map[string]*operengine.ActionInstData, len(operInstData.ActionInstDataMap)),
		ParentOperInstID:  operInstData.ParentOperInstID,
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

	for k, v := range operInstData.ActionInstDataMap {
		actionInstData := &operengine.ActionInstData{
			TriggerID:   v.TriggerID,
			OperInstID:  v.OperInstID,
			Name:        v.Name,
			Index:       v.Index,
			Messages:    make([]operengine.Message, len(v.Messages)),
			Content:     make(map[string]any, len(v.Content)),
			PrivateData: v.PrivateData,
			Lifecycle:   convActInstLifeCycleToCommon(v.Lifecycle),
		}

		for idx, msg := range v.Messages {
			actionInstData.Messages[idx] = operengine.Message{
				Time: msg.Time,
				Text: msg.Text,
			}
		}

		err = json.Unmarshal([]byte(v.Content), &actionInstData.Content)
		if err != nil {
			return nil, err
		}

		data.ActionInstDataMap[k] = actionInstData
	}

	return data, nil
}

// UpdateActionInstData upsert action inst data.
func (h *handler) UpdateActionInstData(ctx context.Context, actionInstData *operengine.ActionInstData) error {
	if ctx == nil {
		return errors.New("ctx is nil")
	}

	if actionInstData == nil {
		return errors.New("actionInstData is nil")
	}

	if actionInstData.OperInstID == "" {
		return errors.New("operation instance id is empty")
	}

	filter := base.AliveFilter()
	opts := []OptFn{
		WithOperInstID(actionInstData.OperInstID),
	}
	for _, opt := range opts {
		filter = opt(filter)
	}

	data, err := convActionInstDataToDB(actionInstData)
	if err != nil {
		return err
	}

	filed := fmt.Sprintf("action_data.%s", actionInstData.Name)
	err = h.dao.updateField(ctx, filter, filed, data)
	if err != nil {
		return err
	}

	return nil
}

// UpdateActInstMsg update action inst msg.
func (h *handler) UpdateActInstMsg(ctx context.Context, operInstID, actionName string,
	msgs []operengine.Message) error {

	if ctx == nil {
		return errors.New("ctx is nil")
	}

	if operInstID == "" {
		return errors.New("operation instance id is empty")
	}

	if actionName == "" {
		return errors.New("actionName is empty")
	}

	filter := base.AliveFilter()
	opts := []OptFn{
		WithOperInstID(operInstID),
	}
	for _, opt := range opts {
		filter = opt(filter)
	}

	filed := fmt.Sprintf("action_data.%s.messages", actionName)
	err := h.dao.updateField(ctx, filter, filed, convMessageToDB(msgs))
	if err != nil {
		return err
	}

	return nil
}

// UpdateActInstLifecycle update action inst lifecycle.
func (h *handler) UpdateActInstLifecycle(ctx context.Context, operInstID, actionName string,
	lifecycle *operengine.ActInstLifeCycle) error {

	if ctx == nil {
		return errors.New("ctx is nil")
	}

	if operInstID == "" {
		return errors.New("operation instance id is empty")
	}

	if actionName == "" {
		return errors.New("actionName is empty")
	}

	filter := base.AliveFilter()
	opts := []OptFn{
		WithOperInstID(operInstID),
	}
	for _, opt := range opts {
		filter = opt(filter)
	}

	filed := fmt.Sprintf("action_data.%s.life_cycle", actionName)
	err := h.dao.updateField(ctx, filter, filed, convActInstLifeCycleToDB(lifecycle))
	if err != nil {
		return err
	}

	return nil
}

// GetActionInstData find one action inst data.
func (h *handler) GetActionInstData(ctx context.Context, operInstID string,
	actionName string) (*operengine.ActionInstData, error) {

	if ctx == nil {
		return nil, errors.New("ctx is nil")
	}

	if operInstID == "" {
		return nil, errors.New("operation instance id is empty")
	}

	if actionName == "" {
		return nil, errors.New("actionName is empty")
	}

	filter := base.AliveFilter()
	opts := []OptFn{
		WithOperInstID(operInstID),
	}
	for _, opt := range opts {
		filter = opt(filter)
	}

	field := fmt.Sprintf("action_data.%s", actionName)
	operInstData, err := h.dao.findOne(ctx, filter, field)
	if err != nil {
		return nil, err
	}

	actionInstData, ok := operInstData.ActionInstDataMap[actionName]
	if !ok {
		return nil, errors.New("action inst data not found")
	}

	data := &operengine.ActionInstData{
		TriggerID:   actionInstData.TriggerID,
		OperInstID:  actionInstData.OperInstID,
		Name:        actionInstData.Name,
		Index:       actionInstData.Index,
		PrivateData: actionInstData.PrivateData,
		Lifecycle:   convActInstLifeCycleToCommon(actionInstData.Lifecycle),
	}

	for _, msg := range actionInstData.Messages {
		data.Messages = append(data.Messages, operengine.Message{
			Time: msg.Time,
			Text: msg.Text,
		})
	}

	err = json.Unmarshal([]byte(actionInstData.Content), &data.Content)
	if err != nil {
		return nil, err
	}

	return data, nil
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
		TriggerID:        operInstData.TriggerID,
		OperInstID:       operInstData.OperInstID,
		OperDefName:      operInstData.OperDefName,
		ActionNames:      operInstData.ActionNames,
		ParentOperInstID: operInstData.ParentOperInstID,
		Timeout:          operInstData.Timeout,
		Lifecycle:        convLifecycleToCommon(operInstData.Lifecycle),
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

	err := h.dao.updateField(ctx, filter, "lifecycle", convLifecycleToDB(lifecycle))
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
		TriggerID:        data.TriggerID,
		OperInstID:       data.OperInstID,
		ActionNames:      data.ActionNames,
		OperDefName:      data.OperDefName,
		ParentOperInstID: data.ParentOperInstID,
		Timeout:          data.Timeout,
		Lifecycle:        convLifecycleToDB(data.Lifecycle),
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

// PushActInstMsgs push act inst msg.
func (h *handler) PushActInstMsgs(ctx context.Context, operInstID string, actionName string,
	msgs ...operengine.Message) error {

	if ctx == nil {
		return errors.New("ctx is nil")
	}

	if operInstID == "" {
		return errors.New("operation instance id is empty")
	}

	if len(msgs) == 0 {
		return errors.New("msgs is empty")
	}

	filter := base.AliveFilter()
	opts := []OptFn{
		WithOperInstID(operInstID),
	}
	for _, opt := range opts {
		filter = opt(filter)
	}

	field := fmt.Sprintf("action_data.%s.messages", actionName)
	for _, msg := range convMessageToDB(msgs) {
		err := h.dao.pushField(ctx, filter, field, msg)
		if err != nil {
			return err
		}
	}

	return nil

}

// AddActInstPrivateData add act inst private data.
func (h *handler) AddActInstPrivateData(ctx context.Context, operInstID string, actionName string,
	data map[string]any) error {

	if ctx == nil {
		return errors.New("ctx is nil")
	}

	if operInstID == "" {
		return errors.New("operation instance id is empty")
	}

	if actionName == "" {
		return errors.New("actionName is empty")
	}

	if len(data) == 0 {
		return errors.New("data is empty")
	}

	filter := base.AliveFilter()
	opts := []OptFn{
		WithOperInstID(operInstID),
	}
	for _, opt := range opts {
		filter = opt(filter)
	}

	for k, v := range data {
		field := fmt.Sprintf("action_data.%s.private_data.%s", actionName, k)
		err := h.dao.updateField(ctx, filter, field, v)
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateActionInstContent update action instance content.
func (h *handler) UpdateActionInstContent(ctx context.Context, operInstID string, actionName string,
	content map[string]any) error {

	if ctx == nil {
		return errors.New("ctx is nil")
	}

	if operInstID == "" {
		return errors.New("operation instance id is empty")
	}

	if actionName == "" {
		return errors.New("actionName is empty")
	}

	filter := base.AliveFilter()
	opts := []OptFn{
		WithOperInstID(operInstID),
	}
	for _, opt := range opts {
		filter = opt(filter)
	}

	bytes, err := json.Marshal(content)
	if err != nil {
		return err
	}

	filed := fmt.Sprintf("action_data.%s.content", actionName)
	err = h.dao.updateField(ctx, filter, filed, string(bytes))
	if err != nil {
		return err
	}

	return nil
}
