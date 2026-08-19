/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package packageworkflow

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"

	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler defines package workflow DAO operations.
type IHandler interface {
	// Get gets package workflow by id.
	Get(nCtx contextx.IContext, workflowID string) (*types.PackageWorkflow, error)

	// GetStatus gets the status of a package workflow.
	GetStatus(nCtx contextx.IContext, workflowID string) (types.PackageWorkflowStatus, error)

	// Count counts package workflow by opts.
	Count(nCtx contextx.IContext, opts ...OptFn) (int64, error)

	// List lists package workflow by page and opts.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.PackageWorkflow, int64, error)

	// Create creates a new package workflow.
	Create(nCtx contextx.IContext, workflow *types.PackageWorkflow) error

	// UpdateStatus updates the status of a package workflow.
	UpdateStatus(nCtx contextx.IContext, workflowID string, status types.PackageWorkflowStatus) error

	// UpdateFinishTime updates the finish time of a package workflow.
	UpdateFinishTime(nCtx contextx.IContext, workflowID string, finishTime time.Time) error
}

// Handler operates package workflow tables.
type Handler struct {
	client *mongo.Database
	daoMap sync.Map
}

func (h *Handler) tenantDao(tenantID string) *dao {
	if d, ok := h.daoMap.Load(tenantID); ok {
		return d.(*dao) // nolint: forcetypeassert
	}

	newDaoClient := newDao(h.client, tenantID)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).With("tenant-id", tenantID).Warn("failed to ensure package workflow indexes")
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	return d.(*dao) // nolint: forcetypeassert
}

// New creates a package workflow Handler.
func New(client *mongo.Database) *Handler {
	return &Handler{
		client: client,
		daoMap: sync.Map{},
	}
}

// Count counts package workflow by opts.
func (h *Handler) Count(nCtx contextx.IContext, opts ...OptFn) (int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return 0, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.tenantDao(tenantID).Count(nCtx, filter)
}

// List lists package workflow by page and opts.
func (h *Handler) List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.PackageWorkflow, int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.tenantDao(tenantID).Count(nCtx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := base.ParsePage(page)

	datas, err := h.tenantDao(tenantID).List(nCtx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}

	workflows := make([]*types.PackageWorkflow, len(datas))
	for idx, data := range datas {
		workflows[idx] = convertPackageWorkflowToTypes(data)
	}

	return workflows, num, nil
}

// Create creates a new package workflow.
func (h *Handler) Create(nCtx contextx.IContext, workflow *types.PackageWorkflow) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()

	if workflow == nil {
		return base.ErrEmptyParamData()
	}

	if workflow.TenantID != "" && workflow.TenantID != tenantID {
		return fmt.Errorf("workflow tenant id %s does not match context tenant id %s", workflow.TenantID, tenantID)
	}

	if workflow.WorkflowID == "" {
		return errors.New("workflow id should not be empty")
	}

	if workflow.TriggerID == "" {
		return errors.New("trigger id should not be empty")
	}

	return h.tenantDao(tenantID).Create(nCtx, convertPackageWorkflowFromTypes(workflow))
}

// Get gets package workflow by id.
func (h *Handler) Get(nCtx contextx.IContext, workflowID string) (*types.PackageWorkflow, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()

	if workflowID == "" {
		return nil, errors.New("workflow id should not be empty")
	}

	filter := base.AliveFilter()
	filter = WithWorkflowID(workflowID)(filter)
	data, err := h.tenantDao(tenantID).Get(nCtx, filter)
	if err != nil {
		return nil, err
	}

	return convertPackageWorkflowToTypes(data), nil
}

// GetStatus gets the status of a package workflow.
func (h *Handler) GetStatus(nCtx contextx.IContext, workflowID string) (types.PackageWorkflowStatus, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return "", err
	}

	tenantID := nCtx.TenantID()

	if workflowID == "" {
		return "", errors.New("workflow id should not be empty")
	}

	filter := base.AliveFilter()
	filter = WithWorkflowID(workflowID)(filter)
	data, err := h.tenantDao(tenantID).Get(nCtx, filter)
	if err != nil {
		return "", err
	}

	return types.PackageWorkflowStatus(data.Status), nil
}

// UpdateStatus updates the status of a package workflow.
func (h *Handler) UpdateStatus(nCtx contextx.IContext, workflowID string, status types.PackageWorkflowStatus) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()

	if workflowID == "" {
		return errors.New("workflow id should not be empty")
	}

	if err := status.Validate(); err != nil {
		return err
	}

	filter := base.AliveFilter()
	filter = WithWorkflowID(workflowID)(filter)

	return h.tenantDao(tenantID).UpdateField(nCtx, filter, FieldKeyStatus, string(status))
}

// UpdateFinishTime updates the finish time of a package workflow.
func (h *Handler) UpdateFinishTime(nCtx contextx.IContext, workflowID string, finishTime time.Time) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()

	if workflowID == "" {
		return errors.New("workflow id should not be empty")
	}

	if finishTime.IsZero() {
		return errors.New("finish time should not be zero")
	}

	filter := base.AliveFilter()
	filter = WithWorkflowID(workflowID)(filter)

	return h.tenantDao(tenantID).UpdateField(nCtx, filter, FieldKeyFinishTime, finishTime)
}

func convertPackageWorkflowToTypes(data *Data) *types.PackageWorkflow {
	if data == nil {
		return nil
	}

	return &types.PackageWorkflow{
		TenantID:    data.TenantID,
		WorkflowID:  data.WorkflowID,
		TriggerID:   data.TriggerID,
		Type:        types.PackageWorkflowType(data.Type),
		Operator:    data.Operator,
		OperateTime: data.OperateTime,
		FinishTime:  data.FinishTime,
		Status:      types.PackageWorkflowStatus(data.Status),
	}
}

func convertPackageWorkflowFromTypes(workflow *types.PackageWorkflow) *Data {
	if workflow == nil {
		return nil
	}

	return &Data{
		TenantID:    workflow.TenantID,
		WorkflowID:  workflow.WorkflowID,
		TriggerID:   workflow.TriggerID,
		Type:        string(workflow.Type),
		Operator:    workflow.Operator,
		OperateTime: workflow.OperateTime,
		FinishTime:  workflow.FinishTime,
		Status:      string(workflow.Status),
	}
}
