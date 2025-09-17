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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler host handler interface.
type IHandler interface {
	// ListAll lists all hosts.
	ListAll(ctx context.Context) ([]*types.Host, error)

	// Count count hosts by conditions.
	Count(ctx context.Context, opts ...OptFn) (int64, error)

	// List lists hosts by page and conditions.
	List(ctx context.Context, page types.Page, opts ...OptFn) ([]*types.Host, int64, error)

	// DeleteMany deletes hosts by hostIDs.
	DeleteMany(ctx context.Context, hostIDs ...int64) error

	// UpsertMany updates or inserts hosts.
	UpsertMany(ctx context.Context, hosts ...*types.Host) error

	// UpsertStaticMany updates or inserts host statics.
	UpsertStaticMany(ctx context.Context, hosts ...*types.Host) error

	// UpdateDynamicMany updates host dynamics. will not insert.
	UpdateDynamicMany(ctx context.Context, hosts ...*types.Host) error

	// FindWithDynamic finds hosts with dynamic fields.
	FindWithDynamic(ctx context.Context, page types.Page, opts ...OptFn) ([]*types.Host, error)

	// UpdateDynamicFields updates host dynamic fields.
	UpdateDynamicFields(ctx context.Context, fields types.HostDynamicFields, hosts ...*types.Host) error

	IDistinctor
}

// IDistinctor host distinct handler interface.
type IDistinctor interface {
	// DistinctBizID distincts with field biz-id.
	DistinctBizID(ctx context.Context, opts ...OptFn) ([]int64, error)

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
	logger logger.ILogger
	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

func (h *handler) tenantDao(tenantID string) *dao {
	if d, ok := h.daoMap.Load(tenantID); ok {
		return d.(*dao) // nolint: forcetypeassert
	}

	newDaoClient := newDao(tenantID, h.client, h.logger)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		h.logger.Warnf("failed to ensure host indexes: %v", errors.Join(base.ErrEnsureIndexesFailed(), err))
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao) // nolint: forcetypeassert
}

// New create a new host handler.
func New(client *mongo.Database, logger logger.ILogger) IHandler {
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

	findOpt := base.ParsePage(page)

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

// DistinctBizID distincts with field biz-id.
func (h *handler) DistinctBizID(ctx context.Context, opts ...OptFn) ([]int64, error) {
	result, err := h.distinctInt64(ctx, FieldKeyStaticBizID, opts...)
	if err != nil {
		return nil, err
	}

	return result, nil
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
			OSTypeCCID:    host.Static.OSTypeCCID,
			Arch:          host.Static.Arch,
			Addressing:    string(host.Static.Addressing),
			RegionID:      host.Static.RegionID,
			CityID:        host.Static.CityID,
			SyncedAgentID: host.Static.SyncedAgentID,
		}
	}

	dynamic := &HostDynamic{}
	if host.Dynamic != nil {
		dynamic = &HostDynamic{
			NodeRole:            string(host.Dynamic.NodeRole),
			NodeStatus:          string(host.Dynamic.NodeStatus),
			NodeVersion:         host.Dynamic.NodeVersion,
			NodeGeneration:      int64(host.Dynamic.NodeGeneration),
			NodeCPUArch:         string(host.Dynamic.NodeCPUArch),
			NodeOsType:          string(host.Dynamic.NodeOsType),
			AgentID:             host.Dynamic.AgentID,
			NetworkUnitID:       host.Dynamic.NetworkUnitID,
			ProxyAccessDisabled: host.Dynamic.ProxyAccessDisabled,
			ProxyTags: func() []string {
				tags := make([]string, len(host.Dynamic.ProxyTags))
				for idx := range host.Dynamic.ProxyTags {
					tags[idx] = string(host.Dynamic.ProxyTags[idx])
				}

				return tags
			}(),
			ProxyClusterPort:  host.Dynamic.ProxyClusterPort,
			ProxyDataPort:     host.Dynamic.ProxyDataPort,
			ProxyFilePort:     host.Dynamic.ProxyFilePort,
			LoginIP:           host.Dynamic.LoginIP,
			LoginPort:         host.Dynamic.LoginPort,
			LoginUser:         host.Dynamic.LoginUser,
			LoginMode:         string(host.Dynamic.LoginMode),
			LoginCreditID:     host.Dynamic.LoginCreditID,
			ExportIP:          host.Dynamic.ExportIP,
			AdvertiseIP:       host.Dynamic.AdvertiseIP,
			RelayDownloadPort: host.Dynamic.RelayDownloadPort,
			RelayCallbackPort: host.Dynamic.RelayCallbackPort,
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
			BizID:         host.Static.BizID,
			NetworkAreaID: host.Static.NetworkAreaID,
			RegionID:      host.Static.RegionID,
			CityID:        host.Static.CityID,
			HostName:      host.Static.HostName,
			DeptName:      host.Static.DeptName,
			InnerIP:       host.Static.InnerIP,
			InnerIPV6:     host.Static.InnerIPV6,
			OuterIP:       host.Static.OuterIP,
			OuterIPV6:     host.Static.OuterIPV6,
			Mac:           host.Static.Mac,
			OSTypeCCID:    host.Static.OSTypeCCID,
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
			NodeGeneration: types.Generation(host.Dynamic.NodeGeneration),
			NodeCPUArch:    criteria.CPUArch(host.Dynamic.NodeCPUArch),
			NodeOsType:     criteria.OSType(host.Dynamic.NodeOsType),
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
			LoginIP:             host.Dynamic.LoginIP,
			LoginPort:           host.Dynamic.LoginPort,
			LoginUser:           host.Dynamic.LoginUser,
			LoginMode:           types.LoginMode(host.Dynamic.LoginMode),
			LoginCreditID:       host.Dynamic.LoginCreditID,
			ExportIP:            host.Dynamic.ExportIP,
			AdvertiseIP:         host.Dynamic.AdvertiseIP,
			RelayDownloadPort:   host.Dynamic.RelayDownloadPort,
			RelayCallbackPort:   host.Dynamic.RelayCallbackPort,
		}
	}

	return &types.Host{
		HostID:   host.HostID,
		TenantID: host.TenantID,
		Static:   static,
		Dynamic:  dynamic,
	}
}

