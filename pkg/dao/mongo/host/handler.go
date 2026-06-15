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
	"fmt"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler host handler interface.
// nolint: interfacebloat
type IHandler interface {
	// ListAll lists all hosts.
	ListAll(nCtx contextx.IContext) ([]*types.Host, error)

	// Count count hosts by conditions.
	Count(nCtx contextx.IContext, opts ...OptFn) (int64, error)

	// CountGroupByNetworkUnitID count hosts by networkunit id.
	CountGroupByNetworkUnitID(nCtx contextx.IContext, opts ...OptFn) (map[int64]int64, error)

	// CountGroupByBizID count hosts by biz id.
	CountGroupByBizID(nCtx contextx.IContext, opts ...OptFn) (map[int64]int64, error)

	// CountGroupBySetID count hosts by set id.
	CountGroupBySetID(nCtx contextx.IContext, opts ...OptFn) (map[int64]int64, error)

	// CountGroupByModuleID count hosts by module id.
	CountGroupByModuleID(nCtx contextx.IContext, opts ...OptFn) (map[int64]int64, error)

	// Exist check a host exist by conditions.
	Exist(nCtx contextx.IContext, opts ...OptFn) (bool, error)

	// List lists hosts by page and conditions.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.Host, int64, error)

	// DeleteMany deletes hosts by hostIDs.
	DeleteMany(nCtx contextx.IContext, hostIDs ...int64) error

	// UpsertMany updates or inserts hosts.
	UpsertMany(nCtx contextx.IContext, hosts ...*types.Host) error

	// UpsertStaticMany updates or inserts host statics.
	UpsertStaticMany(nCtx contextx.IContext, hosts ...*types.Host) error

	// UpdateDynamicMany updates host dynamics. will not insert.
	UpdateDynamicMany(nCtx contextx.IContext, hosts ...*types.Host) error

	// FindWithDynamic finds hosts with dynamic fields.
	FindWithDynamic(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.Host, error)

	// UpdateStaticFields updates host static fields.
	UpdateStaticFields(nCtx contextx.IContext, fields types.HostStaticFields, hosts ...*types.Host) error

	// UpdateDynamicFields updates host dynamic fields.
	UpdateDynamicFields(nCtx contextx.IContext, fields types.HostDynamicFields, hosts ...*types.Host) error

	// GetHostDistributionByNodeRole get host distribution by node role.
	GetHostDistributionByNodeRole(nCtx contextx.IContext, opts ...OptFn) (map[string]int64, error)

	// GetHostDistributionByNetworkAreaID get host distribution by network area id.
	GetHostDistributionByNetworkAreaID(nCtx contextx.IContext, opts ...OptFn) (map[int64]int64, error)

	// GetHostDistributionByNodeVersion get host distribution by node version.
	GetHostDistributionByNodeVersion(nCtx contextx.IContext, opts ...OptFn) (map[string]int64, error)

	// ListWithFields lists hosts with fields.
	ListWithFields(nCtx contextx.IContext, page types.Page, selection *types.HostFieldSelection, opts ...OptFn) ([]*types.Host, int64, error)

	// GetRelayInfosInNetworkUnit gets available Relay Infos in the specified network unit.
	// Returns RelayInfo list with DedicatedInstaller tag and Running status.
	// Uses MongoDB projection to only query required fields (6 fields instead of 40+).
	GetRelayInfosInNetworkUnit(nCtx contextx.IContext, networkUnitID int64) ([]*types.RelayInfo, error)

	// TouchOperationUpdatedAt sets operation_updated_at to now for the given host IDs.
	TouchOperationUpdatedAt(nCtx contextx.IContext, hostIDs ...int64) error

	IDistinctor
}

