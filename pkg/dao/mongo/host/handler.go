/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package host ...
package host

import (
	"context"
	"errors"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Handler host handler interface.
type Handler interface {
	// ListAll lists all hosts.
	ListAll(ctx context.Context) ([]*types.Host, error)

	// Count count hosts by conditions.
	Count(ctx context.Context, opts ...OptFn) (int64, error)

	// List lists hosts by page and conditions.
	List(ctx context.Context, page types.Page, opts ...OptFn) ([]*types.Host, int64, error)

	// UpsertMany updates or inserts hosts.
	UpsertMany(ctx context.Context, hosts ...*types.Host) error

	// UpsertStaticMany updates or inserts host statics.
	UpsertStaticMany(ctx context.Context, hosts ...*types.Host) error

	// UpdateDynamicMany updates host dynamics. will not insert.
	UpdateDynamicMany(ctx context.Context, hosts ...*types.Host) error
}

type handler struct {
	client *mongo.Database
	logger logger.Logger
	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

func (h *handler) tenantDao(tenantID string) *dao {
	if d, ok := h.daoMap.Load(tenantID); ok {
		return d.(*dao)
	}

	newDaoClient := newDao(tenantID, h.client, h.logger)
	if err := newDaoClient.ensureIndexes(); err != nil {
		h.logger.Warnf("failed to ensure host indexes, err: %v", errors.Join(base.ErrEnsureIndexesFailed(), err))
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao)
}

// New create a new host handler.
func New(client *mongo.Database, logger logger.Logger) Handler {
	return &handler{
		client: client,
		logger: logger,
		daoMap: sync.Map{},
	}
}

// ListAll list all host.
func (h *handler) ListAll(ctx context.Context) ([]*types.Host, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	hosts, err := h.tenantDao(tenantID).listAll(ctx)
	if err != nil {
		return nil, err
	}

	data := make([]*types.Host, len(hosts))
	for idx, host := range hosts {
		data[idx] = convertHostToTypes(host)
	}

	return data, nil
}

// Count count host by conditions.
func (h *handler) Count(ctx context.Context, opts ...OptFn) (int64, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return 0, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.tenantDao(tenantID).count(ctx, filter)
}

// List list host by page and conditions.
func (h *handler) List(ctx context.Context, page types.Page, opts ...OptFn) (
	[]*types.Host, int64, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, 0, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.tenantDao(tenantID).count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := new(options.FindOptions)
	if page.Offset > 0 {
		findOpt.SetSkip(int64(page.Offset))
	}
	if page.Limit > 0 {
		findOpt.SetLimit(int64(page.Limit))
	}

	bizs, err := h.tenantDao(tenantID).list(ctx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}

	data := make([]*types.Host, len(bizs))
	for idx, host := range bizs {
		data[idx] = convertHostToTypes(host)
	}

	return data, num, nil
}

// UpsertMany updates or inserts hosts.
func (h *handler) UpsertMany(ctx context.Context, hosts ...*types.Host) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	if len(hosts) == 0 {
		return base.ErrEmptyParamData()
	}

	data := make([]*Host, len(hosts))
	for idx, host := range hosts {
		if host == nil {
			return base.ErrInvalidItemInParamList()
		}

		data[idx] = convertHostFromTypes(host)

		if err = base.CheckTenantIDMatched(tenantID, data[idx].TenantID); err != nil {
			return err
		}
	}

	if err := h.tenantDao(tenantID).upsertMany(ctx, data); err != nil {
		return err
	}

	return nil
}

// UpsertStaticMany updates or inserts host statics.
func (h *handler) UpsertStaticMany(ctx context.Context, hosts ...*types.Host) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	if len(hosts) == 0 {
		return base.ErrEmptyParamData()
	}

	data := make([]*Host, len(hosts))
	for idx, host := range hosts {
		if host == nil {
			return base.ErrInvalidItemInParamList()
		}

		data[idx] = convertHostFromTypes(host)

		if err = base.CheckTenantIDMatched(tenantID, data[idx].TenantID); err != nil {
			return err
		}
	}

	if err := h.tenantDao(tenantID).upsertStaticMany(ctx, data); err != nil {
		return err
	}

	return nil
}