// DeleteMany delete many hosts.
func (h *handler) DeleteMany(ctx context.Context, hostIDs ...int64) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	if len(hostIDs) == 0 {
		return base.ErrEmptyParamData()
	}

	filter := base.AliveFilter()
	filter = WithHostID(hostIDs...)(filter)
	if err := h.tenantDao(tenantID).DeleteMany(ctx, filter); err != nil {
		return err
	}

	return nil
}

// FindWithDynamic finds hosts with dynamic fields.
func (h *handler) FindWithDynamic(ctx context.Context, page types.Page, opts ...OptFn) ([]*types.Host, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	findOpt := base.ParsePage(page)
	findOpt.SetProjection(bson.D{{Key: FieldKeyHostID, Value: 1}, {Key: FieldKeyDynamic, Value: 1}})

	hosts, err := h.tenantDao(tenantID).List(ctx, filter, findOpt)
	if err != nil {
		return nil, err
	}

	data := make([]*types.Host, len(hosts))
	for idx, host := range hosts {
		data[idx] = convertHostToTypes(host)
	}

	return data, nil
}

// UpdateDynamicFields updates host dynamic fields.
func (h *handler) UpdateDynamicFields(ctx context.Context, fields types.HostDynamicFields, hosts ...*types.Host) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	docs := make([]*base.DocumentFieldUpdate, 0, len(hosts))
	for _, host := range hosts {
		if host == nil {
			return base.ErrInvalidItemInParamList()
		}

		docs = append(docs, &base.DocumentFieldUpdate{
			Filter: bson.D{bson.E{Key: FieldKeyHostID, Value: host.HostID}},
			Fields: generateHostDynamicUpdates(fields, host),
		})
	}

	if err := h.tenantDao(tenantID).UpdateFieldsBulk(ctx, docs); err != nil {
		h.logger.Errorf("failed to update host dynamic version and status: %v", err)
		return err
	}

	return nil
}

func generateHostDynamicUpdates(fields types.HostDynamicFields, host *types.Host) map[string]any {
	updates := make(map[string]any)

	if fields.NodeRole {
		updates[FieldKeyDynamicNodeRole] = host.Dynamic.NodeRole
	}
	if fields.NodeStatus {
		updates[FieldKeyDynamicNodeStatus] = host.Dynamic.NodeStatus
	}
	if fields.NodeVersion {
		updates[FieldKeyDynamicNodeVersion] = host.Dynamic.NodeVersion
	}
	if fields.NodeGeneration {
		updates[FieldKeyDynamicNodeGeneration] = host.Dynamic.NodeGeneration
	}
	if fields.NodeCPUArch {
		updates[FieldKeyDynamicNodeCPUArch] = host.Dynamic.NodeCPUArch
	}
	if fields.NodeOsType {
		updates[FieldKeyDynamicNodeOsType] = host.Dynamic.NodeOsType
	}
	if fields.AgentID {
		updates[FieldKeyDynamicAgentID] = host.Dynamic.AgentID
	}
	if fields.NetworkUnitID {
		updates[FieldKeyDynamicNetworkUnitID] = host.Dynamic.NetworkUnitID
	}

	if fields.LoginIP {
		updates[FieldKeyDynamicLoginIP] = host.Dynamic.LoginIP
	}
	if fields.LoginPort {
		updates[FieldKeyDynamicLoginPort] = host.Dynamic.LoginPort
	}
	if fields.LoginUser {
		updates[FieldKeyDynamicLoginUser] = host.Dynamic.LoginUser
	}
	if fields.LoginMode {
		updates[FieldKeyDynamicLoginMode] = host.Dynamic.LoginMode
	}
	if fields.LoginCreditID {
		updates[FieldKeyDynamicLoginCreditID] = host.Dynamic.LoginCreditID
	}
	if fields.ExportIP {
		updates[FieldKeyDynamicExportIP] = host.Dynamic.ExportIP
	}
	if fields.AdvertiseIP {
		updates[FieldKeyDynamicAdvertiseIP] = host.Dynamic.AdvertiseIP
	}

	if fields.ProxyTags {
		updates[FieldKeyDynamicProxyTags] = host.Dynamic.ProxyTags
	}
	if fields.ProxyClusterPort {
		updates[FieldKeyDynamicProxyClusterPort] = host.Dynamic.ProxyClusterPort
	}
	if fields.ProxyFilePort {
		updates[FieldKeyDynamicProxyFilePort] = host.Dynamic.ProxyFilePort
	}
	if fields.ProxyDataPort {
		updates[FieldKeyDynamicProxyDataPort] = host.Dynamic.ProxyDataPort
	}
	if fields.RelayDownloadPort {
		updates[FieldKeyDynamicRelayDownloadPort] = host.Dynamic.RelayDownloadPort
	}
	if fields.RelayCallbackPort {
		updates[FieldKeyDynamicRelayCallbackPort] = host.Dynamic.RelayCallbackPort
	}

	return updates
}