// IDistinctor host distinct handler interface.
type IDistinctor interface {
	// DistinctBizID distincts with field biz-id.
	DistinctBizID(nCtx contextx.IContext, opts ...OptFn) ([]int64, error)

	// DistinctNodeRole distincts with field node-role.
	DistinctNodeRole(nCtx contextx.IContext, opts ...OptFn) ([]types.NodeRole, error)

	// DistinctNodeStatus distincts with field node-status.
	DistinctNodeStatus(nCtx contextx.IContext, opts ...OptFn) ([]types.NodeStatus, error)

	// DistinctNodeVersion distincts with field node-version.
	DistinctNodeVersion(nCtx contextx.IContext, opts ...OptFn) ([]string, error)

	// DistinctDeptName distincts with field dept-name.
	DistinctDeptName(nCtx contextx.IContext, opts ...OptFn) ([]string, error)

	// DistinctOSType distincts with field os-type.
	DistinctOSType(nCtx contextx.IContext, opts ...OptFn) ([]string, error)

	// DistinctArch distincts with field arch.
	DistinctArch(nCtx contextx.IContext, opts ...OptFn) ([]string, error)

	// DistinctAddressing distincts with field addressing.
	DistinctAddressing(nCtx contextx.IContext, opts ...OptFn) ([]string, error)

	// DistinctNetworkAreaID distincts with field networkarea-id.
	DistinctNetworkAreaID(nCtx contextx.IContext, opts ...OptFn) ([]int64, error)

	// DistinctNetworkUnitID distincts with field networkunit-id.
	DistinctNetworkUnitID(nCtx contextx.IContext, opts ...OptFn) ([]int64, error)
}

type handler struct {
	client *mongo.Database
	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

func (h *handler) tenantDao(tenantID string) *dao {
	if d, ok := h.daoMap.Load(tenantID); ok {
		return d.(*dao) // nolint: forcetypeassert
	}

	newDaoClient := newDao(tenantID, h.client)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).With("tenant-id", tenantID).Warn("failed to ensure host indexes")
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao) // nolint: forcetypeassert
}

// New create a new host handler.
func New(client *mongo.Database) IHandler {
	return &handler{
		client: client,
		daoMap: sync.Map{},
	}
}

