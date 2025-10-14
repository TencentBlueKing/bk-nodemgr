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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/common"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// IOperationInstData this define the crud interface.
// nolint: interfacebloat
type IOperationInstData interface {
	// Upsert updates or inserts an OperInstData.
	Upsert(nCtx contextx.IContext, data *operation.InstanceData) error

	// FindOne find one OperInstData.
	FindOne(nCtx contextx.IContext, opts ...OptFn) (*operation.InstanceData, error)

	// Count count OperationDatas by conditions.
	Count(nCtx contextx.IContext, opts ...OptFn) (int64, error)

	// ListFullData find all full OperInstData.
	ListFullData(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*operation.InstanceData, int64, error)

	// ListWithoutActInst find all OperInstData whithout actions.
	ListWithoutActInst(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*operation.InstanceBriefData, int64, error)

	// FindOneWithoutActionData find one OperInstData without action data.
	FindOneWithoutActionData(nCtx contextx.IContext, opts ...OptFn) (*operation.InstanceData, error)

	// UpdateLifeCycle updates or inserts an InstanceData's LifeCycle.
	UpdateLifeCycle(nCtx contextx.IContext, operInstID string, lifeCycle *operation.Lifecycle) error

	// UpdateExtraExecutionMessages updates operation instance extra execution messages.
	UpdateExtraExecutionMessages(nCtx contextx.IContext, operInstID string, messages ...common.Message) error

	// ListAllLastOperInst find all last OperInstData in their operation.
	ListAllLastOperInst(nCtx contextx.IContext, opts ...OptFn) ([]*operation.InstanceBriefData, error)

	// Delete deletes operation instance data by given operation instance IDs.
	Delete(nCtx contextx.IContext, operInstIDs ...string) error

	// DeleteByTriggerID deletes operation instance data by given trigger IDs.
	DeleteByTriggerID(nCtx contextx.IContext, triggerIDs ...string) error
}

// Upsert updates or inserts an OperInstData.
func (h *Handler) Upsert(nCtx contextx.IContext, data *operation.InstanceData) error {
	if nCtx == nil {
		return errors.New("nCtx is nil")
	}

	if data == nil {
		return errors.New("data is nil")
	}

	operInstData, err := ConvOpInstanceDataToDB(data)
	if err != nil {
		return fmt.Errorf("convert oper inst data to db data error: %v", err)
	}

	err = h.dao.upsert(nCtx, operInstData)
	if err != nil {
		return err
	}

	return nil
}

