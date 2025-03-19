package cmdb

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// This document is responsible for processing the conversion of original requests and responses
// from the third-party system and the internal data of the nodeman system.

// Handler the handler of cmdb.
type Handler interface {
	// ListBizHosts list biz hosts
	ListBizHosts(ctx context.Context, bizID int64, page types.Page) ([]*types.Host, error)

	// SearchBusiness search business
	SearchBusiness(ctx context.Context, page types.Page) ([]*types.Business, error)

	// SearchNetworkArea search network area
	SearchNetworkArea(ctx context.Context, page types.Page) ([]*types.NetworkArea, error)

	// CreateNetworkArea create network area
	CreateNetworkArea(ctx context.Context, networkAreaName string, cloudVendor string) (*types.NetworkArea, error)

	// UpdateNetworkArea update network area
	UpdateNetworkArea(ctx context.Context, id int64, networkAreaName string, cloudVendor string) error

	// DeleteNetworkArea delete network area
	DeleteNetworkArea(ctx context.Context, id int64) error

	// UpdateHostNetworkAreaField update host network area field
	UpdateHostNetworkAreaField(ctx context.Context, hostIDs []int64, bizID int64, networkAreaID int64) error

	// SearchBizInstTopo search business instance topo
	SearchBizInstTopo(ctx context.Context, bizID int64) ([]*types.BusinessInstanceTopo, error)

	// GetBizInternalModule get biz internal module
	GetBizInternalModule(ctx context.Context, bizID int64) (*types.BusinessInternalModule, error)

	// FindTopoNodePaths find topo node paths
	FindTopoNodePaths(ctx context.Context, bizID int64, topoNodes []*types.TopoNode) ([]*types.TopoNodePath, error)

	// FindModuleBatch find module batch
	FindModuleBatch(ctx context.Context, bizID int64, ids []int64, fields []string) ([]*types.Module, error)

	// SearchCloudVendor search cloud vendor
	SearchCloudVendor(ctx context.Context) ([]*types.CloudVendor, error)

	// SearchOsType search os type
	SearchOsType(ctx context.Context) ([]*types.OsType, error)

	// BindHostAgent bind host agent
	BindHostAgent(ctx context.Context, hostAgentID []*types.HostAgentID) error

	// UnbindHostAgent bind host agent
	UnbindHostAgent(ctx context.Context, hostAgentID []*types.HostAgentID) error

	// AddHostToBusinessIdle add host to business idle
	AddHostToBusinessIdle(ctx context.Context, bizID int64, hosts []*types.CreateHostInfo) ([]int64, error)

	// PushHostIdentifier push host identifier
	PushHostIdentifier(ctx context.Context, hostIDs []int64) (*types.PushHostIdentifierTaskInfo, error)

	// FindHostIdentifierPushResult find host identifier push result
	FindHostIdentifierPushResult(ctx context.Context, taskID string) (*types.PushHostIdentifiersTaskResult, error)
}

type handler struct {
	cli *cli
}

// New initialize a new cmdb handler.
func New(c *client.Capability, conf *Config) (Handler, error) {
	cli, err := newClient(c, conf)
	if err != nil {
		return nil, err
	}

	return &handler{cli: cli}, nil
}

// ListBizHosts list biz hosts.
func (h *handler) ListBizHosts(ctx context.Context, bizID int64, page types.Page) ([]*types.Host, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ListBizHostsReq{
		TenantID: tenantID,
		BKBizID:  bizID,
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}

	resp, err := h.cli.listBizHosts(ctx, req)
	if err != nil {
		return nil, err
	}

	hosts := make([]*types.Host, len(resp.Info))
	for idx, host := range resp.Info {
		hosts[idx] = &types.Host{
			HostID:   host.BKHostID,
			TenantID: tenantID,
			Static: &types.HostStatic{
				BizID:         bizID,
				NetworkAreaID: host.BKCloudID,
				HostName:      host.BKHostName,
				DeptName:      host.DeptName,
				InnerIP:       host.BKHostInnerIPV4,
				InnerIPV6:     host.BKHostInnerIPV6,
				OuterIP:       host.BKHostOuterIPV4,
				OuterIPV6:     host.BKHostOuterIPV6,
				Mac:           host.BKMac,
				OSType:        host.BKOsType,
				SyncedAgentID: host.BKAgentID,
			},
			Dynamic: types.NewBlankNodeDynamic(),
		}
	}

	return hosts, nil
}