// UpdateDynamicMany updates host dynamics. will not insert.
func (h *handler) UpdateDynamicMany(ctx context.Context, hosts ...*types.Host) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	data := make([]*Host, len(hosts))
	for idx, host := range hosts {
		if host == nil {
			return base.ErrInvalidItemInParamList()
		}

		data[idx] = convertHostFromTypes(host)

		if err = base.CheckTenantIDMatched(tenantID, data[idx].TenantID); err != nil {
			return err
		}
	}

	if err := h.tenantDao(tenantID).updateDynamicMany(ctx, data); err != nil {
		return err
	}

	return nil
}

func convertHostFromTypes(host *types.Host) *Host {
	static := &HostStatic{}
	if host.Static != nil {
		static = &HostStatic{
			NetworkAreaID: host.Static.NetworkAreaID,
			BizID:         host.Static.BizID,
			HostName:      host.Static.HostName,
			DeptName:      host.Static.DeptName,
			InnerIP:       host.Static.InnerIP,
			InnerIPV6:     host.Static.InnerIPV6,
			OuterIP:       host.Static.OuterIP,
			OuterIPV6:     host.Static.OuterIPV6,
			Mac:           host.Static.Mac,
			OSType:        host.Static.OSType,
		}
	}

	dynamic := &HostDynamic{}
	if host.Dynamic != nil {
		dynamic = &HostDynamic{
			NodeRole:         string(host.Dynamic.NodeRole),
			NodeStatus:       string(host.Dynamic.NodeStatus),
			NodeVersion:      host.Dynamic.NodeVersion,
			NodeGeneration:   host.Dynamic.NodeGeneration,
			AgentID:          host.Dynamic.AgentID,
			NetworkUnitID:    host.Dynamic.NetworkUnitID,
			Tag:              host.Dynamic.Tag,
			ProxyClusterPort: host.Dynamic.ProxyClusterPort,
			ProxyDataPort:    host.Dynamic.ProxyDataPort,
			ProxyFilePort:    host.Dynamic.ProxyFilePort,
		}
	}

	return &Host{
		HostID:   host.HostID,
		TenantID: host.TenantID,
		Static:   static,
		Dynamic:  dynamic,
	}
}

func convertHostToTypes(host *Host) *types.Host {
	static := &types.HostStatic{}
	if host.Static != nil {
		static = &types.HostStatic{
			NetworkAreaID: host.Static.NetworkAreaID,
			BizID:         host.Static.BizID,
			HostName:      host.Static.HostName,
			DeptName:      host.Static.DeptName,
			InnerIP:       host.Static.InnerIP,
			InnerIPV6:     host.Static.InnerIPV6,
			OuterIP:       host.Static.OuterIP,
			OuterIPV6:     host.Static.OuterIPV6,
			Mac:           host.Static.Mac,
			OSType:        host.Static.OSType,
		}
	}

	dynamic := &types.HostDynamic{}
	if host.Dynamic != nil {
		dynamic = &types.HostDynamic{
			NodeRole:         types.NodeRole(host.Dynamic.NodeRole),
			NodeStatus:       types.NodeStatus(host.Dynamic.NodeStatus),
			NodeVersion:      host.Dynamic.NodeVersion,
			NodeGeneration:   host.Dynamic.NodeGeneration,
			AgentID:          host.Dynamic.AgentID,
			NetworkUnitID:    host.Dynamic.NetworkUnitID,
			Tag:              host.Dynamic.Tag,
			ProxyClusterPort: host.Dynamic.ProxyClusterPort,
			ProxyDataPort:    host.Dynamic.ProxyDataPort,
			ProxyFilePort:    host.Dynamic.ProxyFilePort,
		}
	}

	return &types.Host{
		HostID:   host.HostID,
		TenantID: host.TenantID,
		Static:   static,
		Dynamic:  dynamic,
	}
}
