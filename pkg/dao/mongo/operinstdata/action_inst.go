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
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

// IActionInstData defines the interface for action instance data operations.
type IActionInstData interface {
	// GetActionInstData find one ActionInstData.
	GetActionInstData(ctx context.Context, operInstID string, actionName string) (*action.InstanceData, error)

	// UpdateActionInstData updates or inserts an ActionInstData.
	UpdateActionInstData(ctx context.Context, data *action.InstanceData) error

	// UpdateActInstMsg updates action inst data messages.
	UpdateActInstMsg(ctx context.Context, operInstID, actionName string, msgs []action.Message) error

	// GetActInstLifecycle get action inst data lifecycle.
	GetActInstLifecycle(ctx context.Context, operInstID string, actionName string) (*action.Lifecycle, error)

	// UpdateActInstLifecycle updates action inst data lifecycle.
	UpdateActInstLifecycle(ctx context.Context, operInstID, actionName string,
		lifecycle *action.Lifecycle) error

	// PushActionInstanceMessage push a message to the action_inst_data's msg queue.
	PushActionInstanceMessage(ctx context.Context, operInstID string, actionName string, msgs ...action.Message) error

	// AddActInstPrivateData add action inst data private data.
	AddActInstPrivateData(ctx context.Context, operInstID string, actionName string, data map[string]any) error

	// UpdateActionInstContent update the action_inst_data's content.
	UpdateActionInstContent(ctx context.Context, operInstID string, actionName string, content map[string]any) error

	// UpdateActionInstStatus updates the action_inst_data's status.
	UpdateActionInstStatus(ctx context.Context, operInstID string, actionName string,
		status action.State) error
}

// UpdateActInstMsg update action inst msg.
func (h *handler) UpdateActInstMsg(ctx context.Context, operInstID, actionName string,
	msgs []action.Message) error {

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

	filed := FieldKeyActionInstMessages(actionName)
	err := h.dao.updateField(ctx, filter, filed, convMessageToDB(msgs))
	if err != nil {
		return err
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

	filed := FieldKeyActionInstContent(actionName)
	err = h.dao.updateField(ctx, filter, filed, string(bytes))
	if err != nil {
		return err
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
		field := fmt.Sprintf("%s.%s", FieldKeyActInstPrivateData(actionName), k)
		err := h.dao.updateField(ctx, filter, field, v)
		if err != nil {
			return err
		}
	}

	return nil
}

// PushActionInstanceMessage push act inst msg.
func (h *handler) PushActionInstanceMessage(ctx context.Context, operationInstanceID, actionName string,
	messages ...action.Message) error {

	if ctx == nil {
		return errors.New("ctx is nil")
	}

	if operationInstanceID == "" {
		return errors.New("operation instance id is empty")
	}

	if len(messages) == 0 {
		return errors.New("msgs is empty")
	}

	filter := base.AliveFilter()
	opts := []OptFn{
		WithOperInstID(operationInstanceID),
	}
	for _, opt := range opts {
		filter = opt(filter)
	}

	field := fmt.Sprintf("action_data.%s.messages", actionName)
	for _, msg := range convMessageToDB(messages) {
		err := h.dao.pushField(ctx, filter, field, msg)
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateActionInstData upsert action inst data.
func (h *handler) UpdateActionInstData(ctx context.Context, actionInstData *action.InstanceData) error {
	if ctx == nil {
		return errors.New("ctx is nil")
	}

	if actionInstData == nil {
		return errors.New("actionInstData is nil")
	}

	if actionInstData.OperationInstanceID == "" {
		return errors.New("operation instance id is empty")
	}

	filter := base.AliveFilter()
	opts := []OptFn{
		WithOperInstID(actionInstData.OperationInstanceID),
	}
	for _, opt := range opts {
		filter = opt(filter)
	}

	data, err := ConvActionInstDataToDB(actionInstData)
	if err != nil {
		return err
	}
	filed := FieldKeyActionInstData(actionInstData.Name)

	return h.dao.updateField(ctx, filter, filed, data)
}

// UpdateActInstLifecycle update action inst lifecycle.
func (h *handler) UpdateActInstLifecycle(ctx context.Context, operInstID, actionName string,
	lifecycle *action.Lifecycle) error {

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

	filed := FieldKeyActionInstLifeCycle(actionName)
	err := h.dao.updateField(ctx, filter, filed, ConvActInstLifeCycleToDB(lifecycle))
	if err != nil {
		return err
	}

	return nil
}

// GetActionInstData find one action inst data.
func (h *handler) GetActionInstData(ctx context.Context, operInstID string,
	actionName string) (*action.InstanceData, error) {

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

	field := FieldKeyActionInstData(actionName)
	operInstData, err := h.dao.get(ctx, filter, field)
	if err != nil {
		return nil, err
	}

	actionInstData, ok := operInstData.ActionInstDataMap[actionName]
	if !ok {
		return nil, errors.New("action inst data not found")
	}

	data := &action.InstanceData{
		TriggerID:           actionInstData.TriggerID,
		OperationInstanceID: actionInstData.OperInstID,
		OperationID:         actionInstData.OperationID,
		OperationDefName:    actionInstData.OperDefName,
		Name:                actionInstData.Name,
		Index:               actionInstData.Index,
		TotalIndex:          actionInstData.TotalIndex,
		PrivateData:         actionInstData.PrivateData,
		Lifecycle:           ConvActInstLifeCycleFromDB(actionInstData.Lifecycle),
	}

	for k, v := range actionInstData.PrivateData {
		data.PrivateData[k] = v
	}

	for _, msg := range actionInstData.Messages {
		data.Messages = append(data.Messages, action.Message{
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

// GetActInstLifecycle find one action inst data.
func (h *handler) GetActInstLifecycle(ctx context.Context, operInstID string,
	actionName string) (*action.Lifecycle, error) {

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

	field := FieldKeyActionInstLifeCycle(actionName)
	operInstData, err := h.dao.get(ctx, filter, field)
	if err != nil {
		return nil, err
	}

	actionInstData, ok := operInstData.ActionInstDataMap[actionName]
	if !ok {
		return nil, errors.New("action inst data not found")
	}

	lifecycle := ConvActInstLifeCycleFromDB(actionInstData.Lifecycle)

	return lifecycle, nil
}

// UpdateActionInstStatus update action inst status.
func (h *handler) UpdateActionInstStatus(ctx context.Context, operInstID string, actionName string,
	status action.State) error {

	if ctx == nil {
		return base.ErrInvalidContext()
	}

	if operInstID == "" {
		return base.ErrInvalidID()
	}

	if actionName == "" {
		return errors.New("actionName is empty")
	}

	if err := status.Validate(); err != nil {
		return err
	}

	filter := base.AliveFilter()
	filter = WithOperInstID(operInstID)(filter)
	field := FieldKeyActionInstState(actionName)
	err := h.dao.updateField(ctx, filter, field, status)
	if err != nil {
		return err
	}

	return nil
}