// ListAll list all host.
func (h *handler) ListAll(nCtx contextx.IContext) ([]*types.Host, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()

	hosts, err := h.tenantDao(tenantID).List(nCtx, base.AliveFilter(), nil)
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
func (h *handler) Count(nCtx contextx.IContext, opts ...OptFn) (int64, error) {
	if nCtx == nil {
		return 0, base.ErrInvalidContext()
	}

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

// CountGroupByNetworkUnitID count hosts by networkunit id.
func (h *handler) CountGroupByNetworkUnitID(nCtx contextx.IContext, opts ...OptFn) (map[int64]int64, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.tenantDao(tenantID).CountGroupByInt64(nCtx, filter, FieldKeyDynamicNetworkUnitID)
}

// CountGroupByBizID count hosts by biz id.
func (h *handler) CountGroupByBizID(nCtx contextx.IContext, opts ...OptFn) (map[int64]int64, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.tenantDao(tenantID).CountGroupByInt64(nCtx, filter, FieldKeyStaticBizID)
}

// CountGroupBySetID count hosts by set id.
func (h *handler) CountGroupBySetID(nCtx contextx.IContext, opts ...OptFn) (map[int64]int64, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.tenantDao(tenantID).CountGroupByInt64(nCtx, filter, FieldKeyStaticSetID)
}

// CountGroupByModuleID count hosts by module id.
func (h *handler) CountGroupByModuleID(nCtx contextx.IContext, opts ...OptFn) (map[int64]int64, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.tenantDao(tenantID).CountGroupByInt64(nCtx, filter, FieldKeyStaticModuleID)
}

// Exist count host by conditions.
func (h *handler) Exist(nCtx contextx.IContext, opts ...OptFn) (bool, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return false, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.tenantDao(tenantID).Exist(nCtx, filter)
}

// List list host by page and conditions.
func (h *handler) List(nCtx contextx.IContext, page types.Page, opts ...OptFn) (
	[]*types.Host, int64, error) {

	if nCtx == nil {
		return nil, 0, base.ErrInvalidContext()
	}

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

	hosts, err := h.tenantDao(tenantID).List(nCtx, filter, findOpt)
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
func (h *handler) DistinctBizID(nCtx contextx.IContext, opts ...OptFn) ([]int64, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	result, err := h.distinctInt64(nCtx, FieldKeyStaticBizID, opts...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DistinctNodeRole distincts with field node-role.
func (h *handler) DistinctNodeRole(nCtx contextx.IContext, opts ...OptFn) ([]types.NodeRole, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	result, err := h.distinctString(nCtx, FieldKeyDynamicNodeRole, opts...)
	if err != nil {
		return nil, err
	}

	return types.StringListToNodeRoleList(result), nil
}

// DistinctNodeStatus distincts with field node-status.
func (h *handler) DistinctNodeStatus(nCtx contextx.IContext, opts ...OptFn) ([]types.NodeStatus, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	result, err := h.distinctString(nCtx, FieldKeyDynamicNodeStatus, opts...)
	if err != nil {
		return nil, err
	}

	return types.StringListToNodeStatusList(result), nil
}

// DistinctNodeVersion distincts with field node-version.
func (h *handler) DistinctNodeVersion(nCtx contextx.IContext, opts ...OptFn) ([]string, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	return h.distinctString(nCtx, FieldKeyDynamicNodeVersion, opts...)
}

// DistinctDeptName distincts with field dept-name.
func (h *handler) DistinctDeptName(nCtx contextx.IContext, opts ...OptFn) ([]string, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	return h.distinctString(nCtx, FieldKeyStaticDeptName, opts...)
}

// DistinctOSType distincts with field os-type.
func (h *handler) DistinctOSType(nCtx contextx.IContext, opts ...OptFn) ([]string, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	return h.distinctString(nCtx, FieldKeyStaticOSType, opts...)
}

// DistinctArch distincts with field arch.
func (h *handler) DistinctArch(nCtx contextx.IContext, opts ...OptFn) ([]string, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	return h.distinctString(nCtx, FieldKeyStaticArch, opts...)
}

// DistinctAddressing distincts with field addressing.
func (h *handler) DistinctAddressing(nCtx contextx.IContext, opts ...OptFn) ([]string, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	return h.distinctString(nCtx, FieldKeyStaticAddressing, opts...)
}

// DistinctNetworkAreaID distincts with field networkarea-id.
func (h *handler) DistinctNetworkAreaID(nCtx contextx.IContext, opts ...OptFn) ([]int64, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	return h.distinctInt64(nCtx, FieldKeyStaticNetworkAreaID, opts...)
}

// DistinctNetworkUnitID distincts with field networkunit-id.
func (h *handler) DistinctNetworkUnitID(nCtx contextx.IContext, opts ...OptFn) ([]int64, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	return h.distinctInt64(nCtx, FieldKeyDynamicNetworkUnitID, opts...)
}

// distinctInt64 returns distinct values of specified field.
func (h *handler) distinctInt64(nCtx contextx.IContext, key string, opts ...OptFn) ([]int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.tenantDao(tenantID).distinctInt64(nCtx, key, filter, nil)
}

// distinctString returns distinct values of specified field.
func (h *handler) distinctString(nCtx contextx.IContext, key string, opts ...OptFn) ([]string, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.tenantDao(tenantID).distinctString(nCtx, key, filter, nil)
}

// UpsertMany updates or inserts hosts.
func (h *handler) UpsertMany(nCtx contextx.IContext, hosts ...*types.Host) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()

	if len(hosts) == 0 {
		return base.ErrEmptyParamData()
	}

	data := make([]*Host, len(hosts))
	for idx, host := range hosts {
		if host == nil {
			return base.ErrInvalidItemInParamList()
		}

		data[idx] = convertHostFromTypes(host)

		if err := base.CheckTenantIDMatched(tenantID, data[idx].TenantID); err != nil {
			return fmt.Errorf("failed to upsert host(%d): %w", host.HostID, err)
		}
	}

	if err := h.tenantDao(tenantID).upsertMany(nCtx, data); err != nil {
		return err
	}

	return nil
}

// UpsertStaticMany updates or inserts host statics.
func (h *handler) UpsertStaticMany(nCtx contextx.IContext, hosts ...*types.Host) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()

	if len(hosts) == 0 {
		return base.ErrEmptyParamData()
	}

	data := make([]*Host, len(hosts))
	for idx, host := range hosts {
		if host == nil {
			return base.ErrInvalidItemInParamList()
		}

		data[idx] = convertHostFromTypes(host)

		if err := base.CheckTenantIDMatched(tenantID, data[idx].TenantID); err != nil {
			return err
		}
	}

	if err := h.tenantDao(tenantID).upsertStaticMany(nCtx, data); err != nil {
		return err
	}

	return nil
}

// UpdateDynamicMany updates host dynamics. will not insert.
func (h *handler) UpdateDynamicMany(nCtx contextx.IContext, hosts ...*types.Host) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()

	data := make([]*Host, len(hosts))
	for idx, host := range hosts {
		if host == nil {
			return base.ErrInvalidItemInParamList()
		}

		data[idx] = convertHostFromTypes(host)

		if err := base.CheckTenantIDMatched(tenantID, data[idx].TenantID); err != nil {
			return err
		}
	}

	if err := h.tenantDao(tenantID).updateDynamicMany(nCtx, data); err != nil {
		return err
	}

	return nil
}

func convertHostFromTypes(host *types.Host) *Host {
	static := &HostStatic{}
	if host.Static != nil {
		static = &HostStatic{
			BizID:                    host.Static.BizID,
			SetID:                    host.Static.SetID,
			ModuleID:                 host.Static.ModuleID,
			NetworkAreaID:            host.Static.NetworkAreaID,
			HostName:                 host.Static.HostName,
			DeptName:                 host.Static.DeptName,
			InnerIPList:              host.Static.InnerIPList,
			InnerIPV6List:            host.Static.InnerIPV6List,
			OuterIPList:              host.Static.OuterIPList,
			OuterIPV6List:            host.Static.OuterIPV6List,
			Mac:                      host.Static.Mac,
			Operator:                 host.Static.Operator,
			OSType:                   host.Static.OSType,
			OSTypeCCID:               host.Static.OSTypeCCID,
			Arch:                     host.Static.Arch,
			Addressing:               string(host.Static.Addressing),
			RegionID:                 host.Static.RegionID,
			CityID:                   host.Static.CityID,
			CPUNum:                   host.Static.CPUNum,
			MemCap:                   host.Static.MemCap,
			SyncedAgentID:            host.Static.SyncedAgentID,
			SyncedOpsConsoleHostID:   host.Static.SyncedOpsConsoleHostID,
			SyncedOpsOutBandType:     host.Static.SyncedOpsOutBandType,
			SyncedOpsOutBandProtocol: host.Static.SyncedOpsOutBandProtocol,
			SyncedOpsBMCIP:           host.Static.SyncedOpsBMCIP,
			SyncedOpsBMCPort:         host.Static.SyncedOpsBMCPort,
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
			ProxyClusterPort:         host.Dynamic.ProxyClusterPort,
			ProxyDataPort:            host.Dynamic.ProxyDataPort,
			ProxyFilePort:            host.Dynamic.ProxyFilePort,
			LoginIP:                  host.Dynamic.LoginIP,
			LoginPort:                host.Dynamic.LoginPort,
			LoginUser:                host.Dynamic.LoginUser,
			LoginMode:                string(host.Dynamic.LoginMode),
			LoginCreditID:            host.Dynamic.LoginCreditID,
			ExportIP:                 host.Dynamic.ExportIP,
			ExportIPV6:               host.Dynamic.ExportIPV6,
			AdvertiseIP:              host.Dynamic.AdvertiseIP,
			AdvertiseIPV6:            host.Dynamic.AdvertiseIPV6,
			RelayDownloadPort:        host.Dynamic.RelayDownloadPort,
			RelayCallbackPort:        host.Dynamic.RelayCallbackPort,
			ProxyInstallOriginUnitID: host.Dynamic.ProxyInstallOriginUnitID,
			ConnCycleTime:            host.Dynamic.ConnCycleTime,
			OpsConsoleHostID:         host.Dynamic.OpsConsoleHostID,
			OpsOutBandType:           host.Dynamic.OpsOutBandType,
			OpsOutBandProtocol:       host.Dynamic.OpsOutBandProtocol,
			OpsBMCIP:                 host.Dynamic.OpsBMCIP,
			OpsBMCPort:               host.Dynamic.OpsBMCPort,
		}
	}

	result := &Host{
		HostID:   host.HostID,
		TenantID: host.TenantID,
		Static:   static,
		Dynamic:  dynamic,
	}

	if !host.OperationUpdatedAt.IsZero() {
		t := host.OperationUpdatedAt
		result.OperationUpdatedAt = &t
	}

	return result
}

func convertHostToTypes(host *Host) *types.Host {
	static := &types.HostStatic{}
	if host.Static != nil {
		static = &types.HostStatic{
			BizID:                    host.Static.BizID,
			SetID:                    host.Static.SetID,
			ModuleID:                 host.Static.ModuleID,
			NetworkAreaID:            host.Static.NetworkAreaID,
			RegionID:                 host.Static.RegionID,
			CityID:                   host.Static.CityID,
			HostName:                 host.Static.HostName,
			DeptName:                 host.Static.DeptName,
			InnerIPList:              host.Static.InnerIPList,
			InnerIPV6List:            host.Static.InnerIPV6List,
			OuterIPList:              host.Static.OuterIPList,
			OuterIPV6List:            host.Static.OuterIPV6List,
			Operator:                 host.Static.Operator,
			Mac:                      host.Static.Mac,
			OSTypeCCID:               host.Static.OSTypeCCID,
			OSType:                   host.Static.OSType,
			Arch:                     host.Static.Arch,
			Addressing:               types.Addressing(host.Static.Addressing),
			CPUNum:                   host.Static.CPUNum,
			MemCap:                   host.Static.MemCap,
			SyncedAgentID:            host.Static.SyncedAgentID,
			SyncedOpsConsoleHostID:   host.Static.SyncedOpsConsoleHostID,
			SyncedOpsOutBandType:     host.Static.SyncedOpsOutBandType,
			SyncedOpsOutBandProtocol: host.Static.SyncedOpsOutBandProtocol,
			SyncedOpsBMCIP:           host.Static.SyncedOpsBMCIP,
			SyncedOpsBMCPort:         host.Static.SyncedOpsBMCPort,
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
			ProxyAccessDisabled:      host.Dynamic.ProxyAccessDisabled,
			ProxyClusterPort:         host.Dynamic.ProxyClusterPort,
			ProxyDataPort:            host.Dynamic.ProxyDataPort,
			ProxyFilePort:            host.Dynamic.ProxyFilePort,
			LoginIP:                  host.Dynamic.LoginIP,
			LoginPort:                host.Dynamic.LoginPort,
			LoginUser:                host.Dynamic.LoginUser,
			LoginMode:                types.LoginMode(host.Dynamic.LoginMode),
			LoginCreditID:            host.Dynamic.LoginCreditID,
			ExportIP:                 host.Dynamic.ExportIP,
			ExportIPV6:               host.Dynamic.ExportIPV6,
			AdvertiseIP:              host.Dynamic.AdvertiseIP,
			AdvertiseIPV6:            host.Dynamic.AdvertiseIPV6,
			RelayDownloadPort:        host.Dynamic.RelayDownloadPort,
			RelayCallbackPort:        host.Dynamic.RelayCallbackPort,
			ProxyInstallOriginUnitID: host.Dynamic.ProxyInstallOriginUnitID,
			ConnCycleTime:            host.Dynamic.ConnCycleTime,
			OpsConsoleHostID:         host.Dynamic.OpsConsoleHostID,
			OpsOutBandType:           host.Dynamic.OpsOutBandType,
			OpsOutBandProtocol:       host.Dynamic.OpsOutBandProtocol,
			OpsBMCIP:                 host.Dynamic.OpsBMCIP,
			OpsBMCPort:               host.Dynamic.OpsBMCPort,
		}
	}

	result := &types.Host{
		HostID:   host.HostID,
		TenantID: host.TenantID,
		Static:   static,
		Dynamic:  dynamic,
	}

	if host.OperationUpdatedAt != nil {
		result.OperationUpdatedAt = *host.OperationUpdatedAt
	}

	return result
}

// TouchOperationUpdatedAt sets operation_updated_at to the current time for
// the given host IDs. This marks hosts as recently operated by a user or API
// action, which drives the host list sort order.
func (h *handler) TouchOperationUpdatedAt(nCtx contextx.IContext, hostIDs ...int64) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	if len(hostIDs) == 0 {
		return nil
	}

	tenantID := nCtx.TenantID()
	nowTime := time.Now()

	d := h.tenantDao(tenantID)
	filter := base.AliveFilter()
	filter = WithHostID(hostIDs...)(filter)

	// UpdateField also sets basic.updated_at and basic.is_deleted=false (ORM
	// buildUpdateField); business touches therefore refresh the document-wide
	// updated time as well as operation_updated_at.
	return d.UpdateField(nCtx, filter, FieldKeyOperationUpdatedAt, nowTime)
}

// DeleteMany delete many hosts.
func (h *handler) DeleteMany(nCtx contextx.IContext, hostIDs ...int64) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()

	if len(hostIDs) == 0 {
		return base.ErrEmptyParamData()
	}

	filter := base.AliveFilter()
	filter = WithHostID(hostIDs...)(filter)
	if err := h.tenantDao(tenantID).DeleteMany(nCtx, filter); err != nil {
		return err
	}

	return nil
}

