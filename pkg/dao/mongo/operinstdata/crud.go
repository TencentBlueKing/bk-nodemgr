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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
)

// ICrud this define the crud interface.
type ICrud interface {
	// Upsert updates or inserts an OperInstData.
	Upsert(ctx context.Context, data *operengine.OperInstData) error

	// FindOne find one OperInstData.
	FindOne(ctx context.Context, opts ...OptFn) (*operengine.OperInstData, error)
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
		OperationID:       operInstData.OperationID,
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

	// nolint: varnamelen
	for k, v := range operInstData.ActionInstDataMap {
		actionInstData := &operengine.ActionInstData{
			TriggerID:   v.TriggerID,
			OperInstID:  v.OperInstID,
			OperationID: v.OperationID,
			OperDefName: v.OperDefName,
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
