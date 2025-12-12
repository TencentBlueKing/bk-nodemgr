/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package backend

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// IHandler is interface for nodeman backend Handler.
// nolint: interfacebloat
type IHandler interface {
	IHandlerHost
	IHandlerNetworkArea
	IHandlerNetworkUnit
	IHandlerNodeAgent
	IHandlerNodeProxy
	IHandlerNodeWorkflow
	IHandlerRelease
	IHandlerReleasePlugin
	IHandlerPackage
	IHandlerConfigPolicy
	IHandlerPlugin
	IHandlerProcess
	IHandlerConfigPolicyEvent
	IHandlerGraph

	// ListBusiness list business within specified tenant in contextx.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return business list with page and the total count with filter.
	ListBusiness(ctx contextx.IContext, page types.Page, condition *types.BusinessCondition) (
		[]*types.Business, int64, error)

	// ListTopoEvent list topo events by page and conditions.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return the topo-event list with page and the total count with filter.
	ListTopoEvent(ctx contextx.IContext, page types.Page, condition *types.TopoEventCondition) (
		[]*types.TopoEvent, int64, error)

	// CountTopoEvent count topo events by condition.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the topo-event count with filter.
	CountTopoEvent(ctx contextx.IContext, condition *types.TopoEventCondition) (int64, error)

	// DistinctTopoEvent distinct topo-event by condition.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the topo-event distinct result.
	DistinctTopoEvent(
		ctx contextx.IContext, condition *types.TopoEventCondition) (
		*types.TopoEventDistinctResult, error)

	// ListAccessPoint list access points by page and conditions.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return the access-point list with page and the total count with filter.
	ListAccessPoint(ctx contextx.IContext, page types.Page, condition *types.AccessPointCondition) (
		[]*types.AccessPoint, int64, error)

	// GetConstant get constant by fields.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param fields describes the fields to get.
	// @return the constant result.
	GetConstant(ctx contextx.IContext, fields types.TopoConstantFields) (*types.TopoConstant, error)
}

// IHandlerGraph defines the graph Handler.
type IHandlerGraph interface {
	// GetGraphNode get graph node.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param networkUnitIDs the network unit ids.
	// @return the graph node list.
	GetGraphNode(ctx contextx.IContext, networkUnitIDs []int64) ([]*types.GraphNodeInfo, error)
}

// IHandlerHost defines the host Handler.
type IHandlerHost interface {
	// ListHost list host within specified tenant in contextx.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return host list with page and the total count with filter.
	ListHost(ctx contextx.IContext, page types.Page, condition *types.HostCondition) ([]*types.Host, int64, error)

	// DistinctHost distinct host by condition.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the host distinct result.
	DistinctHost(
		ctx contextx.IContext, condition *types.HostCondition) (
		*types.HostDistinctResult, error)

	// CountHost count host within specified tenant in contextx.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the host count with filter.
	CountHost(ctx contextx.IContext, condition *types.HostCondition) (int64, error)

	// GetHostDistributionByNodeRole get host distribution by node role.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the host distribution by node role.
	GetHostDistributionByNodeRole(ctx contextx.IContext, condition *types.HostCondition) (map[types.NodeRole]int64, error)

	// GetHostDistributionByNetworkAreaID get host distribution by node role.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the host distribution by node role.
	GetHostDistributionByNetworkAreaID(ctx contextx.IContext, condition *types.HostCondition) (map[int64]int64, error)
}

// IHandlerNetworkArea defines the network area Handler.
type IHandlerNetworkArea interface {
	// CreateNetworkArea create network area within specified tenant in contextx.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param networkArea the network area to create.
	// @return the network-area id.
	CreateNetworkArea(ctx contextx.IContext, networkArea *types.NetworkArea) (int64, error)

	// UpdateNetworkArea update network area within specified tenant in contextx.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param networkArea the network area to update.
	// @return the error.
	UpdateNetworkArea(ctx contextx.IContext, networkArea *types.NetworkArea) error

	// ListNetworkArea list network area within specified tenant in contextx.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return the network-area list with page and the total count with filter.
	ListNetworkArea(ctx contextx.IContext, page types.Page, condition *types.NetworkAreaCondition) (
		[]*types.NetworkArea, int64, error)

	// DeleteNetworkArea delete network area within specified tenant in contextx.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param networkAreaID the network area id.
	// @return the error.
	DeleteNetworkArea(ctx contextx.IContext, networkAreaID int64) error

	// GetNetworkArea get specific network area.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param networkAreaID the network area id.
	// @return the network-area.
	GetNetworkArea(ctx contextx.IContext, networkAreaID int64) (*types.NetworkArea, error)
}