// FindWithDynamic finds hosts with dynamic fields.
func (h *handler) FindWithDynamic(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.Host, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	findOpt := base.ParsePage(page)
	findOpt.SetProjection(bson.D{{Key: FieldKeyHostID, Value: 1}, {Key: FieldKeyDynamic, Value: 1}})

	hosts, err := h.tenantDao(tenantID).List(nCtx, filter, findOpt)
	if err != nil {
		return nil, err
	}

	data := make([]*types.Host, len(hosts))
	for idx, host := range hosts {
		data[idx] = convertHostToTypes(host)
	}

	return data, nil
}

// UpdateStaticFields updates host static fields.
func (h *handler) UpdateStaticFields(nCtx contextx.IContext, fields types.HostStaticFields, hosts ...*types.Host) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()

	docs := make([]*base.DocumentFieldUpdate, 0, len(hosts))
	for _, host := range hosts {
		if host == nil || host.Static == nil {
			return base.ErrInvalidItemInParamList()
		}

		updates := generateHostStaticUpdates(fields, host)
		docs = append(docs, &base.DocumentFieldUpdate{
			Filter: bson.D{{Key: FieldKeyHostID, Value: host.HostID}},
			Fields: updates,
		})
	}

	if err := h.tenantDao(tenantID).UpdateFieldsBulk(nCtx, docs); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to update host static fields")

		return err
	}

	return nil
}

