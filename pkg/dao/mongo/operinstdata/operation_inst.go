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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// IOperationInstData this define the crud interface.
type IOperationInstData interface {
	// Upsert updates or inserts an OperInstData.
	Upsert(ctx context.Context, data *operation.InstanceData) error

	// FindOne find one OperInstData.
	FindOne(ctx context.Context, opts ...OptFn) (*operation.InstanceData, error)

	// Count count OperationDatas by conditions.
	Count(ctx context.Context, opts ...OptFn) (int64, error)

	// ListFullData find all full OperInstData.
	ListFullData(ctx context.Context, page types.Page, opts ...OptFn) ([]*operation.InstanceData, int64, error)

	// ListWithoutActInst find all OperInstData whithout actions.
	ListWithoutActInst(ctx context.Context, page types.Page, opts ...OptFn) ([]*operation.InstanceBriefData, int64, error)

	// FindOneWithoutActionData find one OperInstData without action data.
	FindOneWithoutActionData(ctx context.Context, opts ...OptFn) (*operation.InstanceData, error)

	// UpdateLifeCycle updates or inserts an InstanceData's LifeCycle.
	UpdateLifeCycle(ctx context.Context, operInstID string, LifeCycle *operation.Lifecycle) error
}

// Upsert updates or inserts an OperInstData.
func (h *handler) Upsert(ctx context.Context, data *operation.InstanceData) error {
	if ctx == nil {
		return errors.New("ctx is nil")
	}

	if data == nil {
		return errors.New("data is nil")
	}

	operInstData, err := ConvOpInstanceDataToDB(data)
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
func (h *handler) FindOne(ctx context.Context, opts ...OptFn) (*operation.InstanceData, error) {
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
			Messages:            make([]action.Message, len(v.Messages)),
			Content:             make(map[string]any, len(v.Content)),
			PrivateData:         v.PrivateData,
			Lifecycle:           ConvActInstLifeCycleFromDB(v.Lifecycle),
		}

		for idx, msg := range v.Messages {
			actionInstData.Messages[idx] = action.Message{
				Time: msg.Time,
				Text: msg.Text,
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

// List find all OperInstData.
func (h *handler) ListFullData(ctx context.Context, page types.Page, opts ...OptFn) (
	[]*operation.InstanceData, int64, error) {

	if ctx == nil {
		return nil, 0, errors.New("ctx is nil")
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.dao.baseOrm.Count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := base.ParsePage(page)

	operaInstDatas, err := h.dao.baseOrm.List(ctx, filter, findOpt)
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
				Messages:            make([]action.Message, len(v.Messages)),
				Content:             make(map[string]any, len(v.Content)),
				PrivateData:         v.PrivateData,
				Lifecycle:           ConvActInstLifeCycleFromDB(v.Lifecycle),
			}

			for idx, msg := range v.Messages {
				actionInstData.Messages[idx] = action.Message{
					Time: msg.Time,
					Text: msg.Text,
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
func (h *handler) FindOneWithoutActionData(ctx context.Context, opts ...OptFn) (*operation.InstanceData, error) {
	if ctx == nil {
		return nil, errors.New("ctx is nil")
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	field := FieldOfActionData
	InstanceDatas, err := h.dao.findWithoutFields(ctx, filter, types.SingleItemPage(), field)
	if err != nil {
		return nil, err
	}

	if len(InstanceDatas) == 0 {
		return nil, errors.New("not found")
	}

	return ConvAOperaInstDataWithoutActionFromDB(InstanceDatas[0])
}

// ListWithoutActInst List InstanceDatas without action data.
func (h *handler) ListWithoutActInst(ctx context.Context, page types.Page, opts ...OptFn) (
	[]*operation.InstanceBriefData, int64, error) {

	if ctx == nil {
		return nil, 0, errors.New("ctx is nil")
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.dao.baseOrm.Count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	if num == 0 {
		return nil, 0, nil
	}

	field := FieldOfActionData
	operaInstDatas, err := h.dao.findWithoutFields(ctx, filter, page, field)
	if err != nil {
		return nil, 0, err
	}

	if len(operaInstDatas) == 0 {
		return nil, 0, errors.New("not found")
	}

	data := make([]*operation.InstanceBriefData, len(operaInstDatas))
	for idx, opera := range operaInstDatas {
		data[idx], err = ConvAOpeInstBreiefDataFromDB(opera)
		if err != nil {
			return nil, 0, err
		}
	}

	return data, num, nil
}

// Count count OperationDatas by conditions.
func (h *handler) Count(ctx context.Context, opts ...OptFn) (int64, error) {
	if ctx == nil {
		return 0, errors.New("ctx is nil")
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.dao.baseOrm.Count(ctx, filter)
}

// UpdateLifeCycle update operation instance's LifeCycle.
func (h *handler) UpdateLifeCycle(ctx context.Context, operInstID string, lifeCycle *operation.Lifecycle) error {
	if ctx == nil {
		return errors.New("ctx is nil")
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

	err := h.dao.updateField(ctx, filter, FieldKeyLifeCycle, ConvOperaLifeCycleToDB(lifeCycle))
	if err != nil {
		return err
	}

	return nil
}