// SearchBusiness search business.
func (h *handler) SearchBusiness(ctx context.Context, page types.Page) ([]*types.Business, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &SearchBusinessReq{
		TenantID: tenantID,
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
		Fields: nil,
	}

	resp, err := h.cli.searchBusiness(ctx, req)
	if err != nil {
		return nil, err
	}

	bizs := make([]*types.Business, len(resp.Info))
	for idx, business := range resp.Info {
		bizs[idx] = &types.Business{
			TenantID: tenantID,
			BizID:    business.BKBizID,
			BizName:  business.BKBizName,
		}
	}

	return bizs, nil
}

// SearchNetworkArea search network area.
func (h *handler) SearchNetworkArea(ctx context.Context, page types.Page) ([]*types.NetworkArea, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &SearchCloudAreaReq{
		TenantID: tenantID,
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}

	resp, err := h.cli.searchCloudArea(ctx, req)
	if err != nil {
		return nil, err
	}

	netAreas := make([]*types.NetworkArea, len(resp.Info))
	for idx, netArea := range resp.Info {
		netAreas[idx] = &types.NetworkArea{
			TenantID:    tenantID,
			ID:          netArea.BKCloudID,
			Name:        netArea.BKCloudName,
			CloudVendor: netArea.BKCloudVendor,
		}
	}

	return netAreas, nil
}

// CreateNetworkArea create network area.
func (h *handler) CreateNetworkArea(ctx context.Context, networkAreaName string, cloudVendor string) (
	*types.NetworkArea, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &CreateCloudAreaReq{
		TenantID:      tenantID,
		BKCloudName:   networkAreaName,
		BKCloudVendor: cloudVendor,
	}

	resp, err := h.cli.createCloudArea(ctx, req)
	if err != nil {
		return nil, err
	}

	netArea := &types.NetworkArea{
		TenantID:    tenantID,
		ID:          resp.Created.ID,
		Name:        networkAreaName,
		CloudVendor: cloudVendor,
	}

	return netArea, nil
}