// IHandlerNetworkUnit defines the network unit Handler.
type IHandlerNetworkUnit interface {
	// CreateNetworkUnit create network unit within specified tenant in contextx.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param networkUnit the network unit to create.
	// @param accessPoints the access points of the network unit.
	// @return the network-unit id.
	CreateNetworkUnit(ctx contextx.IContext, networkUnit *types.NetworkUnit, accessPoints ...*types.AccessPoint) (
		int64, error)

	// UpdateNetworkUnit update network unit within specified tenant in contextx.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param networkUnit the network unit to update.
	// @param accessPoints the access points of the network unit.
	// @return the network-unit id.
	UpdateNetworkUnit(ctx contextx.IContext, networkUnit *types.NetworkUnit, accessPoints ...*types.AccessPoint) error

	// ListNetworkUnit list network unit within specified tenant in contextx.
	// @param ctx content, contains tenant-id.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return the network-unit list with page and the total count with filter.
	ListNetworkUnit(ctx contextx.IContext, page types.Page, condition *types.NetworkUnitCondition) (
		[]*types.NetworkUnit, int64, error)

	// GetNetworkUnit get specific network unit.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param networkUnitID the network unit id.
	// @return the network-unit and its accesspoints.
	GetNetworkUnit(ctx contextx.IContext, networkUnitID int64) (*types.NetworkUnit, map[int64]*types.AccessPoint, error)

	// DeleteNetworkUnit delete network unit within specified tenant in contextx.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param networkUnitID the network unit id.
	// @return the network-unit id.
	DeleteNetworkUnit(ctx contextx.IContext, networkUnitID int64) error
}

// IHandlerNodeAgent defines the node agent Handler.
type IHandlerNodeAgent interface {
	// InstallAgent node agent.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param installParam the install param.
	// @return the installing workflow-ids and error.
	InstallAgent(ctx contextx.IContext, installParam *types.NodeAgentInstallParam) (string, error)

	// UpgradeAgent node agent.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param upgradeParam the upgrade param.
	// @return the upgrading workflow-ids and error.
	UpgradeAgent(ctx contextx.IContext, upgradeParam *types.NodeAgentUpgradeParam) (string, error)

	// ReconfigAgent node agent.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param reconfigParam the reconfig param.
	// @return the reconfig workflow-ids and error.
	ReconfigAgent(ctx contextx.IContext, reconfigParam *types.NodeAgentReconfigParam) (string, error)

	// RestartProxy node agent.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param restartParam the restart param.
	// @return the restarting workflow-ids and error.
	RestartAgent(ctx contextx.IContext, restartParam *types.NodeAgentRestartParam) (string, error)

	// CheckAgentInstall node agent.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param checkParam the check param.
	// @return the check result and error.
	CheckAgentInstall(ctx contextx.IContext, checkParam []*types.NodeAgentInstallCheckInfo) ([]*types.NodeAgentInstallCheckResult, error)

	// UninstallAgent node agent.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param uninstallParam the uninstall param.
	// @return the restarting workflow-ids and error.
	UninstallAgent(ctx contextx.IContext, uninstallParam *types.NodeAgentUninstallParam) (string, error)
}

// IHandlerNodeProxy defines the node proxy Handler.
type IHandlerNodeProxy interface {
	// InstallProxy node proxy.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param installParam the install param.
	// @return the installing workflow-ids and error.
	InstallProxy(ctx contextx.IContext, installParam *types.NodeProxyInstallParam) (string, error)

	// UpgradeProxy node proxy.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param upgradeParam the upgrade param.
	// @return the upgrading workflow-ids and error.
	UpgradeProxy(ctx contextx.IContext, upgradeParam *types.NodeProxyUpgradeParam) (string, error)

	// RestartProxy node proxy.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param restartParam the restart param.
	// @return the restarting workflow-ids and error.
	RestartProxy(ctx contextx.IContext, restartParam *types.NodeProxyRestartParam) (string, error)

	// ReconfigProxy node proxy.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param reconfigParam the reconfig param.
	// @return the reconfig workflow-ids and error.
	ReconfigProxy(ctx contextx.IContext, reconfigParam *types.NodeProxyReconfigParam) (string, error)

	// UpdateProxy update node proxy.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param updateParam the update param.
	// @return the error.
	UpdateProxy(ctx contextx.IContext, updateParam *types.NodeProxyUpdateParam) error

	// UninstallProxy node proxy.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param uninstallParam the uninstall param.
	// @return the restarting workflow-ids and error.
	UninstallProxy(ctx contextx.IContext, uninstallParm *types.NodeProxyUninstallParam) (string, error)
}