// nolint: gocognit, gocyclo, cyclop
func generateHostStaticUpdates(fields types.HostStaticFields, host *types.Host) map[string]any {
	updates := make(map[string]any)

	if fields.BizID {
		updates[FieldKeyStaticBizID] = host.Static.BizID
	}
	if fields.SetID {
		updates[FieldKeyStaticSetID] = host.Static.SetID
	}
	if fields.ModuleID {
		updates[FieldKeyStaticModuleID] = host.Static.ModuleID
	}
	if fields.NetworkAreaID {
		updates[FieldKeyStaticNetworkAreaID] = host.Static.NetworkAreaID
	}
	if fields.RegionID {
		updates[FieldKeyStaticRegionID] = host.Static.RegionID
	}
	if fields.CityID {
		updates[FieldKeyStaticCityID] = host.Static.CityID
	}
	if fields.HostName {
		updates[FieldKeyStaticHostName] = host.Static.HostName
	}
	if fields.DeptName {
		updates[FieldKeyStaticDeptName] = host.Static.DeptName
	}
	if fields.InnerIPList {
		updates[FieldKeyStaticInnerIPList] = host.Static.InnerIPList
	}
	if fields.InnerIPV6List {
		updates[FieldKeyStaticInnerIPV6List] = host.Static.InnerIPV6List
	}
	if fields.OuterIPList {
		updates[FieldKeyStaticOuterIPList] = host.Static.OuterIPList
	}
	if fields.OuterIPV6List {
		updates[FieldKeyStaticOuterIPV6List] = host.Static.OuterIPV6List
	}
	if fields.Operator {
		updates[FieldKeyStaticOperator] = host.Static.Operator
	}
	if fields.Mac {
		updates[FieldKeyStaticMac] = host.Static.Mac
	}
	if fields.OSTypeCCID {
		updates[FieldKeyStaticOSTypeCCID] = host.Static.OSTypeCCID
	}
	if fields.OSType {
		updates[FieldKeyStaticOSType] = host.Static.OSType
	}
	if fields.Arch {
		updates[FieldKeyStaticArch] = host.Static.Arch
	}
	if fields.Addressing {
		updates[FieldKeyStaticAddressing] = host.Static.Addressing
	}
	if fields.CPUNum {
		updates[FieldKeyStaticCPUNum] = host.Static.CPUNum
	}
	if fields.MemCap {
		updates[FieldKeyStaticMemCap] = host.Static.MemCap
	}
	if fields.SyncedAgentID {
		updates[FieldKeyStaticSyncedAgentID] = host.Static.SyncedAgentID
	}
	if fields.SyncedOpsConsoleHostID {
		updates[FieldKeyStaticSyncedOpsConsoleHostID] = host.Static.SyncedOpsConsoleHostID
	}
	if fields.SyncedOpsOutBandType {
		updates[FieldKeyStaticSyncedOpsOutBandType] = host.Static.SyncedOpsOutBandType
	}
	if fields.SyncedOpsOutBandProtocol {
		updates[FieldKeyStaticSyncedOpsOutBandProtocol] = host.Static.SyncedOpsOutBandProtocol
	}
	if fields.SyncedOpsBMCIP {
		updates[FieldKeyStaticSyncedOpsBMCIP] = host.Static.SyncedOpsBMCIP
	}
	if fields.SyncedOpsBMCPort {
		updates[FieldKeyStaticSyncedOpsBMCPort] = host.Static.SyncedOpsBMCPort
	}

	return updates
}

