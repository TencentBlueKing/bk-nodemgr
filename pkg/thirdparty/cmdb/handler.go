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

	// CreateDynamicGroup create dynamic group
	CreateDynamicGroup(ctx context.Context, group *types.DynamicGroup) (string, error)

	// ExecuteDynamicGroup execute dynamic group
	ExecuteDynamicGroup(ctx context.Context, bizID int64, groupID string, fields []string, page types.Page) (
		[]*types.Host, error)

	// GetDynamicGroup get dynamic group
	GetDynamicGroup(ctx context.Context, bizID int64, groupID string) (*types.DynamicGroup, error)

	// SearchDynamicGroup search dynamic group
	SearchDynamicGroup(ctx context.Context, bizID int64, name string, page types.Page) (
		[]*types.DynamicGroup, error)

	// UpdateDynamicGroup update dynamic group
	UpdateDynamicGroup(ctx context.Context, group *types.DynamicGroup) error

	// DeleteDynamicGroup delete dynamic group
	DeleteDynamicGroup(ctx context.Context, bizID int64, groupID string) error

	// ListServiceTemplate list service template
	ListServiceTemplate(ctx context.Context, bizID int64, serviceCategoryID int64, serviceTemplateName string,
		serviceTemplateIDs []int64, isExact bool, page types.Page) ([]*types.ServiceTemplate, error)

	// ListServiceInstance list service instance
	ListServiceInstance(ctx context.Context, bizID int64, moduleID int64, hostIDs []int64, svcInstFuzzyName string,
		page types.Page) ([]*types.ServiceInstance, error)

	// ListProcessInstance list process instance
	ListProcessInstance(ctx context.Context, bizID int64, serviceInstanceID int64) ([]*types.ProcessInstance, error)

	// ListProcTemplate list process template
	ListProcTemplate(ctx context.Context, bizID int64, serviceTemplateID int64, processTemplateIDs []int64) (
		[]*types.ProcessTemplate, error)

	// FindSetBatch find set batch
	FindSetBatch(ctx context.Context, bizID int64, setIDs []int64, fields []string) ([]*types.Set, error)

	// SearchSet search set
	SearchSet(ctx context.Context, bizID int64, fields []string, page types.Page) ([]*types.Set, error)

	// SearchModule search module
	SearchModule(ctx context.Context, bizID int64, setID int64, fields []string, page types.Page) (
		[]*types.Module, error)

	// FindHostTopoRelation find host topo relation
	FindHostTopoRelation(ctx context.Context, bizID int64, setIDs []int64, moduleIDs []int64, hostIDs []int64,
		page types.Page) ([]*types.HostTopoRelation, error)

	// FindHostBizRelations find host biz relations
	FindHostBizRelations(ctx context.Context, bizID int64, hostIDs []int64) ([]*types.HostTopoRelation, error)

	// FindHostByServiceTemplate find host by service template
	FindHostByServiceTemplate(ctx context.Context, bizID int64, serviceTemplateID []int64, moduleIDs []int64,
		fields []string, page types.Page) ([]*types.Host, error)

	// FindHostBySetTemplate find host by set template
	FindHostBySetTemplate(ctx context.Context, bizID int64, setTemplateID []int64, setIDs []int64, fields []string,
		page types.Page) ([]*types.Host, error)

	// FindHostByTopo find host by topo
	FindHostByTopo(ctx context.Context, bizID int64, objID string, instID int64, fields []string, page types.Page) (
		[]*types.Host, error)

	// FindHostRelationsWithTopo find host relations with topo
	FindHostRelationsWithTopo(ctx context.Context, bizID int64, objID string, instIDs []int64, fields []string,
		page types.Page) ([]*types.HostTopoRelation, error)

	// ListServiceInstanceDetail list service instance detail
	ListServiceInstanceDetail(ctx context.Context, bizID int64, page types.Page) ([]*types.ServiceInstanceDetail, error)

	// GetMainlineObjectTopo get mainline object topo
	GetMainlineObjectTopo(ctx context.Context) ([]*types.MainlineObjectTopo, error)

	// ListBizHostsTopo list biz hosts topo
	ListBizHostsTopo(ctx context.Context, bizID int64, fields []string, page types.Page) ([]*types.HostWithHostTopo,
		error)

	// ListServiceInstanceByHost list service instance by host
	ListServiceInstanceByHost(ctx context.Context, bizID int64, hostID int64, page types.Page) (
		[]*types.ServiceInstance, error)

	// ListServiceInstanceBySetTemplate list service instance by set template
	ListServiceInstanceBySetTemplate(ctx context.Context, bizID int64, setTemplate int64, page types.Page) (
		[]*types.ServiceInstance, error)

	// ListSetTemplate list set template
	ListSetTemplate(ctx context.Context, bizID int64, setTemplateIDs []int64, page types.Page) (
		[]*types.ServiceInstance, error)

	// BatchUpdateHost batch update host
	BatchUpdateHost(ctx context.Context, hostsUpdateInfo []*types.UpdateHostProperties) error

	// BindHostAgent bind host agent
	BindHostAgent(ctx context.Context, hostAgentID []*types.HostAgentID) error

	// UnbindHostAgent bind host agent
	UnbindHostAgent(ctx context.Context, hostAgentID []*types.HostAgentID) error

	// AddHostToBusinessIdle add host to business idle
	AddHostToBusinessIdle(ctx context.Context, bizID int64, hosts []*types.CreateHostInfo) ([]int64, error)

	// PushHostIdentifier push host identifier
	PushHostIdentifier(ctx context.Context, hostIDs []int64) (*types.PushHostIdentifierTaskResult, error)

	// HostResourceWatch host resource watch
	HostResourceWatch(ctx context.Context, event types.ResourceWatchEvent) ([]*types.HostResourceWatchEvent, error)

	// HostRelationResourceWatch host relation resource watch
	HostRelationResourceWatch(ctx context.Context, event types.ResourceWatchEvent) (
		[]*types.HostRelationResourceWatchEvent, error)

	// ProcessResourceWatch process resource watch
	ProcessResourceWatch(ctx context.Context, event types.ResourceWatchEvent) ([]*types.ProcessResourceWatchEvent,
		error)
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