// IHandlerNodeWorkflow defines the node workflow Handler.
// nolint: interfacebloat
type IHandlerNodeWorkflow interface {
	// ListNodeWorkflow list node workflow within specified tenant in contextx.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return the node-workflow list with page and the total count with filter.
	ListNodeWorkflow(ctx contextx.IContext, page types.Page, condition *types.NodeWorkflowCondition) (
		[]*types.NodeWorkflow, int64, error)

	// CountNodeWorkflow count node workflow by conditions.
	//	@param ctx contextx, contains tenant-id.
	//	@param condition the filter conditions.
	//	@return the node-workflow count with filter.
	CountNodeWorkflow(ctx contextx.IContext, condition *types.NodeWorkflowCondition) (int64, error)

	// DistinctNodeWorkflow distinct node workflow by conditions.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param request the node workflow distinct request.
	// @param conditions the filter conditions.
	// @return the node-workflow distinct result.
	DistinctNodeWorkflow(ctx contextx.IContext, request types.NodeWorkflowDistinctRequest,
		condition *types.NodeWorkflowCondition) (*types.NodeWorkflowDistinctResult, error)

	// ListNodeWorkflowOperation list network area operation.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param page describes the page info when listing.
	// @param workflowID the workflow id.
	// @return the operation list with page and the total count with filter.
	ListNodeWorkflowOperation(ctx contextx.IContext, page types.Page, condition *types.NodeWorkflowOperationCondition) (
		[]*types.NodeWorkflowListOperationResult, int64, error)

	// CountNodeWorkflowOperation count network area operation.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param workflowID the workflow id.
	// @return the operation count with filter.
	CountNodeWorkflowOperation(ctx contextx.IContext, condition *types.NodeWorkflowOperationCondition) (int64, error)

	// ListNodeWorkflowOperationInstance list network area operation instance.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param operationID the operation id.
	// @return the operation instance list with page and the total count with filter.
	ListNodeWorkflowOperationInstance(ctx contextx.IContext, operationID ...string) (
		[]*operation.InstanceBriefData, int64, error)

	// CountNodeWorkflowOperationInstance count network area operation instance.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param operationID the operation id.
	// @return the operation instance count with filter.
	CountNodeWorkflowOperationInstance(ctx contextx.IContext, operationID ...string) (int64, error)

	// GetNodeWorkflowOperationInstanceLog distinct node workflow by conditions.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param request the node workflow distinct request.
	// @param condition the filter conditions.
	// @return the node-workflow distinct result.
	GetNodeWorkflowOperationInstanceLog(ctx contextx.IContext, instanceID string) (*operation.InstanceData, error)

	// ListNodeWorkflowOperationInstanceStatus list network area operation instance status.
	// @param ctx contextx, contains tenant-id.
	// @param triggerID the trigger id.
	// @return the operation instance status list.
	ListNodeWorkflowOperationInstanceStatus(ctx contextx.IContext,
		condition *types.NodeWorkflowOperInstanceStatusCondition) ([]*operation.InstanceStatus, error)

	// TerminateNodeWorkflowOperation terminate node operation.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param param the terminate param.
	// @return the error.
	TerminateNodeWorkflowOperation(ctx contextx.IContext, terminateParam *types.NodeWorkflowOperationTerminateParam) error

	// RetryNodeWorkflowOperation retry node operation.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param retryParam the retry param.
	// @return the error.
	RetryNodeWorkflowOperation(ctx contextx.IContext, retryParam *types.NodeWorkflowOperationRetryParam) error
}

// IHandlerConfigPolicy defines the backend Handler for config policy.
type IHandlerConfigPolicy interface {
	// ListConfigPolicy lists config policy by page and conditions.
	ListConfigPolicy(ctx contextx.IContext, page types.Page, condition *types.ConfigPolicyCondition) (
		[]*types.ConfigPolicy, int64, error)

	// GetConfigPolicy gets config policy by config policy id.
	GetConfigPolicy(ctx contextx.IContext, configPolicyID int64) (*types.ConfigPolicy, error)

	// CountConfigPolicy counts config policy by conditions.
	CountConfigPolicy(ctx contextx.IContext, condition *types.ConfigPolicyCondition) (int64, error)

	// CreateConfigPolicy creates config policy.
	CreateConfigPolicy(ctx contextx.IContext, configPolicy *types.ConfigPolicy) (int64, error)

	// UpdateConfigPolicy updates config policy.
	UpdateConfigPolicy(ctx contextx.IContext, configPolicy *types.ConfigPolicy) (int64, error)

	// EnableConfigPolicy enables config policy by config policy ids.
	EnableConfigPolicy(ctx contextx.IContext, configPolicyIDs ...int64) error

	// DisableConfigPolicy disables config policy by config policy ids.
	DisableConfigPolicy(ctx contextx.IContext, configPolicyIDs ...int64) error

	// DeleteConfigPolicy deletes config policy by config policy ids.
	DeleteConfigPolicy(ctx contextx.IContext, configPolicyIDs ...int64) error
}

// IHandlerPackage defines the backend Handler for package.
type IHandlerPackage interface {
	// ListPackageEvent list package events by page and conditions.
	ListPackageEvent(ctx contextx.IContext, page types.Page, condition *types.PackageEventCondition) (
		[]*types.PackageEvent, int64, error)

	// CountTopoEvent count package events by condition.
	CountPackageEvent(ctx contextx.IContext, condition *types.PackageEventCondition) (int64, error)

	// DistinctPackageEvent distinct package event by condition.
	DistinctPackageEvent(
		ctx contextx.IContext, condition *types.PackageEventCondition) (
		*types.PackageEventDistinctResult, error)
}