func (h *handler) UpdateDynamicFields(nCtx contextx.IContext, fields types.HostDynamicFields, hosts ...*types.Host) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()

	docs := make([]*base.DocumentFieldUpdate, 0, len(hosts))
	for _, host := range hosts {
		if host == nil || host.Dynamic == nil {
			return base.ErrInvalidItemInParamList()
		}

		updates := generateHostDynamicUpdates(fields, host)
		docs = append(docs, &base.DocumentFieldUpdate{
			Filter: bson.D{{Key: FieldKeyHostID, Value: host.HostID}},
			Fields: updates,
		})
	}

	if err := h.tenantDao(tenantID).UpdateFieldsBulk(nCtx, docs); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to update host dynamic version and status")

		return err
	}

	return nil
}

// nolint: gocognit, gocyclo, cyclop
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
	if fields.ExportIPV6 {
		updates[FieldKeyDynamicExportIPV6] = host.Dynamic.ExportIPV6
	}
	if fields.AdvertiseIP {
		updates[FieldKeyDynamicAdvertiseIP] = host.Dynamic.AdvertiseIP
	}
	if fields.AdvertiseIPV6 {
		updates[FieldKeyDynamicAdvertiseIPV6] = host.Dynamic.AdvertiseIPV6
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

	if fields.ProxyInstallOriginUnitID {
		updates[FieldKeyDynamicProxyInstallOriginUnitID] = host.Dynamic.ProxyInstallOriginUnitID
	}

	if fields.ConnCycleTime {
		updates[FieldKeyDynamicConnCycleTime] = host.Dynamic.ConnCycleTime
	}

	if fields.OpsConsoleHostID {
		updates[FieldKeyDynamicOpsConsoleHostID] = host.Dynamic.OpsConsoleHostID
	}
	if fields.OpsOutBandType {
		updates[FieldKeyDynamicOpsOutBandType] = host.Dynamic.OpsOutBandType
	}
	if fields.OpsOutBandProtocol {
		updates[FieldKeyDynamicOpsOutBandProtocol] = host.Dynamic.OpsOutBandProtocol
	}
	if fields.OpsBMCIP {
		updates[FieldKeyDynamicOpsBMCIP] = host.Dynamic.OpsBMCIP
	}
	if fields.OpsBMCPort {
		updates[FieldKeyDynamicOpsBMCPort] = host.Dynamic.OpsBMCPort
	}

	return updates
}