// UpdateNetworkArea update network area.
func (h *handler) UpdateNetworkArea(
	ctx context.Context, id int64, networkAreaName string, cloudVendor string) error {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	req := &UpdateCloudAreaReq{
		TenantID:      tenantID,
		BKCloudID:     id,
		BKCloudName:   networkAreaName,
		BKCloudVendor: cloudVendor,
	}

	err = h.cli.updateCloudArea(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// DeleteNetworkArea delete network area.
func (h *handler) DeleteNetworkArea(ctx context.Context, id int64) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	req := &DeleteCloudAreaReq{
		TenantID:  tenantID,
		BKCloudID: id,
	}

	err = h.cli.deleteCloudArea(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// UpdateHostNetworkAreaField update host network area field.
func (h *handler) UpdateHostNetworkAreaField(
	ctx context.Context, hostIDs []int64, bizID int64, networkAreaID int64) error {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	req := &UpdateHostCloudAreaFieldReq{
		TenantID:  tenantID,
		BKBizID:   bizID,
		BKCloudID: networkAreaID,
		BKHostIDs: hostIDs,
	}

	err = h.cli.updateHostCloudAreaField(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// SearchBizInstTopo search business instance topo.
func (h *handler) SearchBizInstTopo(ctx context.Context, bizID int64) ([]*types.BusinessInstanceTopo, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &SearchBizInstTopoReq{
		TenantID: tenantID,
		BKBizID:  bizID,
	}

	resp, err := h.cli.searchBizInstTopo(ctx, req)
	if err != nil {
		return nil, err
	}

	var coverFunc func(*BizInstTopo) *types.BusinessInstanceTopo
	coverFunc = func(src *BizInstTopo) (dst *types.BusinessInstanceTopo) {
		if src == nil {
			return nil
		}

		dst = &types.BusinessInstanceTopo{
			InstID:   src.BKInstID,
			InstName: src.BKInstName,
			ObjID:    src.BKObjID,
			ObjName:  src.BKObjName,
		}

		if len(src.Children) > 0 {
			dst.Children = make([]*types.BusinessInstanceTopo, len(src.Children))
			for i, child := range src.Children {
				dst.Children[i] = coverFunc(child)
			}
		}

		return
	}

	result := make([]*types.BusinessInstanceTopo, len(*resp))
	for i, topo := range *resp {
		result[i] = coverFunc(topo)
	}

	return result, nil
}

// GetBizInternalModule get biz internal module.
func (h *handler) GetBizInternalModule(ctx context.Context, bizID int64) (*types.BusinessInternalModule, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &GetBizInternalModuleReq{
		TenantID: tenantID,
		BKBizID:  bizID,
	}
	resp, err := h.cli.getBizInternalModule(ctx, req)
	if err != nil {
		return nil, err
	}

	result := &types.BusinessInternalModule{
		TenantID: tenantID,
		SetID:    resp.BKSetID,
		SetName:  resp.BKSetName,
	}
	result.Module = make([]*types.Module, len(resp.Module))
	for i, module := range resp.Module {
		result.Module[i] = &types.Module{
			ModuleID:          module.BKModuleID,
			ModuleName:        module.BKModuleName,
			SetID:             module.BKSetID,
			BakOperator:       module.BKBakOperator,
			BizID:             module.BKBizID,
			ModuleType:        module.BKModuleType,
			ParentID:          module.BKParentID,
			HostApplyEnabled:  module.HostApplyEnabled,
			ServiceCategoryID: module.ServiceCategoryID,
			ServiceTemplateID: module.ServiceTemplateID,
			SetTemplateID:     module.SetTemplateID,
			SupplierAccount:   module.BKSupplierAccount,
			CreatedBy:         module.BKCreatedBy,
			Operator:          module.Operator,
			LastTime:          module.LastTime,
			CreateTime:        module.CreateTime,
			CreatedAt:         module.BKCreatedAt,
			UpdatedAt:         module.BKUpdatedAt,
		}
	}

	return result, nil
}

// FindTopoNodePaths find topo node paths.
func (h *handler) FindTopoNodePaths(ctx context.Context, bizID int64, topoNodes []*types.TopoNode) (
	[]*types.TopoNodePath, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	node := make([]*Node, len(topoNodes))
	for i, topoNode := range topoNodes {
		node[i] = &Node{
			BKObjID:  topoNode.ObjID,
			BKInstID: topoNode.InstID,
		}
	}
	req := &FindTopoNodePathsReq{
		TenantID: tenantID,
		BKBizID:  bizID,
		BKNodes:  node,
	}

	resp, err := h.cli.findTopoNodePaths(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.TopoNodePath, len(*resp))
	for index, nodePath := range *resp {
		result[index] = &types.TopoNodePath{
			ObjID:    nodePath.BKObjID,
			InstID:   nodePath.BKInstID,
			InstName: nodePath.BKInstName,
		}

		if len(nodePath.BKPaths) > 0 {
			result[index].Paths = make([][]*types.TopoNode, len(nodePath.BKPaths))
			for i, path := range nodePath.BKPaths {
				result[index].Paths[i] = make([]*types.TopoNode, len(path))
				for j, node := range path {
					result[index].Paths[i][j] = &types.TopoNode{
						ObjID:    node.BKObjID,
						InstID:   node.BKInstID,
						InstName: node.BKInstName,
					}
				}
			}
		}
	}

	return result, nil
}

// FindModuleBatch find module batch.
func (h *handler) FindModuleBatch(ctx context.Context, bizID int64, ids []int64, fields []string) (
	[]*types.Module, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &FindModuleBatchReq{
		TenantID: tenantID,
		BKBizID:  bizID,
		Fields:   fields,
		BKIDs:    ids,
	}
	resp, err := h.cli.findModuleBatch(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.Module, len(*resp))
	for index, module := range *resp {
		result[index] = &types.Module{
			ModuleID:          module.BKModuleID,
			ModuleName:        module.BKModuleName,
			SetID:             module.BKSetID,
			BakOperator:       module.BKBakOperator,
			BizID:             module.BKBizID,
			ModuleType:        module.BKModuleType,
			ParentID:          module.BKParentID,
			HostApplyEnabled:  module.HostApplyEnabled,
			ServiceCategoryID: module.ServiceCategoryID,
			ServiceTemplateID: module.ServiceTemplateID,
			SetTemplateID:     module.SetTemplateID,
			SupplierAccount:   module.BKSupplierAccount,
			CreatedBy:         module.BKCreatedBy,
			Operator:          module.Operator,
			LastTime:          module.LastTime,
			CreateTime:        module.CreateTime,
			CreatedAt:         module.BKCreatedAt,
			UpdatedAt:         module.BKUpdatedAt,
		}
	}

	return result, nil
}

// searchObjectAttributeEnumOption search cmdb object attribute's option, like bk_cloud_vendor and bk_os_type.
func (h *handler) searchObjectAttributeEnumOption(ctx context.Context, objID string, bizID int64, objAttrID string) (
	[]*EnumOption, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &SearchObjectAttributeReq{
		TenantID: tenantID,
		BKObjID:  objID,
		BKBizID:  bizID,
	}

	resp, err := h.cli.searchObjectAttribute(ctx, req)
	if err != nil {
		return nil, err
	}

	var result []*EnumOption
	for _, objAttribute := range *resp {
		if objAttribute.BKPropertyID != objAttrID {
			continue
		}

		options, ok := objAttribute.Option.([]any)
		if !ok {
			return nil, fmt.Errorf("try to convert type to []any failed, bk_property_id(%s), option(%v)", objAttrID,
				objAttribute.Option)
		}

		result = make([]*EnumOption, len(options))
		for index, option := range options {
			mapOption, ok := option.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("try to convert type to map[string]any failed, bk_property_id(%s), option(%v)",
					objAttrID, option)
			}

			key, ok := mapOption["id"].(string)
			if !ok {
				return nil, fmt.Errorf("try to convert type to string failed, bk_property_id(%s), option[id](%v)",
					objAttrID, mapOption["id"])
			}

			name, ok := mapOption["name"].(string)
			if !ok {
				return nil, fmt.Errorf("try to convert type to string failed, bk_property_id(%s), option[name](%v)",
					objAttrID, mapOption["name"])
			}

			result[index] = &EnumOption{
				Key:   key,
				Value: name,
			}
		}
	}

	return result, nil
}

// SearchCloudVendor search cloud vendor.
func (h *handler) SearchCloudVendor(ctx context.Context) ([]*types.CloudVendor, error) {
	result, err := h.searchObjectAttributeEnumOption(ctx, "plat", DefaultBusinessID, "bk_cloud_vendor")
	if err != nil {
		return nil, err
	}

	cloudVendors := make([]*types.CloudVendor, len(result))
	for idx, vendor := range result {
		cloudVendors[idx] = &types.CloudVendor{
			Key:  vendor.Key,
			Name: vendor.Value,
		}
	}

	return cloudVendors, nil
}

// SearchOsType search ostype.
func (h *handler) SearchOsType(ctx context.Context) ([]*types.OsType, error) {
	result, err := h.searchObjectAttributeEnumOption(ctx, "host", DefaultBusinessID, "bk_os_type")
	if err != nil {
		return nil, err
	}

	osTypes := make([]*types.OsType, len(result))
	for idx, osType := range result {
		osTypes[idx] = &types.OsType{
			Key:  osType.Key,
			Name: osType.Value,
		}
	}

	return osTypes, nil
}

// BindHostAgent bind host agent.
func (h *handler) BindHostAgent(ctx context.Context, hostAgentID []*types.HostAgentID) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	req := &BindHostAgentReq{TenantID: tenantID}
	for _, hostAgent := range hostAgentID {
		req.List = append(req.List, &HostAgentIDInfo{
			BKHostID:  hostAgent.HostID,
			BKAgentID: hostAgent.AgentID,
		})
	}
	err = h.cli.bindHostAgent(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// UnbindHostAgent bind host agent.
func (h *handler) UnbindHostAgent(ctx context.Context, hostAgentID []*types.HostAgentID) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	req := &UnbindHostAgentReq{TenantID: tenantID}
	for _, hostAgent := range hostAgentID {
		req.List = append(req.List, &HostAgentIDInfo{
			BKHostID:  hostAgent.HostID,
			BKAgentID: hostAgent.AgentID,
		})
	}
	err = h.cli.unbindHostAgent(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// AddHostToBusinessIdle add host to business idle.
func (h *handler) AddHostToBusinessIdle(ctx context.Context, bizID int64, hosts []*types.CreateHostInfo) (
	[]int64, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &AddHostToBusinessIdleReq{TenantID: tenantID, BKBizID: bizID}
	for _, host := range hosts {
		req.BKHostList = append(req.BKHostList, &CreateHostInfo{
			BKCloudID:         host.NetworkAreaID,
			BKHostInnerIP:     host.InnerIP,
			BKHostInnerIPV6:   host.InnerIPV6,
			BKHostOuterIP:     host.OuterIP,
			BKHostOuterIPV6:   host.OuterIPV6,
			BKOSType:          host.OSType,
			BKCpuArchitecture: host.Arch,
			BKAddressing:      string(host.Addressing),
		})
	}
	resp, err := h.cli.addHostToBusinessIdle(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.BKHostIDs, nil
}

// PushHostIdentifier push host identifier.
func (h *handler) PushHostIdentifier(ctx context.Context, hostIDs []int64) (
	*types.PushHostIdentifierTaskInfo, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &PushHostIdentifierReq{
		TenantID:  tenantID,
		BKHostIDs: hostIDs,
	}
	resp, err := h.cli.pushHostIdentifier(ctx, req)
	if err != nil {
		return nil, err
	}

	result := &types.PushHostIdentifierTaskInfo{
		TaskID: resp.TaskID,
	}
	result.HostInfo = make([]*types.HostIdentification, len(resp.HostInfos))
	for index, hostInfo := range resp.HostInfos {
		result.HostInfo[index] = &types.HostIdentification{
			HostID:         hostInfo.BKHostID,
			Identification: hostInfo.Identification,
		}
	}

	return result, nil
}

// FindHostIdentifierPushResult find host identifier push result.
func (h *handler) FindHostIdentifierPushResult(ctx context.Context, taskID string) (
	*types.PushHostIdentifiersTaskResult, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &FindHostIdentifierPushResultReq{
		TenantID: tenantID,

		TaskID: taskID,
	}
	resp, err := h.cli.findHostIdentifierPushResult(ctx, req)
	if err != nil {
		return nil, err
	}

	result := &types.PushHostIdentifiersTaskResult{
		SuccessHostIDList: resp.SuccessList,
		FailedHostIDList:  resp.FailedList,
		PendingHostIDList: resp.PendingList,
	}

	return result, nil
}