// FindOne find one OperInstData.
func (h *Handler) FindOne(nCtx contextx.IContext, opts ...OptFn) (*operation.InstanceData, error) {
	if nCtx == nil {
		return nil, errors.New("nCtx is nil")
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	operInstDatas, err := h.dao.find(nCtx, filter)
	if err != nil {
		return nil, err
	}

	if len(operInstDatas) == 0 {
		return nil, errors.New("not found")
	}
	operInstData := operInstDatas[0]

	data, err := ConvAOperaInstDataWithoutActionFromDB(operInstData)
	if err != nil {
		return nil, err
	}
	data.ActionInstanceDataMap = make(map[string]*action.InstanceData, len(operInstData.ActionInstDataMap))
	// nolint: varnamelen
	for k, v := range operInstData.ActionInstDataMap {
		actionInstData := &action.InstanceData{
			TriggerID:           v.TriggerID,
			OperationInstanceID: v.OperInstID,
			OperationID:         v.OperationID,
			OperationDefName:    v.OperDefName,
			Name:                v.Name,
			Index:               v.Index,
			TotalIndex:          v.TotalIndex,
			Messages:            make([]common.Message, len(v.Messages)),
			Content:             make(map[string]any, len(v.Content)),
			PrivateData:         v.PrivateData,
			Lifecycle:           ConvActInstLifeCycleFromDB(v.Lifecycle),
		}

		for idx, msg := range v.Messages {
			actionInstData.Messages[idx] = common.Message{
				Time:  msg.Time,
				Text:  msg.Text,
				Level: msg.Level,
			}
		}

		if len(v.Content) != 0 {
			err = json.Unmarshal([]byte(v.Content), &actionInstData.Content)
			if err != nil {
				return nil, err
			}
		}

		data.ActionInstanceDataMap[k] = actionInstData
	}

	return data, nil
}

// ListFullData find all OperInstData.
func (h *Handler) ListFullData(nCtx contextx.IContext, page types.Page, opts ...OptFn) (
	[]*operation.InstanceData, int64, error) {

	if nCtx == nil {
		return nil, 0, errors.New("nCtx is nil")
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.dao.Count(nCtx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := base.ParsePage(page)

	operaInstDatas, err := h.dao.List(nCtx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}

	data := make([]*operation.InstanceData, len(operaInstDatas))
	for idx, opera := range operaInstDatas {
		data[idx], err = ConvAOperaInstDataWithoutActionFromDB(opera)
		if err != nil {
			return nil, 0, err
		}
		data[idx].ActionInstanceDataMap = make(map[string]*action.InstanceData, len(opera.ActionInstDataMap))

		for k, v := range opera.ActionInstDataMap {
			actionInstData := &action.InstanceData{
				TriggerID:           v.TriggerID,
				OperationInstanceID: v.OperInstID,
				OperationID:         v.OperationID,
				OperationDefName:    v.OperDefName,
				Name:                v.Name,
				Index:               v.Index,
				TotalIndex:          v.TotalIndex,
				Messages:            make([]common.Message, len(v.Messages)),
				Content:             make(map[string]any, len(v.Content)),
				PrivateData:         v.PrivateData,
				Lifecycle:           ConvActInstLifeCycleFromDB(v.Lifecycle),
			}

			for idx, msg := range v.Messages {
				actionInstData.Messages[idx] = common.Message{
					Time:  msg.Time,
					Text:  msg.Text,
					Level: msg.Level,
				}
			}

			err = json.Unmarshal([]byte(v.Content), &actionInstData.Content)
			if err != nil {
				return nil, 0, err
			}

			data[idx].ActionInstanceDataMap[k] = actionInstData
		}
	}

	return data, num, nil
}

// FindOneWithoutActionData find InstanceData without action data.
func (h *Handler) FindOneWithoutActionData(nCtx contextx.IContext, opts ...OptFn) (*operation.InstanceData, error) {
	if nCtx == nil {
		return nil, errors.New("nCtx is nil")
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	field := FieldOfActionData
	InstanceDatas, err := h.dao.findWithoutFields(nCtx, filter, types.SingleItemPage(), field)
	if err != nil {
		return nil, err
	}

	if len(InstanceDatas) == 0 {
		return nil, errors.New("not found")
	}

	return ConvAOperaInstDataWithoutActionFromDB(InstanceDatas[0])
}

// ListWithoutActInst List InstanceDatas without action data.
func (h *Handler) ListWithoutActInst(nCtx contextx.IContext, page types.Page, opts ...OptFn) (
	[]*operation.InstanceBriefData, int64, error) {

	if nCtx == nil {
		return nil, 0, errors.New("nCtx is nil")
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.dao.Count(nCtx, filter)
	if err != nil {
		return nil, 0, err
	}

	if num == 0 {
		return nil, 0, nil
	}

	field := FieldOfActionData
	operaInstDatas, err := h.dao.findWithoutFields(nCtx, filter, page, field)
	if err != nil {
		return nil, 0, err
	}

	if len(operaInstDatas) == 0 {
		return nil, 0, errors.New("not found")
	}

	data := make([]*operation.InstanceBriefData, len(operaInstDatas))
	for idx, opera := range operaInstDatas {
		data[idx], err = ConvOpeInstBriefDataFromDB(opera)
		if err != nil {
			return nil, 0, err
		}
	}

	return data, num, nil
}

// Count count OperationDatas by conditions.
func (h *Handler) Count(nCtx contextx.IContext, opts ...OptFn) (int64, error) {
	if nCtx == nil {
		return 0, errors.New("nCtx is nil")
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.dao.Count(nCtx, filter)
}

// UpdateLifeCycle update operation instance's LifeCycle.
func (h *Handler) UpdateLifeCycle(nCtx contextx.IContext, operInstID string, lifeCycle *operation.Lifecycle) error {
	if nCtx == nil {
		return errors.New("nCtx is nil")
	}

	if operInstID == "" {
		return errors.New("operation instance id is empty")
	}

	if lifeCycle == nil {
		return errors.New("LifeCycle is nil")
	}

	filter := base.AliveFilter()
	opts := []OptFn{
		WithOperInstID(operInstID),
	}
	for _, opt := range opts {
		filter = opt(filter)
	}

	err := h.dao.updateField(nCtx, filter, FieldKeyLifeCycle, ConvOperaLifeCycleToDB(lifeCycle))
	if err != nil {
		return err
	}

	return nil
}

// UpdateExtraExecutionMessages updates operation instance extra execution messages.
func (h *Handler) UpdateExtraExecutionMessages(
	nCtx contextx.IContext, operInstID string, messages ...common.Message) error {

	if nCtx == nil {
		return errors.New("nCtx is nil")
	}

	if operInstID == "" {
		return errors.New("operation instance id is empty")
	}

	if len(messages) == 0 {
		return errors.New("messages is empty")
	}

	filter := base.AliveFilter()
	opts := []OptFn{
		WithOperInstID(operInstID),
	}
	for _, opt := range opts {
		filter = opt(filter)
	}

	for _, msg := range convMessageToDB(messages) {
		err := h.dao.pushField(nCtx, filter, "extra_execution_messages", msg)
		if err != nil {
			return err
		}
	}

	return nil
}

// ListAllLastOperInst find all last OperInstData in their operation.
func (h *Handler) ListAllLastOperInst(nCtx contextx.IContext, opts ...OptFn) (
	[]*operation.InstanceBriefData, error) {

	if nCtx == nil {
		return nil, errors.New("nCtx is nil")
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	operaInstDatas, err := h.dao.listALLLastOperInst(nCtx, filter)
	if err != nil {
		return nil, err
	}
	data := make([]*operation.InstanceBriefData, len(operaInstDatas))
	for idx, opera := range operaInstDatas {
		data[idx], err = ConvOpeInstBriefDataFromDB(opera)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

// Delete deletes operation instance data by given operation instance IDs.
func (h *Handler) Delete(nCtx contextx.IContext, operInstIDs ...string) error {
	if nCtx == nil {
		return errors.New("nCtx is nil")
	}

	if len(operInstIDs) == 0 {
		return errors.New("operation instance ids is empty")
	}

	return h.dao.delete(nCtx, operInstIDs...)
}

// DeleteByTriggerID deletes operation instance data by given trigger IDs.
func (h *Handler) DeleteByTriggerID(nCtx contextx.IContext, triggerIDs ...string) error {
	if nCtx == nil {
		return errors.New("nCtx is nil")
	}

	if len(triggerIDs) == 0 {
		return errors.New("trigger ids is empty")
	}

	return h.dao.deleteByTriggerID(nCtx, triggerIDs...)
}