// GetHostDistributionByNodeRole gets the host distribution by node role.
func (h *handler) GetHostDistributionByNodeRole(nCtx contextx.IContext, opts ...OptFn) (map[string]int64, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, fmt.Errorf("failed to get host distribution by node role: %w", err)
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	results, err := h.tenantDao(tenantID).getHostDistributionByNodeRole(nCtx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get host distribution by node role: %w", err)
	}

	nodeRoleDistribution := make(map[string]int64)
	for _, result := range results {
		nodeRoleDistribution[result.NodeRole] = result.HostCount
	}

	return nodeRoleDistribution, nil
}

// GetHostDistributionByNetworkAreaID gets the host distribution by network area id.
func (h *handler) GetHostDistributionByNetworkAreaID(nCtx contextx.IContext, opts ...OptFn) (map[int64]int64, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, fmt.Errorf("failed to get host distribution by network area id: %w", err)
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	results, err := h.tenantDao(tenantID).getHostDistributionByNetworkAreaID(nCtx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get host distribution by network area id: %w", err)
	}

	nodeRoleDistribution := make(map[int64]int64)
	for _, result := range results {
		nodeRoleDistribution[result.NetworkAreaID] = result.HostCount
	}

	return nodeRoleDistribution, nil
}

// GetHostDistributionByNodeVersion gets the host distribution by node version.
func (h *handler) GetHostDistributionByNodeVersion(nCtx contextx.IContext, opts ...OptFn) (map[string]int64, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, fmt.Errorf("failed to get host distribution by node version: %w", err)
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	results, err := h.tenantDao(tenantID).getHostDistributionByNodeVersion(nCtx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get host distribution by node version: %w", err)
	}

	nodeVersionDistribution := make(map[string]int64)
	for _, result := range results {
		nodeVersionDistribution[result.NodeVersion] = result.HostCount
	}

	return nodeVersionDistribution, nil
}