// IHandlerPlugin defines the backend Handler for plugin.
type IHandlerPlugin interface {
	// ListPlugins lists plugins by page and conditions.
	ListPlugins(ctx contextx.IContext, page types.Page, condition *types.PluginCondition) (
		[]*types.Plugin, int64, error)

	// CountPlugins counts plugins by conditions.
	CountPlugins(ctx contextx.IContext, condition *types.PluginCondition) (int64, error)

	// InstallPlugin install plugin.
	InstallPlugin(ctx contextx.IContext, installParam ...*types.PluginInstallParam) (string, error)

	// ApplyPluginSubConfig apply plugin sub config.
	ApplyPluginSubConfig(ctx contextx.IContext, applyParam ...*types.PluginApplySubConfigParam) (string, error)
}

// IHandlerProcess defines the backend Handler for process.
type IHandlerProcess interface {
	// ListProcesses lists processes by page and conditions.
	ListProcesses(ctx contextx.IContext, page types.Page, condition *types.ProcessCondition) (
		[]*types.Process, int64, error)

	// CountProcesses counts processes by conditions.
	CountProcesses(ctx contextx.IContext, condition *types.ProcessCondition) (int64, error)

	// GetProcessDistributionByHostID gets process distribution by host id.
	GetProcessDistributionByHostID(nCtx contextx.IContext, condition *types.ProcessCondition) (map[int64]int64, error)

	// GetProcessDistributionByPluginName gets process distribution by plugin name.
	GetProcessDistributionByPluginName(nCtx contextx.IContext, condition *types.ProcessCondition) (map[string]int64, error)
}

// IHandlerConfigPolicyEvent defines the backend Handler for policy event.
type IHandlerConfigPolicyEvent interface {
	// ListConfigPolicyEvent list policy events by page and conditions.
	ListConfigPolicyEvent(ctx contextx.IContext, page types.Page, condition *types.ConfigPolicyEventCondition) (
		[]*types.ConfigPolicyEvent, int64, error)

	// CountConfigPolicyEvent count policy events by condition.
	CountConfigPolicyEvent(ctx contextx.IContext, condition *types.ConfigPolicyEventCondition) (int64, error)

	// DistinctConfigPolicyEvent distinct policy event by condition.
	DistinctConfigPolicyEvent(
		ctx contextx.IContext, condition *types.ConfigPolicyEventCondition) (
		*types.ConfigPolicyEventDistinctResult, error)
}

var _ IHandler = &Handler{}

// Handler defines the backend handler.
type Handler struct {
	cli *cli
}

// New initialize a new nodeman backend Handler.
func New(c *restclient.Capability, conf Config) (*Handler, error) {
	cli, err := newClient(c, conf)
	if err != nil {
		return nil, err
	}

	return &Handler{cli: cli}, nil
}

