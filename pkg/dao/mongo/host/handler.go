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

// IHandler host handler interface.
type IHandler interface {
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

	IDistinctor
}

// IDistinctor host distinct handler interface.
type IDistinctor interface {
	// DistinctNodeRole distincts with field node-role.
	DistinctNodeRole(ctx context.Context, opts ...OptFn) ([]types.NodeRole, error)

	// DistinctNodeStatus distincts with field node-status.
	DistinctNodeStatus(ctx context.Context, opts ...OptFn) ([]types.NodeStatus, error)

	// DistinctNodeVersion distincts with field node-version.
	DistinctNodeVersion(ctx context.Context, opts ...OptFn) ([]string, error)

	// DistinctDeptName distincts with field dept-name.
	DistinctDeptName(ctx context.Context, opts ...OptFn) ([]string, error)

	// DistinctOSType distincts with field os-type.
	DistinctOSType(ctx context.Context, opts ...OptFn) ([]string, error)

	// DistinctArch distincts with field arch.
	DistinctArch(ctx context.Context, opts ...OptFn) ([]string, error)

	// DistinctAddressing distincts with field addressing.
	DistinctAddressing(ctx context.Context, opts ...OptFn) ([]string, error)

	// DistinctNetworkAreaID distincts with field networkarea-id.
	DistinctNetworkAreaID(ctx context.Context, opts ...OptFn) ([]int64, error)

	// DistinctNetworkUnitID distincts with field networkunit-id.
	DistinctNetworkUnitID(ctx context.Context, opts ...OptFn) ([]int64, error)
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
	if err := newDaoClient.EnsureIndexes(); err != nil {
		h.logger.Warnf("failed to ensure host indexes, err: %v", errors.Join(base.ErrEnsureIndexesFailed(), err))
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao)
}

// New create a new host handler.
func New(client *mongo.Database, logger logger.Logger) IHandler {
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

	hosts, err := h.tenantDao(tenantID).List(ctx, base.AliveFilter(), nil)
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

	return h.tenantDao(tenantID).Count(ctx, filter)
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

	num, err := h.tenantDao(tenantID).Count(ctx, filter)
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

	hosts, err := h.tenantDao(tenantID).List(ctx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}

	data := make([]*types.Host, len(hosts))
	for idx, host := range hosts {
		data[idx] = convertHostToTypes(host)
	}

	return data, num, nil
}

// DistinctNodeRole distincts with field node-role.
func (h *handler) DistinctNodeRole(ctx context.Context, opts ...OptFn) ([]types.NodeRole, error) {
	result, err := h.distinctString(ctx, FieldKeyDynamicNodeRole, opts...)
	if err != nil {
		return nil, err
	}

	return types.StringListToNodeRoleList(result), nil
}

// DistinctNodeStatus distincts with field node-status.
func (h *handler) DistinctNodeStatus(ctx context.Context, opts ...OptFn) ([]types.NodeStatus, error) {
	result, err := h.distinctString(ctx, FieldKeyDynamicNodeStatus, opts...)
	if err != nil {
		return nil, err
	}

	return types.StringListToNodeStatusList(result), nil
}

// DistinctNodeVersion distincts with field node-version.
func (h *handler) DistinctNodeVersion(ctx context.Context, opts ...OptFn) ([]string, error) {
	return h.distinctString(ctx, FieldKeyDynamicNodeVersion, opts...)
}

// DistinctDeptName distincts with field dept-name.
func (h *handler) DistinctDeptName(ctx context.Context, opts ...OptFn) ([]string, error) {
	return h.distinctString(ctx, FieldKeyStaticDeptName, opts...)
}

// DistinctOSType distincts with field os-type.
func (h *handler) DistinctOSType(ctx context.Context, opts ...OptFn) ([]string, error) {
	return h.distinctString(ctx, FieldKeyStaticOSType, opts...)
}

// DistinctArch distincts with field arch.
func (h *handler) DistinctArch(ctx context.Context, opts ...OptFn) ([]string, error) {
	return h.distinctString(ctx, FieldKeyStaticArch, opts...)
}

// DistinctAddressing distincts with field addressing.
func (h *handler) DistinctAddressing(ctx context.Context, opts ...OptFn) ([]string, error) {
	return h.distinctString(ctx, FieldKeyStaticAddressing, opts...)
}