func (h *handler) ListHostsWithoutBiz(ctx context.Context, fields []string, page types.Page) ([]*types.Host, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ListHostsWithoutBusinessReq{
		TenantID: tenantID,
		Fields:   fields,
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}

	resp, err := h.cli.listHostsWithoutBusiness(ctx, req)
	if err != nil {
		return nil, err
	}

	hosts := make([]*types.Host, len(resp.Info))
	for idx, host := range resp.Info {
		hosts[idx] = &types.Host{
			HostID:   host.BKHostID,
			TenantID: tenantID,
			Static: &types.HostStatic{
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

// CreateDynamicGroup create dynamic group.
func (h *handler) CreateDynamicGroup(ctx context.Context, group *types.DynamicGroup) (string, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return "", err
	}

	req := &CreateDynamicGroupReq{}
	req.TenantID = tenantID
	req.BKBizID = group.BizID
	req.BKObjID = group.ObjID
	req.Name = group.Name
	for _, condition := range group.Condition {
		reqCondition := &DynamicGroupCondition{BKObjID: group.ObjID}
		reqCondition.Condition = make([]*FieldCondition, len(condition.Condition))
		for index, cond := range condition.Condition {
			reqCondition.Condition[index] = &FieldCondition{
				Field:    cond.Field,
				Operator: cond.Operator,
				Value:    cond.Value,
			}
		}
		req.Info.Condition = append(req.Info.Condition, reqCondition)
	}

	for _, condition := range group.VariableCondition {
		reqCondition := &DynamicGroupCondition{BKObjID: group.ObjID}
		reqCondition.Condition = make([]*FieldCondition, len(condition.Condition))
		for index, cond := range condition.Condition {
			reqCondition.Condition[index] = &FieldCondition{
				Field:    cond.Field,
				Operator: cond.Operator,
				Value:    cond.Value,
			}
		}
		req.Info.VariableCondition = append(req.Info.VariableCondition, reqCondition)
	}

	resp, err := h.cli.createDynamicGroup(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

// ExecuteDynamicGroup execute dynamic group.
func (h *handler) ExecuteDynamicGroup(ctx context.Context, bizID int64, groupID string, fields []string,
	page types.Page) ([]*types.Host, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	if len(fields) == 0 {
		fields = []string{"bk_host_id", "bk_host_name", "bk_cloud_id", "bk_host_innerip"}
	}
	req := &ExecuteDynamicGroupReq{
		TenantID:       tenantID,
		BKBizID:        bizID,
		ID:             groupID,
		Fields:         fields,
		DisableCounter: false,
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}

	resp, err := h.cli.executeHostDynamicGroup(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.Host, len(resp.Info))
	for index, host := range resp.Info {
		result[index] = &types.Host{
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
			},
			Dynamic: types.NewBlankNodeDynamic(),
		}
	}

	return result, nil
}

// SearchDynamicGroup search dynamic group.
func (h *handler) SearchDynamicGroup(ctx context.Context, bizID int64, name string, page types.Page) (
	[]*types.DynamicGroup, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &SearchDynamicGroupReq{
		TenantID: tenantID,
		BKBizID:  bizID,
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}
	req.Condition.Name = name

	resp, err := h.cli.searchDynamicGroup(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.DynamicGroup, len(resp.Info))
	for index, group := range resp.Info {
		result[index] = &types.DynamicGroup{
			ID:    group.ID,
			BizID: group.BKBizID,
			ObjID: group.BKObjID,
			Name:  group.Name,
		}

		result[index].Condition = make([]*types.DynamicGroupCondition, len(group.Info.Condition))
		for condIdx, condition := range group.Info.Condition {
			result[index].Condition[condIdx] = &types.DynamicGroupCondition{
				ObjID: group.BKObjID,
			}
			result[index].Condition[condIdx].Condition = make([]*types.FieldCondition, len(condition.Condition))
			for fieldIndex, field := range condition.Condition {
				result[index].Condition[condIdx].Condition[fieldIndex] = &types.FieldCondition{
					Field:    field.Field,
					Operator: field.Operator,
					Value:    field.Value,
				}
			}
		}

		result[index].VariableCondition = make([]*types.DynamicGroupCondition, len(group.Info.VariableCondition))
		for condIdx, condition := range group.Info.VariableCondition {
			result[index].VariableCondition[condIdx] = &types.DynamicGroupCondition{
				ObjID: group.BKObjID,
			}
			result[index].VariableCondition[condIdx].Condition = make([]*types.FieldCondition, len(condition.Condition))
			for fieldIndex, field := range condition.Condition {
				result[index].VariableCondition[condIdx].Condition[fieldIndex] = &types.FieldCondition{
					Field:    field.Field,
					Operator: field.Operator,
					Value:    field.Value,
				}
			}
		}
	}

	return result, nil
}

// UpdateDynamicGroup update dynamic group.
func (h *handler) UpdateDynamicGroup(ctx context.Context, group *types.DynamicGroup) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	req := &UpdateDynamicGroupReq{}
	req.TenantID = tenantID
	req.ID = group.ID
	req.BKBizID = group.BizID
	req.BKObjID = group.ObjID
	req.Name = group.Name
	for _, condition := range group.Condition {
		reqCondition := &DynamicGroupCondition{BKObjID: group.ObjID}
		reqCondition.Condition = make([]*FieldCondition, len(condition.Condition))
		for index, cond := range condition.Condition {
			reqCondition.Condition[index] = &FieldCondition{
				Field:    cond.Field,
				Operator: cond.Operator,
				Value:    cond.Value,
			}
		}
		req.Info.Condition = append(req.Info.Condition, reqCondition)
	}

	for _, condition := range group.VariableCondition {
		reqCondition := &DynamicGroupCondition{BKObjID: group.ObjID}
		reqCondition.Condition = make([]*FieldCondition, len(condition.Condition))
		for index, cond := range condition.Condition {
			reqCondition.Condition[index] = &FieldCondition{
				Field:    cond.Field,
				Operator: cond.Operator,
				Value:    cond.Value,
			}
		}
		req.Info.VariableCondition = append(req.Info.VariableCondition, reqCondition)
	}

	err = h.cli.updateDynamicGroup(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// DeleteDynamicGroup delete dynamic group.
func (h *handler) DeleteDynamicGroup(ctx context.Context, bizID int64, groupID string) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	req := &DeleteDynamicGroupReq{
		TenantID: tenantID,
		BKBizID:  bizID,
		ID:       groupID,
	}
	err = h.cli.deleteDynamicGroup(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// GetDynamicGroup get dynamic group.
func (h *handler) GetDynamicGroup(ctx context.Context, bizID int64, groupID string) (*types.DynamicGroup, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &GetDynamicGroupReq{
		TenantID: tenantID,
		BKBizID:  bizID,
		ID:       groupID,
	}
	resp, err := h.cli.getDynamicGroup(ctx, req)
	if err != nil {
		return nil, err
	}

	result := &types.DynamicGroup{
		ID:    resp.ID,
		BizID: resp.BKBizID,
		ObjID: resp.BKObjID,
		Name:  resp.Name,
	}

	result.Condition = make([]*types.DynamicGroupCondition, len(resp.Info.Condition))
	for condIdx, condition := range resp.Info.Condition {
		result.Condition[condIdx] = &types.DynamicGroupCondition{
			ObjID: resp.BKObjID,
		}
		result.Condition[condIdx].Condition = make([]*types.FieldCondition, len(condition.Condition))
		for fieldIndex, field := range condition.Condition {
			result.Condition[condIdx].Condition[fieldIndex] = &types.FieldCondition{
				Field:    field.Field,
				Operator: field.Operator,
				Value:    field.Value,
			}
		}
	}

	result.VariableCondition = make([]*types.DynamicGroupCondition, len(resp.Info.VariableCondition))
	for condIdx, condition := range resp.Info.VariableCondition {
		result.VariableCondition[condIdx] = &types.DynamicGroupCondition{
			ObjID: resp.BKObjID,
		}
		result.VariableCondition[condIdx].Condition = make([]*types.FieldCondition, len(condition.Condition))
		for fieldIndex, field := range condition.Condition {
			result.VariableCondition[condIdx].Condition[fieldIndex] = &types.FieldCondition{
				Field:    field.Field,
				Operator: field.Operator,
				Value:    field.Value,
			}
		}
	}

	return result, nil
}

// ListServiceTemplate list service template.
func (h *handler) ListServiceTemplate(ctx context.Context, bizID int64, serviceCategoryID int64,
	serviceTemplateName string, serviceTemplateIDs []int64, isExact bool, page types.Page) (
	[]*types.ServiceTemplate, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ListServiceTemplateReq{
		TenantID: tenantID,

		BKBizID:             bizID,
		ServiceCategoryID:   serviceCategoryID,
		ServiceTemplateName: serviceTemplateName,
		ServiceTemplateIDs:  serviceTemplateIDs,
		IsExact:             isExact,
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}

	resp, err := h.cli.listServiceTemplate(ctx, req)
	if err != nil {
		return nil, err
	}

	serviceTemplates := make([]*types.ServiceTemplate, len(resp.Info))
	for index, serviceTemplate := range resp.Info {
		serviceTemplates[index] = &types.ServiceTemplate{
			ID:                  serviceTemplate.ID,
			BizID:               serviceTemplate.BKBizID,
			ServiceTemplateName: serviceTemplate.ServiceTemplateName,
			ServiceCategoryID:   serviceTemplate.ServiceCategoryID,
			HostApplyEnabled:    serviceTemplate.HostApplyEnabled,
		}
	}

	return serviceTemplates, nil
}

// ListServiceInstance list service instance.
func (h *handler) ListServiceInstance(ctx context.Context, bizID int64, moduleID int64, hostIDs []int64,
	svcInstFuzzyName string, page types.Page) ([]*types.ServiceInstance, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ListServiceInstanceReq{
		TenantID:                 tenantID,
		BKBizID:                  bizID,
		BKModuleID:               moduleID,
		BKHostIDs:                hostIDs,
		ServiceInstanceFuzzyName: svcInstFuzzyName,
		Page: Page{
			Start: page.Offset,
			Sort:  page.Sort,
			Limit: page.Limit,
		},
	}
	resp, err := h.cli.listServiceInstance(ctx, req)
	if err != nil {
		return nil, err
	}

	serviceInstances := make([]*types.ServiceInstance, len(resp.Info))
	for index, serviceInstance := range resp.Info {
		serviceInstances[index] = &types.ServiceInstance{
			ID:                  serviceInstance.ID,
			BizID:               serviceInstance.BKBizID,
			ServiceInstanceName: serviceInstance.ServiceInstanceName,
			HostID:              serviceInstance.BKHostID,
			ModuleID:            serviceInstance.BKModuleID,
		}
	}

	return serviceInstances, nil
}

// ListProcessInstance list process instance.
func (h *handler) ListProcessInstance(ctx context.Context, bizID int64, serviceInstanceID int64) (
	[]*types.ProcessInstance, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ListProcessInstanceReq{
		TenantID:          tenantID,
		BKBizID:           bizID,
		ServiceInstanceID: serviceInstanceID,
	}

	resp, err := h.cli.listProcessInstance(ctx, req)
	if err != nil {
		return nil, err
	}

	processInstances := make([]*types.ProcessInstance, len(*resp))
	for index, processInstance := range *resp {
		processInstances[index] = &types.ProcessInstance{
			Property: &types.ProcessProperty{
				AutoStart:       processInstance.Property.AutoStart,
				BizID:           processInstance.Property.BKBizID,
				FuncName:        processInstance.Property.BKFuncName,
				ProcessID:       processInstance.Property.BKProcessID,
				ProcessName:     processInstance.Property.BKProcessName,
				StartParamRegex: processInstance.Property.BKStartParamRegex,
				Description:     processInstance.Property.Description,
				FaceStopCMD:     processInstance.Property.FaceStopCMD,
				PidFile:         processInstance.Property.PidFile,
				Priority:        processInstance.Property.Priority,
				ProcNum:         processInstance.Property.ProcNum,
				ReloadCMD:       processInstance.Property.ReloadCMD,
				RestartCMD:      processInstance.Property.RestartCMD,
				StartCMD:        processInstance.Property.StartCMD,
				StopCMD:         processInstance.Property.StopCMD,
				Timeout:         processInstance.Property.Timeout,
				User:            processInstance.Property.User,
				WorkPath:        processInstance.Property.WorkPath,
			},
			Relation: &types.ProcessRelation{
				BizID:             processInstance.Relation.BKBizID,
				ProcessID:         processInstance.Relation.BKProcessID,
				ServiceInstanceID: processInstance.Relation.ServiceInstanceID,
				ProcessTemplateID: processInstance.Relation.ProcessTemplateID,
				HostID:            processInstance.Relation.BKHostID,
			},
		}

		processInstances[index].Property.BindInfo = make([]*types.BindInfo, len(processInstance.Property.BindInfo))
		for bindInfoIdx, bindInfo := range processInstance.Property.BindInfo {
			processInstances[index].Property.BindInfo[bindInfoIdx] = &types.BindInfo{
				Enable:        bindInfo.Enable,
				IP:            bindInfo.IP,
				Port:          bindInfo.Port,
				Protocol:      bindInfo.Protocol,
				TemplateRowID: bindInfo.TemplateRowID,
			}
		}
	}

	return processInstances, nil
}

// ListProcTemplate list process template.
func (h *handler) ListProcTemplate(ctx context.Context, bizID int64, serviceTemplateID int64,
	processTemplateIDs []int64) ([]*types.ProcessTemplate, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ListProcTemplateReq{
		TenantID:           tenantID,
		BKBizID:            bizID,
		ServiceTemplateID:  serviceTemplateID,
		ProcessTemplateIDs: processTemplateIDs,
	}
	resp, err := h.cli.listProcTemplate(ctx, req)
	if err != nil {
		return nil, err
	}

	templates := make([]*types.ProcessTemplate, len(resp.Info))
	for index, procTemplate := range resp.Info {
		templates[index] = &types.ProcessTemplate{
			ID:                  procTemplate.ID,
			BizID:               procTemplate.BKBizID,
			ProcessTemplateName: procTemplate.BKProcessTemplateName,
			ServiceTemplateID:   procTemplate.ServiceTemplateID,
			Property: &types.ProcessProperty{
				AutoStart:       procTemplate.Property.AutoStart.Value,
				BizID:           procTemplate.Property.BKBizID.Value,
				FuncName:        procTemplate.Property.BKFuncName.Value,
				ProcessID:       procTemplate.Property.BKProcessID.Value,
				ProcessName:     procTemplate.Property.BKProcessName.Value,
				StartParamRegex: procTemplate.Property.BKStartParamRegex.Value,
				Description:     procTemplate.Property.Description.Value,
				FaceStopCMD:     procTemplate.Property.FaceStopCMD.Value,
				PidFile:         procTemplate.Property.PidFile.Value,
				Priority:        procTemplate.Property.Priority.Value,
				ProcNum:         procTemplate.Property.ProcNum.Value,
				ReloadCMD:       procTemplate.Property.ReloadCMD.Value,
				RestartCMD:      procTemplate.Property.RestartCMD.Value,
				StartCMD:        procTemplate.Property.StartCMD.Value,
				StopCMD:         procTemplate.Property.StopCMD.Value,
				Timeout:         procTemplate.Property.Timeout.Value,
				User:            procTemplate.Property.User.Value,
				WorkPath:        procTemplate.Property.WorkPath.Value,
			},
		}

		templates[index].Property.BindInfo = make([]*types.BindInfo, len(procTemplate.Property.BindInfo.Value))
		for bindInfoIdx, bindInfo := range procTemplate.Property.BindInfo.Value {
			templates[index].Property.BindInfo[bindInfoIdx] = &types.BindInfo{
				Enable:        bindInfo.Enable.Value,
				IP:            bindInfo.IP.Value,
				Port:          bindInfo.Port.Value,
				Protocol:      bindInfo.Protocol.Value,
				TemplateRowID: bindInfo.TemplateRowID,
			}
		}
	}

	return templates, nil
}

// FindSetBatch find set batch.
func (h *handler) FindSetBatch(ctx context.Context, bizID int64, setIDs []int64, fields []string) (
	[]*types.Set, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &FindSetBatchReq{
		TenantID: tenantID,
		BKBizID:  bizID,
		Fields:   fields,
		BKIDs:    setIDs,
	}

	resp, err := h.cli.findSetBatch(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.Set, len(*resp))
	for index, set := range *resp {
		result[index] = &types.Set{
			SetID:              set.BKSetID,
			SetName:            set.BKSetName,
			SetDesc:            set.BKSetDesc,
			SetEnv:             set.BKSetEnv,
			BizID:              set.BKBizID,
			Capacity:           set.BKCapacity,
			ParentID:           set.BKParentID,
			Description:        set.Description,
			SetTemplateID:      set.SetTemplateID,
			SetTemplateVersion: set.SetTemplateVersion,
			ServiceStatus:      set.BKServiceStatus,
		}
	}

	return result, nil
}

// SearchSet search set.
func (h *handler) SearchSet(ctx context.Context, bizID int64, fields []string, page types.Page) ([]*types.Set, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &SearchSetReq{
		TenantID: tenantID,
		BKBizID:  bizID,
		Fields:   fields,
		Page: Page{
			Start: page.Offset,
			Sort:  page.Sort,
			Limit: page.Limit,
		},
	}
	resp, err := h.cli.searchSet(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.Set, len(resp.Info))
	for index, set := range resp.Info {
		result[index] = &types.Set{
			SetID:              set.BKSetID,
			SetName:            set.BKSetName,
			SetDesc:            set.BKSetDesc,
			SetEnv:             set.BKSetEnv,
			BizID:              set.BKBizID,
			Capacity:           set.BKCapacity,
			ParentID:           set.BKParentID,
			Description:        set.Description,
			SetTemplateID:      set.SetTemplateID,
			SetTemplateVersion: set.SetTemplateVersion,
			ServiceStatus:      set.BKServiceStatus,
		}
	}

	return result, nil
}

// SearchModule search module.
func (h *handler) SearchModule(ctx context.Context, bizID int64, setID int64, fields []string, page types.Page) (
	[]*types.Module, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &SearchModuleReq{
		TenantID: tenantID,
		BKBizID:  bizID,
		BKSetID:  setID,
		Fields:   fields,
		Page: Page{
			Start: page.Offset,
			Sort:  page.Sort,
			Limit: page.Limit,
		},
	}

	resp, err := h.cli.searchModule(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.Module, len(resp.Info))
	for index, module := range resp.Info {
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
		}
	}

	return result, nil
}

// FindHostTopoRelation find host topo relation.
func (h *handler) FindHostTopoRelation(ctx context.Context, bizID int64, setIDs []int64, moduleIDs []int64,
	hostIDs []int64, page types.Page) ([]*types.HostTopoRelation, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &FindHostTopoRelationReq{
		TenantID:    tenantID,
		BKBizID:     bizID,
		BKSetIDs:    setIDs,
		BKModuleIDs: moduleIDs,
		BKHostIDs:   hostIDs,
		Page: Page{
			Start: page.Offset,
			Sort:  page.Sort,
			Limit: page.Limit,
		},
	}
	resp, err := h.cli.findHostTopoRelation(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.HostTopoRelation, len(resp.Data))
	for index, relation := range resp.Data {
		result[index] = &types.HostTopoRelation{
			BizID:    relation.BKBizID,
			HostID:   relation.BKHostID,
			SetID:    relation.BKSetID,
			ModuleID: relation.BKModuleID,
		}
	}

	return result, nil
}

// FindHostBizRelations find host biz relations.
func (h *handler) FindHostBizRelations(ctx context.Context, bizID int64, hostIDs []int64) ([]*types.HostTopoRelation,
	error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &FindHostBizRelationsReq{
		TenantID: tenantID,

		BKBizID:  bizID,
		BKHostID: hostIDs,
	}

	resp, err := h.cli.findHostBizRelations(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.HostTopoRelation, len(*resp))
	for index, relation := range *resp {
		result[index] = &types.HostTopoRelation{
			BizID:    relation.BKBizID,
			HostID:   relation.BKHostID,
			SetID:    relation.BKSetID,
			ModuleID: relation.BKModuleID,
		}
	}

	return result, nil
}

// FindHostByServiceTemplate find host by service template.
func (h *handler) FindHostByServiceTemplate(ctx context.Context, bizID int64, serviceTemplateID []int64, moduleIDs []int64,
	fields []string, page types.Page) ([]*types.Host, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &FindHostByServiceTemplateReq{
		TenantID: tenantID,

		BKBizID:              bizID,
		BKServiceTemplateIDs: serviceTemplateID,
		BKModuleIDs:          moduleIDs,
		Fields:               fields,
		Page: Page{
			Start: page.Offset,
			Sort:  page.Sort,
			Limit: page.Limit,
		},
	}

	resp, err := h.cli.findHostByServiceTemplate(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.Host, len(resp.Info))
	for index, host := range resp.Info {
		result[index] = &types.Host{
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

	return result, nil
}

// FindHostBySetTemplate find host by set template.
func (h *handler) FindHostBySetTemplate(ctx context.Context, bizID int64, setTemplateID []int64, setIDs []int64,
	fields []string, page types.Page) ([]*types.Host, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &FindHostBySetTemplateReq{
		TenantID: tenantID,

		BKBizID:          bizID,
		BKSetTemplateIDs: setTemplateID,
		BKSetIDs:         setIDs,
		Fields:           fields,
		Page: Page{
			Start: page.Offset,
			Sort:  page.Sort,
			Limit: page.Limit,
		},
	}
	resp, err := h.cli.findHostBySetTemplate(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.Host, len(resp.Info))
	for index, host := range resp.Info {
		result[index] = &types.Host{
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

	return result, nil
}

// FindHostByTopo find host by topo.
func (h *handler) FindHostByTopo(ctx context.Context, bizID int64, objID string, instID int64, fields []string,
	page types.Page) ([]*types.Host, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &FindHostByTopoReq{
		TenantID: tenantID,

		BKBizID:  bizID,
		BKObjID:  objID,
		BKInstID: instID,
		Fields:   fields,
		Page: Page{
			Start: page.Offset,
			Sort:  page.Sort,
			Limit: page.Limit,
		},
	}

	resp, err := h.cli.findHostByTopo(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.Host, len(resp.Info))
	for index, host := range resp.Info {
		result[index] = &types.Host{
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

	return result, nil
}

// FindHostRelationsWithTopo find host relations with topo.
func (h *handler) FindHostRelationsWithTopo(ctx context.Context, bizID int64, objID string, instIDs []int64,
	fields []string, page types.Page) ([]*types.HostTopoRelation, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &FindHostRelationsWithTopoReq{
		TenantID:  tenantID,
		BKBizID:   bizID,
		BKObjID:   objID,
		BKInstIDs: instIDs,
		Fields:    fields,
		Page: Page{
			Start: page.Offset,
			Sort:  page.Sort,
			Limit: page.Limit,
		},
	}

	resp, err := h.cli.findHostRelationsWithTopo(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.HostTopoRelation, len(resp.Info))
	for index, relation := range resp.Info {
		result[index] = &types.HostTopoRelation{
			BizID:    relation.BKBizID,
			HostID:   relation.BKHostID,
			SetID:    relation.BKSetID,
			ModuleID: relation.BKModuleID,
		}
	}

	return result, nil
}

// ListServiceInstanceDetail list service instance detail.
func (h *handler) ListServiceInstanceDetail(ctx context.Context, bizID int64, page types.Page) (
	[]*types.ServiceInstanceDetail, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ListServiceInstanceDetailReq{
		TenantID: tenantID,

		BKBizID: bizID,
		Page: Page{
			Start: page.Offset,
			Sort:  page.Sort,
			Limit: page.Limit,
		},
	}
	resp, err := h.cli.listServiceInstanceDetail(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.ServiceInstanceDetail, len(resp.Info))
	for index, serviceInstance := range resp.Info {
		result[index] = &types.ServiceInstanceDetail{
			ID:                  serviceInstance.ID,
			BizID:               serviceInstance.BKBizID,
			ServiceTemplateName: serviceInstance.Name,
			ServiceTemplateID:   serviceInstance.ServiceTemplateID,
			HostID:              serviceInstance.BKHostID,
			ModuleID:            serviceInstance.BKModuleID,
			ServiceCategoryID:   serviceInstance.ServiceCategoryID,
		}

		result[index].ProcessInstances = make([]*types.ProcessInstance, len(serviceInstance.ProcessInstances))
		for svcInstIndex, svcInst := range serviceInstance.ProcessInstances {
			result[index].ProcessInstances[svcInstIndex] = &types.ProcessInstance{
				Property: &types.ProcessProperty{
					AutoStart:       svcInst.Process.AutoStart,
					BizID:           svcInst.Process.BKBizID,
					FuncName:        svcInst.Process.BKFuncName,
					ProcessID:       svcInst.Process.BKProcessID,
					ProcessName:     svcInst.Process.BKProcessName,
					StartParamRegex: svcInst.Process.BKStartParamRegex,
					Description:     svcInst.Process.Description,
					FaceStopCMD:     svcInst.Process.FaceStopCMD,
					PidFile:         svcInst.Process.PidFile,
					Priority:        svcInst.Process.Priority,
					ProcNum:         svcInst.Process.ProcNum,
					ReloadCMD:       svcInst.Process.ReloadCMD,
					RestartCMD:      svcInst.Process.RestartCMD,
					StartCMD:        svcInst.Process.StartCMD,
					StopCMD:         svcInst.Process.StopCMD,
					Timeout:         svcInst.Process.Timeout,
					User:            svcInst.Process.User,
					WorkPath:        svcInst.Process.WorkPath,
				},
				Relation: &types.ProcessRelation{
					BizID:             svcInst.Relation.BKBizID,
					ProcessID:         svcInst.Relation.BKProcessID,
					ServiceInstanceID: svcInst.Relation.ServiceInstanceID,
					ProcessTemplateID: svcInst.Relation.ProcessTemplateID,
					HostID:            svcInst.Relation.BKHostID,
				},
			}
			result[index].ProcessInstances[svcInstIndex].Property.BindInfo = make([]*types.BindInfo,
				len(svcInst.Process.BindInfo))
			for bindInfoIndex, bindInfo := range svcInst.Process.BindInfo {
				result[index].ProcessInstances[svcInstIndex].Property.BindInfo[bindInfoIndex] = &types.BindInfo{
					Enable:        bindInfo.Enable,
					IP:            bindInfo.IP,
					Port:          bindInfo.Port,
					Protocol:      bindInfo.Protocol,
					TemplateRowID: bindInfo.TemplateRowID,
				}
			}
		}
	}

	return result, nil
}

// GetMainlineObjectTopo get mainline object topo.
func (h *handler) GetMainlineObjectTopo(ctx context.Context) ([]*types.MainlineObjectTopo, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}
	req := &GetMainlineObjectTopoReq{
		TenantID: tenantID,
	}
	resp, err := h.cli.getMainlineObjectTopo(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.MainlineObjectTopo, len(*resp))
	for index, topo := range *resp {
		result[index] = &types.MainlineObjectTopo{
			ObjID:      topo.BKObjID,
			ObjName:    topo.BKObjName,
			NextObj:    topo.BKNextObj,
			NextName:   topo.BKNextName,
			PreObjID:   topo.BKPreObjID,
			PreObjName: topo.BKPreObjName,
		}
	}

	return result, nil
}

// ListBizHostsTopo list biz hosts topo.
func (h *handler) ListBizHostsTopo(ctx context.Context, bizID int64, fields []string, page types.Page) (
	[]*types.HostWithHostTopo, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ListBizHostsTopoReq{
		TenantID: tenantID,
		BKBizID:  bizID,
		Fields:   fields,
		Page: Page{
			Start: page.Offset,
			Sort:  page.Sort,
			Limit: page.Limit,
		},
	}
	resp, err := h.cli.listBizHostsTopo(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.HostWithHostTopo, len(resp.Info))
	for index, host := range resp.Info {
		result[index] = &types.HostWithHostTopo{
			Host: &types.Host{
				HostID:   host.Host.BKHostID,
				TenantID: tenantID,
				Static: &types.HostStatic{
					BizID:         bizID,
					NetworkAreaID: host.Host.BKCloudID,
					HostName:      host.Host.BKHostName,
					DeptName:      host.Host.DeptName,
					InnerIP:       host.Host.BKHostInnerIPV4,
					InnerIPV6:     host.Host.BKHostInnerIPV6,
					OuterIP:       host.Host.BKHostOuterIPV4,
					OuterIPV6:     host.Host.BKHostOuterIPV6,
					Mac:           host.Host.BKMac,
					OSType:        host.Host.BKOsType,
					SyncedAgentID: host.Host.BKAgentID,
				},
				Dynamic: types.NewBlankNodeDynamic(),
			},
		}
		result[index].HostTopo = make([]*types.HostTopo, len(host.Topo))
		for topoIndex, topo := range host.Topo {
			result[index].HostTopo[topoIndex] = &types.HostTopo{
				SetID:   topo.BKSetID,
				SetName: topo.BKSetName,
			}
			result[index].HostTopo[topoIndex].Module = make([]*struct {
				ModuleID   int64
				ModuleName string
			}, len(topo.Module))
			for moduleIndex, module := range topo.Module {
				result[index].HostTopo[topoIndex].Module[moduleIndex] = &struct {
					ModuleID   int64
					ModuleName string
				}{
					ModuleID:   module.BKModuleID,
					ModuleName: module.BKModuleName,
				}
			}
		}
	}

	return result, nil
}

// ListServiceInstanceByHost list service instance by host.
func (h *handler) ListServiceInstanceByHost(ctx context.Context, bizID int64, hostID int64, page types.Page) (
	[]*types.ServiceInstance, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ListServiceInstanceByHostReq{
		TenantID: tenantID,

		BKBizID:  bizID,
		BKHostID: hostID,
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}
	resp, err := h.cli.listServiceInstanceByHost(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.ServiceInstance, len(resp.Info))
	for index, serviceInstance := range resp.Info {
		result[index] = &types.ServiceInstance{
			ID:                  serviceInstance.ID,
			BizID:               serviceInstance.BKBizID,
			ServiceInstanceName: serviceInstance.ServiceInstanceName,
			HostID:              serviceInstance.BKHostID,
			ModuleID:            serviceInstance.BKModuleID,
		}
	}

	return result, nil
}

// ListServiceInstanceBySetTemplate list service instance by set template.
func (h *handler) ListServiceInstanceBySetTemplate(ctx context.Context, bizID int64, setTemplate int64,
	page types.Page) ([]*types.ServiceInstance, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ListServiceInstanceBySetTemplateReq{
		TenantID: tenantID,

		BKBizID:       bizID,
		SetTemplateID: setTemplate,
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}
	resp, err := h.cli.listServiceInstanceBySetTemplate(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.ServiceInstance, len(resp.Info))
	for index, serviceInstance := range resp.Info {
		result[index] = &types.ServiceInstance{
			ID:                  serviceInstance.ID,
			BizID:               serviceInstance.BKBizID,
			ServiceInstanceName: serviceInstance.ServiceInstanceName,
			HostID:              serviceInstance.BKHostID,
			ModuleID:            serviceInstance.BKModuleID,
		}
	}

	return result, nil
}

// ListSetTemplate list set template.
func (h *handler) ListSetTemplate(ctx context.Context, bizID int64, setTemplateIDs []int64, page types.Page) (
	[]*types.ServiceInstance, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ListSetTemplateReq{
		TenantID: tenantID,

		BKBizID:        bizID,
		SetTemplateIDs: setTemplateIDs,
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}
	resp, err := h.cli.listSetTemplate(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.ServiceInstance, len(resp.Info))
	for index, serviceInstance := range resp.Info {
		result[index] = &types.ServiceInstance{
			ID:                  serviceInstance.ID,
			BizID:               serviceInstance.BKBizID,
			ServiceInstanceName: serviceInstance.ServiceInstanceName,
			HostID:              serviceInstance.BKHostID,
			ModuleID:            serviceInstance.BKModuleID,
		}
	}

	return result, nil
}

// BatchUpdateHost batch update host.
func (h *handler) BatchUpdateHost(ctx context.Context, hostsUpdateInfo []*types.UpdateHostProperties) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	req := &BatchUpdateHostReq{TenantID: tenantID}
	for _, hostUpdateInfo := range hostsUpdateInfo {
		updateInfo := &UpdateHostProperties{}
		updateInfo.BKHostID = hostUpdateInfo.HostID
		updateInfo.Properties.BKHostName = hostUpdateInfo.HostName
		updateInfo.Properties.BKComment = hostUpdateInfo.Comment
		updateInfo.Properties.Operator = hostUpdateInfo.Operator
		updateInfo.Properties.BKIspName = hostUpdateInfo.IspName
		req.Update = append(req.Update, updateInfo)
	}
	err = h.cli.batchUpdateHost(ctx, req)
	if err != nil {
		return err
	}

	return nil
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
	*types.PushHostIdentifierTaskResult, error) {

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

	result := &types.PushHostIdentifierTaskResult{
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

// HostResourceWatch host resource watch.
func (h *handler) HostResourceWatch(ctx context.Context, event types.ResourceWatchEvent) (
	[]*types.HostResourceWatchEvent, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	if len(event.Fields) == 0 {
		event.Fields = []string{"bk_host_id", "bk_host_name", "bk_os_type", "bk_host_innerip_v4",
			"bk_host_innerip_v6", "bk_host_outerip_v4", "bk_host_outerip_v6", "bk_mac", "bk_cloud_id"}
	}

	req := &ResourceWatchReq{
		TenantID:    tenantID,
		BKResource:  string(event.Resource),
		BKCursor:    event.Cursor,
		BKFields:    event.Fields,
		BKStartFrom: event.StartFrom,
	}
	for _, eventType := range event.EventType {
		req.BKEventTypes = append(req.BKEventTypes, string(eventType))
	}

	resp, err := h.cli.hostResourceWatch(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.HostResourceWatchEvent, len(resp.BKEvents))
	for index, hostEvent := range resp.BKEvents {
		result[index] = &types.HostResourceWatchEvent{
			EventType: types.ResourceWatchEventType(hostEvent.BKEventType),
			Resource:  types.ResourceWatchEventResource(hostEvent.BKResource),
			Cursor:    hostEvent.BKCursor,
		}

		if !resp.BKWatched {
			result[index].Detail = nil
			continue
		}

		result[index].Detail = &types.Host{
			HostID: hostEvent.BKDetail.BKHostID,
			Static: &types.HostStatic{
				HostName:      hostEvent.BKDetail.BKHostName,
				OSType:        hostEvent.BKDetail.BKOsType,
				InnerIP:       hostEvent.BKDetail.BKHostInnerIPV4,
				InnerIPV6:     hostEvent.BKDetail.BKHostInnerIPV6,
				OuterIP:       hostEvent.BKDetail.BKHostOuterIPV4,
				OuterIPV6:     hostEvent.BKDetail.BKHostOuterIPV6,
				Mac:           hostEvent.BKDetail.BKMac,
				NetworkAreaID: hostEvent.BKDetail.BKCloudID,
			},
		}
	}

	return result, nil
}

// HostRelationResourceWatch host relation resource watch.
func (h *handler) HostRelationResourceWatch(ctx context.Context, event types.ResourceWatchEvent) (
	[]*types.HostRelationResourceWatchEvent, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ResourceWatchReq{
		TenantID:    tenantID,
		BKResource:  string(event.Resource),
		BKCursor:    event.Cursor,
		BKFields:    event.Fields,
		BKStartFrom: event.StartFrom,
	}
	for _, eventType := range event.EventType {
		req.BKEventTypes = append(req.BKEventTypes, string(eventType))
	}

	resp, err := h.cli.hostRelationResourceWatch(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.HostRelationResourceWatchEvent, len(resp.BKEvents))
	for index, hostEvent := range resp.BKEvents {
		result[index] = &types.HostRelationResourceWatchEvent{
			EventType: types.ResourceWatchEventType(hostEvent.BKEventType),
			Resource:  types.ResourceWatchEventResource(hostEvent.BKResource),
			Cursor:    hostEvent.BKCursor,
		}

		if !resp.BKWatched {
			result[index].Detail = nil
			continue
		}

		result[index].Detail = &types.HostTopoRelation{
			HostID:   hostEvent.BKDetail.BKHostID,
			ModuleID: hostEvent.BKDetail.BKModuleID,
			SetID:    hostEvent.BKDetail.BKSetID,
			BizID:    hostEvent.BKDetail.BKBizID,
		}
	}

	return result, nil
}

// ProcessResourceWatch host relation resource watch.
func (h *handler) ProcessResourceWatch(ctx context.Context, event types.ResourceWatchEvent) (
	[]*types.ProcessResourceWatchEvent, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ResourceWatchReq{
		TenantID:    tenantID,
		BKResource:  string(event.Resource),
		BKCursor:    event.Cursor,
		BKFields:    event.Fields,
		BKStartFrom: event.StartFrom,
	}
	for _, eventType := range event.EventType {
		req.BKEventTypes = append(req.BKEventTypes, string(eventType))
	}

	resp, err := h.cli.processResourceWatch(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.ProcessResourceWatchEvent, len(resp.BKEvents))
	for index, processEvent := range resp.BKEvents {
		result[index] = &types.ProcessResourceWatchEvent{
			EventType: types.ResourceWatchEventType(processEvent.BKEventType),
			Resource:  types.ResourceWatchEventResource(processEvent.BKResource),
			Cursor:    processEvent.BKCursor,
		}

		if !resp.BKWatched {
			result[index].Detail = nil
			continue
		}

		result[index].Detail = &types.ProcessProperty{
			AutoStart:       processEvent.BKDetail.AutoStart,
			BizID:           processEvent.BKDetail.BKBizID,
			FuncName:        processEvent.BKDetail.BKFuncName,
			ProcessID:       processEvent.BKDetail.BKProcessID,
			ProcessName:     processEvent.BKDetail.BKProcessName,
			StartParamRegex: processEvent.BKDetail.BKStartParamRegex,
			Description:     processEvent.BKDetail.Description,
			FaceStopCMD:     processEvent.BKDetail.FaceStopCMD,
			PidFile:         processEvent.BKDetail.PidFile,
			Priority:        processEvent.BKDetail.Priority,
			ProcNum:         processEvent.BKDetail.ProcNum,
			ReloadCMD:       processEvent.BKDetail.ReloadCMD,
			RestartCMD:      processEvent.BKDetail.RestartCMD,
			StartCMD:        processEvent.BKDetail.StartCMD,
			StopCMD:         processEvent.BKDetail.StopCMD,
			Timeout:         processEvent.BKDetail.Timeout,
			User:            processEvent.BKDetail.User,
			WorkPath:        processEvent.BKDetail.WorkPath,
		}

		for _, bindInfo := range processEvent.BKDetail.BindInfo {
			result[index].Detail.BindInfo = append(result[index].Detail.BindInfo, &types.BindInfo{
				Enable:        bindInfo.Enable,
				IP:            bindInfo.IP,
				Port:          bindInfo.Port,
				Protocol:      bindInfo.Protocol,
				TemplateRowID: bindInfo.TemplateRowID,
			})
		}
	}

	return result, nil
}