// ListWithFields lists hosts with fields.
func (h *handler) ListWithFields(nCtx contextx.IContext, page types.Page, selection *types.HostFieldSelection, opts ...OptFn) (
	[]*types.Host, int64, error) {

	if nCtx == nil {
		return nil, 0, base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, fmt.Errorf("failed to list hosts with fields: %w", err)
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

	fields := convertHostFieldSelectionToFields(selection)

	findOpt := base.ParsePage(page)

	hosts, err := h.tenantDao(tenantID).List(nCtx, filter, findOpt, fields...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list hosts with fields: %w", err)
	}

	data := make([]*types.Host, len(hosts))
	for idx, host := range hosts {
		data[idx] = convertHostToTypes(host)
	}

	return data, num, nil
}

func convertHostFieldSelectionToFields(selection *types.HostFieldSelection) []string {
	fields := make([]string, 0)
	if selection.HostID {
		fields = append(fields, FieldKeyHostID)
	}
	if selection.BizID {
		fields = append(fields, FieldKeyStaticBizID)
	}
	if selection.NetworkAreaID {
		fields = append(fields, FieldKeyStaticNetworkAreaID)
	}
	if selection.InnerIPList {
		fields = append(fields, FieldKeyStaticInnerIPList)
	}
	if selection.InnerIPV6List {
		fields = append(fields, FieldKeyStaticInnerIPV6List)
	}
	if selection.LoginUser {
		fields = append(fields, FieldKeyDynamicLoginUser)
	}

	return fields
}

// GetRelayInfosInNetworkUnit gets available Relay Infos in the specified network unit.
// Returns RelayInfo list with DedicatedInstaller tag and Running status.
// Uses MongoDB projection to only query required fields (6 fields instead of 40+).
func (h *handler) GetRelayInfosInNetworkUnit(nCtx contextx.IContext, networkUnitID int64) ([]*types.RelayInfo, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, fmt.Errorf("failed to check tenant id: %w", err)
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	filter = WithDynamicNetworkUnitID(networkUnitID)(filter)
	filter = WithDynamicNodeRole(types.NodeRoleProxy)(filter)
	filter = WithDynamicNodeStatus(types.NodeStatusRunning)(filter)
	filter = WithDynamicProxyTags(types.ProxyTagDedicatedInstaller)(filter)

	fields := []string{
		FieldKeyHostID,
		FieldKeyDynamicAgentID,
		FieldKeyDynamicAdvertiseIP,
		FieldKeyDynamicAdvertiseIPV6,
		FieldKeyDynamicRelayDownloadPort,
		FieldKeyDynamicRelayCallbackPort,
	}

	hosts, err := h.tenantDao(tenantID).List(nCtx, filter, nil, fields...)
	if err != nil {
		return nil, fmt.Errorf("failed to get relay hosts: %w", err)
	}

	relayInfos := make([]*types.RelayInfo, 0, len(hosts))
	for _, host := range hosts {
		relayInfo := &types.RelayInfo{
			HostID: host.HostID,
		}

		if host.Dynamic != nil {
			relayInfo.AgentID = host.Dynamic.AgentID
			relayInfo.DownloadSvcPort = host.Dynamic.RelayDownloadPort
			relayInfo.CallbackSvcPort = host.Dynamic.RelayCallbackPort
			relayInfo.AdvertiseIP = host.Dynamic.AdvertiseIP
			relayInfo.AdvertiseIPV6 = host.Dynamic.AdvertiseIPV6
		}

		relayInfos = append(relayInfos, relayInfo)
	}

	return relayInfos, nil
}