// DistinctNetworkAreaID distincts with field networkarea-id.
func (h *handler) DistinctNetworkAreaID(ctx context.Context, opts ...OptFn) ([]int64, error) {
	return h.distinctInt64(ctx, FieldKeyStaticNetworkAreaID, opts...)
}

// DistinctNetworkUnitID distincts with field networkunit-id.
func (h *handler) DistinctNetworkUnitID(ctx context.Context, opts ...OptFn) ([]int64, error) {
	return h.distinctInt64(ctx, FieldKeyDynamicNetworkUnitID, opts...)
}

// distinctInt64 returns distinct values of specified field.
func (h *handler) distinctInt64(ctx context.Context, key string, opts ...OptFn) ([]int64, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.tenantDao(tenantID).distinctInt64(ctx, key, filter, nil)
}

// distinctString returns distinct values of specified field.
func (h *handler) distinctString(ctx context.Context, key string, opts ...OptFn) ([]string, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.tenantDao(tenantID).distinctString(ctx, key, filter, nil)
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
			BizID:         host.Static.BizID,
			NetworkAreaID: host.Static.NetworkAreaID,
			HostName:      host.Static.HostName,
			DeptName:      host.Static.DeptName,
			InnerIP:       host.Static.InnerIP,
			InnerIPV6:     host.Static.InnerIPV6,
			OuterIP:       host.Static.OuterIP,
			OuterIPV6:     host.Static.OuterIPV6,
			Mac:           host.Static.Mac,
			OSType:        host.Static.OSType,
			Arch:          host.Static.Arch,
			Addressing:    string(host.Static.Addressing),
			SyncedAgentID: host.Static.SyncedAgentID,
		}
	}

	dynamic := &HostDynamic{}
	if host.Dynamic != nil {
		dynamic = &HostDynamic{
			NodeRole:       string(host.Dynamic.NodeRole),
			NodeStatus:     string(host.Dynamic.NodeStatus),
			NodeVersion:    host.Dynamic.NodeVersion,
			NodeGeneration: int64(host.Dynamic.NodeGeneration),
			AgentID:        host.Dynamic.AgentID,
			NetworkUnitID:  host.Dynamic.NetworkUnitID,
			ProxyTags: func() []string {
				tags := make([]string, len(host.Dynamic.ProxyTags))
				for idx := range host.Dynamic.ProxyTags {
					tags[idx] = string(host.Dynamic.ProxyTags[idx])
				}

				return tags
			}(),
			ProxyAccessDisabled: host.Dynamic.ProxyAccessDisabled,
			ProxyClusterPort:    host.Dynamic.ProxyClusterPort,
			ProxyDataPort:       host.Dynamic.ProxyDataPort,
			ProxyFilePort:       host.Dynamic.ProxyFilePort,
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
			Arch:          host.Static.Arch,
			Addressing:    types.Addressing(host.Static.Addressing),
			SyncedAgentID: host.Static.SyncedAgentID,
		}
	}

	dynamic := &types.HostDynamic{}
	if host.Dynamic != nil {
		dynamic = &types.HostDynamic{
			NodeRole:       types.NodeRole(host.Dynamic.NodeRole),
			NodeStatus:     types.NodeStatus(host.Dynamic.NodeStatus),
			NodeVersion:    host.Dynamic.NodeVersion,
			NodeGeneration: types.NodeGeneration(host.Dynamic.NodeGeneration),
			AgentID:        host.Dynamic.AgentID,
			NetworkUnitID:  host.Dynamic.NetworkUnitID,
			ProxyTags: func() []types.ProxyTag {
				tags := make([]types.ProxyTag, len(host.Dynamic.ProxyTags))
				for idx := range host.Dynamic.ProxyTags {
					tags[idx] = types.ProxyTag(host.Dynamic.ProxyTags[idx])
				}

				return tags
			}(),
			ProxyAccessDisabled: host.Dynamic.ProxyAccessDisabled,
			ProxyClusterPort:    host.Dynamic.ProxyClusterPort,
			ProxyDataPort:       host.Dynamic.ProxyDataPort,
			ProxyFilePort:       host.Dynamic.ProxyFilePort,
		}
	}

	return &types.Host{
		HostID:   host.HostID,
		TenantID: host.TenantID,
		Static:   static,
		Dynamic:  dynamic,
	}
}
