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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/common"
)

// IActionInstData defines the interface for action instance data operations.
type IActionInstData interface {
	// GetActionInstData find one ActionInstData.
	GetActionInstData(nCtx contextx.IContext, operInstID string, actionName string) (*action.InstanceData, error)

	// UpdateActionInstData updates or inserts an ActionInstData.
	UpdateActionInstData(nCtx contextx.IContext, data *action.InstanceData) error

	// UpdateActInstMsg updates action inst data messages.
	UpdateActInstMsg(nCtx contextx.IContext, operInstID, actionName string, msgs []common.Message) error

	// GetActInstLifecycle get action inst data lifecycle.
	GetActInstLifecycle(nCtx contextx.IContext, operInstID string, actionName string) (*action.Lifecycle, error)

	// UpdateActInstLifecycle updates action inst data lifecycle.
	UpdateActInstLifecycle(nCtx contextx.IContext, operInstID, actionName string, lifecycle *action.Lifecycle) error

	// PushActionInstanceMessage push a message to the action_inst_data's msg queue.
	PushActionInstanceMessage(nCtx contextx.IContext, operInstID string, actionName string, msgs ...common.Message) error

	// GetActInstPrivateData get action inst data private data.
	GetActInstPrivateData(nCtx contextx.IContext, operInstID string, actionName string) (map[string]any, error)

	// PushActInstPrivateData add action inst data private data.
	PushActInstPrivateData(nCtx contextx.IContext, operInstID string, actionName string, data map[string]any) error

	// UpdateActionInstContent update the action_inst_data's content.
	UpdateActionInstContent(nCtx contextx.IContext, operInstID string, actionName string, content map[string]any) error

	// UpdateActionInstStatus updates the action_inst_data's status.
	UpdateActionInstStatus(nCtx contextx.IContext, operInstID string, actionName string, status action.State) error
}

// UpdateActInstMsg update action inst msg.
func (h *Handler) UpdateActInstMsg(nCtx contextx.IContext, operInstID, actionName string, msgs []common.Message) error {
	if nCtx == nil {
		return errors.New("nCtx is nil")
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
	err := h.dao.updateField(nCtx, filter, filed, convMessageToDB(msgs))
	if err != nil {
		return err
	}

	return nil
}

// UpdateActionInstContent update action instance content.
func (h *Handler) UpdateActionInstContent(
	nCtx contextx.IContext, operInstID string, actionName string, content map[string]any) error {

	if nCtx == nil {
		return errors.New("nCtx is nil")
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
	err = h.dao.updateField(nCtx, filter, filed, string(bytes))
	if err != nil {
		return err
	}

	return nil
}

// PushActInstPrivateData push act inst private data.
func (h *Handler) PushActInstPrivateData(
	nCtx contextx.IContext, operInstID string, actionName string, data map[string]any) error {

	if nCtx == nil {
		return errors.New("nCtx is nil")
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

	operInstData, err := h.dao.get(nCtx, filter)
	if err != nil {
		return err
	}

	for k, v := range data {
		operInstData.ActionInstDataMap[actionName].PrivateData[k] = v
	}

	err = h.dao.updateField(nCtx, filter,
		FieldKeyActionInstData(actionName), operInstData.ActionInstDataMap[actionName])

	if err != nil {
		return err
	}

	return nil
}

// PushActionInstanceMessage push act inst msg.
func (h *Handler) PushActionInstanceMessage(
	nCtx contextx.IContext, operationInstanceID, actionName string, messages ...common.Message) error {

	if nCtx == nil {
		return errors.New("nCtx is nil")
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
		err := h.dao.pushField(nCtx, filter, field, msg)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetActInstPrivateData get action inst data private data.
func (h *Handler) GetActInstPrivateData(
	nCtx contextx.IContext, operInstID string, actionName string) (map[string]any, error) {

	if nCtx == nil {
		return nil, errors.New("nCtx is nil")
	}

	if operInstID == "" {
		return nil, errors.New("operation instance id is empty")
	}

	if actionName == "" {
		return nil, errors.New("action name is empty")
	}

	filter := base.AliveFilter()
	opts := []OptFn{
		WithOperInstID(operInstID),
	}
	for _, opt := range opts {
		filter = opt(filter)
	}

	field := FieldKeyActInstPrivateData(actionName)
	operInstData, err := h.dao.get(nCtx, filter, field)
	if err != nil {
		return nil, err
	}

	actionInstData, ok := operInstData.ActionInstDataMap[actionName]
	if !ok {
		return nil, errors.New("action inst private data not found")
	}

	return actionInstData.PrivateData, nil
}

// UpdateActionInstData upsert action inst data.
func (h *Handler) UpdateActionInstData(nCtx contextx.IContext, actionInstData *action.InstanceData) error {
	if nCtx == nil {
		return errors.New("nCtx is nil")
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

	return h.dao.updateField(nCtx, filter, filed, data)
}

// UpdateActInstLifecycle update action inst lifecycle.
func (h *Handler) UpdateActInstLifecycle(
	nCtx contextx.IContext, operInstID, actionName string, lifecycle *action.Lifecycle) error {

	if nCtx == nil {
		return errors.New("nCtx is nil")
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
	err := h.dao.updateField(nCtx, filter, filed, ConvActInstLifeCycleToDB(lifecycle))
	if err != nil {
		return err
	}

	return nil
}

// GetActionInstData find one action inst data.
func (h *Handler) GetActionInstData(
	nCtx contextx.IContext, operInstID string, actionName string) (*action.InstanceData, error) {

	if nCtx == nil {
		return nil, errors.New("nCtx is nil")
	}

	if operInstID == "" {
		return nil, errors.New("operation instance id is empty")
	}

	if actionName == "" {
		return nil, errors.New("action name is empty")
	}

	filter := base.AliveFilter()
	opts := []OptFn{
		WithOperInstID(operInstID),
	}
	for _, opt := range opts {
		filter = opt(filter)
	}

	field := FieldKeyActionInstData(actionName)
	operInstData, err := h.dao.get(nCtx, filter, field)
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
		DisplayNameZh:       actionInstData.DisplayNameZh,
		DisplayNameEn:       actionInstData.DisplayNameEn,
		Index:               actionInstData.Index,
		TotalIndex:          actionInstData.TotalIndex,
		PrivateData:         actionInstData.PrivateData,
		Lifecycle:           ConvActInstLifeCycleFromDB(actionInstData.Lifecycle),
	}

	for _, msg := range actionInstData.Messages {
		data.Messages = append(data.Messages, common.Message{
			Time:   msg.Time,
			TextZh: msg.TextZh,
			TextEn: msg.TextEn,
			Level:  msg.Level,
		})
	}

	err = json.Unmarshal([]byte(actionInstData.Content), &data.Content)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// GetActInstLifecycle find one action inst data.
func (h *Handler) GetActInstLifecycle(
	nCtx contextx.IContext, operInstID string, actionName string) (*action.Lifecycle, error) {

	if nCtx == nil {
		return nil, errors.New("nCtx is nil")
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
	operInstData, err := h.dao.get(nCtx, filter, field)
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
func (h *Handler) UpdateActionInstStatus(
	nCtx contextx.IContext, operInstID string, actionName string, status action.State) error {

	if nCtx == nil {
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
	err := h.dao.updateField(nCtx, filter, field, status)
	if err != nil {
		return err
	}

	return nil
}