// ListBusiness list business within specified tenant in contextx.
func (h *Handler) ListBusiness(ctx contextx.IContext, page types.Page, condition *types.BusinessCondition) (
	[]*types.Business, int64, error) {

	req := &protoBackend.TopoBusinessListReq{
		Page: convertPage(page),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listBusiness(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	items := resp.GetItems()
	data := make([]*types.Business, len(items))
	for idx, item := range items {
		data[idx] = &types.Business{
			TenantID: item.GetTenantId(),
			BizID:    item.GetBkBizId(),
			BizName:  item.GetBkBizName(),
		}
	}

	return data, resp.GetTotal(), nil
}

// ListHost list host within specified tenant in contextx.
// nolint: funlen
func (h *Handler) ListHost(ctx contextx.IContext, page types.Page, condition *types.HostCondition) (
	[]*types.Host, int64, error) {

	req := &protoBackend.TopoHostListReq{
		Page: convertPage(page),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listHost(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	total, hosts := resp.ConvertHostsToTypes()

	return hosts, total, nil
}

// DistinctHost distinct host within specified tenant in contextx.
func (h *Handler) DistinctHost(
	ctx contextx.IContext, condition *types.HostCondition) (
	*types.HostDistinctResult, error) {

	req := &protoBackend.TopoHostDistinctReq{}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, err
	}

	resp, err := h.cli.distinctHost(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertResultToTypes(), nil
}

// CountHost count host within specified tenant in contextx.
func (h *Handler) CountHost(ctx contextx.IContext, condition *types.HostCondition) (int64, error) {
	req := &protoBackend.TopoHostListReq{
		OnlyCount: true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}

	resp, err := h.cli.listHost(ctx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// GetHostDistributionByNodeRole count host within specified tenant in contextx.
func (h *Handler) GetHostDistributionByNodeRole(ctx contextx.IContext, condition *types.HostCondition) (map[types.NodeRole]int64, error) {
	req := &protoBackend.TopoGetHostDistributionByNodeRoleReq{}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, err
	}

	resp, err := h.cli.getHostDistributionByNodeRole(ctx, req)
	if err != nil {
		return nil, err
	}

	hostDistributionByNodeRole := make(map[types.NodeRole]int64)
	for nodeRole, hostCount := range resp.GetData() {
		hostDistributionByNodeRole[types.NodeRole(nodeRole)] = hostCount
	}

	return hostDistributionByNodeRole, nil
}

// GetHostDistributionByNetworkAreaID count host within specified tenant in contextx.
func (h *Handler) GetHostDistributionByNetworkAreaID(ctx contextx.IContext, condition *types.HostCondition) (map[int64]int64, error) {
	req := &protoBackend.TopoGetHostDistributionByNetworkAreaIDReq{}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, err
	}

	resp, err := h.cli.getHostDistributionByNetworkAreaID(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.GetData(), nil
}

// CreateNetworkArea creates a new networkarea.
func (h *Handler) CreateNetworkArea(ctx contextx.IContext, networkArea *types.NetworkArea) (int64, error) {
	req := &protoBackend.TopoNetworkAreaCreateReq{
		BkNetworkareaName: networkArea.Name,
		CloudVendor:       networkArea.CloudVendor,
	}

	resp, err := h.cli.createNetworkArea(ctx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetBkNetworkareaId(), nil
}

// UpdateNetworkArea updates an existing networkarea.
func (h *Handler) UpdateNetworkArea(ctx contextx.IContext, networkArea *types.NetworkArea) error {
	req := &protoBackend.TopoNetworkAreaUpdateReq{
		BkNetworkareaId:   networkArea.ID,
		BkNetworkareaName: networkArea.Name,
		CloudVendor:       networkArea.CloudVendor,
	}

	_, err := h.cli.updateNetworkArea(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// ListNetworkArea list network area within specified tenant in contextx.
func (h *Handler) ListNetworkArea(ctx contextx.IContext, page types.Page, condition *types.NetworkAreaCondition) (
	[]*types.NetworkArea, int64, error) {

	req := &protoBackend.TopoNetworkAreaListReq{
		Page: convertPage(page),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listNetworkArea(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	items := resp.GetItems()
	data := make([]*types.NetworkArea, len(items))
	for idx, item := range items {
		data[idx] = &types.NetworkArea{
			TenantID:    item.GetTenantId(),
			ID:          item.GetBkNetworkareaId(),
			Name:        item.GetBkNetworkareaName(),
			CloudVendor: item.GetCloudVendor(),
		}
	}

	return data, resp.GetTotal(), nil
}

// GetNetworkArea gets an existing networkarea.
func (h *Handler) GetNetworkArea(ctx contextx.IContext, networkAreaID int64) (*types.NetworkArea, error) {
	req := &protoBackend.TopoNetworkAreaGetReq{
		BkNetworkareaId: networkAreaID,
	}

	resp, err := h.cli.getNetworkArea(ctx, req)
	if err != nil {
		return nil, err
	}

	return &types.NetworkArea{
		TenantID:    resp.GetData().GetTenantId(),
		ID:          resp.GetData().GetBkNetworkareaId(),
		Name:        resp.GetData().GetBkNetworkareaName(),
		CloudVendor: resp.GetData().GetCloudVendor(),
	}, nil
}

// DeleteNetworkArea deletes an existing networkarea.
func (h *Handler) DeleteNetworkArea(ctx contextx.IContext, networkAreaID int64) error {
	req := &protoBackend.TopoNetworkAreaDeleteReq{
		BkNetworkareaId: networkAreaID,
	}

	_, err := h.cli.deleteNetworkArea(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// CreateNetworkUnit creates a new networkunit.
func (h *Handler) CreateNetworkUnit(
	ctx contextx.IContext, networkUnit *types.NetworkUnit, accessPoints ...*types.AccessPoint) (int64, error) {

	req := new(protoBackend.TopoNetworkUnitCreateReq)
	req.ConvertNetworkUnitFromTypes(networkUnit, accessPoints...)

	resp, err := h.cli.createNetworkUnit(ctx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetBkNetworkunitId(), nil
}

// UpdateNetworkUnit updates an existing networkunit.
func (h *Handler) UpdateNetworkUnit(
	ctx contextx.IContext, networkUnit *types.NetworkUnit, accessPoints ...*types.AccessPoint) error {

	req := new(protoBackend.TopoNetworkUnitUpdateReq)
	req.ConvertNetworkUnitFromTypes(networkUnit, accessPoints...)

	_, err := h.cli.updateNetworkUnit(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// GetNetworkUnit gets an existing networkunit.
func (h *Handler) GetNetworkUnit(ctx contextx.IContext, networkUnitID int64) (
	*types.NetworkUnit, map[int64]*types.AccessPoint, error) {

	req := &protoBackend.TopoNetworkUnitGetReq{
		BkNetworkunitId: networkUnitID,
	}

	resp, err := h.cli.getNetworkUnit(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	networkUnit, accessPoints := resp.ConvertNetworkUnitToTypes()

	return networkUnit, accessPoints, nil
}

// ListNetworkUnit list network unit within specified tenant in contextx.
func (h *Handler) ListNetworkUnit(ctx contextx.IContext, page types.Page, condition *types.NetworkUnitCondition) (
	[]*types.NetworkUnit, int64, error) {

	req := &protoBackend.TopoNetworkUnitListReq{
		Page: convertPage(page),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listNetworkUnit(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	total, networkUnits := resp.ConvertNetworkUnitsToTypes()

	return networkUnits, total, nil
}

// DeleteNetworkUnit deletes network unit within specified tenant in contextx.
func (h *Handler) DeleteNetworkUnit(ctx contextx.IContext, networkUnitID int64) error {
	req := &protoBackend.TopoNetworkUnitDeleteReq{
		BkNetworkunitId: networkUnitID,
	}

	_, err := h.cli.deleteNetworkUnit(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// ListTopoEvent list topo event within specified tenant in contextx.
func (h *Handler) ListTopoEvent(ctx contextx.IContext, page types.Page, condition *types.TopoEventCondition) (
	[]*types.TopoEvent, int64, error) {

	req := &protoBackend.TopoEventListReq{
		Page: convertPage(page),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listTopoEvent(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	total, events := resp.ConvertTopoEventsToTypes()

	return events, total, nil
}

// CountTopoEvent count the number of topo events by conditions.
func (h *Handler) CountTopoEvent(ctx contextx.IContext, condition *types.TopoEventCondition) (int64, error) {
	req := &protoBackend.TopoEventListReq{
		OnlyCount: true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}

	resp, err := h.cli.listTopoEvent(ctx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// DistinctTopoEvent distinct the number of topo events by conditions.
func (h *Handler) DistinctTopoEvent(
	ctx contextx.IContext, condition *types.TopoEventCondition) (
	*types.TopoEventDistinctResult, error) {

	req := &protoBackend.TopoEventDistinctReq{}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, err
	}

	resp, err := h.cli.distinctTopoEvent(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertResultToTypes(), nil
}

// ListAccessPoint list access point within specified tenant in contextx.
func (h *Handler) ListAccessPoint(ctx contextx.IContext, page types.Page, condition *types.AccessPointCondition) (
	[]*types.AccessPoint, int64, error) {

	req := &protoBackend.TopoAccessPointListReq{
		Page: convertPage(page),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listAccessPoint(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	total, events := resp.ConvertAccessPointsToTypes()

	return events, total, nil
}

// GetConstant get constant by fields.
func (h *Handler) GetConstant(ctx contextx.IContext, fields types.TopoConstantFields) (*types.TopoConstant, error) {
	req := &protoBackend.TopoConstantGetReq{}
	if err := req.ConvertFieldsFromTypes(fields); err != nil {
		return nil, err
	}

	resp, err := h.cli.getConstant(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertConstantToTypes(), nil
}

// ListNodeWorkflow list node workflow within specified tenant in contextx.
func (h *Handler) ListNodeWorkflow(ctx contextx.IContext, page types.Page, condition *types.NodeWorkflowCondition) (
	[]*types.NodeWorkflow, int64, error) {

	req := &protoBackend.NodeWorkflowListReq{
		Page: convertPage(page),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listNodeWorkflow(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	result, num := resp.ConvertNodeWorkflowsToTypes()

	return result, num, nil
}

// CountNodeWorkflow count host within specified tenant in contextx.
func (h *Handler) CountNodeWorkflow(ctx contextx.IContext, condition *types.NodeWorkflowCondition) (int64, error) {
	req := &protoBackend.NodeWorkflowListReq{
		OnlyCount: true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}

	resp, err := h.cli.listNodeWorkflow(ctx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// DistinctNodeWorkflow distinct node workflow by conditions.
func (h *Handler) DistinctNodeWorkflow(ctx contextx.IContext, _ types.NodeWorkflowDistinctRequest,
	conditions *types.NodeWorkflowCondition) (*types.NodeWorkflowDistinctResult, error) {

	req := &protoBackend.NodeWorkflowDistinctReq{}
	if err := req.ConvertConditionsFromTypes(conditions); err != nil {
		return nil, err
	}

	resp, err := h.cli.distinctNodeWorkflow(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertWorkflowDistinctToTypes(), nil
}

// ListNodeWorkflowOperation list workflow  operation.
func (h *Handler) ListNodeWorkflowOperation(ctx contextx.IContext,
	page types.Page, condition *types.NodeWorkflowOperationCondition) (
	[]*types.NodeWorkflowListOperationResult, int64, error) {

	req := &protoBackend.NodeWorkflowOperationListReq{
		Page:      convertPage(page),
		OnlyCount: false,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listNodeWorkflowOperation(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	operations, total := resp.ConvertWorkflowOperationToTypes()

	return operations, total, nil
}

// CountNodeWorkflowOperation count workflow  operation.
func (h *Handler) CountNodeWorkflowOperation(ctx contextx.IContext,
	condition *types.NodeWorkflowOperationCondition) (int64, error) {

	req := &protoBackend.NodeWorkflowOperationListReq{
		OnlyCount: true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}
	resp, err := h.cli.listNodeWorkflowOperation(ctx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotalCount(), nil
}

// ListNodeWorkflowOperationInstance list workflow operation instance.
func (h *Handler) ListNodeWorkflowOperationInstance(
	ctx contextx.IContext, operationID ...string) ([]*operation.InstanceBriefData, int64, error) {

	req := &protoBackend.NodeWorkflowOperationInstanceListReq{
		OnlyCount:   false,
		OperationId: []string(operationID),
	}

	resp, err := h.cli.listNodeWorkflowOperationInstance(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	operations, total := resp.ConvertWorkflowOperationInstanceToTypes()

	return operations, total, nil
}

// CountNodeWorkflowOperationInstance count workflow operation instance.
func (h *Handler) CountNodeWorkflowOperationInstance(ctx contextx.IContext, operationID ...string) (int64, error) {
	req := &protoBackend.NodeWorkflowOperationInstanceListReq{
		OnlyCount:   true,
		OperationId: []string(operationID),
	}

	resp, err := h.cli.listNodeWorkflowOperationInstance(ctx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// GetNodeWorkflowOperationInstanceLog get workflow operation instance log.
func (h *Handler) GetNodeWorkflowOperationInstanceLog(ctx contextx.IContext, instanceID string) (
	*operation.InstanceData, error) {

	req := &protoBackend.NodeWorkflowOperationInstanceLogGetReq{
		OperInstId: instanceID,
	}

	resp, err := h.cli.getOperationInstanceLog(ctx, req)
	if err != nil {
		return nil, err
	}

	operations := resp.ConvertWorkflowOperationInstanceLogToTypes()

	return operations, nil
}

// ListNodeWorkflowOperationInstanceStatus list workflow operation instance status.
func (h *Handler) ListNodeWorkflowOperationInstanceStatus(ctx contextx.IContext,
	conditions *types.NodeWorkflowOperInstanceStatusCondition) ([]*operation.InstanceStatus, error) {

	req := &protoBackend.NodeWorkflowOperationInstanceListStatusReq{
		Page: &protoBackend.Page{},
	}

	if err := req.ConvertConditionsFromTypes(conditions); err != nil {
		return nil, err
	}

	resp, err := h.cli.listNodeWorkflowOpInstanceStatus(ctx, req)
	if err != nil {
		return nil, err
	}

	instanceStatus := resp.ConvertWorkflowOperationInstanceStatusToTypes()

	return instanceStatus, nil
}

// RetryNodeWorkflowOperation retry node workflow operation.
func (h *Handler) RetryNodeWorkflowOperation(ctx contextx.IContext, retryParam *types.NodeWorkflowOperationRetryParam) error {
	req := &protoBackend.NodeWorkflowOperationRetryReq{}

	req.ConvertOperationRetryParamFromTypes(*retryParam)

	_, err := h.cli.retryOperation(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// TerminateNodeWorkflowOperation terminate node operation.
func (h *Handler) TerminateNodeWorkflowOperation(ctx contextx.IContext, terminateParam *types.NodeWorkflowOperationTerminateParam) error {
	req := &protoBackend.NodeWorkflowOperationTerminateReq{}

	req.ConvertOperationTerminateParamFromTypes(terminateParam)

	_, err := h.cli.terminateOperation(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// ListConfigPolicy lists config policy.
func (h *Handler) ListConfigPolicy(ctx contextx.IContext,
	page types.Page,
	condition *types.ConfigPolicyCondition) ([]*types.ConfigPolicy, int64, error) {

	req := &protoBackend.ConfigPolicyListReq{
		Page: convertPage(page),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listConfigPolicy(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	total, configPolicies := resp.ConvertConfigPoliciesToTypes()

	return configPolicies, total, nil
}

// GetConfigPolicy gets config policy by config policy id.
func (h *Handler) GetConfigPolicy(ctx contextx.IContext, configPolicyID int64) (*types.ConfigPolicy, error) {
	req := &protoBackend.ConfigPolicyGetReq{
		ConfigpolicyId: configPolicyID,
	}

	resp, err := h.cli.getConfigPolicy(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertConfigPolicyToTypes(), nil
}

// CountConfigPolicy counts config policy.
func (h *Handler) CountConfigPolicy(ctx contextx.IContext,
	condition *types.ConfigPolicyCondition) (int64, error) {

	req := &protoBackend.ConfigPolicyListReq{
		OnlyCount: true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}

	resp, err := h.cli.listConfigPolicy(ctx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// CreateConfigPolicy creates config policy.
func (h *Handler) CreateConfigPolicy(ctx contextx.IContext, configPolicy *types.ConfigPolicy) (int64, error) {
	req := new(protoBackend.ConfigPolicyCreateReq)
	req.ConvertConfigPolicyFromTypes(configPolicy)

	resp, err := h.cli.createConfigPolicy(ctx, req)
	if err != nil {
		return -1, err
	}

	return resp.GetData().GetConfigpolicyId(), nil
}

// UpdateConfigPolicy updates config policy.
func (h *Handler) UpdateConfigPolicy(ctx contextx.IContext, configPolicy *types.ConfigPolicy) (int64, error) {
	req := new(protoBackend.ConfigPolicyUpdateReq)
	req.ConvertConfigPolicyFromTypes(configPolicy)

	resp, err := h.cli.updateConfigPolicy(ctx, req)
	if err != nil {
		return -1, err
	}

	return resp.GetData().GetConfigpolicyId(), nil
}

// EnableConfigPolicy enables config policy.
func (h *Handler) EnableConfigPolicy(ctx contextx.IContext, configPolicyIDs ...int64) error {
	req := &protoBackend.ConfigPolicyEnableReq{ConfigpolicyId: configPolicyIDs}

	_, err := h.cli.enableConfigPolicy(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// DisableConfigPolicy disables config policy.
func (h *Handler) DisableConfigPolicy(ctx contextx.IContext, configPolicyIDs ...int64) error {
	req := &protoBackend.ConfigPolicyDisableReq{ConfigpolicyId: configPolicyIDs}

	_, err := h.cli.disableConfigPolicy(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// DeleteConfigPolicy deletes config policy.
func (h *Handler) DeleteConfigPolicy(ctx contextx.IContext, configPolicyIDs ...int64) error {
	req := &protoBackend.ConfigPolicyDeleteReq{ConfigpolicyId: configPolicyIDs}

	_, err := h.cli.deleteConfigPolicy(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// ListPackageEvent list package event within specified tenant in contextx.
func (h *Handler) ListPackageEvent(ctx contextx.IContext, page types.Page, condition *types.PackageEventCondition) (
	[]*types.PackageEvent, int64, error) {

	req := &protoBackend.PackageEventListReq{
		Page: convertPage(page),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listPackageEvent(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	total, events := resp.ConvertPackageEventsToTypes()

	return events, total, nil
}

// CountPackageEvent count the number of package events by conditions.
func (h *Handler) CountPackageEvent(ctx contextx.IContext, condition *types.PackageEventCondition) (int64, error) {
	req := &protoBackend.PackageEventListReq{
		OnlyCount: true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}

	resp, err := h.cli.listPackageEvent(ctx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// DistinctPackageEvent distinct the number of package events by conditions.
func (h *Handler) DistinctPackageEvent(
	ctx contextx.IContext, condition *types.PackageEventCondition) (
	*types.PackageEventDistinctResult, error) {

	req := &protoBackend.PackageEventDistinctReq{}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, err
	}

	resp, err := h.cli.distinctPackageEvent(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertResultToTypes()
}

// ListConfigPolicyEvent list policy event within specified tenant in contextx.
func (h *Handler) ListConfigPolicyEvent(ctx contextx.IContext, page types.Page, condition *types.ConfigPolicyEventCondition) (
	[]*types.ConfigPolicyEvent, int64, error) {

	req := &protoBackend.ConfigPolicyEventListReq{
		Page: convertPage(page),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listConfigPolicyEvent(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	total, events := resp.ConvertConfigPolicyEventsToTypes()

	return events, total, nil
}

// CountConfigPolicyEvent count the number of policy events by conditions.
func (h *Handler) CountConfigPolicyEvent(ctx contextx.IContext, condition *types.ConfigPolicyEventCondition) (int64, error) {
	req := &protoBackend.ConfigPolicyEventListReq{
		OnlyCount: true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}

	resp, err := h.cli.listConfigPolicyEvent(ctx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// DistinctConfigPolicyEvent distinct the number of policy events by conditions.
func (h *Handler) DistinctConfigPolicyEvent(
	ctx contextx.IContext, condition *types.ConfigPolicyEventCondition) (
	*types.ConfigPolicyEventDistinctResult, error) {

	req := &protoBackend.ConfigPolicyEventDistinctReq{}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, err
	}

	resp, err := h.cli.distinctConfigPolicyEvent(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertResultToTypes()
}

// GetGraphNode get graph node.
func (h *Handler) GetGraphNode(ctx contextx.IContext,
	networkUnitIDs []int64) ([]*types.GraphNodeInfo, error) {

	req := &protoBackend.TopoGraphNodeGetReq{
		BkNetworkunitId: networkUnitIDs,
	}

	resp, err := h.cli.getGraphNode(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertResultToTypes(), nil
}
